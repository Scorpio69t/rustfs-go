package acl

import "errors"

var (
	// ErrInvalidCannedACL indicates an unknown canned ACL.
	ErrInvalidCannedACL = errors.New("acl: invalid canned ACL")

	// ErrMixedACL indicates both canned and grant-based ACLs were provided.
	ErrMixedACL = errors.New("acl: cannot mix canned ACL with grant-based ACL")
)
