// Package object object/select.go
package object

import (
	"bytes"
	"context"
	"encoding/xml"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	s3select "github.com/Scorpio69t/rustfs-go/pkg/select"
)

// Select queries object content using S3 Select.
func (s *objectService) Select(ctx context.Context, bucketName, objectName string, opts s3select.Options) (*s3select.Results, error) {
	if err := validateBucketName(bucketName); err != nil {
		return nil, err
	}
	if err := validateObjectName(objectName); err != nil {
		return nil, err
	}

	selectReqBytes, err := xml.Marshal(opts)
	if err != nil {
		return nil, err
	}

	queryValues := url.Values{}
	queryValues.Set("select", "")
	queryValues.Set("select-type", "2")

	meta := core.RequestMetadata{
		BucketName:       bucketName,
		ObjectName:       objectName,
		QueryValues:      queryValues,
		CustomHeader:     opts.Header(),
		ContentMD5Base64: sumMD5Base64(selectReqBytes),
		ContentSHA256Hex: sumSHA256Hex(selectReqBytes),
		ContentBody:      bytes.NewReader(selectReqBytes),
		ContentLength:    int64(len(selectReqBytes)),
	}

	req := core.NewRequest(ctx, http.MethodPost, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	return s3select.NewResults(resp, bucketName, objectName)
}
