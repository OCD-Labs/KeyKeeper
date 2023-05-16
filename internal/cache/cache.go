// Package cache provides a cache interface definition for caching
// and managing session tokens for an HTTP server.
package cache

import (
	"context"
	"time"
)

// Cache defines interfaces required for caching.
type Cache interface {
	// BlacklistSession adds a session token to the blacklist cache with an expiration duration.
	BlacklistSession(ctx context.Context, sessionToken string, expirationTime time.Duration) error

	// IsSessionBlacklisted checks if a session token is blacklisted by querying the Redis cache.
	IsSessionBlacklisted(ctx context.Context, sessionToken string) (bool, error)
}
