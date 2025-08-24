package jwt

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisTokenBlacklist manages blacklisted JWT tokens using Redis
type RedisTokenBlacklist struct {
	client *redis.Client
	prefix string
}

// NewRedisTokenBlacklist creates a new Redis-based token blacklist
func NewRedisTokenBlacklist(client *redis.Client, prefix string) *RedisTokenBlacklist {
	if prefix == "" {
		prefix = "jwt_blacklist:"
	}
	
	return &RedisTokenBlacklist{
		client: client,
		prefix: prefix,
	}
}

// hashToken creates a SHA256 hash of the token for storage
func (rtb *RedisTokenBlacklist) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// BlacklistToken adds a token to the Redis blacklist with TTL based on expiration
func (rtb *RedisTokenBlacklist) BlacklistToken(token string, expiresAt time.Time) error {
	ctx := context.Background()
	tokenHash := rtb.hashToken(token)
	key := rtb.prefix + tokenHash
	
	// Calculate TTL - how long until the token naturally expires
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// Token already expired, no need to blacklist
		return nil
	}
	
	// Store in Redis with TTL
	err := rtb.client.Set(ctx, key, "blacklisted", ttl).Err()
	return err
}

// IsTokenBlacklisted checks if a token is in the Redis blacklist
func (rtb *RedisTokenBlacklist) IsTokenBlacklisted(token string) bool {
	ctx := context.Background()
	tokenHash := rtb.hashToken(token)
	key := rtb.prefix + tokenHash
	
	// Check if key exists in Redis
	exists, err := rtb.client.Exists(ctx, key).Result()
	if err != nil {
		// If Redis is down, we might want to handle this differently
		// For now, we'll assume the token is not blacklisted
		return false
	}
	
	return exists > 0
}

// GetBlacklistedCount returns the number of blacklisted tokens in Redis
func (rtb *RedisTokenBlacklist) GetBlacklistedCount() (int64, error) {
	ctx := context.Background()
	
	// Count keys matching our prefix pattern
	keys, err := rtb.client.Keys(ctx, rtb.prefix+"*").Result()
	if err != nil {
		return 0, err
	}
	
	return int64(len(keys)), nil
}

// ClearExpiredTokens manually removes expired tokens (Redis TTL handles this automatically)
// This method is mainly for compatibility and monitoring
func (rtb *RedisTokenBlacklist) ClearExpiredTokens() error {
	// Redis automatically removes expired keys, so this is a no-op
	// You could use this for logging/monitoring purposes
	return nil
}

// CloseConnection closes the Redis connection
func (rtb *RedisTokenBlacklist) CloseConnection() error {
	return rtb.client.Close()
}
