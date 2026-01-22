package sse

import "errors"

var (
	// ErrInvalidKeySize is returned when SSE-C key is not 256 bits (32 bytes)
	ErrInvalidKeySize = errors.New("sse: encryption key must be 256 bits (32 bytes)")

	// ErrNoEncryptionConfig is returned when no encryption configuration is found
	ErrNoEncryptionConfig = errors.New("sse: bucket has no encryption configuration")
)
