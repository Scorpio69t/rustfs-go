package acl

import (
	"strings"
	"testing"
)

func TestACLToXML(t *testing.T) {
	acl := ACL{
		Owner: Owner{
			ID:          "owner-id",
			DisplayName: "owner",
		},
		Grants: []Grant{
			{
				Grantee: Grantee{
					Type: "CanonicalUser",
					ID:   "grantee-id",
				},
				Permission: PermissionRead,
			},
		},
	}

	data, err := acl.ToXML()
	if err != nil {
		t.Fatalf("ToXML() error = %v", err)
	}
	if !strings.Contains(string(data), "<AccessControlPolicy") {
		t.Fatalf("expected AccessControlPolicy element")
	}
}

func TestACLNormalizeCanned(t *testing.T) {
	acl := ACL{Canned: ACLPublicRead}
	if err := acl.Normalize(); err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
}

func TestACLNormalizeMixed(t *testing.T) {
	acl := ACL{
		Canned: ACLPrivate,
		Owner:  Owner{ID: "owner-id"},
	}
	if err := acl.Normalize(); err != ErrMixedACL {
		t.Fatalf("expected ErrMixedACL, got %v", err)
	}
}
