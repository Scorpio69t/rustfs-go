// Package cse provides simple client-side encryption helpers.
package cse

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

const (
	// MetadataKeyAlgorithm stores the CSE algorithm in user metadata.
	MetadataKeyAlgorithm = "rustfs-cse-algorithm"
	// MetadataKeyNonce stores the base64-encoded nonce in user metadata.
	MetadataKeyNonce = "rustfs-cse-nonce"
	// AlgorithmAESGCM indicates AES-GCM encryption.
	AlgorithmAESGCM = "AES-GCM"
)

// Client provides client-side encryption helpers.
type Client struct {
	key []byte
}

// New creates a new CSE client with a 16, 24, or 32 byte key.
func New(key []byte) (*Client, error) {
	switch len(key) {
	case 16, 24, 32:
		return &Client{key: key}, nil
	default:
		return nil, errors.New("cse key must be 16, 24, or 32 bytes")
	}
}

// Encrypt encrypts the entire reader and returns ciphertext and metadata.
func (c *Client) Encrypt(reader io.Reader) ([]byte, map[string]string, error) {
	plain, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, err
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plain, nil)
	metadata := map[string]string{
		MetadataKeyAlgorithm: AlgorithmAESGCM,
		MetadataKeyNonce:     base64.StdEncoding.EncodeToString(nonce),
	}

	return ciphertext, metadata, nil
}

// Decrypt decrypts the entire reader using metadata.
func (c *Client) Decrypt(reader io.Reader, metadata map[string]string) ([]byte, error) {
	algo := metadata[MetadataKeyAlgorithm]
	if algo != AlgorithmAESGCM {
		return nil, errors.New("unsupported or missing cse algorithm")
	}

	nonceB64 := metadata[MetadataKeyNonce]
	if nonceB64 == "" {
		return nil, errors.New("missing cse nonce")
	}
	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		return nil, err
	}

	ciphertext, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, errors.New("invalid cse nonce size")
	}

	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plain, nil
}
