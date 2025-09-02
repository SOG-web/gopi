package pwreset

import (
	"context"
	"errors"
	"time"

	"gopi.com/internal/domain/model"
	"gorm.io/gorm"
)

// PasswordResetToken represents a password reset token in the database
type PasswordResetToken struct {
	model.Base
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	UserID    string    `gorm:"not null;index" json:"user_id"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	IsUsed    bool      `gorm:"default:false;index" json:"is_used"`
}

// DatabaseService manages password reset tokens using database instead of Redis
type DatabaseService struct {
	db  *gorm.DB
	ttl time.Duration
}

// NewDatabaseService creates a new database-based password reset service
func NewDatabaseService(db *gorm.DB, ttl time.Duration) *DatabaseService {
	// Auto-migrate the table
	db.AutoMigrate(&PasswordResetToken{})

	return &DatabaseService{
		db:  db,
		ttl: ttl,
	}
}

// GenerateToken creates a new secure token for the given user ID and stores it in database
// Returns the token string which should be sent to the user via email link
func (s *DatabaseService) GenerateToken(ctx context.Context, userID string) (string, error) {
	if userID == "" {
		return "", errors.New("userID is required")
	}

	token, err := randomToken(32)
	if err != nil {
		return "", err
	}

	// Create the password reset token record
	resetToken := &PasswordResetToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.ttl),
		IsUsed:    false,
	}

	// Save to database
	if err := s.db.Create(resetToken).Error; err != nil {
		return "", err
	}

	return token, nil
}

// ValidateToken returns the associated userID if the token exists and is not expired
func (s *DatabaseService) ValidateToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", errors.New("token is required")
	}

	var resetToken PasswordResetToken
	err := s.db.Where("token = ? AND expires_at > ? AND is_used = false", token, time.Now()).
		First(&resetToken).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("invalid or expired token")
		}
		return "", err
	}

	return resetToken.UserID, nil
}

// ConsumeToken deletes the token to enforce single-use semantics
func (s *DatabaseService) ConsumeToken(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token is required")
	}

	// Mark the token as used instead of deleting it (for audit purposes)
	result := s.db.Model(&PasswordResetToken{}).
		Where("token = ? AND expires_at > ? AND is_used = false", token, time.Now()).
		Update("is_used", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("token not found or already used")
	}

	return nil
}

// GetTokenCount returns the number of active (unused and not expired) tokens
func (s *DatabaseService) GetTokenCount() (int64, error) {
	var count int64
	err := s.db.Model(&PasswordResetToken{}).
		Where("expires_at > ? AND is_used = false", time.Now()).
		Count(&count).Error

	return count, err
}

// GetExpiredTokenCount returns the number of expired tokens that can be cleaned up
func (s *DatabaseService) GetExpiredTokenCount() (int64, error) {
	var count int64
	err := s.db.Model(&PasswordResetToken{}).
		Where("expires_at <= ?", time.Now()).
		Count(&count).Error

	return count, err
}

// GetUsedTokenCount returns the number of used tokens
func (s *DatabaseService) GetUsedTokenCount() (int64, error) {
	var count int64
	err := s.db.Model(&PasswordResetToken{}).
		Where("is_used = true").
		Count(&count).Error

	return count, err
}

// ClearExpiredTokens removes expired tokens from the database
func (s *DatabaseService) ClearExpiredTokens() error {
	return s.db.Where("expires_at <= ?", time.Now()).Delete(&PasswordResetToken{}).Error
}

// GetTotalTokenCount returns the total number of tokens in the database
func (s *DatabaseService) GetTotalTokenCount() (int64, error) {
	var count int64
	err := s.db.Model(&PasswordResetToken{}).Count(&count).Error
	return count, err
}

// IsTokenUsed checks if a token has been used
func (s *DatabaseService) IsTokenUsed(token string) (bool, error) {
	var resetToken PasswordResetToken
	err := s.db.Where("token = ?", token).First(&resetToken).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("token not found")
		}
		return false, err
	}

	return resetToken.IsUsed, nil
}
