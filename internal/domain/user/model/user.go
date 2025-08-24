package model

import (
	"time"

	"gopi.com/internal/domain/model"
)

type User struct {
	model.Base
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Password      string    `json:"-"` // Never expose password in JSON
	Height        float64   `json:"height"`
	Weight        float64   `json:"weight"`
	OTP           string    `json:"-"` // Never expose OTP in JSON
	IsStaff       bool      `json:"is_staff"`
	IsActive      bool      `json:"is_active"`
	IsSuperuser   bool      `json:"is_superuser"`
	IsVerified    bool      `json:"is_verified"`
	DateJoined    time.Time `json:"date_joined"`
	LastLogin     *time.Time `json:"last_login"` // Can be null
	ProfileImageURL string   `json:"profile_image_url,omitempty"`
}

// GetFullName returns the full name of the user
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// IsAdmin returns true if the user is an admin
func (u *User) IsAdmin() bool {
	return u.IsStaff
}

// SetPassword sets the user's password (to be implemented with bcrypt)
func (u *User) SetPassword(password string) {
	// This will be implemented in the service layer
	u.Password = password
}

// CheckPassword checks if the provided password matches the user's password
func (u *User) CheckPassword(password string) bool {
	// This will be implemented in the service layer with bcrypt
	return u.Password == password
}