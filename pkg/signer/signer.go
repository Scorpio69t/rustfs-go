// Package signer internal/signer/signer.go
// Provides internal signer interfaces and implementations
package signer

import (
	"net/http"
	"time"
)

// Signer defines signer interface
type Signer interface {
	// Sign signs a request
	Sign(req *http.Request, accessKey, secretKey, sessionToken, region string) *http.Request

	// Presign generates a presigned request
	Presign(req *http.Request, accessKey, secretKey, sessionToken, region string, expires time.Duration) *http.Request
}

// SignerType represents signer type
type SignerType int

const (
	SignerV4 SignerType = iota
	SignerV2
	SignerAnonymous
)

// NewSigner creates a signer instance
func NewSigner(signerType SignerType) Signer {
	switch signerType {
	case SignerV2:
		return &V2Signer{}
	case SignerAnonymous:
		return &AnonymousSigner{}
	default:
		return &V4Signer{}
	}
}

// AnonymousSigner signs anonymously
type AnonymousSigner struct{}

// Sign signs request anonymously
func (s *AnonymousSigner) Sign(req *http.Request, accessKey, secretKey, sessionToken, region string) *http.Request {
	return req
}

// Presign presigns request anonymously
func (s *AnonymousSigner) Presign(req *http.Request, accessKey, secretKey, sessionToken, region string, expires time.Duration) *http.Request {
	return req
}
