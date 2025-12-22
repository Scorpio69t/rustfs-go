// Package bucket bucket/list.go
package bucket

import (
	"context"
	"encoding/xml"
	"net/http"

	"github.com/Scorpio69t/rustfs-go/errors"
	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/types"
)

// List all buckets
func (s *bucketService) List(ctx context.Context) ([]types.BucketInfo, error) {
	// prepare request metadata
	meta := core.RequestMetadata{}

	// create GET request
	req := core.NewRequest(ctx, http.MethodGet, meta)

	// execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)

	if resp == nil {
		return nil, errors.ErrNilResponse
	}

	// check response status code
	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp, "", "")
	}

	// parse response
	var result listAllMyBucketsResult
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Buckets.Bucket, nil
}

// GetLocation gets the location of a bucket
func (s *bucketService) GetLocation(ctx context.Context, bucketName string) (string, error) {
	// validate bucket name
	if err := validateBucketName(bucketName); err != nil {
		return "", err
	}

	// check cache
	if s.locationCache != nil {
		if location, ok := s.locationCache.Get(bucketName); ok {
			return location, nil
		}
	}

	// prepare request metadata
	meta := core.RequestMetadata{
		BucketName: bucketName,
		QueryValues: map[string][]string{
			"location": {""},
		},
	}

	// create GET request
	req := core.NewRequest(ctx, http.MethodGet, meta)

	// execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return "", err
	}
	defer closeResponse(resp)

	if resp == nil {
		return "", errors.ErrNilResponse
	}

	// check response status code
	if resp.StatusCode != http.StatusOK {
		return "", parseErrorResponse(resp, bucketName, "")
	}

	// parse response
	var result locationConstraint
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	location := result.Location
	if location == "" {
		location = "us-east-1"
	}

	// cache the location
	if s.locationCache != nil {
		s.locationCache.Set(bucketName, location)
	}

	return location, nil
}

// listAllMyBucketsResult list all buckets result
type listAllMyBucketsResult struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	Owner   owner    `xml:"Owner"`
	Buckets buckets  `xml:"Buckets"`
}

// owner info
type owner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

// buckets list of buckets
type buckets struct {
	Bucket []types.BucketInfo `xml:"Bucket"`
}

// locationConstraint location constraint
type locationConstraint struct {
	XMLName  xml.Name `xml:"LocationConstraint"`
	Location string   `xml:",chardata"`
}
