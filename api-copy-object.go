package rustfs

import (
	"context"
	"io"
	"net/http"
)

// CopyObject - copy a source object into a new object
func (c *Client) CopyObject(ctx context.Context, dst CopyDestOptions, src CopySrcOptions) (UploadInfo, error) {
	if err := src.validate(); err != nil {
		return UploadInfo{}, err
	}

	if err := dst.validate(); err != nil {
		return UploadInfo{}, err
	}

	header := make(http.Header)
	dst.Marshal(header)
	src.Marshal(header)

	resp, err := c.executeMethod(ctx, http.MethodPut, requestMetadata{
		bucketName:   dst.Bucket,
		objectName:   dst.Object,
		customHeader: header,
	})
	if err != nil {
		return UploadInfo{}, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return UploadInfo{}, httpRespToErrorResponse(resp, dst.Bucket, dst.Object)
	}

	// Update the progress properly after successful copy.
	if dst.Progress != nil {
		io.Copy(io.Discard, io.LimitReader(dst.Progress, dst.Size))
	}

	cpObjRes := copyObjectResult{}
	if err = xmlDecoder(resp.Body, &cpObjRes); err != nil {
		return UploadInfo{}, err
	}

	// extract lifecycle expiry date and rule ID
	expTime, ruleID := amzExpirationToExpiryDateRuleID(resp.Header.Get(amzExpiration))

	return UploadInfo{
		Bucket:           dst.Bucket,
		Key:              dst.Object,
		LastModified:     cpObjRes.LastModified,
		ETag:             trimEtag(cpObjRes.ETag),
		VersionID:        resp.Header.Get(amzVersionID),
		Expiration:       expTime,
		ExpirationRuleID: ruleID,
	}, nil
}
