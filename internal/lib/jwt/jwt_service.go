package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"gopi.com/internal/domain/model"
	userModel "gopi.com/internal/domain/user/model"
)

type JWTService struct {
	secretKey     string
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
	blacklist     *RedisTokenBlacklist
}

type Claims struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	IsStaff     bool   `json:"is_staff"`
	IsSuperuser bool   `json:"is_superuser"`
	IsVerified  bool   `json:"is_verified"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func NewJWTService(secretKey string, tokenExpiry, refreshExpiry time.Duration, redisClient *redis.Client) *JWTService {
	return &JWTService{
		secretKey:     secretKey,
		tokenExpiry:   tokenExpiry,
		refreshExpiry: refreshExpiry,
		blacklist:     NewRedisTokenBlacklist(redisClient, "jwt_blacklist:"),
	}
}

// GenerateTokenPair generates both access and refresh tokens for a user
func (j *JWTService) GenerateTokenPair(user *userModel.User) (*TokenPair, error) {
	// Generate access token
	accessToken, err := j.generateToken(user, j.tokenExpiry)
	if err != nil {
		return nil, err
	}

	// Generate refresh token (longer expiry, no detailed claims)
	refreshClaims := &Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   "refresh",
		},
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshTokenObj.SignedString([]byte(j.secretKey))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(j.tokenExpiry.Seconds()),
	}, nil
}

// generateToken creates a JWT token for a user
func (j *JWTService) generateToken(user *userModel.User, expiry time.Duration) (string, error) {
	claims := &Claims{
		UserID:      user.ID,
		Email:       user.Email,
		Username:    user.Username,
		IsStaff:     user.IsStaff,
		IsSuperuser: user.IsSuperuser,
		IsVerified:  user.IsVerified,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   "access",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	// Check if token is blacklisted first
	if j.blacklist.IsTokenBlacklisted(tokenString) {
		return nil, errors.New("token has been invalidated")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken generates a new access token from a valid refresh token
func (j *JWTService) RefreshToken(refreshToken string, user *userModel.User) (string, error) {
	claims, err := j.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Verify this is a refresh token
	if claims.Subject != "refresh" {
		return "", errors.New("invalid refresh token")
	}

	// Verify the user ID matches
	if claims.UserID != user.ID {
		return "", errors.New("token user mismatch")
	}

	// Generate new access token
	return j.generateToken(user, j.tokenExpiry)
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func (j *JWTService) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	// Check if it starts with "Bearer "
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("invalid authorization header format")
	}

	return authHeader[len(bearerPrefix):], nil
}

// GetUserFromToken extracts user information from token claims
func (j *JWTService) GetUserFromToken(tokenString string) (*userModel.User, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Create a user object from claims (minimal info)
	user := &userModel.User{
		Base: model.Base{
			ID: claims.UserID,
		},
		Email:       claims.Email,
		Username:    claims.Username,
		IsStaff:     claims.IsStaff,
		IsSuperuser: claims.IsSuperuser,
		IsVerified:  claims.IsVerified,
	}

	return user, nil
}

// BlacklistToken adds a token to the blacklist
func (j *JWTService) BlacklistToken(tokenString string) error {
	// Parse the token to get its expiration time
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		// Add to blacklist with its expiration time
		j.blacklist.BlacklistToken(tokenString, claims.ExpiresAt.Time)
		return nil
	}

	return errors.New("invalid token claims")
}

// IsTokenBlacklisted checks if a token is blacklisted
func (j *JWTService) IsTokenBlacklisted(tokenString string) bool {
	return j.blacklist.IsTokenBlacklisted(tokenString)
}

// GetBlacklistedTokenCount returns the number of blacklisted tokens
func (j *JWTService) GetBlacklistedTokenCount() (int64, error) {
	return j.blacklist.GetBlacklistedCount()
}
