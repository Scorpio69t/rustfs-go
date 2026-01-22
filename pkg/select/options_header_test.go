package s3select

import (
	"testing"

	"github.com/Scorpio69t/rustfs-go/pkg/sse"
)

func TestOptionsHeaderEmpty(t *testing.T) {
	opts := Options{}
	headers := opts.Header()
	if len(headers) != 0 {
		t.Fatalf("expected empty headers, got %v", headers)
	}
}

func TestOptionsHeaderWithSSE(t *testing.T) {
	opts := Options{
		ServerSideEncryption: sse.NewSSES3(),
	}

	headers := opts.Header()
	if got := headers.Get("X-Amz-Server-Side-Encryption"); got != "AES256" {
		t.Fatalf("expected SSE-S3 header, got %q", got)
	}
}
