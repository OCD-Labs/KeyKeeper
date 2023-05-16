// Package cache provides a Redis cache implementation for caching
// and managing session tokens for an HTTP server.
package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache wraps a redis.Client object.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache.
func NewRedisCache(addr string) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &RedisCache{client: client}, nil
}

// BlacklistSession adds a session token to the blacklist cache with an expiration duration.
func (rc *RedisCache) BlacklistSession(ctx context.Context, sessionToken string, expirationTime time.Duration) error {
	key := "blacklist:" + sessionToken
	return rc.client.Set(ctx, key, "blacklisted", expirationTime).Err()
}

// IsSessionBlacklisted checks if a session token is blacklisted by querying the Redis cache.
func (rc *RedisCache) IsSessionBlacklisted(ctx context.Context, sessionToken string) (bool, error) {
	key := "blacklist:" + sessionToken
	exists, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}
