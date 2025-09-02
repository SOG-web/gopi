package dto

import "time"

// Registration DTOs (Django's RegistrationSerializer equivalent)
type RegistrationRequest struct {
	Username  string  `json:"username" binding:"required"`
	Email     string  `json:"email" binding:"required,email"`
	FirstName string  `json:"first_name" binding:"required"`
	LastName  string  `json:"last_name" binding:"required"`
	Password  string  `json:"password" binding:"required,min=8"`
	Height    float64 `json:"height" binding:"required"`
	Weight    float64 `json:"weight" binding:"required"`
}

type RegistrationResponse struct {
	Response    string    `json:"response"`
	Success     bool      `json:"success"`
	StatusCode  int       `json:"status_code"`
	Data        *UserData `json:"data,omitempty"`
	UserID      string    `json:"user_id,omitempty"`
	Token       string    `json:"token,omitempty"`
	DateJoined  time.Time `json:"date_joined,omitempty"`
	IsVerified  bool      `json:"is_verified,omitempty"`
	ErrorMessage string   `json:"error_message,omitempty"`
}

// Login DTOs (Django's UserLoginSerializer equivalent)
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Message     string `json:"message,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	UserEmail   string `json:"user_email,omitempty"`
	Token       string `json:"token,omitempty"`
	StatusCode  int    `json:"status_code"`
	Success     bool   `json:"success,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// OTP Verification DTOs (Django's VerifyUserSerializer equivalent)
type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

type VerifyOTPResponse struct {
	Success      bool   `json:"success"`
	Code         int    `json:"code"`
	Message      string `json:"message,omitempty"`
	Extra        string `json:"extra,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	Status       int    `json:"status,omitempty"`
}

// Resend OTP DTOs (Django's ResendOTPSerializer equivalent)
type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResendOTPResponse struct {
	Message    string      `json:"message"`
	Sent       bool        `json:"sent"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data"`
}

// Change Password DTOs (Django's ChangeUserPasswordSerializer equivalent)
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ChangePasswordResponse struct {
	Success      bool        `json:"success"`
	StatusCode   int         `json:"status_code"`
	Message      string      `json:"message,omitempty"`
	ErrorMessage string      `json:"error_message,omitempty"`
	Data         interface{} `json:"data,omitempty"`
}

// User Data DTOs (Django's UserSerializer equivalent)
type UserData struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Height      float64    `json:"height"`
	Weight      float64    `json:"weight"`
	IsStaff     bool       `json:"is_staff"`
	IsActive    bool       `json:"is_active"`
	IsSuperuser bool       `json:"is_superuser"`
	IsVerified  bool       `json:"is_verified"`
	DateJoined  time.Time  `json:"date_joined"`
	LastLogin   *time.Time `json:"last_login"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ProfileImageURL string  `json:"profile_image_url,omitempty"`
}

// User Management DTOs
type UpdateUserRequest struct {
	Username  string  `json:"username,omitempty"`
	FirstName string  `json:"first_name,omitempty"`
	LastName  string  `json:"last_name,omitempty"`
	Height    float64 `json:"height,omitempty" binding:"min=0"`
	Weight    float64 `json:"weight,omitempty" binding:"min=0"`
}

type UpdateUserResponse struct {
	Success    bool      `json:"success"`
	StatusCode int       `json:"status_code"`
	Message    string    `json:"message"`
	Data       *UserData `json:"data"`
}

type UserProfileResponse struct {
	Success    bool      `json:"success"`
	StatusCode int       `json:"status_code"`
	Data       *UserData `json:"data"`
}

type GetUsersResponse struct {
	Success    bool        `json:"success"`
	StatusCode int         `json:"status_code"`
	Data       []*UserData `json:"data"`
	Count      int         `json:"count"`
}

// Logout Response
type LogoutResponse struct {
	Message    string `json:"message"`
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
}

// Delete Account Response
type DeleteAccountResponse struct {
	Detail string `json:"detail"`
}

// Admin Management DTOs
type AdminActionResponse struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

type UserStatsResponse struct {
	Success    bool                   `json:"success"`
	StatusCode int                    `json:"status_code"`
	Data       map[string]interface{} `json:"data"`
}

type BulkEmailRequest struct {
	UserIDs []string `json:"user_ids" binding:"required"`
	Subject string   `json:"subject" binding:"required"`
	Content string   `json:"content" binding:"required"`
}

type ApologyEmailRequest struct {
	Users []map[string]string `json:"users" binding:"required"`
}

// Generic Auth Error Response
type AuthErrorResponse struct {
	Error       string `json:"error,omitempty"`
	Success     bool   `json:"success"`
	StatusCode  int    `json:"status_code"`
	ErrorMessage string `json:"error_message,omitempty"`
	Errors      interface{} `json:"errors,omitempty"`
}

// Password Reset DTOs
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type PasswordResetResponse struct {
	Message    string      `json:"message"`
	Success    bool        `json:"success"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data,omitempty"`
}

type PasswordResetConfirmRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type PasswordResetConfirmResponse struct {
	Message    string      `json:"message"`
	Success    bool        `json:"success"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data,omitempty"`
}
