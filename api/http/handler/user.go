package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gopi.com/api/http/dto"
	userService "gopi.com/internal/app/user"
	userModel "gopi.com/internal/domain/user/model"
	"gopi.com/internal/lib/storage"
)

type UserHandler struct {
	userService *userService.UserService
	storage     storage.Storage
}

func NewUserHandler(userService *userService.UserService, storage storage.Storage) *UserHandler {
	return &UserHandler{
		userService: userService,
		storage:     storage,
	}
}

// GetUserProfile gets current user's profile (authenticated user)
// @Summary Get User Profile
// @Description Get the profile information of the authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.UserProfileResponse "User profile retrieved successfully"
// @Failure 401 {object} dto.AuthErrorResponse "Unauthorized - invalid or missing token"
// @Failure 404 {object} dto.AuthErrorResponse "User not found"
// @Failure 500 {object} dto.AuthErrorResponse "Internal server error"
// @Router /api/users/profile [get]
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	userID := c.GetString("user_id") // From auth middleware

	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.AuthErrorResponse{
			Error:      "Unauthorized",
			Success:    false,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.AuthErrorResponse{
			Error:      "User not found",
			Success:    false,
			StatusCode: http.StatusNotFound,
		})
		return
	}

	userData := h.userModelToDTO(user)
	c.JSON(http.StatusOK, dto.UserProfileResponse{
		Success:    true,
		StatusCode: http.StatusOK,
		Data:       userData,
	})
}

// UpdateUserProfile updates current user's profile
// @Summary Update User Profile
// @Description Update the authenticated user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.UpdateUserRequest true "Profile update details"
// @Success 200 {object} dto.UserProfileResponse "Profile updated successfully"
// @Failure 400 {object} dto.AuthErrorResponse "Invalid request format"
// @Failure 401 {object} dto.AuthErrorResponse "Unauthorized - invalid or missing token"
// @Failure 404 {object} dto.AuthErrorResponse "User not found"
// @Failure 500 {object} dto.AuthErrorResponse "Internal server error"
// @Router /api/users/profile [put]
func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	userID := c.GetString("user_id") // From auth middleware

	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.AuthErrorResponse{
			Error:      "Unauthorized",
			Success:    false,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.AuthErrorResponse{
			Error:      err.Error(),
			Success:    false,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.AuthErrorResponse{
			Error:      "User not found",
			Success:    false,
			StatusCode: http.StatusNotFound,
		})
		return
	}

	// Update user fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Username != "" {
		// Check if username is available
		if err := h.userService.ValidateUsername(req.Username); err != nil && req.Username != user.Username {
			c.JSON(http.StatusBadRequest, dto.AuthErrorResponse{
				Error:      err.Error(),
				Success:    false,
				StatusCode: http.StatusBadRequest,
			})
			return
		}
		user.Username = req.Username
	}
	if req.Height > 0 {
		user.Height = req.Height
	}
	if req.Weight > 0 {
		user.Weight = req.Weight
	}

	err = h.userService.UpdateUser(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AuthErrorResponse{
			Error:      err.Error(),
			Success:    false,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	userData := h.userModelToDTO(user)
	c.JSON(http.StatusOK, dto.UpdateUserResponse{
		Success:    true,
		StatusCode: http.StatusOK,
		Message:    "Profile updated successfully",
		Data:       userData,
	})
}

// GetAllUsers returns all users (admin only)
// @Summary Get All Users
// @Description Retrieve all users in the system (admin access required)
// @Tags Users
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "Page number for pagination" default(1)
// @Param limit query int false "Number of users per page" default(20)
// @Success 200 {object} dto.GetUsersResponse "Users retrieved successfully"
// @Failure 401 {object} dto.AuthErrorResponse "Unauthorized - invalid or missing token"
// @Failure 403 {object} dto.AuthErrorResponse "Forbidden - admin access required"
// @Failure 500 {object} dto.AuthErrorResponse "Internal server error"
// @Router /api/users [get]
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.AuthErrorResponse{
			Error:      "Unauthorized",
			Success:    false,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	// Check if user is admin
	currentUser, err := h.userService.GetUserByID(userID)
	if err != nil || !currentUser.IsStaff {
		c.JSON(http.StatusForbidden, dto.AuthErrorResponse{
			Error:      "Access denied",
			Success:    false,
			StatusCode: http.StatusForbidden,
		})
		return
	}

	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.AuthErrorResponse{
			Error:      err.Error(),
			Success:    false,
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	usersData := make([]*dto.UserData, len(users))
	for i, user := range users {
		usersData[i] = h.userModelToDTO(user)
	}

	c.JSON(http.StatusOK, dto.GetUsersResponse{
		Success:    true,
		StatusCode: http.StatusOK,
		Data:       usersData,
		Count:      len(usersData),
	})
}

// GetStaffUsers returns all staff users (admin only)
// @Summary Get Staff Users
// @Description Retrieve all staff users (admin access required)
// @Tags Users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.GetUsersResponse "Staff users retrieved successfully"
// @Failure 401 {object} dto.AuthErrorResponse "Unauthorized - invalid or missing token"
// @Failure 403 {object} dto.AuthErrorResponse "Forbidden - admin access required"
// @Failure 500 {object} dto.AuthErrorResponse "Internal server error"
// @Router /api/users/staff [get]
func (h *UserHandler) GetStaffUsers(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.AuthErrorResponse{
			Error:      "Unauthorized",
			Success:    false,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	// Check if user is admin
	currentUser, err := h.userService.GetUserByID(userID)
	if err != nil || !currentUser.IsStaff {
		c.JSON(http.StatusForbidden, dto.AuthErrorResponse{
			Error:      "Access denied",
			Success:    false,
			StatusCode: http.StatusForbidden,
		})
		return
	}

	users, err := h.userService.GetStaffUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.AuthErrorResponse{
			Error:      err.Error(),
			Success:    false,
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	usersData := make([]*dto.UserData, len(users))
	for i, user := range users {
		usersData[i] = h.userModelToDTO(user)
	}

	c.JSON(http.StatusOK, dto.GetUsersResponse{
		Success:    true,
		StatusCode: http.StatusOK,
		Data:       usersData,
		Count:      len(usersData),
	})
}

// GetVerifiedUsers returns all verified users (admin only)
// @Summary Get Verified Users
// @Description Retrieve all verified users (admin access required)
// @Tags Users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.GetUsersResponse "Verified users retrieved successfully"
// @Failure 401 {object} dto.AuthErrorResponse "Unauthorized - invalid or missing token"
// @Failure 403 {object} dto.AuthErrorResponse "Forbidden - admin access required"
// @Failure 500 {object} dto.AuthErrorResponse "Internal server error"
// @Router /api/users/verified [get]
func (h *UserHandler) GetVerifiedUsers(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.AuthErrorResponse{
			Error:      "Unauthorized",
			Success:    false,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	// Check if user is admin
	currentUser, err := h.userService.GetUserByID(userID)
	if err != nil || !currentUser.IsStaff {
		c.JSON(http.StatusForbidden, dto.AuthErrorResponse{
			Error:      "Access denied",
			Success:    false,
			StatusCode: http.StatusForbidden,
		})
		return
	}

	users, err := h.userService.GetVerifiedUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.AuthErrorResponse{
			Error:      err.Error(),
			Success:    false,
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	usersData := make([]*dto.UserData, len(users))
	for i, user := range users {
		usersData[i] = h.userModelToDTO(user)
	}

	c.JSON(http.StatusOK, dto.GetUsersResponse{
		Success:    true,
		StatusCode: http.StatusOK,
		Data:       usersData,
		Count:      len(usersData),
	})
}

// GetUnverifiedUsers returns all unverified users (admin only)
// @Summary Get Unverified Users
// @Description Retrieve all unverified users (admin access required)
// @Tags Users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.GetUsersResponse "Unverified users retrieved successfully"
// @Failure 401 {object} dto.AuthErrorResponse "Unauthorized - invalid or missing token"
// @Failure 403 {object} dto.AuthErrorResponse "Forbidden - admin access required"
// @Failure 500 {object} dto.AuthErrorResponse "Internal server error"
// @Router /api/users/unverified [get]
func (h *UserHandler) GetUnverifiedUsers(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.AuthErrorResponse{
			Error:      "Unauthorized",
			Success:    false,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	// Check if user is admin
	currentUser, err := h.userService.GetUserByID(userID)
	if err != nil || !currentUser.IsStaff {
		c.JSON(http.StatusForbidden, dto.AuthErrorResponse{
			Error:      "Access denied",
			Success:    false,
			StatusCode: http.StatusForbidden,
		})
		return
	}

	users, err := h.userService.GetUnverifiedUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.AuthErrorResponse{
			Error:      err.Error(),
			Success:    false,
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	usersData := make([]*dto.UserData, len(users))
	for i, user := range users {
		usersData[i] = h.userModelToDTO(user)
	}

	c.JSON(http.StatusOK, dto.GetUsersResponse{
		Success:    true,
		StatusCode: http.StatusOK,
		Data:       usersData,
		Count:      len(usersData),
	})
}

// GetUserByID returns a specific user by ID (admin only)
// @Summary Get User by ID
// @Description Retrieve a specific user by their ID (admin access required)
// @Tags Users
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserProfileResponse "User retrieved successfully"
// @Failure 400 {object} dto.AuthErrorResponse "Invalid user ID"
// @Failure 401 {object} dto.AuthErrorResponse "Unauthorized - invalid or missing token"
// @Failure 403 {object} dto.AuthErrorResponse "Forbidden - admin access required"
// @Failure 404 {object} dto.AuthErrorResponse "User not found"
// @Failure 500 {object} dto.AuthErrorResponse "Internal server error"
// @Router /api/users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.AuthErrorResponse{
			Error:      "Unauthorized",
			Success:    false,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	// Check if user is admin
	currentUser, err := h.userService.GetUserByID(userID)
	if err != nil || !currentUser.IsStaff {
		c.JSON(http.StatusForbidden, dto.AuthErrorResponse{
			Error:      "Access denied",
			Success:    false,
			StatusCode: http.StatusForbidden,
		})
		return
	}

	targetUserID := c.Param("id")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, dto.AuthErrorResponse{
			Error:      "User ID is required",
			Success:    false,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	targetUser, err := h.userService.GetUserByID(targetUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.AuthErrorResponse{
			Error:      "User not found",
			Success:    false,
			StatusCode: http.StatusNotFound,
		})
		return
	}

	userData := h.userModelToDTO(targetUser)
	c.JSON(http.StatusOK, dto.UserProfileResponse{
		Success:    true,
		StatusCode: http.StatusOK,
		Data:       userData,
	})
}

// Helper function to convert user model to DTO
func (h *UserHandler) userModelToDTO(user *userModel.User) *dto.UserData {
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
        ProfileImageURL: user.ProfileImageURL,
    }
}
