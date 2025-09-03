package user_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopi.com/api/http/dto"
	"gopi.com/api/http/handler"
	userService "gopi.com/internal/app/user"
	"gopi.com/internal/domain/model"
	userModel "gopi.com/internal/domain/user/model"
	"gopi.com/internal/lib/jwt"

	"gopi.com/tests/mocks"
)

// Mock JWT Service for testing
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateTokenPair(user *userModel.User) (*jwt.TokenPair, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.TokenPair), args.Error(1)
}

func (m *MockJWTService) ValidateToken(tokenString string) (*jwt.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Claims), args.Error(1)
}

func (m *MockJWTService) RefreshToken(refreshToken string, user *userModel.User) (string, error) {
	args := m.Called(refreshToken, user)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) BlacklistToken(tokenString string) error {
	args := m.Called(tokenString)
	return args.Error(0)
}

func (m *MockJWTService) IsTokenBlacklisted(tokenString string) bool {
	args := m.Called(tokenString)
	return args.Bool(0)
}

func (m *MockJWTService) ExtractTokenFromHeader(authHeader string) (string, error) {
	args := m.Called(authHeader)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) GetUserFromToken(tokenString string) (*userModel.User, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

// Test setup helper for auth endpoints
func setupAuthTest(t *testing.T) (*gin.Engine, *mocks.MockUserRepository, *MockEmailService, jwt.JWTServiceInterface) {
	gin.SetMode(gin.TestMode)

	// Create mock repositories
	mockUserRepo := new(mocks.MockUserRepository)
	mockEmailService := new(MockEmailService)
	mockJWTService := new(MockJWTService)

	// Create services with mock repositories
	userSvc := userService.NewUserService(mockUserRepo, mockEmailService)

	// Create handler with services
	authHandler := handler.NewAuthHandler(userSvc, mockJWTService)

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup auth routes (following the pattern from auth_routes.go)
	auth := router.Group("/api/auth")
	{
		// User registration
		auth.POST("/register/", authHandler.UserRegister)

		// User login
		auth.POST("/login/", authHandler.UserLogin)

		// User logout - requires authentication
		auth.GET("/logout/", func(c *gin.Context) {
			// Mock auth middleware - set user_id in context
			c.Set("user_id", "test-user-id")
			c.Next()
		}, authHandler.UserLogout)

		// OTP verification
		auth.POST("/verify/", authHandler.VerifyOTP)

		// Delete account - requires authentication
		auth.DELETE("/delete/", func(c *gin.Context) {
			// Mock auth middleware - set user_id in context
			c.Set("user_id", "test-user-id")
			c.Next()
		}, authHandler.DeleteAccount)

		// Change password - requires authentication
		auth.PUT("/change-password/", func(c *gin.Context) {
			// Mock auth middleware - set user_id in context
			c.Set("user_id", "test-user-id")
			c.Next()
		}, authHandler.ChangePassword)

		// Resend OTP
		auth.PUT("/resend-otp/:id/", authHandler.ResendOTP)
	}

	return router, mockUserRepo, mockEmailService, mockJWTService
}

func TestAuthHandler_UserRegister(t *testing.T) {
	router, mockUserRepo, mockEmailService, jwtService := setupAuthTest(t)
	mockJWTService := jwtService.(*MockJWTService)

	tests := []struct {
		name           string
		requestBody    dto.RegistrationRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful user registration",
			requestBody: dto.RegistrationRequest{
				Username:  "testuser",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
				Password:  "password123",
				Height:    175.0,
				Weight:    70.0,
			},
			expectedStatus: http.StatusCreated,
			mockSetup: func() {
				// Create mock user that will be returned by RegisterUser
				mockUser := &userModel.User{
					Base: model.Base{
						ID:        "test-user-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:   "testuser",
					Email:      "test@example.com",
					FirstName:  "Test",
					LastName:   "User",
					Height:     175.0,
					Weight:     70.0,
					IsVerified: false,
					DateJoined: time.Now(),
				}

				// Mock validation checks
				mockUserRepo.On("EmailExists", "test@example.com").Return(false, nil)
				mockUserRepo.On("UsernameExists", "testuser").Return(false, nil)

				// Mock user creation - return the mock user
				mockUserRepo.On("Create", mock.Anything).Run(func(args mock.Arguments) {
					// This will be called with the user created in RegisterUser
				}).Return(nil)

				// Mock GetByID for ResendOTP (called after user registration)
				mockUserRepo.On("GetByID", mock.AnythingOfType("string")).Return(mockUser, nil)

				// Mock UpdateOTP for ResendOTP
				mockUserRepo.On("UpdateOTP", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)

				// Mock email sending (for initial registration and resend)
				mockEmailService.On("SendOTPEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)

			},
		},
		{
			name: "email already exists",
			requestBody: dto.RegistrationRequest{
				Username:  "testuser",
				Email:     "existing@example.com",
				FirstName: "Test",
				LastName:  "User",
				Password:  "password123",
				Height:    175.0,
				Weight:    70.0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				mockUserRepo.On("EmailExists", "existing@example.com").Return(true, nil)
			},
		},
		{
			name: "username already exists",
			requestBody: dto.RegistrationRequest{
				Username:  "existinguser",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
				Password:  "password123",
				Height:    175.0,
				Weight:    70.0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				mockUserRepo.On("EmailExists", "test@example.com").Return(false, nil)
				mockUserRepo.On("UsernameExists", "existinguser").Return(true, nil)
			},
		},
		{
			name: "invalid request body - missing email",
			requestBody: dto.RegistrationRequest{
				Username:  "testuser",
				FirstName: "Test",
				LastName:  "User",
				Password:  "password123",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// No mocks needed - validation happens before service calls
			},
		},
		{
			name: "invalid request body - short password",
			requestBody: dto.RegistrationRequest{
				Username:  "testuser",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
				Password:  "123", // Too short
				Height:    175.0,
				Weight:    70.0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// No mocks needed - validation happens before service calls
			},
		},
		{
			name: "repository error on user creation",
			requestBody: dto.RegistrationRequest{
				Username:  "testuser",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
				Password:  "password123",
				Height:    175.0,
				Weight:    70.0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				mockUserRepo.On("EmailExists", "test@example.com").Return(false, nil)
				mockUserRepo.On("UsernameExists", "testuser").Return(false, nil)
				mockUserRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all mock expectations first
			mockUserRepo.ExpectedCalls = nil
			mockEmailService.ExpectedCalls = nil
			mockJWTService.ExpectedCalls = nil

			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/api/auth/register/", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockUserRepo.AssertExpectations(t)
			mockEmailService.AssertExpectations(t)
			mockJWTService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_UserLogin(t *testing.T) {
	router, mockUserRepo, mockEmailService, jwtService := setupAuthTest(t)
	mockJWTService := jwtService.(*MockJWTService)

	tests := []struct {
		name           string
		requestBody    dto.LoginRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful login",
			requestBody: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				// Create a temporary service to hash the password
				tempUserSvc := userService.NewUserService(mockUserRepo, mockEmailService)
				hashedPassword, _ := tempUserSvc.HashPassword("password123")

				// Create expected user
				expectedUser := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:   "testuser",
					Email:      "test@example.com",
					Password:   hashedPassword,
					IsVerified: true,
					IsActive:   true,
				}

				mockUserRepo.On("GetByEmail", "test@example.com").Return(expectedUser, nil)
				mockUserRepo.On("UpdateLastLogin", "user123").Return(nil)

				// Mock JWT token generation
				expectedTokenPair := &jwt.TokenPair{
					AccessToken:  "mock-access-token",
					RefreshToken: "mock-refresh-token",
					ExpiresIn:    3600,
				}
				mockJWTService.On("GenerateTokenPair", expectedUser).Return(expectedTokenPair, nil)
			},
		},
		{
			name: "invalid credentials - wrong password",
			requestBody: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusForbidden,
			mockSetup: func() {
				hashedPassword, _ := userService.NewUserService(mockUserRepo, nil).HashPassword("password123")
				user := &userModel.User{
					Base: model.Base{
						ID: "user123",
					},
					Email:    "test@example.com",
					Password: hashedPassword,
				}

				mockUserRepo.On("GetByEmail", "test@example.com").Return(user, nil)
			},
		},
		{
			name: "user not found",
			requestBody: dto.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusForbidden,
			mockSetup: func() {
				mockUserRepo.On("GetByEmail", "nonexistent@example.com").Return(nil, errors.New("user not found"))
			},
		},
		{
			name: "user not verified",
			requestBody: dto.LoginRequest{
				Email:    "unverified@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			mockSetup: func() {
				hashedPassword, _ := userService.NewUserService(mockUserRepo, nil).HashPassword("password123")
				user := &userModel.User{
					Base: model.Base{
						ID: "user123",
					},
					Email:      "unverified@example.com",
					Password:   hashedPassword,
					IsVerified: false,
					IsActive:   true,
				}

				mockUserRepo.On("GetByEmail", "unverified@example.com").Return(user, nil)
			},
		},
		{
			name: "user not active",
			requestBody: dto.LoginRequest{
				Email:    "inactive@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusForbidden,
			mockSetup: func() {
				hashedPassword, _ := userService.NewUserService(mockUserRepo, nil).HashPassword("password123")
				user := &userModel.User{
					Base: model.Base{
						ID: "user123",
					},
					Email:      "inactive@example.com",
					Password:   hashedPassword,
					IsVerified: true,
					IsActive:   false,
				}

				mockUserRepo.On("GetByEmail", "inactive@example.com").Return(user, nil)
			},
		},
		{
			name: "invalid request body",
			requestBody: dto.LoginRequest{
				Email: "", // Missing email
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// No mocks needed - validation happens before service calls
			},
		},
		{
			name: "JWT token generation error",
			requestBody: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				// Create a temporary service to hash the password
				tempUserSvc := userService.NewUserService(mockUserRepo, mockEmailService)
				hashedPassword, _ := tempUserSvc.HashPassword("password123")

				expectedUser := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:   "testuser",
					Email:      "test@example.com",
					Password:   hashedPassword,
					IsVerified: true,
					IsActive:   true,
				}

				mockUserRepo.On("GetByEmail", "test@example.com").Return(expectedUser, nil)
				mockUserRepo.On("UpdateLastLogin", "user123").Return(nil)
				mockJWTService.On("GenerateTokenPair", expectedUser).Return(nil, errors.New("JWT generation failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockUserRepo.ExpectedCalls = nil
			mockEmailService.ExpectedCalls = nil
			mockJWTService.ExpectedCalls = nil

			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/api/auth/login/", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockJWTService.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_UserLogout(t *testing.T) {
	router, _, _, jwtService := setupAuthTest(t)
	mockJWTService := jwtService.(*MockJWTService)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful logout with token",
			authHeader:     "Bearer mock-jwt-token",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				mockJWTService.On("BlacklistToken", "mock-jwt-token").Return(nil)
			},
		},
		{
			name:           "successful logout without token",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				// No JWT calls expected when no token provided
			},
		},
		{
			name:           "logout with invalid auth header format",
			authHeader:     "InvalidFormat mock-jwt-token",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				// No JWT calls expected since header format is invalid
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockJWTService.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodGet, "/api/auth/logout/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockJWTService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_VerifyOTP(t *testing.T) {
	router, mockUserRepo, mockEmailService, _ := setupAuthTest(t)

	tests := []struct {
		name           string
		requestBody    dto.VerifyOTPRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful OTP verification",
			requestBody: dto.VerifyOTPRequest{
				Email: "test@example.com",
				OTP:   "123456",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				user := &userModel.User{
					Base: model.Base{
						ID: "user123",
					},
					Email:      "test@example.com",
					FirstName:  "Test",
					IsVerified: false,
				}
				mockUserRepo.On("GetByOTP", "test@example.com", "123456").Return(user, nil)
				mockUserRepo.On("MarkAsVerified", "user123").Return(nil)
				mockEmailService.On("SendWelcomeEmail", "test@example.com", "Test").Return(nil)
			},
		},
		{
			name: "user already verified",
			requestBody: dto.VerifyOTPRequest{
				Email: "verified@example.com",
				OTP:   "123456",
			},
			expectedStatus: http.StatusForbidden,
			mockSetup: func() {
				mockUserRepo.On("GetByOTP", "verified@example.com", "123456").Return(&userModel.User{
					Base: model.Base{
						ID: "user123",
					},
					Email:      "verified@example.com",
					IsVerified: true,
				}, nil)
			},
		},
		{
			name: "invalid OTP or user not found",
			requestBody: dto.VerifyOTPRequest{
				Email: "test@example.com",
				OTP:   "wrongotp",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				mockUserRepo.On("GetByOTP", mock.Anything, mock.Anything).Return(nil, errors.New("user not found or incorrect OTP"))
			},
		},
		{
			name: "invalid request body",
			requestBody: dto.VerifyOTPRequest{
				Email: "", // Missing email
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// No mocks needed - validation happens before service calls
			},
		},
		{
			name: "mark as verified error",
			requestBody: dto.VerifyOTPRequest{
				Email: "test@example.com",
				OTP:   "123456",
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockUserRepo.On("GetByOTP", "test@example.com", "123456").Return(&userModel.User{
					Base: model.Base{
						ID: "user123",
					},
					Email:      "test@example.com",
					IsVerified: false,
				}, nil)
				mockUserRepo.On("MarkAsVerified", "user123").Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockUserRepo.ExpectedCalls = nil
			mockEmailService.ExpectedCalls = nil

			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/api/auth/verify/", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockUserRepo.AssertExpectations(t)
			mockEmailService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_ChangePassword(t *testing.T) {
	router, mockUserRepo, mockEmailService, jwtService := setupAuthTest(t)
	mockJWTService := jwtService.(*MockJWTService)

	tests := []struct {
		name           string
		requestBody    dto.ChangePasswordRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful password change",
			requestBody: dto.ChangePasswordRequest{
				OldPassword: "oldpassword",
				NewPassword: "newpassword123",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				// Create user with hashed old password
				oldHashedPassword, _ := userService.NewUserService(mockUserRepo, mockEmailService).HashPassword("oldpassword")
				user := &userModel.User{
					Base:     model.Base{ID: "test-user-id"},
					Password: oldHashedPassword,
				}
				mockUserRepo.On("GetByID", "test-user-id").Return(user, nil)
				mockUserRepo.On("UpdatePassword", "test-user-id", mock.AnythingOfType("string")).Return(nil)
			},
		},
		{
			name: "incorrect old password",
			requestBody: dto.ChangePasswordRequest{
				OldPassword: "wrongpassword",
				NewPassword: "newpassword123",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// Create user with hashed old password
				oldHashedPassword, _ := userService.NewUserService(mockUserRepo, mockEmailService).HashPassword("oldpassword")
				user := &userModel.User{
					Base:     model.Base{ID: "test-user-id"},
					Password: oldHashedPassword,
				}
				mockUserRepo.On("GetByID", "test-user-id").Return(user, nil)
				// No UpdatePassword call expected since password check will fail
			},
		},
		{
			name: "invalid request body",
			requestBody: dto.ChangePasswordRequest{
				OldPassword: "", // Missing old password
				NewPassword: "newpassword123",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// No mocks needed - validation happens before service calls
			},
		},
		{
			name: "repository error",
			requestBody: dto.ChangePasswordRequest{
				OldPassword: "oldpassword",
				NewPassword: "newpassword123",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// Create user with hashed old password
				oldHashedPassword, _ := userService.NewUserService(mockUserRepo, mockEmailService).HashPassword("oldpassword")
				user := &userModel.User{
					Base:     model.Base{ID: "test-user-id"},
					Password: oldHashedPassword,
				}
				mockUserRepo.On("GetByID", "test-user-id").Return(user, nil)
				mockUserRepo.On("UpdatePassword", "test-user-id", mock.AnythingOfType("string")).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockUserRepo.ExpectedCalls = nil
			mockEmailService.ExpectedCalls = nil

			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/api/auth/change-password/", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockJWTService.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_DeleteAccount(t *testing.T) {
	router, mockUserRepo, _, jwtService := setupAuthTest(t)
	mockJWTService := jwtService.(*MockJWTService)

	tests := []struct {
		name           string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful account deletion",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				mockUserRepo.On("Delete", "test-user-id").Return(nil)
			},
		},
		{
			name:           "repository error",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				mockUserRepo.On("Delete", "test-user-id").Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodDelete, "/api/auth/delete/", nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockJWTService.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_ResendOTP(t *testing.T) {
	router, mockUserRepo, mockEmailService, jwtService := setupAuthTest(t)
	mockJWTService := jwtService.(*MockJWTService)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful OTP resend",
			userID:         "user123",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				user := &userModel.User{
					Base: model.Base{
						ID: "user123",
					},
					Email:     "test@example.com",
					FirstName: "Test",
				}
				mockUserRepo.On("GetByID", "user123").Return(user, nil)
				mockUserRepo.On("UpdateOTP", "user123", mock.AnythingOfType("string")).Return(nil)
				mockEmailService.On("SendOTPEmail", "test@example.com", "Test", mock.AnythingOfType("string")).Return(nil)
			},
		},
		{
			name:           "user not found",
			userID:         "nonexistent",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				mockUserRepo.On("GetByID", "nonexistent").Return(nil, errors.New("user not found"))
			},
		},
		{
			name:           "missing user ID",
			userID:         "",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// No mocks needed - validation happens before service calls
			},
		},
		{
			name:           "resend OTP error",
			userID:         "user123",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				user := &userModel.User{
					Base: model.Base{
						ID: "user123",
					},
					Email:     "test@example.com",
					FirstName: "Test",
				}
				mockUserRepo.On("GetByID", "user123").Return(user, nil)
				mockUserRepo.On("UpdateOTP", "user123", mock.AnythingOfType("string")).Return(errors.New("resend error"))
			},
		},
		{
			name:           "get user error after successful resend",
			userID:         "user123",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				mockUserRepo.On("GetByID", "user123").Return(nil, errors.New("user not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockUserRepo.ExpectedCalls = nil
			mockEmailService.ExpectedCalls = nil

			tt.mockSetup()

			url := "/api/auth/resend-otp/" + tt.userID + "/"
			if tt.userID == "" {
				url = "/api/auth/resend-otp//"
			}

			req, _ := http.NewRequest(http.MethodPut, url, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockUserRepo.AssertExpectations(t)
			mockEmailService.AssertExpectations(t)
			mockJWTService.AssertExpectations(t)
		})
	}
}
