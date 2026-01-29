// Package object object/multipart_list.go
package object

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/types"
)

// UserIdentity represents an owner/initiator in multipart listing.
type UserIdentity struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

// MultipartUpload represents a multipart upload entry.
type MultipartUpload struct {
	Key          string       `xml:"Key"`
	UploadID     string       `xml:"UploadId"`
	Initiator    UserIdentity `xml:"Initiator"`
	Owner        UserIdentity `xml:"Owner"`
	StorageClass string       `xml:"StorageClass"`
	Initiated    string       `xml:"Initiated"`
}

// ListMultipartUploadsResult represents ListMultipartUploads response.
type ListMultipartUploadsResult struct {
	XMLName            xml.Name          `xml:"ListMultipartUploadsResult"`
	Bucket             string            `xml:"Bucket"`
	KeyMarker          string            `xml:"KeyMarker"`
	UploadIDMarker     string            `xml:"UploadIdMarker"`
	NextKeyMarker      string            `xml:"NextKeyMarker"`
	NextUploadIDMarker string            `xml:"NextUploadIdMarker"`
	MaxUploads         int               `xml:"MaxUploads"`
	IsTruncated        bool              `xml:"IsTruncated"`
	Uploads            []MultipartUpload `xml:"Upload"`
	CommonPrefixes     []CommonPrefix    `xml:"CommonPrefixes"`
}

// ListPartsResult represents ListParts response.
type ListPartsResult struct {
	XMLName              xml.Name         `xml:"ListPartsResult"`
	Bucket               string           `xml:"Bucket"`
	Key                  string           `xml:"Key"`
	UploadID             string           `xml:"UploadId"`
	PartNumberMarker     int              `xml:"PartNumberMarker"`
	NextPartNumberMarker int              `xml:"NextPartNumberMarker"`
	MaxParts             int              `xml:"MaxParts"`
	IsTruncated          bool             `xml:"IsTruncated"`
	Initiator            UserIdentity     `xml:"Initiator"`
	Owner                UserIdentity     `xml:"Owner"`
	StorageClass         string           `xml:"StorageClass"`
	Parts                []types.ObjectPart `xml:"Part"`
}

// ListMultipartUploads lists in-progress multipart uploads for a bucket.
func (s *objectService) ListMultipartUploads(ctx context.Context, bucketName string, opts ...MultipartListOption) (ListMultipartUploadsResult, error) {
	if err := validateBucketName(bucketName); err != nil {
		return ListMultipartUploadsResult{}, err
	}

	options := applyListMultipartUploadsOptions(opts)

	queryValues := url.Values{}
	queryValues.Set("uploads", "")
	if options.Prefix != "" {
		queryValues.Set("prefix", options.Prefix)
	}
	if options.Delimiter != "" {
		queryValues.Set("delimiter", options.Delimiter)
	}
	if options.KeyMarker != "" {
		queryValues.Set("key-marker", options.KeyMarker)
	}
	if options.UploadIDMarker != "" {
		queryValues.Set("upload-id-marker", options.UploadIDMarker)
	}
	maxUploads := options.MaxUploads
	if maxUploads <= 0 {
		maxUploads = 1000
	}
	queryValues.Set("max-uploads", strconv.Itoa(maxUploads))

	meta := core.RequestMetadata{
		BucketName:  bucketName,
		QueryValues: queryValues,
	}

	req := core.NewRequest(ctx, http.MethodGet, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return ListMultipartUploadsResult{}, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return ListMultipartUploadsResult{}, parseErrorResponse(resp, bucketName, "")
	}

	var result ListMultipartUploadsResult
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return ListMultipartUploadsResult{}, fmt.Errorf("failed to decode list multipart uploads response: %w", err)
	}
	return result, nil
}

// ListObjectParts lists parts for a specific multipart upload.
func (s *objectService) ListObjectParts(ctx context.Context, bucketName, objectName, uploadID string, opts ...ListPartsOption) (ListPartsResult, error) {
	if err := validateBucketName(bucketName); err != nil {
		return ListPartsResult{}, err
	}
	if err := validateObjectName(objectName); err != nil {
		return ListPartsResult{}, err
	}
	if uploadID == "" {
		return ListPartsResult{}, errors.New("uploadID must not be empty")
	}

	options := applyListPartsOptions(opts)

	queryValues := url.Values{}
	queryValues.Set("uploadId", uploadID)
	if options.PartNumberMarker > 0 {
		queryValues.Set("part-number-marker", strconv.Itoa(options.PartNumberMarker))
	}
	maxParts := options.MaxParts
	if maxParts <= 0 {
		maxParts = 1000
	}
	queryValues.Set("max-parts", strconv.Itoa(maxParts))

	meta := core.RequestMetadata{
		BucketName:  bucketName,
		ObjectName:  objectName,
		QueryValues: queryValues,
	}

	req := core.NewRequest(ctx, http.MethodGet, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return ListPartsResult{}, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return ListPartsResult{}, parseErrorResponse(resp, bucketName, objectName)
	}

	var result ListPartsResult
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return ListPartsResult{}, fmt.Errorf("failed to decode list parts response: %w", err)
	}
	return result, nil
}
