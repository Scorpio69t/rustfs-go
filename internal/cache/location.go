// Package cache internal/cache/location.go
package cache

import (
	"sync"
	"time"
)

// LocationCache 桶位置缓存
type LocationCache struct {
	mu      sync.RWMutex
	entries map[string]locationEntry
	ttl     time.Duration
}

type locationEntry struct {
	location  string
	expiresAt time.Time
}

// NewLocationCache 创建位置缓存
func NewLocationCache(ttl time.Duration) *LocationCache {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &LocationCache{
		entries: make(map[string]locationEntry),
		ttl:     ttl,
	}
}

// Get 获取桶位置
func (c *LocationCache) Get(bucketName string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[bucketName]
	if !ok {
		return "", false
	}

	if time.Now().After(entry.expiresAt) {
		return "", false
	}

	return entry.location, true
}

// Set 设置桶位置
func (c *LocationCache) Set(bucketName, location string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[bucketName] = locationEntry{
		location:  location,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Delete 删除桶位置
func (c *LocationCache) Delete(bucketName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, bucketName)
}

// Clear 清空缓存
func (c *LocationCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]locationEntry)
}
