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

// DefaultTransport 创建默认 HTTP 传输
// 与 http.DefaultTransport 类似，但禁用压缩以避免自动解压缩 gzip 编码的内容
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
		// 禁用压缩，避免底层传输自动解码 content-encoding=gzip 的对象
		// 参考: https://golang.org/src/net/http/transport.go?h=roundTrip#L1843
		DisableCompression: true,
	}

	if secure {
		tr.TLSClientConfig = &tls.Config{
			// 安全考虑：
			// - 不使用 SSLv3（POODLE 和 BEAST 漏洞）
			// - 不使用 TLSv1.0（使用 CBC 密码的 POODLE 和 BEAST）
			// - 不使用 TLSv1.1（RC4 密码使用）
			MinVersion: tls.VersionTLS12,
		}

		// 支持自定义 CA 证书（通过 SSL_CERT_FILE 环境变量）
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

// TransportOptions 传输选项
type TransportOptions struct {
	// TLS 配置
	TLSConfig *tls.Config

	// 超时设置
	DialTimeout           time.Duration
	DialKeepAlive         time.Duration
	ResponseHeaderTimeout time.Duration
	ExpectContinueTimeout time.Duration
	TLSHandshakeTimeout   time.Duration

	// 连接池设置
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration

	// 代理设置
	Proxy func(*http.Request) (*url.URL, error)

	// 启用压缩（默认禁用）
	EnableCompression bool

	// 禁用保持连接
	DisableKeepAlives bool
}

// NewTransport 创建自定义传输
func NewTransport(opts TransportOptions) *http.Transport {
	// 设置默认值
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

	// 确定是否禁用压缩（默认禁用，除非明确启用）
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

	// 设置 TLS 配置
	if opts.TLSConfig != nil {
		tr.TLSClientConfig = opts.TLSConfig
	}

	// 设置代理
	if opts.Proxy != nil {
		tr.Proxy = opts.Proxy
	} else {
		tr.Proxy = http.ProxyFromEnvironment
	}

	return tr
}

// mustGetSystemCertPool 返回系统 CA 证书池，如果出错则返回空池
func mustGetSystemCertPool() *x509.CertPool {
	pool, err := x509.SystemCertPool()
	if err != nil {
		// 在某些系统（如 Windows）上可能失败，返回空池
		return x509.NewCertPool()
	}
	return pool
}

// NewHTTPClient 创建配置好的 HTTP 客户端
func NewHTTPClient(transport *http.Transport, timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 0 // 无超时
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
