package rustfs

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go/v1/pkg/s3signer"
)

// PresignedGetObject - generate presigned GET URL
func (c *Client) PresignedGetObject(ctx context.Context, bucketName, objectName string, expiry time.Duration, reqParams url.Values) (*url.URL, error) {
	// Validate inputs
	if bucketName == "" {
		return nil, fmt.Errorf("bucket name cannot be empty")
	}
	if objectName == "" {
		return nil, fmt.Errorf("object name cannot be empty")
	}

	// Build query values
	queryValues := make(url.Values)
	for k, v := range reqParams {
		queryValues[k] = v
	}

	// Build URL
	u, err := c.makeTargetURL(bucketName, objectName, queryValues)
	if err != nil {
		return nil, err
	}

	// Create request for signing
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Sign request
	if c.accessKey != "" && c.secretKey != "" {
		expiryTime := time.Now().Add(expiry)
		err = s3signer.SignV4Presigned(req, c.accessKey, c.secretKey, c.region, "s3", expiryTime)
		if err != nil {
			return nil, err
		}
	}

	return req.URL, nil
}

// PresignedPutObject - generate presigned PUT URL
func (c *Client) PresignedPutObject(ctx context.Context, bucketName, objectName string, expiry time.Duration) (*url.URL, error) {
	// Validate inputs
	if bucketName == "" {
		return nil, fmt.Errorf("bucket name cannot be empty")
	}
	if objectName == "" {
		return nil, fmt.Errorf("object name cannot be empty")
	}

	// Build URL
	u, err := c.makeTargetURL(bucketName, objectName, make(url.Values))
	if err != nil {
		return nil, err
	}

	// Create request for signing
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Sign request
	if c.accessKey != "" && c.secretKey != "" {
		expiryTime := time.Now().Add(expiry)
		err = s3signer.SignV4Presigned(req, c.accessKey, c.secretKey, c.region, "s3", expiryTime)
		if err != nil {
			return nil, err
		}
	}

	return req.URL, nil
}

// PostPolicy - POST policy for presigned POST
type PostPolicy struct {
	Expiration time.Time
	Conditions []map[string]interface{}
}

// PresignedPostPolicy - generate presigned POST URL and form data
func (c *Client) PresignedPostPolicy(ctx context.Context, policy *PostPolicy) (*url.URL, map[string]string, error) {
	if policy == nil {
		return nil, nil, fmt.Errorf("policy cannot be nil")
	}

	// Build form data
	formData := make(map[string]string)

	// Add policy conditions
	for _, condition := range policy.Conditions {
		for k, v := range condition {
			formData[k] = fmt.Sprintf("%v", v)
		}
	}

	// Add expiration
	formData["expiration"] = policy.Expiration.Format(time.RFC3339)

	// Add credentials
	if c.accessKey != "" {
		formData["AWSAccessKeyId"] = c.accessKey
	}

	// Generate signature
	if c.secretKey != "" {
		// Create policy string
		policyStr := fmt.Sprintf(`{"expiration":"%s","conditions":[`, policy.Expiration.Format(time.RFC3339))
		for i, condition := range policy.Conditions {
			if i > 0 {
				policyStr += ","
			}
			policyStr += "{"
			first := true
			for k, v := range condition {
				if !first {
					policyStr += ","
				}
				policyStr += fmt.Sprintf(`"%s":"%v"`, k, v)
				first = false
			}
			policyStr += "}"
		}
		policyStr += "]}"

		// Sign policy
		signature := s3signer.SignPolicy(policyStr, c.secretKey)
		formData["signature"] = signature
	}

	// Build URL (use endpoint)
	scheme := "http"
	if c.secure {
		scheme = "https"
	}
	u := &url.URL{
		Scheme: scheme,
		Host:   c.endpoint,
		Path:   "/",
	}

	return u, formData, nil
}
