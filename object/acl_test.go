// Package object object/acl_test.go
package object

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Scorpio69t/rustfs-go/pkg/acl"
)

func TestObjectACLSetGet(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<AccessControlPolicy>
  <Owner>
    <ID>owner-id</ID>
    <DisplayName>owner</DisplayName>
  </Owner>
  <AccessControlList>
    <Grant>
      <Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="CanonicalUser">
        <ID>grantee-id</ID>
      </Grantee>
      <Permission>FULL_CONTROL</Permission>
    </Grant>
  </AccessControlList>
</AccessControlPolicy>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["acl"]; !ok {
			t.Fatalf("expected acl query flag")
		}
		switch r.Method {
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), "<AccessControlPolicy") {
				t.Fatalf("expected ACL XML in body, got %s", string(body))
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(responseXML))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	policy := acl.ACL{
		Owner: acl.Owner{ID: "owner-id", DisplayName: "owner"},
		Grants: []acl.Grant{
			{
				Grantee:    acl.Grantee{Type: "CanonicalUser", ID: "grantee-id"},
				Permission: acl.PermissionFullControl,
			},
		},
	}

	if err := service.SetACL(context.Background(), "bucket", "object", policy); err != nil {
		t.Fatalf("SetACL() error = %v", err)
	}

	got, err := service.GetACL(context.Background(), "bucket", "object")
	if err != nil {
		t.Fatalf("GetACL() error = %v", err)
	}
	if len(got.Grants) != 1 || got.Grants[0].Permission != acl.PermissionFullControl {
		t.Fatalf("unexpected ACL grants: %+v", got.Grants)
	}
}
