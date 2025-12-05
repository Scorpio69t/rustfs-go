package rustfs

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go/v1/pkg/s3errors"
	"github.com/Scorpio69t/rustfs-go/v1/pkg/s3signer"
)

// executeMethod - execute HTTP request
func (c *Client) executeMethod(ctx context.Context, method string, metadata requestMetadata) (res *http.Response, err error) {
	// Create request
	req, err := c.buildRequest(ctx, method, metadata)
	if err != nil {
		return nil, err
	}

	// Sign request
	if c.accessKey != "" && c.secretKey != "" {
		err = s3signer.SignV4(req, c.accessKey, c.secretKey, c.region, "s3", time.Now())
		if err != nil {
			return nil, err
		}
	}

	// Execute request
	res, err = c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	// Check for errors
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		defer res.Body.Close()
		return nil, s3errors.ToErrorResponse(res, metadata.bucketName, metadata.objectName)
	}

	return res, nil
}

// requestMetadata - request metadata
type requestMetadata struct {
	bucketName       string
	objectName       string
	queryValues      url.Values
	customHeader     http.Header
	contentBody      io.Reader
	contentLength    int64
	contentMD5Base64 string
	contentSHA256Hex string
}

// buildRequest - build HTTP request
func (c *Client) buildRequest(ctx context.Context, method string, metadata requestMetadata) (*http.Request, error) {
	// Build URL
	u, err := c.makeTargetURL(metadata.bucketName, metadata.objectName, metadata.queryValues)
	if err != nil {
		return nil, err
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, u.String(), metadata.contentBody)
	if err != nil {
		return nil, err
	}

	// Set headers
	if metadata.contentLength > 0 {
		req.ContentLength = metadata.contentLength
	}
	if metadata.contentMD5Base64 != "" {
		req.Header.Set("Content-MD5", metadata.contentMD5Base64)
	}
	if metadata.contentSHA256Hex != "" {
		req.Header.Set("X-Amz-Content-Sha256", metadata.contentSHA256Hex)
	}

	// Set custom headers
	for k, v := range metadata.customHeader {
		for _, val := range v {
			req.Header.Add(k, val)
		}
	}

	// Set default headers
	req.Header.Set("User-Agent", "rustfs-go/1.0.0")
	req.Header.Set("Accept", "*/*")

	return req, nil
}

// makeTargetURL - make target URL
func (c *Client) makeTargetURL(bucketName, objectName string, queryValues url.Values) (*url.URL, error) {
	// Determine scheme
	scheme := "http"
	if c.secure {
		scheme = "https"
	}

	// Build path
	path := "/"
	if bucketName != "" {
		if c.bucketLookup == BucketLookupPath {
			path += bucketName + "/"
		} else {
			// DNS style
			host := bucketName + "." + c.endpoint
			u := &url.URL{
				Scheme:   scheme,
				Host:     host,
				Path:     "/" + objectName,
				RawQuery: queryValues.Encode(),
			}
			return u, nil
		}
	}
	if objectName != "" {
		path += objectName
	}

	// Build URL
	u := &url.URL{
		Scheme:   scheme,
		Host:     c.endpoint,
		Path:     path,
		RawQuery: queryValues.Encode(),
	}

	return u, nil
}

// parseResponse - parse XML response
func parseResponse(body io.Reader, v interface{}) error {
	if body == nil {
		return fmt.Errorf("response body is nil")
	}
	return xml.NewDecoder(body).Decode(v)
}

// readAll - read all data from reader
func readAll(r io.Reader) ([]byte, error) {
	if r == nil {
		return nil, fmt.Errorf("reader is nil")
	}
	return io.ReadAll(r)
}

// closeResponse - close response body
func closeResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// getObjectResponse - get object response wrapper
type getObjectResponse struct {
	Body       io.ReadCloser
	Headers    http.Header
	StatusCode int
}

// putObjectResponse - put object response
type putObjectResponse struct {
	ETag         string
	VersionID    string
	LastModified time.Time
}

// uploadPartResponse - upload part response
type uploadPartResponse struct {
	ETag       string
	PartNumber int
}
