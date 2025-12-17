package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewTraceContext(t *testing.T) {
	tests := []struct {
		name     string
		hook     TraceHook
		wantNil  bool
	}{
		{
			name: "With hook",
			hook: func(info TraceInfo) {
				// Do nothing
			},
			wantNil: false,
		},
		{
			name:    "Without hook",
			hook:    nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			traceCtx := NewTraceContext(ctx, tt.hook)

			if traceCtx == nil && !tt.wantNil {
				t.Error("Expected non-nil context")
			}
		})
	}
}

func TestTraceInfo_GetTimings(t *testing.T) {
	now := time.Now()

	trace := &TraceInfo{
		DNSStart:          now,
		DNSDone:           now.Add(10 * time.Millisecond),
		ConnectStart:      now.Add(10 * time.Millisecond),
		ConnectDone:       now.Add(20 * time.Millisecond),
		TLSHandshakeStart: now.Add(20 * time.Millisecond),
		TLSHandshakeDone:  now.Add(50 * time.Millisecond),
		WroteHeaders:      now.Add(50 * time.Millisecond),
		WroteRequest:      now.Add(55 * time.Millisecond),
		GotFirstResponse:  now.Add(100 * time.Millisecond),
		GotConn:           now.Add(50 * time.Millisecond),
	}

	timings := trace.GetTimings()

	tests := []struct {
		name     string
		key      string
		expected time.Duration
	}{
		{"DNS lookup", "dns_lookup", 10 * time.Millisecond},
		{"TCP connect", "tcp_connect", 10 * time.Millisecond},
		{"TLS handshake", "tls_handshake", 30 * time.Millisecond},
		{"Request write", "request_write", 5 * time.Millisecond},
		{"Server processing", "server_processing", 45 * time.Millisecond},
		{"Total request", "total_request", 50 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, ok := timings[tt.key]; !ok {
				t.Errorf("Expected timing for %s", tt.key)
			} else if got != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestTraceInfo_TotalDuration(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		trace    *TraceInfo
		expected time.Duration
	}{
		{
			name: "Reused connection",
			trace: &TraceInfo{
				ConnReused:       true,
				GotConn:          now,
				GotFirstResponse: now.Add(100 * time.Millisecond),
			},
			expected: 100 * time.Millisecond,
		},
		{
			name: "New connection with DNS",
			trace: &TraceInfo{
				ConnReused:       false,
				DNSStart:         now,
				GotFirstResponse: now.Add(200 * time.Millisecond),
			},
			expected: 200 * time.Millisecond,
		},
		{
			name: "New connection without DNS",
			trace: &TraceInfo{
				ConnReused:       false,
				ConnectStart:     now,
				GotFirstResponse: now.Add(150 * time.Millisecond),
			},
			expected: 150 * time.Millisecond,
		},
		{
			name:     "Empty trace",
			trace:    &TraceInfo{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.trace.TotalDuration()
			if got != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestTraceWithHTTPRequest(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // 模拟处理时间
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// 追踪信息（使用指针以便在 hook 中更新）
	var capturedTrace *TraceInfo
	hook := func(info TraceInfo) {
		// 保存追踪信息的副本
		trace := info
		capturedTrace = &trace
	}

	// 创建带追踪的请求
	ctx := context.Background()
	traceCtx := NewTraceContext(ctx, hook)

	req, err := http.NewRequestWithContext(traceCtx, "GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 验证追踪信息
	if capturedTrace == nil {
		t.Fatal("Expected trace to be captured")
	}

	if capturedTrace.GotConn.IsZero() {
		t.Error("Expected GotConn to be set")
	}

	// 注意：在 HTTP/1.1 本地连接中，某些追踪事件可能不会触发
	// 所以我们只检查基本的连接信息
	t.Logf("Trace info: ConnReused=%v, WasIdle=%v",
		capturedTrace.ConnReused, capturedTrace.ConnWasIdle)
}

func BenchmarkNewTraceContext(b *testing.B) {
	hook := func(info TraceInfo) {
		// Do nothing
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewTraceContext(ctx, hook)
	}
}

func BenchmarkGetTimings(b *testing.B) {
	now := time.Now()
	trace := &TraceInfo{
		DNSStart:          now,
		DNSDone:           now.Add(10 * time.Millisecond),
		ConnectStart:      now.Add(10 * time.Millisecond),
		ConnectDone:       now.Add(20 * time.Millisecond),
		TLSHandshakeStart: now.Add(20 * time.Millisecond),
		TLSHandshakeDone:  now.Add(50 * time.Millisecond),
		WroteHeaders:      now.Add(50 * time.Millisecond),
		WroteRequest:      now.Add(55 * time.Millisecond),
		GotFirstResponse:  now.Add(100 * time.Millisecond),
		GotConn:           now.Add(50 * time.Millisecond),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trace.GetTimings()
	}
}
