// Package transport internal/transport/trace.go
package transport

import (
	"context"
	"crypto/tls"
	"net/http/httptrace"
	"time"
)

// TraceInfo contains HTTP request trace information
type TraceInfo struct {
	// DNS lookup
	DNSStart     time.Time
	DNSDone      time.Time
	DNSError     error
	DNSAddresses []string

	// Connection
	ConnectStart time.Time
	ConnectDone  time.Time
	ConnectError error

	// TLS
	TLSHandshakeStart time.Time
	TLSHandshakeDone  time.Time
	TLSError          error
	TLSVersion        uint16
	TLSCipherSuite    uint16
	TLSServerName     string

	// Request/Response
	WroteHeaders     time.Time
	WroteRequest     time.Time
	GotFirstResponse time.Time
	GotConn          time.Time

	// Connection reuse
	ConnReused   bool
	ConnWasIdle  bool
	ConnIdleTime time.Duration
}

// TraceHook is a callback used to record HTTP trace info
type TraceHook func(TraceInfo)

// NewTraceContext creates a context with HTTP tracing
func NewTraceContext(ctx context.Context, hook TraceHook) context.Context {
	if hook == nil {
		return ctx
	}

	trace := &TraceInfo{}

	clientTrace := &httptrace.ClientTrace{
		// DNS
		DNSStart: func(info httptrace.DNSStartInfo) {
			trace.DNSStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			trace.DNSDone = time.Now()
			trace.DNSError = info.Err
			if info.Addrs != nil {
				for _, addr := range info.Addrs {
					trace.DNSAddresses = append(trace.DNSAddresses, addr.String())
				}
			}
		},

		// Connection
		ConnectStart: func(network, addr string) {
			trace.ConnectStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			trace.ConnectDone = time.Now()
			trace.ConnectError = err
		},

		// TLS
		TLSHandshakeStart: func() {
			trace.TLSHandshakeStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			trace.TLSHandshakeDone = time.Now()
			trace.TLSError = err
			if err == nil {
				trace.TLSVersion = state.Version
				trace.TLSCipherSuite = state.CipherSuite
				trace.TLSServerName = state.ServerName
			}
		},

		// Request writing
		WroteHeaders: func() {
			trace.WroteHeaders = time.Now()
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			trace.WroteRequest = time.Now()
		},

		// Response
		GotFirstResponseByte: func() {
			trace.GotFirstResponse = time.Now()
		},

		// Connection reuse
		GotConn: func(info httptrace.GotConnInfo) {
			trace.GotConn = time.Now()
			trace.ConnReused = info.Reused
			trace.ConnWasIdle = info.WasIdle
			trace.ConnIdleTime = info.IdleTime

			// Call hook to pass trace info
			if hook != nil {
				hook(*trace)
			}
		},
	}

	return httptrace.WithClientTrace(ctx, clientTrace)
}

// GetTimings returns durations for each stage
func (t *TraceInfo) GetTimings() map[string]time.Duration {
	timings := make(map[string]time.Duration)

	if !t.DNSStart.IsZero() && !t.DNSDone.IsZero() {
		timings["dns_lookup"] = t.DNSDone.Sub(t.DNSStart)
	}

	if !t.ConnectStart.IsZero() && !t.ConnectDone.IsZero() {
		timings["tcp_connect"] = t.ConnectDone.Sub(t.ConnectStart)
	}

	if !t.TLSHandshakeStart.IsZero() && !t.TLSHandshakeDone.IsZero() {
		timings["tls_handshake"] = t.TLSHandshakeDone.Sub(t.TLSHandshakeStart)
	}

	if !t.WroteHeaders.IsZero() && !t.WroteRequest.IsZero() {
		timings["request_write"] = t.WroteRequest.Sub(t.WroteHeaders)
	}

	if !t.WroteRequest.IsZero() && !t.GotFirstResponse.IsZero() {
		timings["server_processing"] = t.GotFirstResponse.Sub(t.WroteRequest)
	}

	if !t.GotConn.IsZero() && !t.GotFirstResponse.IsZero() {
		timings["total_request"] = t.GotFirstResponse.Sub(t.GotConn)
	}

	return timings
}

// TotalDuration returns total time from connection start to first byte
func (t *TraceInfo) TotalDuration() time.Duration {
	if t.ConnReused {
		// For reused connections, measure from GotConn to GotFirstResponse
		if !t.GotConn.IsZero() && !t.GotFirstResponse.IsZero() {
			return t.GotFirstResponse.Sub(t.GotConn)
		}
	} else {
		// For new connection, measure from DNSStart or ConnectStart to GotFirstResponse
		start := t.DNSStart
		if start.IsZero() {
			start = t.ConnectStart
		}
		if !start.IsZero() && !t.GotFirstResponse.IsZero() {
			return t.GotFirstResponse.Sub(start)
		}
	}
	return 0
}
