package pwreset

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Service manages password reset tokens using Redis with expiry and single-use semantics.
type Service struct {
	rdb    *redis.Client
	ttl    time.Duration
	prefix string
}

func NewService(rdb *redis.Client, ttl time.Duration) *Service {
	return &Service{rdb: rdb, ttl: ttl, prefix: "pwreset:"}
}

// GenerateToken creates a new secure token for the given user ID and stores it in Redis.
// Returns the token string which should be sent to the user via email link.
func (s *Service) GenerateToken(ctx context.Context, userID string) (string, error) {
	if userID == "" {
		return "", errors.New("userID is required")
	}
	token, err := randomToken(32)
	if err != nil {
		return "", err
	}
	key := s.key(token)
	if err := s.rdb.Set(ctx, key, userID, s.ttl).Err(); err != nil {
		return "", err
	}
	return token, nil
}

// ValidateToken returns the associated userID if the token exists and is not expired.
func (s *Service) ValidateToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", errors.New("token is required")
	}
	val, err := s.rdb.Get(ctx, s.key(token)).Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.New("invalid or expired token")
		}
		return "", err
	}
	return val, nil
}

// ConsumeToken deletes the token to enforce single-use semantics.
func (s *Service) ConsumeToken(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token is required")
	}
	_, err := s.rdb.Del(ctx, s.key(token)).Result()
	return err
}

func (s *Service) key(token string) string {
	return fmt.Sprintf("%s%s", s.prefix, token)
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// URL-safe base64 without padding
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// PasswordResetServiceInterface defines the interface for password reset operations
type PasswordResetServiceInterface interface {
	GenerateToken(ctx context.Context, userID string) (string, error)
	ValidateToken(ctx context.Context, token string) (string, error)
	ConsumeToken(ctx context.Context, token string) error
}

// NewPasswordResetServiceFactory creates password reset service based on environment configuration
func NewPasswordResetServiceFactory(redisClient *redis.Client, db *gorm.DB, ttl time.Duration) PasswordResetServiceInterface {
	// Check environment variable to choose implementation
	useDatabase := os.Getenv("USE_DATABASE_PWRESET") == "true"

	if useDatabase {
		return NewDatabaseService(db, ttl)
	}

	return NewService(redisClient, ttl)
}
