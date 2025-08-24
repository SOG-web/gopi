package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gopi.com/api/http/dto"
	userService "gopi.com/internal/app/user"
	userModel "gopi.com/internal/domain/user/model"
	"gopi.com/internal/lib/jwt"
)

type AuthHandler struct {
	userService *userService.UserService
	jwtService  *jwt.JWTService
}

func NewAuthHandler(userService *userService.UserService, jwtService *jwt.JWTService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtService:  jwtService,
	}
}

// UserRegister handles user registration (Django's user_register equivalent)
// @Summary User Registration
// @Description Register a new user account with email verification
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RegistrationRequest true "Registration details"
// @Success 201 {object} dto.RegistrationResponse "User registered successfully, OTP sent to email"
// @Failure 400 {object} dto.RegistrationResponse "Invalid input or validation error"
// @Failure 409 {object} dto.RegistrationResponse "User already exists"
// @Failure 500 {object} dto.RegistrationResponse "Internal server error"
// @Router /api/auth/register [post]
func (h *AuthHandler) UserRegister(c *gin.Context) {
	var req dto.RegistrationRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.RegistrationResponse{
			Response:     "Error",
			Success:      false,
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	// Validate email
	if err := h.userService.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, dto.RegistrationResponse{
			Response:     "Error",
			Success:      false,
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	// Validate username
	if err := h.userService.ValidateUsername(req.Username); err != nil {
		c.JSON(http.StatusBadRequest, dto.RegistrationResponse{
			Response:     "Error",
			Success:      false,
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	// Create user
	user, err := h.userService.RegisterUser(
		req.Username,
		req.Email,
		req.FirstName,
		req.LastName,
		req.Password,
		req.Height,
		req.Weight,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.RegistrationResponse{
			Response:     "Error",
			Success:      false,
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	// Generate JWT token pair
	tokenPair, err := h.jwtService.GenerateTokenPair(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.RegistrationResponse{
			Response:     "Error",
			Success:      false,
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: "Failed to generate token",
		})
		return
	}

	userData := &dto.UserData{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Height:      user.Height,
		Weight:      user.Weight,
		IsStaff:     user.IsStaff,
		IsActive:    user.IsActive,
		IsSuperuser: user.IsSuperuser,
		IsVerified:  user.IsVerified,
		DateJoined:  user.DateJoined,
		LastLogin:   user.LastLogin,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	c.JSON(http.StatusCreated, dto.RegistrationResponse{
		Response:   "Successful register new user",
		Success:    true,
		StatusCode: http.StatusCreated,
		Data:       userData,
		UserID:     user.ID,
		Token:      tokenPair.AccessToken,
		DateJoined: user.DateJoined,
		IsVerified: user.IsVerified,
	})
}

// UserLogin handles user login (Django's user_login equivalent)
// @Summary User Login
// @Description Authenticate user and return JWT tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse "Login successful with JWT tokens"
// @Failure 400 {object} dto.LoginResponse "Invalid request format"
// @Failure 401 {object} dto.LoginResponse "Invalid credentials"
// @Failure 403 {object} dto.LoginResponse "Account not verified or inactive"
// @Failure 500 {object} dto.LoginResponse "Internal server error"
// @Router /api/auth/login [post]
func (h *AuthHandler) UserLogin(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.LoginResponse{
			ErrorMessage: err.Error(),
			Success:      false,
			StatusCode:   http.StatusBadRequest,
		})
		return
	}

	// Authenticate user
	user, err := h.userService.LoginUser(req.Email, req.Password)
	if err != nil {
		var statusCode int
		switch err.Error() {
		case "user's email is not verified":
			statusCode = http.StatusUnauthorized
		case "invalid user", "incorrect login credentials", "user not active":
			statusCode = http.StatusForbidden
		default:
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, dto.LoginResponse{
			ErrorMessage: err.Error(),
			Success:      false,
			StatusCode:   statusCode,
		})
		return
	}

	// Generate JWT token pair
	tokenPair, err := h.jwtService.GenerateTokenPair(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.LoginResponse{
			ErrorMessage: "Failed to generate token",
			Success:      false,
			StatusCode:   http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		Message:    "User logged in successfully!",
		UserID:     user.ID,
		UserEmail:  user.Email,
		Token:      tokenPair.AccessToken,
		StatusCode: http.StatusOK,
		Success:    true,
	})
}

// UserLogout handles user logout (Django's user_logout equivalent)
// @Summary User Logout
// @Description Log out user and invalidate JWT tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.LogoutResponse "Logout successful"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - invalid or missing token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/auth/logout [post]
func (h *AuthHandler) UserLogout(c *gin.Context) {
	// The JWT middleware should have already validated the token and set user context
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.AuthErrorResponse{
			Error:      "Unauthorized - user not authenticated",
			Success:    false,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	// Get the token from the Authorization header to blacklist it
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString := authHeader[7:]
		
		// Add the token to blacklist
		err := h.jwtService.BlacklistToken(tokenString)
		if err != nil {
			// Log the error but don't fail the logout
			// In production, you might want to log this properly
		}
	}

	// Optional: Update user's last_logout timestamp in the future
	// h.userService.UpdateLastLogout(userID)

	c.JSON(http.StatusOK, dto.LogoutResponse{
		Message:    "User logged out successfully",
		Success:    true,
		StatusCode: http.StatusOK,
	})
}

// VerifyOTP handles OTP verification (Django's verify_otp equivalent)
// @Summary Verify OTP
// @Description Verify email OTP to activate user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.VerifyOTPRequest true "OTP verification details"
// @Success 200 {object} dto.VerifyOTPResponse "OTP verified successfully, account activated"
// @Failure 400 {object} dto.VerifyOTPResponse "Invalid request format"
// @Failure 401 {object} dto.VerifyOTPResponse "Invalid or expired OTP"
// @Failure 404 {object} dto.VerifyOTPResponse "User not found"
// @Failure 500 {object} dto.VerifyOTPResponse "Internal server error"
// @Router /api/auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.VerifyOTPResponse{
			Success:      false,
			Code:         http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	err := h.userService.VerifyOTP(req.Email, req.OTP)
	if err != nil {
		var statusCode int
		switch err.Error() {
		case "user has already been verified":
			statusCode = http.StatusForbidden
		case "user not found or incorrect OTP":
			statusCode = http.StatusNotFound
		default:
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, dto.VerifyOTPResponse{
			Success:      false,
			Status:       statusCode,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.VerifyOTPResponse{
		Success: true,
		Code:    http.StatusOK,
		Message: "Account Created and Verified!",
		Extra:   "You have successfully signed up on GoPadi",
		Data:    req,
	})
}

// DeleteAccount handles account deletion (Django's delete_account equivalent)
// @Summary Delete User Account
// @Description Permanently delete the authenticated user's account
// @Tags Authentication
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.DeleteAccountResponse "Account deleted successfully"
// @Failure 401 {object} dto.AuthErrorResponse "Unauthorized - invalid or missing token"
// @Failure 404 {object} dto.AuthErrorResponse "User not found"
// @Failure 500 {object} dto.AuthErrorResponse "Internal server error"
// @Router /api/auth/delete-account [delete]
func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	// Here you would get the user ID from the authenticated context
	userID := c.GetString("user_id") // This would come from auth middleware

	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.AuthErrorResponse{
			Error:      "Unauthorized",
			Success:    false,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	err := h.userService.DeleteAccount(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AuthErrorResponse{
			Error:      err.Error(),
			Success:    false,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusOK, dto.DeleteAccountResponse{
		Detail: "Account deleted!",
	})
}

// ResendOTP handles OTP resend (Django's ResendOTPAPIView equivalent)
// @Summary Resend OTP
// @Description Resend verification OTP to user's email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.ResendOTPResponse "OTP sent successfully"
// @Failure 400 {object} dto.AuthErrorResponse "Invalid user ID"
// @Failure 404 {object} dto.AuthErrorResponse "User not found"
// @Failure 429 {object} dto.AuthErrorResponse "Too many requests - rate limited"
// @Failure 500 {object} dto.AuthErrorResponse "Internal server error"
// @Router /api/auth/resend-otp/{id} [post]
func (h *AuthHandler) ResendOTP(c *gin.Context) {
	userIDParam := c.Param("id")
	if userIDParam == "" {
		c.JSON(http.StatusBadRequest, dto.AuthErrorResponse{
			Error:      "User ID is required",
			Success:    false,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	err := h.userService.ResendOTP(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AuthErrorResponse{
			Error:      err.Error(),
			Success:    false,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	// Get user email for response
	user, err := h.userService.GetUserByID(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AuthErrorResponse{
			Error:      err.Error(),
			Success:    false,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusOK, dto.ResendOTPResponse{
		Message:    "OTP sent to " + user.Email,
		Sent:       true,
		StatusCode: http.StatusOK,
		Data:       gin.H{"email": user.Email},
	})
}

// ChangePassword handles password change (Django's ChangePasswordView equivalent)
// @Summary Change Password
// @Description Change user password with current password verification
// @Tags Authentication
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.ChangePasswordRequest true "Password change details"
// @Success 200 {object} dto.ChangePasswordResponse "Password changed successfully"
// @Failure 400 {object} dto.ChangePasswordResponse "Invalid request format"
// @Failure 401 {object} dto.ChangePasswordResponse "Unauthorized or invalid current password"
// @Failure 404 {object} dto.ChangePasswordResponse "User not found"
// @Failure 500 {object} dto.ChangePasswordResponse "Internal server error"
// @Router /api/auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ChangePasswordResponse{
			Success:      false,
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	// Here you would get the user ID from the authenticated context
	userID := c.GetString("user_id") // This would come from auth middleware

	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.ChangePasswordResponse{
			Success:      false,
			StatusCode:   http.StatusUnauthorized,
			ErrorMessage: "Unauthorized",
		})
		return
	}

	err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ChangePasswordResponse{
			Success:      false,
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ChangePasswordResponse{
		Success:    true,
		StatusCode: http.StatusOK,
		Message:    "Password updated successfully!",
		Data:       gin.H{"old_password": "", "new_password": ""}, // Don't expose actual passwords
	})
}

// Helper function to convert user model to DTO
func (h *AuthHandler) userModelToDTO(user *userModel.User) *dto.UserData {
	return &dto.UserData{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Height:      user.Height,
		Weight:      user.Weight,
		IsStaff:     user.IsStaff,
		IsActive:    user.IsActive,
		IsSuperuser: user.IsSuperuser,
		IsVerified:  user.IsVerified,
		DateJoined:  user.DateJoined,
		LastLogin:   user.LastLogin,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
