// Package rustfs provides client interface for RustFS object storage
package rustfs

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

// BucketLookupType - type of bucket lookup.
type BucketLookupType int

const (
	// BucketLookupAuto - enable automatic bucket lookup by method.
	BucketLookupAuto BucketLookupType = iota
	// BucketLookupDNS - lookup is performed using DNS style bucket
	// names (default).
	BucketLookupDNS
	// BucketLookupPath - lookup is performed using path style bucket
	// names.
	BucketLookupPath
)

// Options - RustFS client options
type Options struct {
	Creds        *credentials.Credentials
	Secure       bool
	Region       string
	BucketLookup BucketLookupType
	Transport    http.RoundTripper
	CustomMD5    bool
}

// Client - RustFS client structure
type Client struct {
	endpoint     string
	accessKey    string
	secretKey    string
	secure       bool
	region       string
	bucketLookup BucketLookupType
	httpClient   *http.Client
	creds        *credentials.Credentials
}

// New - instantiate a new RustFS client
func New(endpoint string, opts *Options) (*Client, error) {
	if endpoint == "" {
		return nil, ErrInvalidArgument("endpoint cannot be empty")
	}

	client := &Client{
		endpoint:     endpoint,
		secure:       true,
		region:       "us-east-1",
		bucketLookup: BucketLookupDNS,
	}

	if opts != nil {
		if opts.Creds != nil {
			client.creds = opts.Creds
		}
		if opts.Secure {
			client.secure = opts.Secure
		}
		if opts.Region != "" {
			client.region = opts.Region
		}
		if opts.BucketLookup != 0 {
			client.bucketLookup = opts.BucketLookup
		}
		if opts.Transport != nil {
			client.httpClient = &http.Client{
				Transport: opts.Transport,
			}
		}
	}

	// Set default HTTP client if not provided
	if client.httpClient == nil {
		tr := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		if client.secure {
			tr.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: false,
			}
		}

		client.httpClient = &http.Client{
			Transport: tr,
			Timeout:   0, // No timeout
		}
	}

	// Get credentials
	if client.creds != nil {
		creds, err := client.creds.Get()
		if err != nil {
			return nil, err
		}
		client.accessKey = creds.AccessKeyID
		client.secretKey = creds.SecretAccessKey
	}

	return client, nil
}

// SetAppInfo - add application details to user agent.
func (c *Client) SetAppInfo(appName string, appVersion string) {
	// Implementation for setting app info
	// This can be used to track SDK usage
}

// GetRegion - get current region.
func (c *Client) GetRegion() string {
	return c.region
}

// SetRegion - set new region.
func (c *Client) SetRegion(region string) {
	c.region = region
}

// ErrInvalidArgument - invalid argument error
func ErrInvalidArgument(message string) error {
	return &ErrorResponse{
		Code:    "InvalidArgument",
		Message: message,
	}
}

// ErrorResponse - error response
type ErrorResponse struct {
	Code    string
	Message string
}

func (e *ErrorResponse) Error() string {
	return e.Message
}
