package user_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	userService "gopi.com/internal/app/user"
	"gopi.com/internal/domain/model"
	userModel "gopi.com/internal/domain/user/model"
	userMocks "gopi.com/tests/mocks/user"
)

// Mock Email Service for testing
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendOTPEmail(email, firstName, otp string) error {
	args := m.Called(email, firstName, otp)
	return args.Error(0)
}

func (m *MockEmailService) SendWelcomeEmail(email, firstName string) error {
	args := m.Called(email, firstName)
	return args.Error(0)
}

func (m *MockEmailService) SendPasswordResetEmail(email, resetLink string) error {
	args := m.Called(email, resetLink)
	return args.Error(0)
}

func (m *MockEmailService) SendApologyEmail(email, username string) error {
	args := m.Called(email, username)
	return args.Error(0)
}

func (m *MockEmailService) SendBulkEmail(emails []string, subject, htmlContent string) error {
	args := m.Called(emails, subject, htmlContent)
	return args.Error(0)
}

func (m *MockEmailService) TestEmailConnection() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEmailService) GetQueueLength() int {
	args := m.Called()
	return args.Int(0)
}

func TestUserService_RegisterUser(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	mockEmailService := new(MockEmailService)
	userSvc := userService.NewUserService(mockUserRepo, mockEmailService)

	tests := []struct {
		name              string
		username          string
		email             string
		firstName         string
		lastName          string
		password          string
		height            float64
		weight            float64
		expectedErr       error
		expectedErrString string
		mockSetup         func()
	}{
		{
			name:        "successful user registration",
			username:    "testuser",
			email:       "test@example.com",
			firstName:   "Test",
			lastName:    "User",
			password:    "password123",
			height:      175.0,
			weight:      70.0,
			expectedErr: nil,
			mockSetup: func() {
				mockUserRepo.On("EmailExists", "test@example.com").Return(false, nil)
				mockUserRepo.On("UsernameExists", "testuser").Return(false, nil)
				mockUserRepo.On("Create", mock.MatchedBy(func(u *userModel.User) bool {
					return u.Username == "testuser" &&
						u.Email == "test@example.com" &&
						u.FirstName == "Test" &&
						u.LastName == "User" &&
						u.Height == 175.0 &&
						u.Weight == 70.0 &&
						u.IsActive == true &&
						u.IsVerified == false &&
						u.IsStaff == false
				})).Return(nil)
				mockEmailService.On("SendOTPEmail", "test@example.com", "Test", mock.AnythingOfType("string")).Return(nil)
			},
		},
		{
			name:              "email already exists",
			username:          "testuser",
			email:             "existing@example.com",
			firstName:         "Test",
			lastName:          "User",
			password:          "password123",
			height:            175.0,
			weight:            70.0,
			expectedErrString: "the email has already been taken",
			mockSetup: func() {
				mockUserRepo.On("EmailExists", "existing@example.com").Return(true, nil)
			},
		},
		{
			name:              "username already exists",
			username:          "existinguser",
			email:             "test@example.com",
			firstName:         "Test",
			lastName:          "User",
			password:          "password123",
			height:            175.0,
			weight:            70.0,
			expectedErrString: "the username has already been taken",
			mockSetup: func() {
				mockUserRepo.On("EmailExists", "test@example.com").Return(false, nil)
				mockUserRepo.On("UsernameExists", "existinguser").Return(true, nil)
			},
		},
		{
			name:              "email exists check error",
			username:          "testuser",
			email:             "test@example.com",
			firstName:         "Test",
			lastName:          "User",
			password:          "password123",
			height:            175.0,
			weight:            70.0,
			expectedErrString: "repository error",
			mockSetup: func() {
				mockUserRepo.On("EmailExists", "test@example.com").Return(false, errors.New("repository error"))
			},
		},
		{
			name:              "create user error",
			username:          "testuser",
			email:             "test@example.com",
			firstName:         "Test",
			lastName:          "User",
			password:          "password123",
			height:            175.0,
			weight:            70.0,
			expectedErrString: "repository error",
			mockSetup: func() {
				mockUserRepo.On("EmailExists", "test@example.com").Return(false, nil)
				mockUserRepo.On("UsernameExists", "testuser").Return(false, nil)
				mockUserRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockUserRepo.ExpectedCalls = nil
			mockEmailService.ExpectedCalls = nil

			tt.mockSetup()

			user, err := userSvc.RegisterUser(tt.username, tt.email, tt.firstName, tt.lastName, tt.password, tt.height, tt.weight)

			if tt.expectedErrString != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrString)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.username, user.Username)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.firstName, user.FirstName)
				assert.Equal(t, tt.lastName, user.LastName)
			}

			mockUserRepo.AssertExpectations(t)
			mockEmailService.AssertExpectations(t)
		})
	}
}

func TestUserService_LoginUser(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	tests := []struct {
		name              string
		email             string
		password          string
		expectedErrString string
		mockSetup         func()
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func() {
				hashedPassword, _ := userSvc.HashPassword("password123")
				user := &userModel.User{
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

				mockUserRepo.On("GetByEmail", "test@example.com").Return(user, nil)
				mockUserRepo.On("UpdateLastLogin", "user123").Return(nil)
			},
		},
		{
			name:              "user not found",
			email:             "nonexistent@example.com",
			password:          "password123",
			expectedErrString: "invalid user",
			mockSetup: func() {
				mockUserRepo.On("GetByEmail", "nonexistent@example.com").Return(nil, errors.New("user not found"))
			},
		},
		{
			name:              "incorrect password",
			email:             "test@example.com",
			password:          "wrongpassword",
			expectedErrString: "incorrect login credentials",
			mockSetup: func() {
				hashedPassword, _ := userSvc.HashPassword("password123")
				user := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Email:    "test@example.com",
					Password: hashedPassword,
				}

				mockUserRepo.On("GetByEmail", "test@example.com").Return(user, nil)
			},
		},
		{
			name:              "user not verified",
			email:             "test@example.com",
			password:          "password123",
			expectedErrString: "user's email is not verified",
			mockSetup: func() {
				hashedPassword, _ := userSvc.HashPassword("password123")
				user := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Email:      "test@example.com",
					Password:   hashedPassword,
					IsVerified: false,
					IsActive:   true,
				}

				mockUserRepo.On("GetByEmail", "test@example.com").Return(user, nil)
			},
		},
		{
			name:              "user not active",
			email:             "test@example.com",
			password:          "password123",
			expectedErrString: "user not active",
			mockSetup: func() {
				hashedPassword, _ := userSvc.HashPassword("password123")
				user := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Email:      "test@example.com",
					Password:   hashedPassword,
					IsVerified: true,
					IsActive:   false,
				}

				mockUserRepo.On("GetByEmail", "test@example.com").Return(user, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			user, err := userSvc.LoginUser(tt.email, tt.password)

			if tt.expectedErrString != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrString)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.NotNil(t, user.LastLogin)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_VerifyOTP(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	mockEmailService := new(MockEmailService)
	userSvc := userService.NewUserService(mockUserRepo, mockEmailService)

	tests := []struct {
		name              string
		email             string
		otp               string
		expectedErrString string
		mockSetup         func()
	}{
		{
			name:  "successful OTP verification",
			email: "test@example.com",
			otp:   "123456",
			mockSetup: func() {
				user := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
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
			name:              "user not found or incorrect OTP",
			email:             "test@example.com",
			otp:               "wrongotp",
			expectedErrString: "user not found or incorrect OTP",
			mockSetup: func() {
				mockUserRepo.On("GetByOTP", "test@example.com", "wrongotp").Return(nil, errors.New("not found"))
			},
		},
		{
			name:              "user already verified",
			email:             "test@example.com",
			otp:               "123456",
			expectedErrString: "user has already been verified",
			mockSetup: func() {
				user := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Email:      "test@example.com",
					IsVerified: true,
				}

				mockUserRepo.On("GetByOTP", "test@example.com", "123456").Return(user, nil)
			},
		},
		{
			name:              "mark as verified error",
			email:             "test@example.com",
			otp:               "123456",
			expectedErrString: "repository error",
			mockSetup: func() {
				user := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Email:      "test@example.com",
					IsVerified: false,
				}

				mockUserRepo.On("GetByOTP", "test@example.com", "123456").Return(user, nil)
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

			err := userSvc.VerifyOTP(tt.email, tt.otp)

			if tt.expectedErrString != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrString)
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
			mockEmailService.AssertExpectations(t)
		})
	}
}

func TestUserService_ChangePassword(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	tests := []struct {
		name              string
		userID            string
		oldPassword       string
		newPassword       string
		expectedErrString string
		mockSetup         func()
	}{
		{
			name:        "successful password change",
			userID:      "user123",
			oldPassword: "oldpassword",
			newPassword: "newpassword123",
			mockSetup: func() {
				oldHashedPassword, _ := userSvc.HashPassword("oldpassword")
				user := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Password: oldHashedPassword,
				}

				mockUserRepo.On("GetByID", "user123").Return(user, nil)
				mockUserRepo.On("UpdatePassword", "user123", mock.MatchedBy(func(pwd string) bool {
					return userSvc.CheckPassword("newpassword123", pwd)
				})).Return(nil)
			},
		},
		{
			name:              "user not found",
			userID:            "nonexistent",
			oldPassword:       "oldpassword",
			newPassword:       "newpassword123",
			expectedErrString: "user not found",
			mockSetup: func() {
				mockUserRepo.On("GetByID", "nonexistent").Return(nil, errors.New("user not found"))
			},
		},
		{
			name:              "incorrect old password",
			userID:            "user123",
			oldPassword:       "wrongpassword",
			newPassword:       "newpassword123",
			expectedErrString: "incorrect password",
			mockSetup: func() {
				oldHashedPassword, _ := userSvc.HashPassword("oldpassword")
				user := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Password: oldHashedPassword,
				}

				mockUserRepo.On("GetByID", "user123").Return(user, nil)
			},
		},
		{
			name:              "update password error",
			userID:            "user123",
			oldPassword:       "oldpassword",
			newPassword:       "newpassword123",
			expectedErrString: "repository error",
			mockSetup: func() {
				oldHashedPassword, _ := userSvc.HashPassword("oldpassword")
				user := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Password: oldHashedPassword,
				}

				mockUserRepo.On("GetByID", "user123").Return(user, nil)
				mockUserRepo.On("UpdatePassword", "user123", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			err := userSvc.ChangePassword(tt.userID, tt.oldPassword, tt.newPassword)

			if tt.expectedErrString != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrString)
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	t.Run("successful user update", func(t *testing.T) {
		// Clear previous expectations
		mockUserRepo.ExpectedCalls = nil

		user := &userModel.User{
			Base: model.Base{
				ID:        "user123",
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

		mockUserRepo.On("Update", mock.MatchedBy(func(u *userModel.User) bool {
			return u.ID == "user123"
		})).Return(nil)

		err := userSvc.UpdateUser(user)

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("update user error", func(t *testing.T) {
		// Clear previous expectations
		mockUserRepo.ExpectedCalls = nil

		user := &userModel.User{
			Base: model.Base{
				ID: "user123",
			},
		}

		mockUserRepo.On("Update", user).Return(errors.New("repository error"))

		err := userSvc.UpdateUser(user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository error")
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	t.Run("successful get user by id", func(t *testing.T) {
		// Clear previous expectations
		mockUserRepo.ExpectedCalls = nil

		expectedUser := &userModel.User{
			Base: model.Base{
				ID:        "user123",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Username: "testuser",
			Email:    "test@example.com",
		}

		mockUserRepo.On("GetByID", "user123").Return(expectedUser, nil)

		user, err := userSvc.GetUserByID("user123")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user123", user.ID)
		assert.Equal(t, "testuser", user.Username)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		// Clear previous expectations
		mockUserRepo.ExpectedCalls = nil

		mockUserRepo.On("GetByID", "nonexistent").Return(nil, errors.New("user not found"))

		user, err := userSvc.GetUserByID("nonexistent")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")

		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByEmail(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	t.Run("successful get user by email", func(t *testing.T) {
		// Clear previous expectations
		mockUserRepo.ExpectedCalls = nil

		expectedUser := &userModel.User{
			Base: model.Base{
				ID:        "user123",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Username: "testuser",
			Email:    "test@example.com",
		}

		mockUserRepo.On("GetByEmail", "test@example.com").Return(expectedUser, nil)

		user, err := userSvc.GetUserByEmail("test@example.com")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByUsername(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	t.Run("successful get user by username", func(t *testing.T) {
		// Clear previous expectations
		mockUserRepo.ExpectedCalls = nil

		expectedUser := &userModel.User{
			Base: model.Base{
				ID:        "user123",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Username: "testuser",
			Email:    "test@example.com",
		}

		mockUserRepo.On("GetByUsername", "testuser").Return(expectedUser, nil)

		user, err := userSvc.GetUserByUsername("testuser")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetAllUsers(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	t.Run("successful get all users", func(t *testing.T) {
		// Clear previous expectations
		mockUserRepo.ExpectedCalls = nil

		expectedUsers := []*userModel.User{
			{
				Base: model.Base{
					ID:        "user1",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Username: "user1",
				Email:    "user1@example.com",
			},
			{
				Base: model.Base{
					ID:        "user2",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Username: "user2",
				Email:    "user2@example.com",
			},
		}

		mockUserRepo.On("GetAllUsers").Return(expectedUsers, nil)

		users, err := userSvc.GetAllUsers()

		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "user1", users[0].Username)
		assert.Equal(t, "user2", users[1].Username)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetStaffUsers(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	t.Run("successful get staff users", func(t *testing.T) {
		// Clear previous expectations
		mockUserRepo.ExpectedCalls = nil

		staffUsers := []*userModel.User{
			{
				Base: model.Base{
					ID:        "staff1",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Username: "staff1",
				Email:    "staff1@example.com",
				IsStaff:  true,
			},
		}

		mockUserRepo.On("GetStaffUsers").Return(staffUsers, nil)

		users, err := userSvc.GetStaffUsers()

		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.True(t, users[0].IsStaff)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetVerifiedUsers(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	t.Run("successful get verified users", func(t *testing.T) {
		// Clear previous expectations
		mockUserRepo.ExpectedCalls = nil

		verifiedUsers := []*userModel.User{
			{
				Base: model.Base{
					ID:        "verified1",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Username:   "verified1",
				Email:      "verified1@example.com",
				IsVerified: true,
			},
		}

		mockUserRepo.On("GetVerifiedUsers").Return(verifiedUsers, nil)

		users, err := userSvc.GetVerifiedUsers()

		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.True(t, users[0].IsVerified)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUnverifiedUsers(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	t.Run("successful get unverified users", func(t *testing.T) {
		// Clear previous expectations
		mockUserRepo.ExpectedCalls = nil

		unverifiedUsers := []*userModel.User{
			{
				Base: model.Base{
					ID:        "unverified1",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Username:   "unverified1",
				Email:      "unverified1@example.com",
				IsVerified: false,
			},
		}

		mockUserRepo.On("GetUnverifiedUsers").Return(unverifiedUsers, nil)

		users, err := userSvc.GetUnverifiedUsers()

		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.False(t, users[0].IsVerified)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_DeleteAccount(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	t.Run("successful account deletion", func(t *testing.T) {
		// Clear previous expectations
		mockUserRepo.ExpectedCalls = nil

		mockUserRepo.On("Delete", "user123").Return(nil)

		err := userSvc.DeleteAccount("user123")

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("delete account error", func(t *testing.T) {
		// Clear previous expectations
		mockUserRepo.ExpectedCalls = nil

		mockUserRepo.On("Delete", "user123").Return(errors.New("repository error"))

		err := userSvc.DeleteAccount("user123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository error")
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserList(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	t.Run("successful get user list", func(t *testing.T) {
		// Clear previous expectations
		mockUserRepo.ExpectedCalls = nil

		expectedUsers := []*userModel.User{
			{
				Base: model.Base{
					ID:        "user1",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Username: "user1",
			},
			{
				Base: model.Base{
					ID:        "user2",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Username: "user2",
			},
		}

		mockUserRepo.On("List", 10, 0).Return(expectedUsers, nil)

		users, err := userSvc.GetUserList(10, 0)

		assert.NoError(t, err)
		assert.Len(t, users, 2)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_ValidateEmail(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	tests := []struct {
		name              string
		email             string
		expectedErrString string
		mockSetup         func()
	}{
		{
			name:  "valid email - not taken",
			email: "new@example.com",
			mockSetup: func() {
				mockUserRepo.On("EmailExists", "new@example.com").Return(false, nil)
			},
		},
		{
			name:              "email already taken",
			email:             "taken@example.com",
			expectedErrString: "the email has already been taken",
			mockSetup: func() {
				mockUserRepo.On("EmailExists", "taken@example.com").Return(true, nil)
			},
		},
		{
			name:              "email exists check error",
			email:             "error@example.com",
			expectedErrString: "repository error",
			mockSetup: func() {
				mockUserRepo.On("EmailExists", "error@example.com").Return(false, errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			err := userSvc.ValidateEmail(tt.email)

			if tt.expectedErrString != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrString)
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_ValidateUsername(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	tests := []struct {
		name              string
		username          string
		expectedErrString string
		mockSetup         func()
	}{
		{
			name:     "valid username - not taken",
			username: "newuser",
			mockSetup: func() {
				mockUserRepo.On("UsernameExists", "newuser").Return(false, nil)
			},
		},
		{
			name:              "username already taken",
			username:          "takenuser",
			expectedErrString: "the username has already been taken",
			mockSetup: func() {
				mockUserRepo.On("UsernameExists", "takenuser").Return(true, nil)
			},
		},
		{
			name:              "username exists check error",
			username:          "erroruser",
			expectedErrString: "repository error",
			mockSetup: func() {
				mockUserRepo.On("UsernameExists", "erroruser").Return(false, errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			err := userSvc.ValidateUsername(tt.username)

			if tt.expectedErrString != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrString)
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_HashPassword(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	t.Run("successful password hashing", func(t *testing.T) {
		password := "testpassword123"

		hashedPassword, err := userSvc.HashPassword(password)

		assert.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
		assert.NotEqual(t, password, hashedPassword)

		// Verify the hash can be checked
		isValid := userSvc.CheckPassword(password, hashedPassword)
		assert.True(t, isValid)
	})
}

func TestUserService_CheckPassword(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	userSvc := userService.NewUserService(mockUserRepo, nil)

	t.Run("correct password", func(t *testing.T) {
		password := "testpassword123"
		hashedPassword, _ := userSvc.HashPassword(password)

		isValid := userSvc.CheckPassword(password, hashedPassword)

		assert.True(t, isValid)
	})

	t.Run("incorrect password", func(t *testing.T) {
		password := "testpassword123"
		wrongPassword := "wrongpassword"
		hashedPassword, _ := userSvc.HashPassword(password)

		isValid := userSvc.CheckPassword(wrongPassword, hashedPassword)

		assert.False(t, isValid)
	})
}
