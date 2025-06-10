package i18n

import (
	"astrovista-api/cache"
	"sync"
	"time"
)

// TranslationCache implements an in-memory cache for translations
// to reduce requests to the translation API
type TranslationCache struct {
	translations map[string]cacheEntry
	mutex        sync.RWMutex
	maxSize      int
	expiration   time.Duration
	// Optional functions for Redis integration
	redisEnabled bool
}

// cacheEntry represents a cache entry with expiration information
type cacheEntry struct {
	value      string
	expiration time.Time
}

// NewTranslationCache creates a new instance of the translation cache
func NewTranslationCache() *TranslationCache {
	return &TranslationCache{
		translations: make(map[string]cacheEntry),
		maxSize:      1000,           // Limits the maximum number of entries
		expiration:   24 * time.Hour, // Default expiration time
		redisEnabled: false,          // By default, does not use Redis
	}
}

// EnableRedisCache configures the cache to also use Redis
func (c *TranslationCache) EnableRedisCache() {
	c.redisEnabled = cache.Client != nil
}

// Get retrieves a translation from the cache
func (c *TranslationCache) Get(key string) (string, bool) {
	// If Redis is enabled, try to fetch from Redis cache first
	if c.redisEnabled && cache.Client != nil {
		redisCache := NewRedisTranslationCache()
		if value, found := redisCache.Get(key); found {
			return value, true
		}
	}

	// Otherwise, use the in-memory cache
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, found := c.translations[key]
	if !found {
		return "", false
	}

	// Check if the entry has expired
	if time.Now().After(entry.expiration) {
		// Could do the removal here, but to avoid
		// lock upgrading, we leave it to the periodic cleanup process
		return "", false
	}

	return entry.value, true
}

// Set stores a translation in the cache
func (c *TranslationCache) Set(key string, value string) {
	// If Redis is enabled, also store in Redis
	if c.redisEnabled && cache.Client != nil {
		redisCache := NewRedisTranslationCache()
		redisCache.Set(key, value)
	}

	// Also store in the in-memory cache for quick access
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// If the cache is full and the key doesn't exist, do a cleanup
	if len(c.translations) >= c.maxSize && c.translations[key].value == "" {
		c.cleanupLocked()
	}

	c.translations[key] = cacheEntry{
		value:      value,
		expiration: time.Now().Add(c.expiration),
	}
}

// Clear clears the entire cache
func (c *TranslationCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.translations = make(map[string]cacheEntry)
}

// cleanupLocked removes expired entries from the cache (assumes the lock is already obtained)
func (c *TranslationCache) cleanupLocked() {
	now := time.Now()

	// Remove expired entries
	for key, entry := range c.translations {
		if now.After(entry.expiration) {
			delete(c.translations, key)
		}
	}

	// If it's still too large, remove the oldest ones
	// This is a simple implementation, not a complete LRU
	if len(c.translations) >= c.maxSize {
		// We remove about 25% of the cache to avoid doing this frequently
		toRemove := c.maxSize / 4
		removed := 0

		for key := range c.translations {
			delete(c.translations, key)
			removed++
			if removed >= toRemove {
				break
			}
		}
	}
}
