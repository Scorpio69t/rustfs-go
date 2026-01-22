// Package object object/restore.go
package object

import (
	"bytes"
	"context"
	"encoding/xml"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/pkg/restore"
)

// Restore initiates a restore request for an archived object.
func (s *objectService) Restore(ctx context.Context, bucketName, objectName, versionID string, req restore.RestoreRequest) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	if err := validateObjectName(objectName); err != nil {
		return err
	}

	req.Normalize()

	restoreRequestBytes, err := xml.Marshal(req)
	if err != nil {
		return err
	}

	queryValues := url.Values{}
	queryValues.Set("restore", "")
	if versionID != "" {
		queryValues.Set("versionId", versionID)
	}

	meta := core.RequestMetadata{
		BucketName:       bucketName,
		ObjectName:       objectName,
		QueryValues:      queryValues,
		ContentMD5Base64: sumMD5Base64(restoreRequestBytes),
		ContentSHA256Hex: sumSHA256Hex(restoreRequestBytes),
		ContentBody:      bytes.NewReader(restoreRequestBytes),
		ContentLength:    int64(len(restoreRequestBytes)),
	}

	reqObj := core.NewRequest(ctx, http.MethodPost, meta)
	resp, err := s.executor.Execute(ctx, reqObj)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp, bucketName, objectName)
	}

	return nil
}
