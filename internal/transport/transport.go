// Package transport internal/transport/transport.go
package transport

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

// DefaultTransport creates default HTTP transport
// Similar to http.DefaultTransport but disables compression to avoid auto-decompressing gzip bodies
func DefaultTransport(secure bool) (*http.Transport, error) {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          256,
		MaxIdleConnsPerHost:   16,
		ResponseHeaderTimeout: time.Minute,
		IdleConnTimeout:       time.Minute,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
		// Disable compression to avoid automatic decoding of content-encoding=gzip
		// Ref: https://golang.org/src/net/http/transport.go?h=roundTrip#L1843
		DisableCompression: true,
	}

	if secure {
		tr.TLSClientConfig = &tls.Config{
			// Security considerations:
			// - No SSLv3 (POODLE/BEAST)
			// - No TLSv1.0 (CBC-related POODLE/BEAST)
			// - No TLSv1.1 (RC4 usage)
			MinVersion: tls.VersionTLS12,
		}

		// Support custom CA certificate via SSL_CERT_FILE environment variable
		if certFile := os.Getenv("SSL_CERT_FILE"); certFile != "" {
			rootCAs := mustGetSystemCertPool()
			data, err := os.ReadFile(certFile)
			if err == nil {
				rootCAs.AppendCertsFromPEM(data)
			}
			tr.TLSClientConfig.RootCAs = rootCAs
		}
	}

	return tr, nil
}

// TransportOptions transport options
type TransportOptions struct {
	// TLS config
	TLSConfig *tls.Config

	// Timeout settings
	DialTimeout           time.Duration
	DialKeepAlive         time.Duration
	ResponseHeaderTimeout time.Duration
	ExpectContinueTimeout time.Duration
	TLSHandshakeTimeout   time.Duration

	// Connection pool settings
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration

	// Proxy settings
	Proxy func(*http.Request) (*url.URL, error)

	// Enable compression (disabled by default)
	EnableCompression bool

	// Disable keep-alives
	DisableKeepAlives bool
}

// NewTransport creates custom transport
func NewTransport(opts TransportOptions) *http.Transport {
	// Set defaults
	dialTimeout := opts.DialTimeout
	if dialTimeout <= 0 {
		dialTimeout = 30 * time.Second
	}

	dialKeepAlive := opts.DialKeepAlive
	if dialKeepAlive <= 0 {
		dialKeepAlive = 30 * time.Second
	}

	maxIdleConns := opts.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 256
	}

	maxIdleConnsPerHost := opts.MaxIdleConnsPerHost
	if maxIdleConnsPerHost <= 0 {
		maxIdleConnsPerHost = 16
	}

	idleConnTimeout := opts.IdleConnTimeout
	if idleConnTimeout <= 0 {
		idleConnTimeout = time.Minute
	}

	responseHeaderTimeout := opts.ResponseHeaderTimeout
	if responseHeaderTimeout <= 0 {
		responseHeaderTimeout = time.Minute
	}

	tlsHandshakeTimeout := opts.TLSHandshakeTimeout
	if tlsHandshakeTimeout <= 0 {
		tlsHandshakeTimeout = 10 * time.Second
	}

	expectContinueTimeout := opts.ExpectContinueTimeout
	if expectContinueTimeout <= 0 {
		expectContinueTimeout = 10 * time.Second
	}

	// Determine whether to disable compression (default disabled unless enabled)
	disableCompression := !opts.EnableCompression

	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   dialTimeout,
			KeepAlive: dialKeepAlive,
		}).DialContext,
		MaxIdleConns:          maxIdleConns,
		MaxIdleConnsPerHost:   maxIdleConnsPerHost,
		IdleConnTimeout:       idleConnTimeout,
		ResponseHeaderTimeout: responseHeaderTimeout,
		TLSHandshakeTimeout:   tlsHandshakeTimeout,
		ExpectContinueTimeout: expectContinueTimeout,
		DisableCompression:    disableCompression,
		DisableKeepAlives:     opts.DisableKeepAlives,
	}

	// Set TLS config
	if opts.TLSConfig != nil {
		tr.TLSClientConfig = opts.TLSConfig
	}

	// Set proxy
	if opts.Proxy != nil {
		tr.Proxy = opts.Proxy
	} else {
		tr.Proxy = http.ProxyFromEnvironment
	}

	return tr
}

// mustGetSystemCertPool returns system CA pool, or empty pool on error
func mustGetSystemCertPool() *x509.CertPool {
	pool, err := x509.SystemCertPool()
	if err != nil {
		// On some systems (e.g., Windows) this may fail; return empty pool
		return x509.NewCertPool()
	}
	return pool
}

// NewHTTPClient creates configured HTTP client
func NewHTTPClient(transport *http.Transport, timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 0 // no timeout
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
