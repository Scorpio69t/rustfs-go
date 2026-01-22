// Package core internal/core/executor.go
package core

import (
	"context"
	"encoding/xml"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Scorpio69t/rustfs-go/errors"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/signer"
	"github.com/Scorpio69t/rustfs-go/types"
)

// Executor executes HTTP requests against RustFS/S3 endpoints
type Executor struct {
	// HTTP client used to perform requests
	httpClient *http.Client

	// Endpoint information
	endpointURL *url.URL

	// Credentials provider
	credentials *credentials.Credentials

	// Target region
	region string

	// Whether to use HTTPS
	secure bool

	// Signature type
	signerType credentials.SignatureType

	// Bucket lookup style
	bucketLookup int

	// Maximum retry attempts
	maxRetries int

	// Bucket location cache
	locationCache LocationCache
}

// ExecutorConfig configures an Executor
type ExecutorConfig struct {
	HTTPClient    *http.Client
	EndpointURL   *url.URL
	Credentials   *credentials.Credentials
	Region        string
	Secure        bool
	BucketLookup  int
	MaxRetries    int
	LocationCache LocationCache
}

// NewExecutor creates a new Executor
func NewExecutor(config ExecutorConfig) *Executor {
	maxRetries := config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 10
	}

	return &Executor{
		httpClient:    config.HTTPClient,
		endpointURL:   config.EndpointURL,
		credentials:   config.Credentials,
		region:        config.Region,
		secure:        config.Secure,
		bucketLookup:  config.BucketLookup,
		maxRetries:    maxRetries,
		locationCache: config.LocationCache,
	}
}

// Execute performs the request with retries and signing
func (e *Executor) Execute(ctx context.Context, req *Request) (*http.Response, error) {
	var (
		resp    *http.Response
		err     error
		httpReq *http.Request
	)

	meta := req.Metadata()

	// Retry loop
	for attempt := 0; attempt < e.maxRetries; attempt++ {
		// Check context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Build HTTP request
		httpReq, err = e.buildHTTPRequest(ctx, req, meta)
		if err != nil {
			return nil, err
		}

		// Execute request
		resp, err = e.httpClient.Do(httpReq)
		if err != nil {
			if e.shouldRetry(err, attempt) {
				if !resetRequestBody(&meta) {
					return nil, err
				}
				e.waitForRetry(ctx, attempt)
				continue
			}
			return nil, err
		}

		// Validate response
		if e.isSuccessStatus(resp.StatusCode, req.metadata.Expect200OKWithError) {
			return resp, nil
		}

		// Retry if the response warrants it
		if e.shouldRetryResponse(resp, attempt) {
			closeResponse(resp)
			if !resetRequestBody(&meta) {
				return resp, nil
			}
			e.waitForRetry(ctx, attempt)
			continue
		}

		// Return error response (non-retryable)
		return resp, nil
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// buildHTTPRequest constructs and signs the outbound HTTP request
func (e *Executor) buildHTTPRequest(ctx context.Context, req *Request, meta RequestMetadata) (*http.Request, error) {
	// Resolve bucket location
	location := meta.BucketLocation
	if location == "" && meta.BucketName != "" {
		location = e.getBucketLocation(ctx, meta.BucketName)
	}
	if location == "" {
		location = e.region
	}

	// Build target URL
	targetURL, err := e.makeTargetURL(meta.BucketName, meta.ObjectName, location, meta.QueryValues)
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method(), targetURL.String(), meta.ContentBody)
	if err != nil {
		return nil, err
	}

	// Set headers
	for k, v := range meta.CustomHeader {
		httpReq.Header[k] = v
	}

	// Add extra presign headers (used for signing only)
	if meta.PresignURL && meta.ExtraPresignHeader != nil {
		for k, v := range meta.ExtraPresignHeader {
			httpReq.Header[k] = v
		}
	}

	// Set Content-Length
	httpReq.ContentLength = meta.ContentLength

	// Set Content-SHA256 header (required for SigV4)
	if meta.ContentSHA256Hex != "" {
		httpReq.Header.Set("X-Amz-Content-Sha256", meta.ContentSHA256Hex)
	}

	// Sign the request
	if err := e.signRequest(httpReq, meta, location); err != nil {
		return nil, err
	}

	return httpReq, nil
}

// Presign builds and signs the request, returning the presigned URL and signed headers without executing it.
func (e *Executor) Presign(ctx context.Context, req *Request) (*url.URL, http.Header, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	httpReq, err := e.buildHTTPRequest(ctx, req, req.Metadata())
	if err != nil {
		return nil, nil, err
	}

	return httpReq.URL, httpReq.Header, nil
}

// GetCredentials returns the resolved credentials using the executor context.
func (e *Executor) GetCredentials(ctx context.Context) (credentials.Value, error) {
	_ = ctx
	if e.credentials == nil {
		return credentials.Value{SignerType: credentials.SignatureAnonymous}, nil
	}
	httpClient := e.httpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	credContext := &credentials.CredContext{
		Client:   httpClient,
		Endpoint: e.endpointURL.String(),
	}
	return e.credentials.GetWithContext(credContext)
}

// ResolveBucketLocation fetches the bucket location and updates the cache.
func (e *Executor) ResolveBucketLocation(ctx context.Context, bucketName string) (string, error) {
	if bucketName == "" {
		return e.region, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if e.locationCache != nil {
		if loc, ok := e.locationCache.Get(bucketName); ok {
			return loc, nil
		}
	}

	meta := RequestMetadata{
		BucketName: bucketName,
		QueryValues: url.Values{
			"location": {""},
		},
	}
	req := NewRequest(ctx, http.MethodGet, meta)
	resp, err := e.Execute(ctx, req)
	if err != nil {
		return "", err
	}
	defer closeResponse(resp)

	if resp == nil {
		return "", errors.NewAPIError(errors.ErrCodeInternalError, "empty response", http.StatusInternalServerError)
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.ParseErrorResponse(resp, bucketName, "")
	}

	var result struct {
		XMLName  xml.Name `xml:"LocationConstraint"`
		Location string   `xml:",chardata"`
	}
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	location := result.Location
	if location == "" {
		location = "us-east-1"
	}
	if e.locationCache != nil {
		e.locationCache.Set(bucketName, location)
	}
	return location, nil
}

// TargetURL returns a resolved target URL for a bucket/object and query.
func (e *Executor) TargetURL(ctx context.Context, bucketName, objectName string, query url.Values) (*url.URL, error) {
	location := e.region
	if bucketName != "" {
		resolved, err := e.ResolveBucketLocation(ctx, bucketName)
		if err != nil {
			return nil, err
		}
		if resolved != "" {
			location = resolved
		}
	}
	return e.makeTargetURL(bucketName, objectName, location, query)
}

// resetRequestBody attempts to rewind the request body for a retry.
// Returns false when the body cannot be replayed.
func resetRequestBody(meta *RequestMetadata) bool {
	if meta.ContentBody == nil {
		return true
	}
	if seeker, ok := meta.ContentBody.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err == nil {
			return true
		}
	}
	return false
}

// makeTargetURL builds the final request URL
func (e *Executor) makeTargetURL(bucketName, objectName, location string, queryValues url.Values) (*url.URL, error) {
	host := e.endpointURL.Host
	scheme := e.endpointURL.Scheme

	// Normalize default ports (strip :80 for HTTP and :443 for HTTPS)
	// Reason: browsers/curl drop default ports, which would break presigned URLs
	if h, p, err := net.SplitHostPort(host); err == nil {
		if (scheme == "http" && p == "80") || (scheme == "https" && p == "443") {
			host = h
			// Wrap IPv6 addresses in brackets
			if ip := net.ParseIP(h); ip != nil && ip.To4() == nil {
				host = "[" + h + "]"
			}
		}
	}

	urlStr := scheme + "://" + host + "/"

	// Build full URL when bucket is present
	if bucketName != "" {
		// Decide virtual-host vs path-style
		isVirtualHost := e.isVirtualHostStyleRequest(bucketName)

		if isVirtualHost {
			// Virtual-host style: http://bucket.host/object
			urlStr = scheme + "://" + bucketName + "." + host + "/"
			if objectName != "" {
				urlStr += encodePath(objectName)
			}
		} else {
			// Path style: http://host/bucket/object
			urlStr = urlStr + bucketName + "/"
			if objectName != "" {
				urlStr += encodePath(objectName)
			}
		}
	}

	// Append query parameters
	if len(queryValues) > 0 {
		urlStr = urlStr + "?" + queryEncode(queryValues)
	}

	return url.Parse(urlStr)
}

// isVirtualHostStyleRequest determines whether to use virtual-hosted style
func (e *Executor) isVirtualHostStyleRequest(bucketName string) bool {
	if bucketName == "" {
		return false
	}

	lookup := types.BucketLookupType(e.bucketLookup)

	switch lookup {
	case types.BucketLookupDNS:
		return true
	case types.BucketLookupPath:
		return false
	case types.BucketLookupAuto:
		// Auto-detect: ensure bucket is DNS compliant
		return isValidVirtualHostBucket(bucketName, e.endpointURL.Scheme == "https")
	}

	return false
}

// isValidVirtualHostBucket checks whether a bucket name is DNS-compliant for virtual-host style
func isValidVirtualHostBucket(bucketName string, https bool) bool {
	if strings.Contains(bucketName, ".") {
		// Buckets with dots break wildcard TLS certificates
		if https {
			return false
		}
	}
	// Length must be 3-63 characters
	if len(bucketName) < 3 || len(bucketName) > 63 {
		return false
	}
	// Reject IP-style names
	if net.ParseIP(bucketName) != nil {
		return false
	}
	return true
}

// encodePath URL-encodes path segments while preserving slashes
func encodePath(pathName string) string {
	if pathName == "" {
		return "/"
	}

	// Preserve slashes but encode other special characters (including '+')
	var encodedPathname strings.Builder
	for _, segment := range strings.Split(pathName, "/") {
		if encodedPathname.Len() > 0 {
			encodedPathname.WriteByte('/')
		}
		// url.PathEscape does not encode '+', so handle manually
		encoded := url.PathEscape(segment)
		encoded = strings.ReplaceAll(encoded, "+", "%2B")
		encodedPathname.WriteString(encoded)
	}

	result := encodedPathname.String()
	if result == "" {
		return "/"
	}
	return result
}

// queryEncode encodes query parameters
func queryEncode(v url.Values) string {
	if v == nil {
		return ""
	}
	// url.Values.Encode() already sorts and encodes
	return v.Encode()
}

// signRequest signs the request using the configured signer
func (e *Executor) signRequest(req *http.Request, meta RequestMetadata, location string) error {
	// Resolve credentials
	if e.credentials == nil {
		return nil // anonymous request
	}

	creds, err := e.GetCredentials(req.Context())
	if err != nil {
		return err
	}

	// Skip signing for anonymous credentials
	if creds.SignerType == credentials.SignatureAnonymous {
		return nil
	}

	// Prefer bucket location over default region
	region := location
	if region == "" {
		region = e.region
	}

	// Handle presign
	if meta.PresignURL {
		expires := time.Duration(meta.Expires) * time.Second
		sn := signer.NewSigner(convertSignerType(creds.SignerType))
		sn.Presign(req, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken, region, expires)
		return nil
	}

	// Sign standard requests
	sn := signer.NewSigner(convertSignerType(creds.SignerType))
	sn.Sign(req, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken, region)
	return nil
}

// convertSignerType maps credentials.SignatureType to signer.SignerType
func convertSignerType(st credentials.SignatureType) signer.SignerType {
	switch st {
	case credentials.SignatureV2:
		return signer.SignerV2
	case credentials.SignatureAnonymous:
		return signer.SignerAnonymous
	default:
		return signer.SignerV4
	}
}

// getBucketLocation returns bucket location from cache or defaults
func (e *Executor) getBucketLocation(ctx context.Context, bucketName string) string {
	if e.locationCache != nil {
		if loc, ok := e.locationCache.Get(bucketName); ok {
			return loc
		}
	}
	return e.region
}

// shouldRetry decides whether an error should be retried
func (e *Executor) shouldRetry(err error, attempt int) bool {
	if attempt >= e.maxRetries-1 {
		return false
	}

	// Check error string patterns
	if err == nil {
		return false
	}

	// Network error patterns
	errStr := err.Error()

	// Retry on connection refused/reset/timeouts and temporary failures
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "TLS handshake timeout") ||
		strings.Contains(errStr, "i/o timeout") ||
		strings.Contains(errStr, "net/http: request canceled") ||
		strings.Contains(errStr, "context deadline exceeded") {
		return true
	}

	// Inspect url.Error
	if urlErr, ok := err.(*url.Error); ok {
		if urlErr.Timeout() {
			return true
		}
		// Recursively inspect wrapped error
		return e.shouldRetry(urlErr.Err, attempt)
	}

	// Inspect net.Error
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}

	return false
}

// shouldRetryResponse decides whether to retry based on HTTP response
func (e *Executor) shouldRetryResponse(resp *http.Response, attempt int) bool {
	if attempt >= e.maxRetries-1 {
		return false
	}
	// Retry on 5xx
	if resp.StatusCode >= 500 {
		return true
	}
	// 429 Too Many Requests
	if resp.StatusCode == 429 {
		return true
	}
	return false
}

// waitForRetry pauses using exponential backoff
func (e *Executor) waitForRetry(ctx context.Context, attempt int) {
	// Exponential backoff
	delay := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
	if delay > 10*time.Second {
		delay = 10 * time.Second
	}

	select {
	case <-ctx.Done():
	case <-time.After(delay):
	}
}

// isSuccessStatus determines if the status code is considered success
func (e *Executor) isSuccessStatus(statusCode int, expect200OKWithError bool) bool {
	if expect200OKWithError {
		return false // body must be inspected for errors
	}
	return statusCode >= 200 && statusCode < 300
}

// LocationCache caches bucket locations
type LocationCache interface {
	Get(bucketName string) (string, bool)
	Set(bucketName, location string)
	Delete(bucketName string)
}

// closeResponse drains and closes the response body
func closeResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}
