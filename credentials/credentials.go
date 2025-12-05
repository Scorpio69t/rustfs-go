// Package credentials provides credential management for RustFS client
package credentials

import (
	"sync"
	"time"
)

// Value is the RustFS credentials value for individual credential fields.
type Value struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	SignerType      SignatureType
}

// IsExpired returns if the credentials are expired.
func (v Value) IsExpired() bool {
	return false
}

// Provider is an interface for retrieving credentials.
type Provider interface {
	Retrieve() (Value, error)
	IsExpired() bool
}

// SignatureType indicates signature type.
type SignatureType int

const (
	// SignatureV4 is V4 signature type.
	SignatureV4 SignatureType = iota
	// SignatureV2 is V2 signature type (deprecated).
	SignatureV2
	// SignatureAnonymous is anonymous signature type.
	SignatureAnonymous
)

// Credentials stores the credentials for accessing RustFS.
type Credentials struct {
	mu        sync.RWMutex
	provider  Provider
	value     Value
	expiry    time.Time
	retrieved time.Time
}

// New returns a pointer to a new Credentials with the provider set.
func New(provider Provider) *Credentials {
	return &Credentials{
		provider: provider,
	}
}

// Get returns the credentials value, or error if the credentials Value failed
// to retrieve. Will return the cached credentials Value if it has not expired.
func (c *Credentials) Get() (Value, error) {
	c.mu.RLock()
	expired := c.isExpired()
	if !expired {
		creds := c.value
		c.mu.RUnlock()
		return creds, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double check if another routine updated
	if !c.isExpired() {
		return c.value, nil
	}

	creds, err := c.provider.Retrieve()
	if err != nil {
		return Value{}, err
	}

	c.value = creds
	c.retrieved = time.Now()
	if c.provider.IsExpired() {
		c.expiry = time.Now()
	} else {
		c.expiry = time.Now().Add(time.Hour * 24)
	}

	return c.value, nil
}

// isExpired helper method that wraps the expiration check
func (c *Credentials) isExpired() bool {
	if c.provider != nil && c.provider.IsExpired() {
		return true
	}
	return !c.expiry.IsZero() && time.Now().After(c.expiry)
}
