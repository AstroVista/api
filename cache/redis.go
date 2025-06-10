package cache

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	// Client is the shared Redis client
	Client *redis.Client
	// DefaultExpiration is the default expiration time for cached items (24 hours)
	DefaultExpiration = 24 * time.Hour
)

// Connect establishes the Redis connection
func Connect() {	// Check if there's a Redis URL in the environment variables (for production use)
	redisURL := os.Getenv("REDIS_URL")

	// If not, use a default for local development
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	// Redis client configuration	Client = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: os.Getenv("REDIS_PASSWORD"), // No password if not defined
		DB:       0,                           // Use database 0
	})
	// Check if the connection is working
	ctx := context.Background()
	_, err := Client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Could not connect to Redis: %v", err)
		log.Println("Cache will be disabled. To enable caching, install Redis and run it on localhost:6379")
		log.Println("A API continuará funcionando normalmente, mas sem o benefício do cache")
		Client = nil
		return
	}

	log.Println("Redis connection established successfully")
}

// Set stores an item in the cache
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if Client == nil {
		return nil // Cache disabled
	}

	// Converting to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	// Storing in Redis
	return Client.Set(ctx, key, data, expiration).Err()
}

// Get retrieves an item from the cache
func Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	if Client == nil {
		return false, nil // Cache disabled
	}

	// Buscando do Redis
	data, err := Client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// Item não encontrado no cache
		return false, nil
	} else if err != nil {
		// Erro ao acessar o Redis
		return false, err
	}

	// Convertendo de JSON para o tipo de destino
	if err := json.Unmarshal(data, dest); err != nil {
		return false, err
	}

	return true, nil
}

// Delete remove um item do cache
func Delete(ctx context.Context, key string) error {
	if Client == nil {
		return nil // Cache desativado
	}

	return Client.Del(ctx, key).Err()
}

// Clear limpa todo o cache
func Clear(ctx context.Context) error {
	if Client == nil {
		return nil // Cache desativado
	}

	return Client.FlushAll(ctx).Err()
}
