// Package object object/compose.go
package object

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/types"
)

const (
	composeMaxPartsCount = 10000
	composeMinPartSize   = 5 * 1024 * 1024
	composeMaxPartSize   = 5 * 1024 * 1024 * 1024
	composeMaxObjectSize = 5 * 1024 * 1024 * 1024 * 1024
)

// SourceInfo describes a source object for composition.
type SourceInfo struct {
	Bucket string
	Object string

	VersionID string

	RangeStart int64
	RangeEnd   int64
	RangeSet   bool

	MatchETag     string
	NotMatchETag  string
	MatchModified time.Time
	NotModified   time.Time
}

// DestinationInfo describes the target object for composition.
type DestinationInfo struct {
	Bucket string
	Object string
}

type copyPartResult struct {
	XMLName      xml.Name  `xml:"CopyPartResult"`
	ETag         string    `xml:"ETag"`
	LastModified time.Time `xml:"LastModified"`
}

// Compose creates a new object by composing one or more source objects.
func (s *objectService) Compose(ctx context.Context, dst DestinationInfo, sources []SourceInfo, opts ...PutOption) (types.UploadInfo, error) {
	if err := validateBucketName(dst.Bucket); err != nil {
		return types.UploadInfo{}, err
	}
	if err := validateObjectName(dst.Object); err != nil {
		return types.UploadInfo{}, err
	}
	if len(sources) == 0 || len(sources) > composeMaxPartsCount {
		return types.UploadInfo{}, fmt.Errorf("compose requires between 1 and %d source objects", composeMaxPartsCount)
	}

	srcInfos := make([]types.ObjectInfo, len(sources))
	srcSizes := make([]int64, len(sources))
	var totalSize int64
	var totalParts int64

	for i, src := range sources {
		if err := validateBucketName(src.Bucket); err != nil {
			return types.UploadInfo{}, err
		}
		if err := validateObjectName(src.Object); err != nil {
			return types.UploadInfo{}, err
		}
		if src.RangeSet {
			if src.RangeStart < 0 {
				return types.UploadInfo{}, fmt.Errorf("range start must be >= 0")
			}
			if src.RangeEnd < src.RangeStart {
				return types.UploadInfo{}, fmt.Errorf("range end must be >= range start")
			}
		}

		statOpts := []StatOption{}
		if src.VersionID != "" {
			statOpts = append(statOpts, func(opts *StatOptions) {
				opts.VersionID = src.VersionID
			})
		}
		info, err := s.Stat(ctx, src.Bucket, src.Object, statOpts...)
		if err != nil {
			return types.UploadInfo{}, err
		}
		srcInfos[i] = info

		copySize := info.Size
		if src.RangeSet {
			if src.RangeEnd >= info.Size {
				return types.UploadInfo{}, fmt.Errorf("range end %d exceeds object size %d", src.RangeEnd, info.Size)
			}
			copySize = src.RangeEnd - src.RangeStart + 1
		}

		if copySize < composeMinPartSize && i < len(sources)-1 {
			return types.UploadInfo{}, fmt.Errorf("source %d is too small (%d bytes) and is not the last part", i, copySize)
		}

		totalSize += copySize
		if totalSize > composeMaxObjectSize {
			return types.UploadInfo{}, fmt.Errorf("compose object size %d exceeds max %d", totalSize, composeMaxObjectSize)
		}

		srcSizes[i] = copySize
		totalParts += composePartsRequired(copySize)
		if totalParts > composeMaxPartsCount {
			return types.UploadInfo{}, fmt.Errorf("compose requires more than %d parts", composeMaxPartsCount)
		}
	}

	for i := range sources {
		if !composeHasConditions(sources[i]) && srcInfos[i].ETag != "" {
			sources[i].MatchETag = srcInfos[i].ETag
		}
	}

	// Prefer server-side copy for empty objects or single small sources without ranges.
	if totalSize == 0 || (len(sources) == 1 && !sources[0].RangeSet && totalParts == 1 && totalSize <= composeMaxPartSize) {
		copyOpts := composeCopyOptions(sources[0], opts...)
		copyInfo, err := s.Copy(ctx, dst.Bucket, dst.Object, sources[0].Bucket, sources[0].Object, copyOpts...)
		if err != nil {
			return types.UploadInfo{}, err
		}
		return types.UploadInfo{
			Bucket:            copyInfo.Bucket,
			Key:               copyInfo.Key,
			ETag:              copyInfo.ETag,
			Size:              totalSize,
			LastModified:      copyInfo.LastModified,
			VersionID:         copyInfo.VersionID,
			ChecksumCRC32:     copyInfo.ChecksumCRC32,
			ChecksumCRC32C:    copyInfo.ChecksumCRC32C,
			ChecksumSHA1:      copyInfo.ChecksumSHA1,
			ChecksumSHA256:    copyInfo.ChecksumSHA256,
			ChecksumCRC64NVME: copyInfo.ChecksumCRC64NVME,
		}, nil
	}

	uploadID, err := s.InitiateMultipartUpload(ctx, dst.Bucket, dst.Object, opts...)
	if err != nil {
		return types.UploadInfo{}, err
	}

	parts := make([]types.ObjectPart, 0, int(totalParts))
	partNumber := 1

	for i, src := range sources {
		header := make(http.Header)
		copySource := fmt.Sprintf("%s/%s", src.Bucket, src.Object)
		if src.VersionID != "" {
			copySource += "?versionId=" + src.VersionID
		}
		header.Set("x-amz-copy-source", copySource)

		if src.MatchETag != "" {
			header.Set("x-amz-copy-source-if-match", src.MatchETag)
		}
		if src.NotMatchETag != "" {
			header.Set("x-amz-copy-source-if-none-match", src.NotMatchETag)
		}
		if !src.MatchModified.IsZero() {
			header.Set("x-amz-copy-source-if-modified-since", src.MatchModified.Format(http.TimeFormat))
		}
		if !src.NotModified.IsZero() {
			header.Set("x-amz-copy-source-if-unmodified-since", src.NotModified.Format(http.TimeFormat))
		}

		startIndex, endIndex := composeCalculateSplits(srcSizes[i], src)
		for j, start := range startIndex {
			end := endIndex[j]
			header.Set("x-amz-copy-source-range", fmt.Sprintf("bytes=%d-%d", start, end))

			part, err := s.uploadPartCopy(ctx, dst.Bucket, dst.Object, uploadID, partNumber, header)
			if err != nil {
				return types.UploadInfo{}, err
			}
			part.Size = end - start + 1
			parts = append(parts, part)
			partNumber++
		}
	}

	uploadInfo, err := s.CompleteMultipartUpload(ctx, dst.Bucket, dst.Object, uploadID, parts, opts...)
	if err != nil {
		return types.UploadInfo{}, err
	}
	uploadInfo.Size = totalSize

	return uploadInfo, nil
}

func (s *objectService) uploadPartCopy(ctx context.Context, bucketName, objectName, uploadID string, partNumber int, headers http.Header) (types.ObjectPart, error) {
	if headers == nil {
		headers = make(http.Header)
	}
	if uploadID == "" {
		return types.ObjectPart{}, fmt.Errorf("upload ID cannot be empty")
	}
	if partNumber <= 0 {
		return types.ObjectPart{}, fmt.Errorf("part number must be greater than 0")
	}

	queryValues := url.Values{}
	queryValues.Set("partNumber", fmt.Sprintf("%d", partNumber))
	queryValues.Set("uploadId", uploadID)

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		ObjectName:   objectName,
		QueryValues:  queryValues,
		CustomHeader: headers,
	}

	req := core.NewRequest(ctx, http.MethodPut, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return types.ObjectPart{}, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return types.ObjectPart{}, parseErrorResponse(resp, bucketName, objectName)
	}

	var result copyPartResult
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return types.ObjectPart{}, fmt.Errorf("failed to decode upload part copy response: %w", err)
	}

	part := types.ObjectPart{
		PartNumber: partNumber,
		ETag:       trimETag(result.ETag),
	}

	if checksumCRC32 := resp.Header.Get("x-amz-checksum-crc32"); checksumCRC32 != "" {
		part.ChecksumCRC32 = checksumCRC32
	}
	if checksumCRC32C := resp.Header.Get("x-amz-checksum-crc32c"); checksumCRC32C != "" {
		part.ChecksumCRC32C = checksumCRC32C
	}
	if checksumSHA1 := resp.Header.Get("x-amz-checksum-sha1"); checksumSHA1 != "" {
		part.ChecksumSHA1 = checksumSHA1
	}
	if checksumSHA256 := resp.Header.Get("x-amz-checksum-sha256"); checksumSHA256 != "" {
		part.ChecksumSHA256 = checksumSHA256
	}

	return part, nil
}

func composeCopyOptions(src SourceInfo, opts ...PutOption) []CopyOption {
	putOptions := applyPutOptions(opts)
	replaceMetadata := putOptions.contentTypeSet ||
		putOptions.ContentEncoding != "" ||
		putOptions.ContentDisposition != "" ||
		putOptions.CacheControl != "" ||
		!putOptions.Expires.IsZero() ||
		putOptions.UserMetadata != nil
	replaceTagging := putOptions.UserTags != nil

	return []CopyOption{func(options *CopyOptions) {
		options.SourceVersionID = src.VersionID
		options.MatchETag = src.MatchETag
		options.NotMatchETag = src.NotMatchETag
		options.MatchModified = src.MatchModified
		options.NotModified = src.NotModified

		options.ReplaceMetadata = replaceMetadata
		options.ReplaceTagging = replaceTagging
		options.UserMetadata = putOptions.UserMetadata
		options.UserTags = putOptions.UserTags

		if putOptions.contentTypeSet {
			options.ContentType = putOptions.ContentType
		}
		options.ContentEncoding = putOptions.ContentEncoding
		options.ContentDisposition = putOptions.ContentDisposition
		options.CacheControl = putOptions.CacheControl
		options.Expires = putOptions.Expires
		options.StorageClass = putOptions.StorageClass
		options.CustomHeaders = putOptions.CustomHeaders
	}}
}

func composeHasConditions(src SourceInfo) bool {
	return src.MatchETag != "" ||
		src.NotMatchETag != "" ||
		!src.MatchModified.IsZero() ||
		!src.NotModified.IsZero()
}

func composePartsRequired(size int64) int64 {
	if size <= 0 {
		return 0
	}
	maxPartSize := int64(composeMaxObjectSize / (composeMaxPartsCount - 1))
	reqParts := size / maxPartSize
	if size%maxPartSize > 0 {
		reqParts++
	}
	return reqParts
}

func composeCalculateSplits(size int64, src SourceInfo) (startIndex, endIndex []int64) {
	if size <= 0 {
		return nil, nil
	}

	reqParts := composePartsRequired(size)
	partCount := int(reqParts)
	startIndex = make([]int64, partCount)
	endIndex = make([]int64, partCount)

	start := int64(0)
	if src.RangeSet {
		start = src.RangeStart
	}

	quot, rem := size/reqParts, size%reqParts
	nextStart := start
	for j := 0; j < partCount; j++ {
		curPartSize := quot
		if int64(j) < rem {
			curPartSize++
		}

		curStart := nextStart
		curEnd := curStart + curPartSize - 1
		nextStart = curEnd + 1

		startIndex[j], endIndex[j] = curStart, curEnd
	}

	return startIndex, endIndex
}
