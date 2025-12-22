// Package rustfs client.go - RustFS Go SDK client entrypoint
package rustfs

import (
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/Scorpio69t/rustfs-go/bucket"
	"github.com/Scorpio69t/rustfs-go/internal/cache"
	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/internal/transport"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/types"
	"golang.org/x/net/publicsuffix"
)

// Client is the RustFS client
type Client struct {
	// Core components
	executor      *core.Executor
	locationCache *cache.LocationCache

	// Service modules
	bucketService bucket.Service
	objectService object.Service

	// Client info
	endpointURL *url.URL
	httpClient  *http.Client
	secure      bool
	region      string

	// Application info
	appInfo struct {
		appName    string
		appVersion string
	}
}

// New creates a new RustFS client
//
// Parameters:
//   - endpoint: RustFS server address (e.g., "localhost:9000", "rustfs.example.com")
//   - opts: client configuration options
//
// Returns:
//   - *Client: client instance
//   - error: error details
//
// Example:
//
//	client, err := rustfs.New("localhost:9000", &rustfs.Options{
//	    Credentials: credentials.NewStaticV4("access-key", "secret-key", ""),
//	    Secure:      false,
//	})
func New(endpoint string, opts *Options) (*Client, error) {
	// Validate options
	if err := opts.validate(); err != nil {
		return nil, err
	}

	// Apply defaults
	opts.setDefaults()

	// Parse endpoint URL
	endpointURL, err := parseEndpointURL(endpoint, opts.Secure)
	if err != nil {
		return nil, err
	}

	// If BucketLookup is Auto and endpoint is an IP, force path-style
	if opts.BucketLookup == types.BucketLookupAuto && isIPAddress(endpointURL.Host) {
		opts.BucketLookup = types.BucketLookupPath
	}

	// Create HTTP transport
	var httpTransport http.RoundTripper
	if opts.Transport != nil {
		httpTransport = opts.Transport
	} else {
		httpTransport = transport.NewTransport(transport.TransportOptions{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90,
			EnableCompression:   false,
		})
	}

	// Create cookie jar
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil, err
	}

	// Create HTTP client
	httpClient := &http.Client{
		Jar:       jar,
		Transport: httpTransport,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Determine region
	region := opts.Region
	if region == "" {
		region = detectRegion(endpointURL, opts.CustomRegionViaURL)
	}

	// Create location cache (0 = no expiration)
	locationCache := cache.NewLocationCache(0)

	// Create core executor
	executor := core.NewExecutor(core.ExecutorConfig{
		HTTPClient:   httpClient,
		EndpointURL:  endpointURL,
		Credentials:  opts.Credentials,
		Region:       region,
		BucketLookup: int(opts.BucketLookup),
		MaxRetries:   opts.MaxRetries,
	})

	// Create service instances
	bucketService := bucket.NewService(executor, locationCache)
	objectService := object.NewService(executor, locationCache)

	// Construct client
	client := &Client{
		executor:      executor,
		locationCache: locationCache,
		bucketService: bucketService,
		objectService: objectService,
		endpointURL:   endpointURL,
		httpClient:    httpClient,
		secure:        opts.Secure,
		region:        region,
	}

	return client, nil
}

// Bucket returns the Bucket service interface
//
// Example:
//
//	err := client.Bucket().Create(ctx, "my-bucket")
func (c *Client) Bucket() bucket.Service {
	return c.bucketService
}

// Object returns the Object service interface
//
// Example:
//
//	info, err := client.Object().Put(ctx, "my-bucket", "my-object", reader, size)
func (c *Client) Object() object.Service {
	return c.objectService
}

// EndpointURL returns the client's endpoint URL
func (c *Client) EndpointURL() *url.URL {
	endpoint := *c.endpointURL // copy to avoid mutating internal state
	return &endpoint
}

// Region returns the configured region
func (c *Client) Region() string {
	return c.region
}

// IsSecure reports whether HTTPS is used
func (c *Client) IsSecure() bool {
	return c.secure
}

// SetAppInfo sets application info appended to the User-Agent header
//
// Parameters:
//   - appName: application name
//   - appVersion: application version
func (c *Client) SetAppInfo(appName, appVersion string) {
	if appName != "" && appVersion != "" {
		c.appInfo.appName = appName
		c.appInfo.appVersion = appVersion
	}
}

// parseEndpointURL parses and normalizes the endpoint URL
func parseEndpointURL(endpoint string, secure bool) (*url.URL, error) {
	if endpoint == "" {
		return nil, errInvalidArgument("endpoint cannot be empty")
	}

	// Add default scheme if missing
	scheme := "http"
	if secure {
		scheme = "https"
	}

	// Prepend scheme when absent
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = scheme + "://" + endpoint
	}

	// Parse URL
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	// Validate scheme
	if endpointURL.Scheme != "http" && endpointURL.Scheme != "https" {
		return nil, errInvalidArgument("endpoint scheme must be http or https")
	}

	return endpointURL, nil
}

// detectRegion infers region from endpoint or falls back to default
func detectRegion(endpointURL *url.URL, customRegionFn func(url.URL) string) string {
	if customRegionFn != nil {
		return customRegionFn(*endpointURL)
	}

	// Default region
	return "us-east-1"
}

// isIPAddress checks whether the host (with optional port) is an IP address
func isIPAddress(host string) bool {
	// Strip port if present
	hostOnly := host
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		hostOnly = host[:colonIndex]
	}

	// Check IP format
	return net.ParseIP(hostOnly) != nil
}

// HealthCheck performs a simple HEAD request to verify connectivity
//
// Parameters:
//   - opts: health check options (nil uses defaults)
//
// Returns:
//   - *core.HealthCheckResult: health check result
//
// Example:
//
//	result := client.HealthCheck(nil)
//	if result.Healthy {
//	    fmt.Printf("Service healthy, response time: %v\n", result.ResponseTime)
//	}
func (c *Client) HealthCheck(opts *core.HealthCheckOptions) *core.HealthCheckResult {
	return c.executor.HealthCheck(opts)
}

// HealthCheckWithRetry performs a health check with retries
//
// Parameters:
//   - opts: health check options
//   - maxRetries: maximum retries (<= 0 defaults to 3)
//
// Returns:
//   - *core.HealthCheckResult: final health check result
//
// Example:
//
//	result := client.HealthCheckWithRetry(&core.HealthCheckOptions{
//	    Timeout: 5 * time.Second,
//	}, 3)
func (c *Client) HealthCheckWithRetry(opts *core.HealthCheckOptions, maxRetries int) *core.HealthCheckResult {
	return c.executor.HealthCheckWithRetry(opts, maxRetries)
}
