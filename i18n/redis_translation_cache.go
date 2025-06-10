package i18n

import (
	"astrovista-api/cache"
	"context"
	"fmt"
	"log"
	"time"
)

// RedisTranslationCache implements a translation cache using Redis
// to persist translations between server restarts
type RedisTranslationCache struct {
	// Prefix to avoid collisions with other keys in Redis
	prefix string
	// Expiration time for stored translations
	expiration time.Duration
}

// NewRedisTranslationCache creates a new instance of Redis cache for translations
func NewRedisTranslationCache() *RedisTranslationCache {
	return &RedisTranslationCache{
		prefix:     "translation:",
		expiration: 30 * 24 * time.Hour, // 30 days cache
	}
}

// Get retrieves a translation from Redis cache
func (c *RedisTranslationCache) Get(key string) (string, bool) {
	// If Redis is not configured, return not found
	if cache.Client == nil {
		return "", false
	}

	ctx := context.Background()
	redisKey := c.prefix + key

	var result string
	found, err := cache.Get(ctx, redisKey, &result)
	if err != nil {
		log.Printf("Error accessing Redis cache for translation: %v", err)
		return "", false
	}

	return result, found
}

// Set stores a translation in Redis cache
func (c *RedisTranslationCache) Set(key string, value string) {
	// If Redis is not configured, do nothing
	if cache.Client == nil {
		return
	}

	ctx := context.Background()
	redisKey := c.prefix + key

	if err := cache.Set(ctx, redisKey, value, c.expiration); err != nil {
		log.Printf("Error storing translation in Redis cache: %v", err)
	}
}

// Clear removes all translations from Redis cache with the specified prefix
func (c *RedisTranslationCache) Clear() {
	// If Redis is not configured, do nothing
	if cache.Client == nil {
		return
	}

	ctx := context.Background()
	// Uses the KEYS command to find all keys with the prefix (less efficient but simpler)
	pattern := fmt.Sprintf("%s*", c.prefix)
	keys, err := cache.Client.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("Error fetching translation keys from Redis: %v", err)
		return
	}

	if len(keys) > 0 {
		if err := cache.Client.Del(ctx, keys...).Err(); err != nil {
			log.Printf("Error deleting translation keys from Redis: %v", err)
		}
	}
}
