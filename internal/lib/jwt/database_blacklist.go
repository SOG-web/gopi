package jwt

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
	"gopi.com/internal/domain/model"
)

// BlacklistedToken represents a blacklisted token in the database
type BlacklistedToken struct {
	model.Base
	TokenHash  string    `gorm:"uniqueIndex;not null" json:"token_hash"`
	ExpiresAt  time.Time `gorm:"not null;index" json:"expires_at"`
	Token      string    `gorm:"-" json:"-"` // Don't store the actual token
}

// DatabaseTokenBlacklist manages blacklisted JWT tokens using database
type DatabaseTokenBlacklist struct {
	db     *gorm.DB
	prefix string
}

// NewDatabaseTokenBlacklist creates a new database-based token blacklist
func NewDatabaseTokenBlacklist(db *gorm.DB) *DatabaseTokenBlacklist {
	// Auto-migrate the table
	db.AutoMigrate(&BlacklistedToken{})

	return &DatabaseTokenBlacklist{
		db:     db,
		prefix: "jwt_blacklist:",
	}
}

// hashToken creates a SHA256 hash of the token for storage
func (dtb *DatabaseTokenBlacklist) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// BlacklistToken adds a token to the database blacklist with expiration
func (dtb *DatabaseTokenBlacklist) BlacklistToken(token string, expiresAt time.Time) error {
	tokenHash := dtb.hashToken(token)

	// Create the blacklisted token record
	blacklistedToken := &BlacklistedToken{
		TokenHash: dtb.prefix + tokenHash,
		ExpiresAt: expiresAt,
	}

	// Use GORM's upsert functionality
	return dtb.db.Where(BlacklistedToken{TokenHash: blacklistedToken.TokenHash}).
		Assign(BlacklistedToken{ExpiresAt: expiresAt}).
		FirstOrCreate(blacklistedToken).Error
}

// IsTokenBlacklisted checks if a token is in the database blacklist
func (dtb *DatabaseTokenBlacklist) IsTokenBlacklisted(token string) bool {
	tokenHash := dtb.hashToken(token)

	var count int64
	dtb.db.Model(&BlacklistedToken{}).
		Where("token_hash = ? AND expires_at > ?", dtb.prefix+tokenHash, time.Now()).
		Count(&count)

	return count > 0
}

// GetBlacklistedCount returns the number of active blacklisted tokens
func (dtb *DatabaseTokenBlacklist) GetBlacklistedCount() (int64, error) {
	var count int64
	err := dtb.db.Model(&BlacklistedToken{}).
		Where("expires_at > ?", time.Now()).
		Count(&count).Error

	return count, err
}

// ClearExpiredTokens removes expired tokens from the database
func (dtb *DatabaseTokenBlacklist) ClearExpiredTokens() error {
	return dtb.db.Where("expires_at <= ?", time.Now()).Delete(&BlacklistedToken{}).Error
}

// GetExpiredTokenCount returns the number of expired tokens that can be cleaned up
func (dtb *DatabaseTokenBlacklist) GetExpiredTokenCount() (int64, error) {
	var count int64
	err := dtb.db.Model(&BlacklistedToken{}).
		Where("expires_at <= ?", time.Now()).
		Count(&count).Error

	return count, err
}

// GetTotalTokenCount returns the total number of tokens in the blacklist (including expired)
func (dtb *DatabaseTokenBlacklist) GetTotalTokenCount() (int64, error) {
	var count int64
	err := dtb.db.Model(&BlacklistedToken{}).Count(&count).Error
	return count, err
}
