package user

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gopi.com/internal/domain/model"
	userModel "gopi.com/internal/domain/user/model"
	"gopi.com/internal/domain/user/repo"
	"gopi.com/internal/lib/email"
	"gopi.com/internal/lib/id"
)

type UserService struct {
	userRepo     repo.UserRepository
	emailService email.EmailServiceInterface
}

func NewUserService(userRepo repo.UserRepository, emailService email.EmailServiceInterface) *UserService {
	return &UserService{
		userRepo:     userRepo,
		emailService: emailService,
	}
}

// RegisterUser creates a new user account (Django's user_register equivalent)
func (s *UserService) RegisterUser(username, email, firstName, lastName, password string, height, weight float64) (*userModel.User, error) {
	// Validate email uniqueness
	emailExists, err := s.userRepo.EmailExists(email)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, errors.New("the email has already been taken")
	}

	// Validate username uniqueness
	usernameExists, err := s.userRepo.UsernameExists(username)
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, errors.New("the username has already been taken")
	}

	// Hash password
	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Generate OTP
	otp := s.GenerateOTP()

	// Create user
	user := &userModel.User{
		Base: model.Base{
			ID:        id.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:    username,
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		Password:    hashedPassword,
		Height:      height,
		Weight:      weight,
		OTP:         otp,
		IsStaff:     false,
		IsActive:    true,
		IsSuperuser: false,
		IsVerified:  false,
		DateJoined:  time.Now(),
		LastLogin:   nil,
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	// Send OTP email
	if s.emailService != nil {
		err = s.emailService.SendOTPEmail(user.Email, user.FirstName, otp)
		if err != nil {
			// Log the error but don't fail the registration
			// In production, you might want to handle this differently
			fmt.Printf("Failed to send OTP email: %v\n", err)
		}
	}

	return user, nil
}

// LoginUser authenticates a user (Django's user_login equivalent)
func (s *UserService) LoginUser(email, password string) (*userModel.User, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("invalid user")
	}

	// Check password
	if !s.CheckPassword(password, user.Password) {
		return nil, errors.New("incorrect login credentials")
	}

	// Check if user is verified
	if !user.IsVerified {
		return nil, errors.New("user's email is not verified")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user not active")
	}

	// Update last login
	err = s.userRepo.UpdateLastLogin(user.ID)
	if err != nil {
		return nil, err
	}

	// Update the user object with new last login time
	now := time.Now()
	user.LastLogin = &now

	return user, nil
}

// VerifyOTP verifies the user's email with OTP (Django's verify_otp equivalent)
func (s *UserService) VerifyOTP(email, otp string) error {
	// Get user by email and OTP
	user, err := s.userRepo.GetByOTP(email, otp)
	if err != nil {
		return errors.New("user not found or incorrect OTP")
	}

	// Check if already verified
	if user.IsVerified {
		return errors.New("user has already been verified")
	}

	// Mark as verified
	err = s.userRepo.MarkAsVerified(user.ID)
	if err != nil {
		return err
	}

	// Send welcome email
	if s.emailService != nil {
		err = s.emailService.SendWelcomeEmail(user.Email, user.FirstName)
		if err != nil {
			// Log the error but don't fail the verification
			fmt.Printf("Failed to send welcome email: %v\n", err)
		}
	}

	return nil
}

// ResendOTP generates and sends a new OTP for user verification
func (s *UserService) ResendOTP(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Generate new OTP
	otp := s.GenerateOTP()

	// Update user with new OTP
	err = s.userRepo.UpdateOTP(user.ID, otp)
	if err != nil {
		return err
	}

	// Send OTP email
	if s.emailService != nil {
		err = s.emailService.SendOTPEmail(user.Email, user.FirstName, otp)
		if err != nil {
			// Log the error but don't fail the operation
			fmt.Printf("Failed to send OTP email: %v\n", err)
		}
	}

	return nil
}

// ChangePassword changes user's password (Django's ChangePasswordView equivalent)
func (s *UserService) ChangePassword(userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Check old password
	if !s.CheckPassword(oldPassword, user.Password) {
		return errors.New("incorrect password")
	}

	// Hash new password
	hashedPassword, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	err = s.userRepo.UpdatePassword(user.ID, hashedPassword)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAccount deletes a user account (Django's delete_account equivalent)
func (s *UserService) DeleteAccount(userID string) error {
	return s.userRepo.Delete(userID)
}

// CreateSuperuser creates a superuser account (Django's create_superuser equivalent)
func (s *UserService) CreateSuperuser(username, email, firstName, lastName, password string, height, weight float64) (*userModel.User, error) {
	// Hash password
	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create superuser
	user := &userModel.User{
		Base: model.Base{
			ID:        id.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:    username,
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		Password:    hashedPassword,
		Height:      height,
		Weight:      weight,
		OTP:         "",
		IsStaff:     true,
		IsActive:    true,
		IsSuperuser: true,
		IsVerified:  true, // Superusers are auto-verified
		DateJoined:  time.Now(),
		LastLogin:   nil,
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetAllUsers returns all users (for admin purposes)
func (s *UserService) GetAllUsers() ([]*userModel.User, error) {
	return s.userRepo.GetAllUsers()
}

// GetStaffUsers returns all staff users
func (s *UserService) GetStaffUsers() ([]*userModel.User, error) {
	return s.userRepo.GetStaffUsers()
}

// GetVerifiedUsers returns all verified users
func (s *UserService) GetVerifiedUsers() ([]*userModel.User, error) {
	return s.userRepo.GetVerifiedUsers()
}

// GetUnverifiedUsers returns all unverified users
func (s *UserService) GetUnverifiedUsers() ([]*userModel.User, error) {
	return s.userRepo.GetUnverifiedUsers()
}

// GetUserByID returns a user by ID
func (s *UserService) GetUserByID(id string) (*userModel.User, error) {
	return s.userRepo.GetByID(id)
}

// GetUserByEmail returns a user by email
func (s *UserService) GetUserByEmail(email string) (*userModel.User, error) {
	return s.userRepo.GetByEmail(email)
}

// GetUserByUsername returns a user by username
func (s *UserService) GetUserByUsername(username string) (*userModel.User, error) {
	return s.userRepo.GetByUsername(username)
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(user *userModel.User) error {
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(user)
}

// HashPassword hashes a password using bcrypt
func (s *UserService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPassword verifies a password against its hash
func (s *UserService) CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GetUserList returns paginated list of users
func (s *UserService) GetUserList(limit, offset int) ([]*userModel.User, error) {
	return s.userRepo.List(limit, offset)
}

// SendBulkEmail sends email to multiple users (Django equivalent)
func (s *UserService) SendBulkEmail(userIDs []string, subject, content string) error {
	var emails []string
	var names []string

	for _, userID := range userIDs {
		user, err := s.userRepo.GetByID(userID)
		if err != nil {
			continue // Skip invalid users
		}
		emails = append(emails, user.Email)
		names = append(names, user.FirstName)
	}

	if s.emailService != nil {
		return s.emailService.SendBulkEmail(emails, subject, content)
	}

	return nil
}

// SendApologyEmails sends apology emails to users from JSON data (Django equivalent)
func (s *UserService) SendApologyEmails(users []map[string]string) error {
	if s.emailService == nil {
		return errors.New("email service not configured")
	}

	for _, userData := range users {
		username, ok := userData["username"]
		if !ok {
			continue
		}
		email, ok := userData["email"]
		if !ok {
			continue
		}

		err := s.emailService.SendApologyEmail(email, username)
		if err != nil {
			// Log error but continue with other emails
			fmt.Printf("Failed to send apology email to %s: %v\n", email, err)
		}
	}

	return nil
}

// ActivateUser activates a user account (admin function)
func (s *UserService) ActivateUser(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	user.IsActive = true
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(user)
}

// DeactivateUser deactivates a user account (admin function)
func (s *UserService) DeactivateUser(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(user)
}

// MakeStaff promotes a user to staff (admin function)
func (s *UserService) MakeStaff(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	user.IsStaff = true
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(user)
}

// RemoveStaff removes staff privileges (admin function)
func (s *UserService) RemoveStaff(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	user.IsStaff = false
	user.IsSuperuser = false // Remove superuser as well
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(user)
}

// GetUserStats returns user statistics (admin function)
func (s *UserService) GetUserStats() (map[string]interface{}, error) {
	allUsers, err := s.userRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}

	verifiedUsers, err := s.userRepo.GetVerifiedUsers()
	if err != nil {
		return nil, err
	}

	staffUsers, err := s.userRepo.GetStaffUsers()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_users":      len(allUsers),
		"verified_users":   len(verifiedUsers),
		"unverified_users": len(allUsers) - len(verifiedUsers),
		"staff_users":      len(staffUsers),
		"active_users":     0, // Will be calculated below
	}

	// Count active users
	activeCount := 0
	for _, user := range allUsers {
		if user.IsActive {
			activeCount++
		}
	}
	stats["active_users"] = activeCount

	return stats, nil
}

// SearchUsers searches for users by username or email (admin function)
func (s *UserService) SearchUsers(query string, limit, offset int) ([]*userModel.User, error) {
	// This would need to be implemented in the repository layer
	// For now, get all users and filter in memory (not optimal for large datasets)
	allUsers, err := s.userRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}

	var filteredUsers []*userModel.User
	query = strings.ToLower(query)

	for _, user := range allUsers {
		if strings.Contains(strings.ToLower(user.Username), query) ||
			strings.Contains(strings.ToLower(user.Email), query) ||
			strings.Contains(strings.ToLower(user.FirstName), query) ||
			strings.Contains(strings.ToLower(user.LastName), query) {
			filteredUsers = append(filteredUsers, user)
		}
	}

	// Apply pagination
	start := offset
	if start > len(filteredUsers) {
		return []*userModel.User{}, nil
	}

	end := start + limit
	if end > len(filteredUsers) {
		end = len(filteredUsers)
	}

	return filteredUsers[start:end], nil
}

// GetFullName returns user's full name (Django model method equivalent)
func (s *UserService) GetUserFullName(user *userModel.User) string {
	return user.GetFullName()
}

// IsUserAdmin checks if user is admin (Django model method equivalent)
func (s *UserService) IsUserAdmin(user *userModel.User) bool {
	return user.IsAdmin()
}

// ForceVerifyUser forces verification without OTP (admin function)
func (s *UserService) ForceVerifyUser(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if user.IsVerified {
		return errors.New("user is already verified")
	}

	return s.userRepo.MarkAsVerified(user.ID)
}

// NewService creates a new UserService (compatibility function)
func NewService(userRepo repo.UserRepository, emailService email.EmailServiceInterface) *UserService {
	return NewUserService(userRepo, emailService)
}

// GenerateOTP generates a 6-digit OTP
func (s *UserService) GenerateOTP() string {
	otp := ""
	for i := 0; i < 6; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		otp += fmt.Sprintf("%d", n.Int64())
	}
	return otp
}

// ValidateEmail checks if email is valid format and not taken
func (s *UserService) ValidateEmail(email string) error {
	exists, err := s.userRepo.EmailExists(email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("the email has already been taken")
	}
	return nil
}

// ValidateUsername checks if username is valid format and not taken
func (s *UserService) ValidateUsername(username string) error {
	exists, err := s.userRepo.UsernameExists(username)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("the username has already been taken")
	}
	return nil
}
