// Package rustfs options.go
package rustfs

import (
	"net/http"
	"net/http/httptrace"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/types"
)

// Options contains client configuration options
type Options struct {
	// Credentials is the credential provider
	// Required, used for signing requests
	Credentials *credentials.Credentials

	// Secure indicates whether to use HTTPS
	// Default: false
	Secure bool

	// Region is the region
	// If not set, will be automatically detected
	Region string

	// Transport is a custom HTTP transport
	// If not set, default transport will be used
	Transport http.RoundTripper

	// Trace is the HTTP trace client
	Trace *httptrace.ClientTrace

	// BucketLookup is the bucket lookup type
	// Default: BucketLookupAuto
	BucketLookup types.BucketLookupType

	// CustomRegionViaURL is a custom region lookup function
	CustomRegionViaURL func(u url.URL) string

	// BucketLookupViaURL is a custom bucket lookup function
	BucketLookupViaURL func(u url.URL, bucketName string) types.BucketLookupType

	// TrailingHeaders enables trailing headers (for streaming upload)
	// Requires server support
	TrailingHeaders bool

	// MaxRetries is the maximum number of retries
	// Default: 10, set to 1 to disable retries
	MaxRetries int
}

// validate validates options
func (o *Options) validate() error {
	if o == nil {
		return errInvalidArgument("options cannot be nil")
	}
	if o.Credentials == nil {
		return errInvalidArgument("credentials are required")
	}
	return nil
}

// setDefaults sets default values
func (o *Options) setDefaults() {
	if o.MaxRetries <= 0 {
		o.MaxRetries = 10
	}
	if o.BucketLookup == 0 {
		o.BucketLookup = types.BucketLookupAuto
	}
}

// errInvalidArgument creates an invalid argument error
func errInvalidArgument(message string) error {
	return &invalidArgumentError{message: message}
}

// invalidArgumentError represents an invalid argument error type
type invalidArgumentError struct {
	message string
}

// Error returns the error message
func (e *invalidArgumentError) Error() string {
	return e.message
}
