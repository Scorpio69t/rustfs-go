// Package transport internal/transport/transport_test.go
package transport

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestDefaultTransport(t *testing.T) {
	tests := []struct {
		name   string
		secure bool
	}{
		{
			name:   "Insecure transport",
			secure: false,
		},
		{
			name:   "Secure transport",
			secure: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := DefaultTransport(tt.secure)
			if err != nil {
				t.Fatalf("DefaultTransport() error = %v", err)
			}

			if tr == nil {
				t.Fatal("DefaultTransport() returned nil")
			}

			// Verify basic config
			if tr.MaxIdleConns != 256 {
				t.Errorf("MaxIdleConns = %d, want 256", tr.MaxIdleConns)
			}

			if tr.MaxIdleConnsPerHost != 16 {
				t.Errorf("MaxIdleConnsPerHost = %d, want 16", tr.MaxIdleConnsPerHost)
			}

			if tr.ResponseHeaderTimeout != time.Minute {
				t.Errorf("ResponseHeaderTimeout = %v, want %v", tr.ResponseHeaderTimeout, time.Minute)
			}

			if tr.IdleConnTimeout != time.Minute {
				t.Errorf("IdleConnTimeout = %v, want %v", tr.IdleConnTimeout, time.Minute)
			}

			if tr.TLSHandshakeTimeout != 10*time.Second {
				t.Errorf("TLSHandshakeTimeout = %v, want %v", tr.TLSHandshakeTimeout, 10*time.Second)
			}

			if tr.ExpectContinueTimeout != 10*time.Second {
				t.Errorf("ExpectContinueTimeout = %v, want %v", tr.ExpectContinueTimeout, 10*time.Second)
			}

			if !tr.DisableCompression {
				t.Error("DisableCompression should be true")
			}

			// Verify TLS config
			if tt.secure {
				if tr.TLSClientConfig == nil {
					t.Error("TLSClientConfig should not be nil for secure transport")
				} else {
					if tr.TLSClientConfig.MinVersion != tls.VersionTLS12 {
						t.Errorf("MinVersion = %d, want %d (TLS 1.2)", tr.TLSClientConfig.MinVersion, tls.VersionTLS12)
					}
				}
			}

			// Verify proxy config
			if tr.Proxy == nil {
				t.Error("Proxy should not be nil")
			}
		})
	}
}

func TestNewTransport(t *testing.T) {
	tests := []struct {
		name string
		opts TransportOptions
		want struct {
			maxIdleConns        int
			maxIdleConnsPerHost int
			idleConnTimeout     time.Duration
			disableCompression  bool
		}
	}{
		{
			name: "Default values",
			opts: TransportOptions{},
			want: struct {
				maxIdleConns        int
				maxIdleConnsPerHost int
				idleConnTimeout     time.Duration
				disableCompression  bool
			}{
				maxIdleConns:        256,
				maxIdleConnsPerHost: 16,
				idleConnTimeout:     time.Minute,
				disableCompression:  true,
			},
		},
		{
			name: "Custom values",
			opts: TransportOptions{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
			want: struct {
				maxIdleConns        int
				maxIdleConnsPerHost int
				idleConnTimeout     time.Duration
				disableCompression  bool
			}{
				maxIdleConns:        100,
				maxIdleConnsPerHost: 10,
				idleConnTimeout:     30 * time.Second,
				disableCompression:  true,
			},
		},
		{
			name: "Custom timeouts",
			opts: TransportOptions{
				DialTimeout:           15 * time.Second,
				DialKeepAlive:         15 * time.Second,
				ResponseHeaderTimeout: 30 * time.Second,
				TLSHandshakeTimeout:   5 * time.Second,
				ExpectContinueTimeout: 5 * time.Second,
			},
			want: struct {
				maxIdleConns        int
				maxIdleConnsPerHost int
				idleConnTimeout     time.Duration
				disableCompression  bool
			}{
				maxIdleConns:        256,
				maxIdleConnsPerHost: 16,
				idleConnTimeout:     time.Minute,
				disableCompression:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewTransport(tt.opts)
			if tr == nil {
				t.Fatal("NewTransport() returned nil")
			}

			if tr.MaxIdleConns != tt.want.maxIdleConns {
				t.Errorf("MaxIdleConns = %d, want %d", tr.MaxIdleConns, tt.want.maxIdleConns)
			}

			if tr.MaxIdleConnsPerHost != tt.want.maxIdleConnsPerHost {
				t.Errorf("MaxIdleConnsPerHost = %d, want %d", tr.MaxIdleConnsPerHost, tt.want.maxIdleConnsPerHost)
			}

			if tr.IdleConnTimeout != tt.want.idleConnTimeout {
				t.Errorf("IdleConnTimeout = %v, want %v", tr.IdleConnTimeout, tt.want.idleConnTimeout)
			}

			if tr.DisableCompression != tt.want.disableCompression {
				t.Errorf("DisableCompression = %v, want %v", tr.DisableCompression, tt.want.disableCompression)
			}
		})
	}
}

func TestNewTransportWithTLS(t *testing.T) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}

	opts := TransportOptions{
		TLSConfig: tlsConfig,
	}

	tr := NewTransport(opts)
	if tr == nil {
		t.Fatal("NewTransport() returned nil")
	}

	if tr.TLSClientConfig == nil {
		t.Fatal("TLSClientConfig should not be nil")
	}

	if tr.TLSClientConfig.MinVersion != tls.VersionTLS13 {
		t.Errorf("MinVersion = %d, want %d (TLS 1.3)", tr.TLSClientConfig.MinVersion, tls.VersionTLS13)
	}
}

func TestNewTransportWithProxy(t *testing.T) {
	proxyURL, _ := url.Parse("http://proxy.example.com:8080")
	proxyFunc := func(*http.Request) (*url.URL, error) {
		return proxyURL, nil
	}

	opts := TransportOptions{
		Proxy: proxyFunc,
	}

	tr := NewTransport(opts)
	if tr == nil {
		t.Fatal("NewTransport() returned nil")
	}

	if tr.Proxy == nil {
		t.Fatal("Proxy should not be nil")
	}

	// Test proxy function
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	gotProxyURL, err := tr.Proxy(req)
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}

	if gotProxyURL.String() != proxyURL.String() {
		t.Errorf("Proxy URL = %s, want %s", gotProxyURL.String(), proxyURL.String())
	}
}

func TestNewTransportDisableKeepAlives(t *testing.T) {
	opts := TransportOptions{
		DisableKeepAlives: true,
	}

	tr := NewTransport(opts)
	if tr == nil {
		t.Fatal("NewTransport() returned nil")
	}

	if !tr.DisableKeepAlives {
		t.Error("DisableKeepAlives should be true")
	}
}

func TestNewTransportEnableCompression(t *testing.T) {
	opts := TransportOptions{
		EnableCompression: true,
	}

	tr := NewTransport(opts)
	if tr == nil {
		t.Fatal("NewTransport() returned nil")
	}

	if tr.DisableCompression {
		t.Error("DisableCompression should be false when compression is enabled")
	}
}

func TestNewHTTPClient(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		want    time.Duration
	}{
		{
			name:    "With timeout",
			timeout: 30 * time.Second,
			want:    30 * time.Second,
		},
		{
			name:    "No timeout",
			timeout: 0,
			want:    0,
		},
		{
			name:    "Negative timeout (treated as no timeout)",
			timeout: -1 * time.Second,
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, _ := DefaultTransport(false)
			client := NewHTTPClient(tr, tt.timeout)

			if client == nil {
				t.Fatal("NewHTTPClient() returned nil")
			}

			if client.Transport != tr {
				t.Error("Client Transport should match provided transport")
			}

			if client.Timeout != tt.want {
				t.Errorf("Timeout = %v, want %v", client.Timeout, tt.want)
			}
		})
	}
}

func TestMustGetSystemCertPool(t *testing.T) {
	pool := mustGetSystemCertPool()
	if pool == nil {
		t.Error("mustGetSystemCertPool() should not return nil")
	}
}

func BenchmarkDefaultTransport(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = DefaultTransport(false)
	}
}

func BenchmarkNewTransport(b *testing.B) {
	opts := TransportOptions{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
	}
	for i := 0; i < b.N; i++ {
		_ = NewTransport(opts)
	}
}
