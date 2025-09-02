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
	"gopi.com/tests/mocks"
	userMocks "gopi.com/tests/mocks/user"
)

// Test setup helper
func setupUserTest(t *testing.T) (*gin.Engine, *userMocks.MockUserRepository, *mocks.MockEmailService) {
	gin.SetMode(gin.TestMode)

	// Create mock repositories
	mockUserRepo := new(userMocks.MockUserRepository)
	mockEmailService := new(mocks.MockEmailService)

	// Create services with mock repositories
	userSvc := userService.NewUserService(mockUserRepo, nil)

	// Create handler with services
	userHandler := handler.NewUserHandler(userSvc, nil)

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup routes
	protected := router.Group("/users")
	protected.Use(func(c *gin.Context) {
		// Mock auth middleware - set user_id in context
		c.Set("user_id", "test-user-id")
		c.Next()
	})

	// User routes
	protected.GET("/profile", userHandler.GetUserProfile)
	protected.PUT("/profile", userHandler.UpdateUserProfile)
	protected.GET("", userHandler.GetAllUsers)
	protected.GET("/staff", userHandler.GetStaffUsers)
	protected.GET("/verified", userHandler.GetVerifiedUsers)
	protected.GET("/unverified", userHandler.GetUnverifiedUsers)
	protected.GET("/:id", userHandler.GetUserByID)

	return router, mockUserRepo, mockEmailService
}

func TestUserHandler_GetUserProfile(t *testing.T) {
	router, mockUserRepo, _ := setupUserTest(t)

	tests := []struct {
		name           string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful profile retrieval",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedUser := &userModel.User{
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
					IsStaff:    false,
					IsActive:   true,
					IsVerified: true,
					DateJoined: time.Now(),
				}

				mockUserRepo.On("GetByID", "test-user-id").Return(expectedUser, nil)
			},
		},
		{
			name:           "user not found",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockUserRepo.On("GetByID", "test-user-id").Return(nil, errors.New("user not found"))
			},
		},
		{
			name:           "unauthorized - no user_id in context",
			expectedStatus: http.StatusUnauthorized,
			mockSetup: func() {
				// No setup needed - test without auth middleware
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "unauthorized - no user_id in context" {
				// Test without auth middleware
				router := gin.New()
				router.Use(gin.Recovery())
				userHandler := &handler.UserHandler{}

				router.GET("/profile", userHandler.GetUserProfile)

				req, _ := http.NewRequest(http.MethodGet, "/profile", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.expectedStatus, w.Code)
			} else {
				// Clear previous expectations
				mockUserRepo.ExpectedCalls = nil

				tt.mockSetup()

				req, _ := http.NewRequest(http.MethodGet, "/users/profile", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.expectedStatus, w.Code)
				mockUserRepo.AssertExpectations(t)
			}
		})
	}
}

func TestUserHandler_UpdateUserProfile(t *testing.T) {
	router, mockUserRepo, _ := setupUserTest(t)

	tests := []struct {
		name           string
		requestBody    dto.UpdateUserRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful profile update",
			requestBody: dto.UpdateUserRequest{
				FirstName: "Updated",
				LastName:  "Name",
				Username:  "updateduser",
				Height:    180.0,
				Weight:    75.0,
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				existingUser := &userModel.User{
					Base: model.Base{
						ID:        "test-user-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:  "testuser",
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
					Height:    175.0,
					Weight:    70.0,
				}

				mockUserRepo.On("GetByID", "test-user-id").Return(existingUser, nil)
				mockUserRepo.On("UsernameExists", "updateduser").Return(false, nil)
				mockUserRepo.On("Update", mock.MatchedBy(func(u *userModel.User) bool {
					return u.ID == "test-user-id" &&
						u.FirstName == "Updated" &&
						u.LastName == "Name" &&
						u.Username == "updateduser" &&
						u.Height == 180.0 &&
						u.Weight == 75.0
				})).Return(nil)
			},
		},
		{
			name: "partial update - only first name",
			requestBody: dto.UpdateUserRequest{
				FirstName: "NewFirstName",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				existingUser := &userModel.User{
					Base: model.Base{
						ID:        "test-user-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:  "testuser",
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
					Height:    175.0,
					Weight:    70.0,
				}

				mockUserRepo.On("GetByID", "test-user-id").Return(existingUser, nil)
				mockUserRepo.On("Update", mock.MatchedBy(func(u *userModel.User) bool {
					return u.ID == "test-user-id" &&
						u.FirstName == "NewFirstName" &&
						u.LastName == "User" // unchanged
				})).Return(nil)
			},
		},
		{
			name: "username already taken",
			requestBody: dto.UpdateUserRequest{
				Username: "existinguser",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				existingUser := &userModel.User{
					Base: model.Base{
						ID:        "test-user-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username: "testuser",
					Email:    "test@example.com",
				}

				mockUserRepo.On("GetByID", "test-user-id").Return(existingUser, nil)
				mockUserRepo.On("UsernameExists", "existinguser").Return(true, nil)
			},
		},
		{
			name: "user not found",
			requestBody: dto.UpdateUserRequest{
				FirstName: "Updated",
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockUserRepo.On("GetByID", "test-user-id").Return(nil, errors.New("user not found"))
			},
		},
		{
			name: "invalid request body",
			requestBody: dto.UpdateUserRequest{
				Height: -10, // Invalid height
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// No mocks needed - validation happens before service calls
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/users/profile", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetAllUsers(t *testing.T) {
	router, mockUserRepo, _ := setupUserTest(t)

	// Mock admin user
	adminUser := &userModel.User{
		Base: model.Base{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username: "admin",
		Email:    "admin@example.com",
		IsStaff:  true,
	}
	mockUserRepo.On("GetByID", "test-user-id").Return(adminUser, nil)

	tests := []struct {
		name           string
		expectedStatus int
		mockSetup      func()
		expectedCount  int
	}{
		{
			name:           "successful get all users",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			mockSetup: func() {
				users := []*userModel.User{
					{
						Base: model.Base{
							ID:        "user1",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Username:   "user1",
						Email:      "user1@example.com",
						FirstName:  "User",
						LastName:   "One",
						IsStaff:    false,
						IsVerified: true,
					},
					{
						Base: model.Base{
							ID:        "user2",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Username:   "user2",
						Email:      "user2@example.com",
						FirstName:  "User",
						LastName:   "Two",
						IsStaff:    true,
						IsVerified: false,
					},
				}

				mockUserRepo.On("GetAllUsers").Return(users, nil)
			},
		},
		{
			name:           "access denied - not admin",
			expectedStatus: http.StatusForbidden,
			mockSetup: func() {
				// No additional setup needed - user authentication is handled in test loop
			},
		},
		{
			name:           "repository error",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				mockUserRepo.On("GetAllUsers").Return(nil, errors.New("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all previous expectations
			mockUserRepo.ExpectedCalls = nil

			// Set up user authentication - different users for different test cases
			if tt.name == "access denied - not admin" {
				// Non-admin user for access denied test
				nonAdminUser := &userModel.User{
					Base: model.Base{
						ID: "test-user-id",
					},
					Username: "regularuser",
					IsStaff:  false,
				}
				mockUserRepo.On("GetByID", "test-user-id").Return(nonAdminUser, nil)
			} else {
				// Admin user for successful and error test cases
				adminUser := &userModel.User{
					Base: model.Base{
						ID: "test-user-id",
					},
					Username: "adminuser",
					IsStaff:  true,
				}
				mockUserRepo.On("GetByID", "test-user-id").Return(adminUser, nil)
			}

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodGet, "/users", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK && tt.expectedCount > 0 {
				var response dto.GetUsersResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.Data, tt.expectedCount)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUserByID(t *testing.T) {
	router, mockUserRepo, _ := setupUserTest(t)

	// Mock admin user
	adminUser := &userModel.User{
		Base: model.Base{
			ID: "test-user-id",
		},
		IsStaff: true,
	}
	mockUserRepo.On("GetByID", "test-user-id").Return(adminUser, nil)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful get user by id",
			userID:         "target-user-id",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				targetUser := &userModel.User{
					Base: model.Base{
						ID:        "target-user-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:   "targetuser",
					Email:      "target@example.com",
					FirstName:  "Target",
					LastName:   "User",
					IsStaff:    false,
					IsVerified: true,
				}

				mockUserRepo.On("GetByID", "target-user-id").Return(targetUser, nil)
			},
		},
		{
			name:           "user not found",
			userID:         "nonexistent",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockUserRepo.On("GetByID", "nonexistent").Return(nil, errors.New("user not found"))
			},
		},
		{
			name:           "access denied - not admin",
			userID:         "target-user-id",
			expectedStatus: http.StatusForbidden,
			mockSetup: func() {
				// No additional setup needed - user authentication is handled in test loop
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all previous expectations
			mockUserRepo.ExpectedCalls = nil

			// Set up user authentication - different users for different test cases
			if tt.name == "access denied - not admin" {
				// Non-admin user for access denied test
				nonAdminUser := &userModel.User{
					Base: model.Base{
						ID: "test-user-id",
					},
					Username: "regularuser",
					IsStaff:  false,
				}
				mockUserRepo.On("GetByID", "test-user-id").Return(nonAdminUser, nil)
			} else {
				// Admin user for successful and error test cases
				adminUser := &userModel.User{
					Base: model.Base{
						ID: "test-user-id",
					},
					Username: "adminuser",
					IsStaff:  true,
				}
				mockUserRepo.On("GetByID", "test-user-id").Return(adminUser, nil)
			}

			tt.mockSetup()

			url := "/users/" + tt.userID

			req, _ := http.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetStaffUsers(t *testing.T) {
	router, mockUserRepo, _ := setupUserTest(t)

	// Mock admin user
	adminUser := &userModel.User{
		Base: model.Base{
			ID: "test-user-id",
		},
		IsStaff: true,
	}
	mockUserRepo.On("GetByID", "test-user-id").Return(adminUser, nil)

	tests := []struct {
		name           string
		expectedStatus int
		mockSetup      func()
		expectedCount  int
	}{
		{
			name:           "successful get staff users",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			mockSetup: func() {
				staffUsers := []*userModel.User{
					{
						Base: model.Base{
							ID:        "staff1",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Username:   "staff1",
						Email:      "staff1@example.com",
						FirstName:  "Staff",
						LastName:   "One",
						IsStaff:    true,
						IsVerified: true,
					},
					{
						Base: model.Base{
							ID:        "staff2",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Username:   "staff2",
						Email:      "staff2@example.com",
						FirstName:  "Staff",
						LastName:   "Two",
						IsStaff:    true,
						IsVerified: true,
					},
				}

				mockUserRepo.On("GetStaffUsers").Return(staffUsers, nil)
			},
		},
		{
			name:           "empty staff users list",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			mockSetup: func() {
				mockUserRepo.On("GetStaffUsers").Return([]*userModel.User{}, nil)
			},
		},
		{
			name:           "access denied - not admin",
			expectedStatus: http.StatusForbidden,
			mockSetup: func() {
				nonAdminUser := &userModel.User{
					Base: model.Base{
						ID: "test-user-id",
					},
					IsStaff: false,
				}
				mockUserRepo.ExpectedCalls = nil // Clear previous calls
				mockUserRepo.On("GetByID", "test-user-id").Return(nonAdminUser, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations except for admin check
			if tt.name != "access denied - not admin" {
				mockUserRepo.ExpectedCalls = mockUserRepo.ExpectedCalls[:1]
			}

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodGet, "/users/staff", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response dto.GetUsersResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.Data, tt.expectedCount)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetVerifiedUsers(t *testing.T) {
	router, mockUserRepo, _ := setupUserTest(t)

	// Mock admin user
	adminUser := &userModel.User{
		Base: model.Base{
			ID: "test-user-id",
		},
		IsStaff: true,
	}
	mockUserRepo.On("GetByID", "test-user-id").Return(adminUser, nil)

	t.Run("successful get verified users", func(t *testing.T) {
		verifiedUsers := []*userModel.User{
			{
				Base: model.Base{
					ID:        "verified1",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Username:   "verified1",
				Email:      "verified1@example.com",
				FirstName:  "Verified",
				LastName:   "One",
				IsVerified: true,
			},
		}

		mockUserRepo.ExpectedCalls = mockUserRepo.ExpectedCalls[:1] // Keep admin check
		mockUserRepo.On("GetVerifiedUsers").Return(verifiedUsers, nil)

		req, _ := http.NewRequest(http.MethodGet, "/users/verified", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.GetUsersResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 1)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserHandler_GetUnverifiedUsers(t *testing.T) {
	router, mockUserRepo, _ := setupUserTest(t)

	// Mock admin user
	adminUser := &userModel.User{
		Base: model.Base{
			ID: "test-user-id",
		},
		IsStaff: true,
	}
	mockUserRepo.On("GetByID", "test-user-id").Return(adminUser, nil)

	t.Run("successful get unverified users", func(t *testing.T) {
		unverifiedUsers := []*userModel.User{
			{
				Base: model.Base{
					ID:        "unverified1",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Username:   "unverified1",
				Email:      "unverified1@example.com",
				FirstName:  "Unverified",
				LastName:   "One",
				IsVerified: false,
			},
		}

		mockUserRepo.ExpectedCalls = nil
		mockUserRepo.On("GetByID", "test-user-id").Return(adminUser, nil)
		mockUserRepo.On("GetUnverifiedUsers").Return(unverifiedUsers, nil)

		req, _ := http.NewRequest(http.MethodGet, "/users/unverified", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.GetUsersResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 1)

		mockUserRepo.AssertExpectations(t)
	})
}
