// Package object object/restore_test.go
package object

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Scorpio69t/rustfs-go/pkg/restore"
)

func TestRestoreRequest(t *testing.T) {
	var gotQuery bool
	var gotBody bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if _, ok := r.URL.Query()["restore"]; ok {
			gotQuery = true
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		if bytes.Contains(body, []byte("<RestoreRequest")) {
			gotBody = true
		}

		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	req := restore.RestoreRequest{}
	req.SetDays(2)
	req.SetTier(restore.TierBulk)

	if err := service.Restore(context.Background(), "bucket", "object", "", req); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}
	if !gotQuery {
		t.Fatalf("expected restore query parameter")
	}
	if !gotBody {
		t.Fatalf("expected restore request body")
	}
}
