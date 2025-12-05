// Package rustfs provides client interface for RustFS object storage
package rustfs

import (
	"errors"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptrace"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/kvcache"
	"github.com/Scorpio69t/rustfs-go/pkg/s3utils"
	"github.com/Scorpio69t/rustfs-go/pkg/singleflight"
	md5simd "github.com/minio/md5-simd"
	"golang.org/x/net/publicsuffix"
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

// Client implements Amazon S3 compatible methods.
type Client struct {
	//  Standard options.

	// Parsed endpoint url provided by the user.
	endpointURL *url.URL

	// Holds various credential providers.
	credsProvider *credentials.Credentials

	// Custom signerType value overrides all credentials.
	overrideSignerType credentials.SignatureType

	// User supplied.
	appInfo struct {
		appName    string
		appVersion string
	}

	// Indicate whether we are using https or not
	secure bool

	// Needs allocation.
	httpClient         *http.Client
	httpTrace          *httptrace.ClientTrace
	bucketLocCache     *kvcache.Cache[string, string]
	bucketSessionCache *kvcache.Cache[string, credentials.Value]
	credsGroup         singleflight.Group[string, credentials.Value]

	// Advanced functionality.
	isTraceEnabled  bool
	traceErrorsOnly bool
	traceOutput     io.Writer

	// S3 specific accelerated endpoint.
	s3AccelerateEndpoint string
	// S3 dual-stack endpoints are enabled by default.
	s3DualstackEnabled bool

	// Region endpoint
	region string

	// Random seed.
	random *rand.Rand

	// lookup indicates type of url lookup supported by server. If not specified,
	// default to Auto.
	lookup BucketLookupType

	// lookupFn is a custom function to return URL lookup type supported by the server.
	lookupFn func(u url.URL, bucketName string) BucketLookupType

	// Factory for MD5 hash functions.
	md5Hasher    func() md5simd.Hasher
	sha256Hasher func() md5simd.Hasher

	healthStatus int32

	trailingHeaderSupport bool
	maxRetries            int
}

// New - instantiate minio client with options
func New(endpoint string, opts *Options) (*Client, error) {
	if opts == nil {
		return nil, errors.New("no options provided")
	}
	clnt, err := privateNew(endpoint, opts)
	if err != nil {
		return nil, err
	}
	if s3utils.IsAmazonEndpoint(*clnt.endpointURL) {
		// If Amazon S3 set to signature v4.
		clnt.overrideSignerType = credentials.SignatureV4
		// Amazon S3 endpoints are resolved into dual-stack endpoints by default
		// for backwards compatibility.
		clnt.s3DualstackEnabled = true
	}

	return clnt, nil
}

func privateNew(endpoint string, opts *Options) (*Client, error) {
	// construct endpoint.
	endpointURL, err := getEndpointURL(endpoint, opts.Secure)
	if err != nil {
		return nil, err
	}

	// Initialize cookies to preserve server sent cookies if any and replay
	// them upon each request.
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}

	// instantiate new Client.
	clnt := new(Client)

	// Save the credentials.
	clnt.credsProvider = opts.Creds

	// Remember whether we are using https or not
	clnt.secure = opts.Secure

	// Save endpoint URL, user agent for future uses.
	clnt.endpointURL = endpointURL

	transport := opts.Transport
	if transport == nil {
		transport, err = DefaultTransport(opts.Secure)
		if err != nil {
			return nil, err
		}
	}

	clnt.httpTrace = opts.Trace

	// Instantiate http client and bucket location cache.
	clnt.httpClient = &http.Client{
		Jar:       jar,
		Transport: transport,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Sets custom region, if region is empty bucket location cache is used automatically.
	if opts.Region == "" {
		if opts.CustomRegionViaURL != nil {
			opts.Region = opts.CustomRegionViaURL(*clnt.endpointURL)
		} else {
			opts.Region = s3utils.GetRegionFromURL(*clnt.endpointURL)
		}
	}
	clnt.region = opts.Region

	// Initialize bucket region cache.
	clnt.bucketLocCache = &kvcache.Cache[string, string]{}

	// Initialize bucket session cache (s3 express).
	clnt.bucketSessionCache = &kvcache.Cache[string, credentials.Value]{}

	// Introduce a new locked random seed.
	clnt.random = rand.New(&lockedRandSource{src: rand.NewSource(time.Now().UTC().UnixNano())})

	// Add default md5 hasher.
	clnt.md5Hasher = opts.CustomMD5
	clnt.sha256Hasher = opts.CustomSHA256
	if clnt.md5Hasher == nil {
		clnt.md5Hasher = newMd5Hasher
	}
	if clnt.sha256Hasher == nil {
		clnt.sha256Hasher = newSHA256Hasher
	}

	clnt.trailingHeaderSupport = opts.TrailingHeaders && clnt.overrideSignerType.IsV4()

	// Sets bucket lookup style, whether server accepts DNS or Path lookup. Default is Auto - determined
	// by the SDK. When Auto is specified, DNS lookup is used for Amazon/Google cloud endpoints and Path for all other endpoints.
	clnt.lookup = opts.BucketLookup
	clnt.lookupFn = opts.BucketLookupViaURL

	// healthcheck is not initialized
	clnt.healthStatus = unknown

	clnt.maxRetries = MaxRetry
	if opts.MaxRetries > 0 {
		clnt.maxRetries = opts.MaxRetries
	}

	// Return.
	return clnt, nil
}
