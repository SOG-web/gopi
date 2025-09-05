package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gopi.com/api/http/dto"
	userService "gopi.com/internal/app/user"
	"gopi.com/internal/lib/email"
	"gopi.com/internal/lib/pwreset"
)

// PasswordResetHandler manages password reset request and confirmation endpoints.
type PasswordResetHandler struct {
	userService  *userService.UserService
	pwService    pwreset.PasswordResetServiceInterface
	emailService email.EmailServiceInterface
	publicHost   string
}

func NewPasswordResetHandler(userSvc *userService.UserService, pwSvc pwreset.PasswordResetServiceInterface, emailSvc email.EmailServiceInterface, publicHost string) *PasswordResetHandler {
	return &PasswordResetHandler{
		userService:  userSvc,
		pwService:    pwSvc,
		emailService: emailSvc,
		publicHost:   publicHost,
	}
}

// RequestPasswordReset handles generating a reset token and emailing a link.
// @Summary Request Password Reset
// @Description Generate a password reset token and send reset link to email. Always returns 200 to prevent user enumeration.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.PasswordResetRequest true "Email address to reset"
// @Success 200 {object} dto.PasswordResetResponse "If that email exists, a reset link has been sent."
// @Failure 400 {object} dto.PasswordResetResponse "Invalid request payload"
// @Router /auth/password-reset/request [post]
func (h *PasswordResetHandler) RequestPasswordReset(c *gin.Context) {
	var req dto.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.PasswordResetResponse{
			Message:    err.Error(),
			Success:    false,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	// Always return 200 to avoid user enumeration
	successResp := dto.PasswordResetResponse{
		Message:    "If that email exists, a reset link has been sent.",
		Success:    true,
		StatusCode: http.StatusOK,
	}

	// Try to find user; if not found, still return success
	user, err := h.userService.GetUserByEmail(req.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusOK, successResp)
		return
	}

	// Generate token
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()
	token, err := h.pwService.GenerateToken(ctx, user.ID)
	if err != nil {
		// Don't leak details; still return success
		c.JSON(http.StatusOK, successResp)
		return
	}

	// Build reset link for frontend
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", h.publicHost, token)

	// Send email asynchronously via EmailService (it queues internally)
	if h.emailService != nil {
		_ = h.emailService.SendPasswordResetEmail(user.Email, resetLink)
	}

	c.JSON(http.StatusOK, successResp)
}

// ConfirmPasswordReset handles verifying token and setting a new password.
// @Summary Confirm Password Reset
// @Description Validate a password reset token and set a new password. Token is single-use and expires after TTL.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.PasswordResetConfirmRequest true "Token and new password"
// @Success 200 {object} dto.PasswordResetConfirmResponse "Password has been reset successfully"
// @Failure 400 {object} dto.PasswordResetConfirmResponse "Invalid or expired token, or invalid payload"
// @Router /auth/password-reset/confirm [post]
func (h *PasswordResetHandler) ConfirmPasswordReset(c *gin.Context) {
	var req dto.PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.PasswordResetConfirmResponse{
			Message:    err.Error(),
			Success:    false,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	userID, err := h.pwService.ValidateToken(ctx, req.Token)
	if err != nil || userID == "" {
		c.JSON(http.StatusBadRequest, dto.PasswordResetConfirmResponse{
			Message:    "Invalid or expired token",
			Success:    false,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	// Reset the password via service
	if err := h.userService.ResetPassword(userID, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, dto.PasswordResetConfirmResponse{
			Message:    err.Error(),
			Success:    false,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	// Consume token (single-use)
	_ = h.pwService.ConsumeToken(ctx, req.Token)

	c.JSON(http.StatusOK, dto.PasswordResetConfirmResponse{
		Message:    "Password has been reset successfully",
		Success:    true,
		StatusCode: http.StatusOK,
	})
}
