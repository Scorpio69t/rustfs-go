// Package object object/list.go
package object

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/types"
)

// ListBucketV2Result represents bucket listing V2 response
type ListBucketV2Result struct {
	XMLName               xml.Name           `xml:"ListBucketResult"`
	Name                  string             `xml:"Name"`
	Prefix                string             `xml:"Prefix"`
	KeyCount              int                `xml:"KeyCount"`
	MaxKeys               int                `xml:"MaxKeys"`
	Delimiter             string             `xml:"Delimiter"`
	IsTruncated           bool               `xml:"IsTruncated"`
	Contents              []types.ObjectInfo `xml:"Contents"`
	CommonPrefixes        []CommonPrefix     `xml:"CommonPrefixes"`
	ContinuationToken     string             `xml:"ContinuationToken"`
	NextContinuationToken string             `xml:"NextContinuationToken"`
	StartAfter            string             `xml:"StartAfter"`
}

// CommonPrefix represents common prefix entry
type CommonPrefix struct {
	Prefix string `xml:"Prefix"`
}

// List lists objects (implementation)
func (s *objectService) List(ctx context.Context, bucketName string, opts ...ListOption) <-chan types.ObjectInfo {
	// Create object info channel
	objectCh := make(chan types.ObjectInfo)

	// Start background goroutine for listing
	go func() {
		defer close(objectCh)

		// Validate parameters
		if err := validateBucketName(bucketName); err != nil {
			objectCh <- types.ObjectInfo{Err: err}
			return
		}

		// Apply options
		options := applyListOptions(opts)

		// Switch to version listing if requested
		if options.WithVersions {
			if err := s.streamObjectVersions(ctx, bucketName, &options, objectCh); err != nil {
				objectCh <- types.ObjectInfo{Err: err}
			}
			return
		}

		// Set delimiter
		delimiter := "/"
		if options.Recursive {
			// Recursive listing, no delimiter
			delimiter = ""
		}

		// Save ContinuationToken for next request
		var continuationToken string

		for {
			// Check if context canceled
			select {
			case <-ctx.Done():
				objectCh <- types.ObjectInfo{Err: ctx.Err()}
				return
			default:
			}

			// Query object list (up to 1000)
			result, err := s.listObjectsV2Query(ctx, bucketName, &options, delimiter, continuationToken)
			if err != nil {
				objectCh <- types.ObjectInfo{Err: err}
				return
			}

			// Send content objects
			for _, object := range result.Contents {
				// Remove ETag quotes
				object.ETag = trimETag(object.ETag)

				select {
				case objectCh <- object:
				case <-ctx.Done():
					objectCh <- types.ObjectInfo{Err: ctx.Err()}
					return
				}
			}

			// Send common prefixes (when using delimiter)
			for _, prefix := range result.CommonPrefixes {
				select {
				case objectCh <- types.ObjectInfo{Key: prefix.Prefix}:
				case <-ctx.Done():
					objectCh <- types.ObjectInfo{Err: ctx.Err()}
					return
				}
			}

			// Save next ContinuationToken if present
			if result.NextContinuationToken != "" {
				continuationToken = result.NextContinuationToken
			}

			// End if list not truncated
			if !result.IsTruncated {
				return
			}

			// Prevent infinite loop (some S3 implementations may bug)
			if continuationToken == "" {
				objectCh <- types.ObjectInfo{
					Err: fmt.Errorf("list is truncated without continuation token"),
				}
				return
			}
		}
	}()

	return objectCh
}

// ListVersions lists object versions and delete markers.
func (s *objectService) ListVersions(ctx context.Context, bucketName string, opts ...ListOption) <-chan types.ObjectInfo {
	opts = append(opts, WithListVersions())
	return s.List(ctx, bucketName, opts...)
}

// streamObjectVersions streams object versions and delete markers using ListObjectVersions.
func (s *objectService) streamObjectVersions(ctx context.Context, bucketName string, options *ListOptions, objectCh chan<- types.ObjectInfo) error {
	// Set delimiter depending on recursive flag
	delimiter := "/"
	if options.Recursive {
		delimiter = ""
	}

	var keyMarker, versionIDMarker string
	for {
		// Context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		result, err := s.listObjectVersionsQuery(ctx, bucketName, options, delimiter, keyMarker, versionIDMarker)
		if err != nil {
			return err
		}

		for _, version := range result.Versions {
			version.ETag = trimETag(version.ETag)
			select {
			case objectCh <- version:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		for _, marker := range result.DeleteMarkers {
			select {
			case objectCh <- marker:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Next markers
		keyMarker = result.NextKeyMarker
		versionIDMarker = result.NextVersionIdMarker

		if !result.IsTruncated {
			return nil
		}
		// guard infinite loop
		if keyMarker == "" && versionIDMarker == "" {
			return fmt.Errorf("version list truncated without next markers")
		}
	}
}

// listObjectsV2Query queries object list V2
func (s *objectService) listObjectsV2Query(ctx context.Context, bucketName string, options *ListOptions, delimiter, continuationToken string) (ListBucketV2Result, error) {
	// Build query parameters
	queryValues := url.Values{}

	// Set list-type=2 (V2)
	queryValues.Set("list-type", "2")

	// Set encoding-type
	queryValues.Set("encoding-type", "url")

	// Set prefix
	if options.Prefix != "" {
		queryValues.Set("prefix", options.Prefix)
	}

	// Set delimiter
	if delimiter != "" {
		queryValues.Set("delimiter", delimiter)
	}

	// Set start-after
	if options.StartAfter != "" {
		queryValues.Set("start-after", options.StartAfter)
	}

	// Set continuation-token
	if continuationToken != "" {
		queryValues.Set("continuation-token", continuationToken)
	}

	// Set max-keys
	maxKeys := options.MaxKeys
	if maxKeys <= 0 {
		maxKeys = 1000 // default max
	}
	queryValues.Set("max-keys", strconv.Itoa(maxKeys))

	// Set fetch-owner
	queryValues.Set("fetch-owner", "true")

	// Set metadata
	if options.WithMetadata {
		queryValues.Set("metadata", "true")
	}

	// Build request metadata
	meta := core.RequestMetadata{
		BucketName:   bucketName,
		QueryValues:  queryValues,
		CustomHeader: options.CustomHeaders,
	}

	// Create GET request
	req := core.NewRequest(ctx, http.MethodGet, meta)

	// Execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return ListBucketV2Result{}, err
	}
	defer closeResponse(resp)

	// Check response
	if resp.StatusCode != http.StatusOK {
		return ListBucketV2Result{}, parseErrorResponse(resp, bucketName, "")
	}

	// Decode XML response
	var result ListBucketV2Result
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return ListBucketV2Result{}, fmt.Errorf("failed to decode list objects response: %w", err)
	}

	// URL decode object names (due to encoding-type=url)
	for i := range result.Contents {
		if decodedKey, err := url.QueryUnescape(result.Contents[i].Key); err == nil {
			result.Contents[i].Key = decodedKey
		}
	}

	// URL decode common prefixes
	for i := range result.CommonPrefixes {
		if decodedPrefix, err := url.QueryUnescape(result.CommonPrefixes[i].Prefix); err == nil {
			result.CommonPrefixes[i].Prefix = decodedPrefix
		}
	}

	return result, nil
}

// listObjectVersionsResult represents ListObjectVersions response
type listObjectVersionsResult struct {
	XMLName             xml.Name           `xml:"ListVersionsResult"`
	Name                string             `xml:"Name"`
	Prefix              string             `xml:"Prefix"`
	KeyMarker           string             `xml:"KeyMarker"`
	VersionIdMarker     string             `xml:"VersionIdMarker"`
	NextKeyMarker       string             `xml:"NextKeyMarker"`
	NextVersionIdMarker string             `xml:"NextVersionIdMarker"`
	MaxKeys             int                `xml:"MaxKeys"`
	IsTruncated         bool               `xml:"IsTruncated"`
	Versions            []types.ObjectInfo `xml:"Version"`
	DeleteMarkers       []types.ObjectInfo `xml:"DeleteMarker"`
}

// listObjectVersionsQuery queries versions and delete markers
func (s *objectService) listObjectVersionsQuery(ctx context.Context, bucketName string, options *ListOptions, delimiter, keyMarker, versionIDMarker string) (listObjectVersionsResult, error) {
	queryValues := url.Values{
		"versions": {"true"},
	}

	if options.Prefix != "" {
		queryValues.Set("prefix", options.Prefix)
	}
	if delimiter != "" {
		queryValues.Set("delimiter", delimiter)
	}
	if options.StartAfter != "" && keyMarker == "" {
		queryValues.Set("key-marker", options.StartAfter)
	} else if keyMarker != "" {
		queryValues.Set("key-marker", keyMarker)
	}
	if versionIDMarker != "" {
		queryValues.Set("version-id-marker", versionIDMarker)
	}

	maxKeys := options.MaxKeys
	if maxKeys <= 0 {
		maxKeys = 1000
	}
	queryValues.Set("max-keys", strconv.Itoa(maxKeys))

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		QueryValues:  queryValues,
		CustomHeader: options.CustomHeaders,
	}

	req := core.NewRequest(ctx, http.MethodGet, meta)

	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return listObjectVersionsResult{}, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return listObjectVersionsResult{}, parseErrorResponse(resp, bucketName, "")
	}

	var result listObjectVersionsResult
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return listObjectVersionsResult{}, fmt.Errorf("failed to decode list object versions response: %w", err)
	}

	// Mark delete markers explicitly
	for i := range result.DeleteMarkers {
		result.DeleteMarkers[i].IsDeleteMarker = true
	}

	return result, nil
}

// trimETag removes quotes from ETag
func trimETag(etag string) string {
	if len(etag) >= 2 && etag[0] == '"' && etag[len(etag)-1] == '"' {
		return etag[1 : len(etag)-1]
	}
	return etag
}
