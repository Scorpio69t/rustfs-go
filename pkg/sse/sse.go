// Package sse provides Server-Side Encryption (SSE) support for S3-compatible storage.
//
// This package implements three SSE modes:
//   - SSE-S3: Server-managed encryption with AES-256
//   - SSE-C: Customer-provided encryption keys
//   - SSE-KMS: AWS Key Management Service (KMS) encryption
package sse

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"net/http"
)

// Configuration represents bucket-level default encryption configuration
type Configuration struct {
	XMLName xml.Name `xml:"ServerSideEncryptionConfiguration"`
	Rules   []Rule   `xml:"Rule"`
}

// Rule defines a server-side encryption rule
type Rule struct {
	ApplySSEByDefault ApplySSEByDefault `xml:"ApplyServerSideEncryptionByDefault"`
	BucketKeyEnabled  bool              `xml:"BucketKeyEnabled,omitempty"`
}

// ApplySSEByDefault specifies the default encryption settings
type ApplySSEByDefault struct {
	SSEAlgorithm   string `xml:"SSEAlgorithm"`             // AES256 or aws:kms
	KMSMasterKeyID string `xml:"KMSMasterKeyID,omitempty"` // KMS key ID for aws:kms
}

// Type represents the server-side encryption type
type Type string

const (
	// SSES3 represents S3-managed encryption (AES256)
	SSES3 Type = "AES256"
	// SSEKMS represents KMS-managed encryption
	SSEKMS Type = "aws:kms"
)

// Encrypter is the interface that wraps the ApplyHeaders method.
//
// ApplyHeaders applies the appropriate SSE headers to an HTTP request.
type Encrypter interface {
	ApplyHeaders(h http.Header)
	Type() Type
}

// S3 represents SSE-S3 encryption (server-managed keys)
type S3 struct{}

// NewSSES3 creates a new SSE-S3 encrypter
func NewSSES3() *S3 {
	return &S3{}
}

// ApplyHeaders applies SSE-S3 headers to the request
func (s *S3) ApplyHeaders(h http.Header) {
	h.Set("X-Amz-Server-Side-Encryption", "AES256")
}

// Type returns the encryption type
func (s *S3) Type() Type {
	return SSES3
}

// C represents SSE-C encryption (customer-provided keys)
type C struct {
	Key       []byte
	Algorithm string // Default: AES256
}

// NewSSEC creates a new SSE-C encrypter with the given 256-bit key
func NewSSEC(key []byte) (*C, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKeySize
	}
	return &C{
		Key:       key,
		Algorithm: "AES256",
	}, nil
}

// ApplyHeaders applies SSE-C headers to the request
func (c *C) ApplyHeaders(h http.Header) {
	h.Set("X-Amz-Server-Side-Encryption-Customer-Algorithm", c.Algorithm)
	h.Set("X-Amz-Server-Side-Encryption-Customer-Key", base64.StdEncoding.EncodeToString(c.Key))

	// Calculate and set MD5 of the key
	md5sum := md5.Sum(c.Key)
	h.Set("X-Amz-Server-Side-Encryption-Customer-Key-MD5", base64.StdEncoding.EncodeToString(md5sum[:]))
}

// Type returns the encryption type
func (c *C) Type() Type {
	return SSES3 // SSE-C uses AES256
}

// ApplyCopyHeaders applies SSE-C headers for copy source
func (c *C) ApplyCopyHeaders(h http.Header) {
	h.Set("X-Amz-Copy-Source-Server-Side-Encryption-Customer-Algorithm", c.Algorithm)
	h.Set("X-Amz-Copy-Source-Server-Side-Encryption-Customer-Key", base64.StdEncoding.EncodeToString(c.Key))

	md5sum := md5.Sum(c.Key)
	h.Set("X-Amz-Copy-Source-Server-Side-Encryption-Customer-Key-MD5", base64.StdEncoding.EncodeToString(md5sum[:]))
}

// KMS represents SSE-KMS encryption (AWS KMS-managed keys)
type KMS struct {
	KeyID   string
	Context map[string]string
}

// NewSSEKMS creates a new SSE-KMS encrypter
func NewSSEKMS(keyID string, context map[string]string) *KMS {
	return &KMS{
		KeyID:   keyID,
		Context: context,
	}
}

// ApplyHeaders applies SSE-KMS headers to the request
func (k *KMS) ApplyHeaders(h http.Header) {
	h.Set("X-Amz-Server-Side-Encryption", "aws:kms")

	if k.KeyID != "" {
		h.Set("X-Amz-Server-Side-Encryption-Aws-Kms-Key-Id", k.KeyID)
	}

	if len(k.Context) > 0 {
		ctx, _ := json.Marshal(k.Context)
		h.Set("X-Amz-Server-Side-Encryption-Context", base64.StdEncoding.EncodeToString(ctx))
	}
}

// Type returns the encryption type
func (k *KMS) Type() Type {
	return SSEKMS
}

// NewConfiguration creates a default SSE-S3 bucket encryption configuration
func NewConfiguration() *Configuration {
	return &Configuration{
		Rules: []Rule{
			{
				ApplySSEByDefault: ApplySSEByDefault{
					SSEAlgorithm: "AES256",
				},
				BucketKeyEnabled: false,
			},
		},
	}
}

// NewKMSConfiguration creates an SSE-KMS bucket encryption configuration
func NewKMSConfiguration(keyID string) *Configuration {
	return &Configuration{
		Rules: []Rule{
			{
				ApplySSEByDefault: ApplySSEByDefault{
					SSEAlgorithm:   "aws:kms",
					KMSMasterKeyID: keyID,
				},
				BucketKeyEnabled: true,
			},
		},
	}
}
