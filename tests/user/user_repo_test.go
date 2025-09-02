package user_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gormModel "gopi.com/internal/data/user/model/gorm"
	"gopi.com/internal/data/user/repo"
	"gopi.com/internal/domain/model"
	userModel "gopi.com/internal/domain/user/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Setup in-memory database for testing
func setupUserTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&gormModel.UserGORM{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestUserRepositoryGORM_Create(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	tests := []struct {
		name        string
		user        *userModel.User
		expectedErr error
	}{
		{
			name: "successful user creation",
			user: &userModel.User{
				Base: model.Base{
					ID:        "test-user-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Username:   "testuser",
				Email:      "test@example.com",
				FirstName:  "Test",
				LastName:   "User",
				Password:   "hashedpassword",
				Height:     175.0,
				Weight:     70.0,
				OTP:        "123456",
				IsStaff:    false,
				IsActive:   true,
				IsVerified: false,
				DateJoined: time.Now(),
			},
			expectedErr: nil,
		},
		{
			name: "user creation with minimal fields",
			user: &userModel.User{
				Base: model.Base{
					ID:        "minimal-user-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Username:   "minimaluser",
				Email:      "minimal@example.com",
				FirstName:  "Minimal",
				LastName:   "User",
				Password:   "hashedpassword",
				Height:     170.0,
				Weight:     65.0,
				IsStaff:    false,
				IsActive:   true,
				IsVerified: false,
				DateJoined: time.Now(),
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := userRepo.Create(tt.user)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)

				// Verify user was created by retrieving it
				createdUser, err := userRepo.GetByID(tt.user.ID)
				assert.NoError(t, err)
				assert.NotNil(t, createdUser)
				assert.Equal(t, tt.user.Username, createdUser.Username)
				assert.Equal(t, tt.user.Email, createdUser.Email)
			}
		})
	}
}

func TestUserRepositoryGORM_GetByID(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create a test user first
	testUser := &userModel.User{
		Base: model.Base{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:   "testuser",
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Password:   "hashedpassword",
		Height:     175.0,
		Weight:     70.0,
		IsStaff:    false,
		IsActive:   true,
		IsVerified: false,
		DateJoined: time.Now(),
	}

	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		userID      string
		expectError bool
	}{
		{
			name:        "successful get user by id",
			userID:      "test-user-id",
			expectError: false,
		},
		{
			name:        "user not found",
			userID:      "nonexistent-id",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := userRepo.GetByID(tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.userID, user.ID)
				assert.Equal(t, "testuser", user.Username)
				assert.Equal(t, "test@example.com", user.Email)
			}
		})
	}
}

func TestUserRepositoryGORM_GetByEmail(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create a test user first
	testUser := &userModel.User{
		Base: model.Base{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:   "testuser",
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Password:   "hashedpassword",
		Height:     175.0,
		Weight:     70.0,
		IsStaff:    false,
		IsActive:   true,
		IsVerified: false,
		DateJoined: time.Now(),
	}

	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		email       string
		expectError bool
	}{
		{
			name:        "successful get user by email",
			email:       "test@example.com",
			expectError: false,
		},
		{
			name:        "user not found by email",
			email:       "nonexistent@example.com",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := userRepo.GetByEmail(tt.email)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, "testuser", user.Username)
			}
		})
	}
}

func TestUserRepositoryGORM_GetByUsername(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create a test user first
	testUser := &userModel.User{
		Base: model.Base{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:   "testuser",
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Password:   "hashedpassword",
		Height:     175.0,
		Weight:     70.0,
		IsStaff:    false,
		IsActive:   true,
		IsVerified: false,
		DateJoined: time.Now(),
	}

	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		username    string
		expectError bool
	}{
		{
			name:        "successful get user by username",
			username:    "testuser",
			expectError: false,
		},
		{
			name:        "user not found by username",
			username:    "nonexistentuser",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := userRepo.GetByUsername(tt.username)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.username, user.Username)
				assert.Equal(t, "test@example.com", user.Email)
			}
		})
	}
}

func TestUserRepositoryGORM_Update(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create a test user first
	testUser := &userModel.User{
		Base: model.Base{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:   "testuser",
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Password:   "hashedpassword",
		Height:     175.0,
		Weight:     70.0,
		IsStaff:    false,
		IsActive:   true,
		IsVerified: false,
		DateJoined: time.Now(),
	}

	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	t.Run("successful user update", func(t *testing.T) {
		// Update user fields
		testUser.FirstName = "Updated"
		testUser.LastName = "Name"
		testUser.Height = 180.0
		testUser.Weight = 75.0
		testUser.IsVerified = true
		testUser.UpdatedAt = time.Now()

		err := userRepo.Update(testUser)
		assert.NoError(t, err)

		// Verify update by retrieving user
		updatedUser, err := userRepo.GetByID(testUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated", updatedUser.FirstName)
		assert.Equal(t, "Name", updatedUser.LastName)
		assert.Equal(t, 180.0, updatedUser.Height)
		assert.Equal(t, 75.0, updatedUser.Weight)
		assert.True(t, updatedUser.IsVerified)
	})

	t.Run("update non-existent user", func(t *testing.T) {
		nonExistentUser := &userModel.User{
			Base: model.Base{
				ID: "nonexistent-id",
			},
		}

		err := userRepo.Update(nonExistentUser)
		assert.Error(t, err)
	})
}

func TestUserRepositoryGORM_Delete(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create a test user first
	testUser := &userModel.User{
		Base: model.Base{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:   "testuser",
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Password:   "hashedpassword",
		Height:     175.0,
		Weight:     70.0,
		IsStaff:    false,
		IsActive:   true,
		IsVerified: false,
		DateJoined: time.Now(),
	}

	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	t.Run("successful user deletion", func(t *testing.T) {
		err := userRepo.Delete(testUser.ID)
		assert.NoError(t, err)

		// Verify user was deleted
		deletedUser, err := userRepo.GetByID(testUser.ID)
		assert.Error(t, err)
		assert.Nil(t, deletedUser)
	})

	t.Run("delete non-existent user", func(t *testing.T) {
		err := userRepo.Delete("nonexistent-id")
		assert.NoError(t, err) // GORM doesn't return error for deleting non-existent records
	})
}

func TestUserRepositoryGORM_List(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create multiple test users
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
			Password:   "password1",
			Height:     170.0,
			Weight:     65.0,
			IsStaff:    false,
			IsActive:   true,
			IsVerified: true,
			DateJoined: time.Now(),
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
			Password:   "password2",
			Height:     175.0,
			Weight:     70.0,
			IsStaff:    true,
			IsActive:   true,
			IsVerified: false,
			DateJoined: time.Now(),
		},
		{
			Base: model.Base{
				ID:        "user3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Username:   "user3",
			Email:      "user3@example.com",
			FirstName:  "User",
			LastName:   "Three",
			Password:   "password3",
			Height:     180.0,
			Weight:     75.0,
			IsStaff:    false,
			IsActive:   false,
			IsVerified: true,
			DateJoined: time.Now(),
		},
	}

	for _, user := range users {
		err := userRepo.Create(user)
		assert.NoError(t, err)
	}

	tests := []struct {
		name     string
		limit    int
		offset   int
		expected int
	}{
		{
			name:     "list all users",
			limit:    10,
			offset:   0,
			expected: 3,
		},
		{
			name:     "list with limit",
			limit:    2,
			offset:   0,
			expected: 2,
		},
		{
			name:     "list with offset",
			limit:    10,
			offset:   1,
			expected: 2,
		},
		{
			name:     "list with limit and offset",
			limit:    1,
			offset:   1,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultUsers, err := userRepo.List(tt.limit, tt.offset)
			assert.NoError(t, err)
			assert.Len(t, resultUsers, tt.expected)
		})
	}
}

func TestUserRepositoryGORM_GetByEmailAndPassword(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create a test user first
	testUser := &userModel.User{
		Base: model.Base{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:   "testuser",
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Password:   "hashedpassword",
		Height:     175.0,
		Weight:     70.0,
		IsStaff:    false,
		IsActive:   true,
		IsVerified: false,
		DateJoined: time.Now(),
	}

	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		email       string
		password    string
		expectError bool
	}{
		{
			name:        "successful authentication",
			email:       "test@example.com",
			password:    "hashedpassword",
			expectError: false,
		},
		{
			name:        "wrong email",
			email:       "wrong@example.com",
			password:    "hashedpassword",
			expectError: true,
		},
		{
			name:        "wrong password",
			email:       "test@example.com",
			password:    "wrongpassword",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := userRepo.GetByEmailAndPassword(tt.email, tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
			}
		})
	}
}

func TestUserRepositoryGORM_GetByOTP(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create a test user first
	testUser := &userModel.User{
		Base: model.Base{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:   "testuser",
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Password:   "hashedpassword",
		Height:     175.0,
		Weight:     70.0,
		OTP:        "123456",
		IsStaff:    false,
		IsActive:   true,
		IsVerified: false,
		DateJoined: time.Now(),
	}

	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		email       string
		otp         string
		expectError bool
	}{
		{
			name:        "successful OTP verification",
			email:       "test@example.com",
			otp:         "123456",
			expectError: false,
		},
		{
			name:        "wrong email",
			email:       "wrong@example.com",
			otp:         "123456",
			expectError: true,
		},
		{
			name:        "wrong OTP",
			email:       "test@example.com",
			otp:         "wrongotp",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := userRepo.GetByOTP(tt.email, tt.otp)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.otp, user.OTP)
			}
		})
	}
}

func TestUserRepositoryGORM_UpdatePassword(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create a test user first
	testUser := &userModel.User{
		Base: model.Base{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:   "testuser",
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Password:   "oldpassword",
		Height:     175.0,
		Weight:     70.0,
		IsStaff:    false,
		IsActive:   true,
		IsVerified: false,
		DateJoined: time.Now(),
	}

	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	t.Run("successful password update", func(t *testing.T) {
		newPassword := "newpassword123"

		err := userRepo.UpdatePassword(testUser.ID, newPassword)
		assert.NoError(t, err)

		// Verify password was updated
		updatedUser, err := userRepo.GetByID(testUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, newPassword, updatedUser.Password)
	})

	t.Run("update password for non-existent user", func(t *testing.T) {
		err := userRepo.UpdatePassword("nonexistent-id", "newpassword")
		assert.NoError(t, err) // GORM doesn't return error for updating non-existent records
	})
}

func TestUserRepositoryGORM_UpdateOTP(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create a test user first
	testUser := &userModel.User{
		Base: model.Base{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:   "testuser",
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Password:   "password",
		Height:     175.0,
		Weight:     70.0,
		OTP:        "oldotp",
		IsStaff:    false,
		IsActive:   true,
		IsVerified: false,
		DateJoined: time.Now(),
	}

	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	t.Run("successful OTP update", func(t *testing.T) {
		newOTP := "654321"

		err := userRepo.UpdateOTP(testUser.ID, newOTP)
		assert.NoError(t, err)

		// Verify OTP was updated
		updatedUser, err := userRepo.GetByID(testUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, newOTP, updatedUser.OTP)
	})
}

func TestUserRepositoryGORM_MarkAsVerified(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create a test user first
	testUser := &userModel.User{
		Base: model.Base{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:   "testuser",
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Password:   "password",
		Height:     175.0,
		Weight:     70.0,
		IsStaff:    false,
		IsActive:   true,
		IsVerified: false,
		DateJoined: time.Now(),
	}

	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	t.Run("successful mark as verified", func(t *testing.T) {
		err := userRepo.MarkAsVerified(testUser.ID)
		assert.NoError(t, err)

		// Verify user was marked as verified
		updatedUser, err := userRepo.GetByID(testUser.ID)
		assert.NoError(t, err)
		assert.True(t, updatedUser.IsVerified)
	})
}

func TestUserRepositoryGORM_UpdateLastLogin(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create a test user first
	testUser := &userModel.User{
		Base: model.Base{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:   "testuser",
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Password:   "password",
		Height:     175.0,
		Weight:     70.0,
		IsStaff:    false,
		IsActive:   true,
		IsVerified: false,
		DateJoined: time.Now(),
	}

	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	t.Run("successful update last login", func(t *testing.T) {
		err := userRepo.UpdateLastLogin(testUser.ID)
		assert.NoError(t, err)

		// Verify last login was updated
		updatedUser, err := userRepo.GetByID(testUser.ID)
		assert.NoError(t, err)
		assert.NotNil(t, updatedUser.LastLogin)
	})
}

func TestUserRepositoryGORM_GetAllUsers(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create multiple test users
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
			Password:   "password1",
			Height:     170.0,
			Weight:     65.0,
			IsStaff:    false,
			IsActive:   true,
			IsVerified: true,
			DateJoined: time.Now(),
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
			Password:   "password2",
			Height:     175.0,
			Weight:     70.0,
			IsStaff:    true,
			IsActive:   true,
			IsVerified: false,
			DateJoined: time.Now(),
		},
	}

	for _, user := range users {
		err := userRepo.Create(user)
		assert.NoError(t, err)
	}

	t.Run("successful get all users", func(t *testing.T) {
		resultUsers, err := userRepo.GetAllUsers()
		assert.NoError(t, err)
		assert.Len(t, resultUsers, 2)

		// Verify users are returned correctly
		usernames := make([]string, len(resultUsers))
		for i, user := range resultUsers {
			usernames[i] = user.Username
		}
		assert.Contains(t, usernames, "user1")
		assert.Contains(t, usernames, "user2")
	})
}

func TestUserRepositoryGORM_GetStaffUsers(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create multiple test users
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
			Password:   "password1",
			Height:     170.0,
			Weight:     65.0,
			IsStaff:    false,
			IsActive:   true,
			IsVerified: true,
			DateJoined: time.Now(),
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
			Password:   "password2",
			Height:     175.0,
			Weight:     70.0,
			IsStaff:    true,
			IsActive:   true,
			IsVerified: false,
			DateJoined: time.Now(),
		},
		{
			Base: model.Base{
				ID:        "user3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Username:   "user3",
			Email:      "user3@example.com",
			FirstName:  "User",
			LastName:   "Three",
			Password:   "password3",
			Height:     180.0,
			Weight:     75.0,
			IsStaff:    true,
			IsActive:   true,
			IsVerified: true,
			DateJoined: time.Now(),
		},
	}

	for _, user := range users {
		err := userRepo.Create(user)
		assert.NoError(t, err)
	}

	t.Run("successful get staff users", func(t *testing.T) {
		staffUsers, err := userRepo.GetStaffUsers()
		assert.NoError(t, err)
		assert.Len(t, staffUsers, 2)

		// Verify all returned users are staff
		for _, user := range staffUsers {
			assert.True(t, user.IsStaff)
		}

		// Verify correct users are returned
		usernames := make([]string, len(staffUsers))
		for i, user := range staffUsers {
			usernames[i] = user.Username
		}
		assert.Contains(t, usernames, "user2")
		assert.Contains(t, usernames, "user3")
	})
}

func TestUserRepositoryGORM_GetVerifiedUsers(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create multiple test users
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
			Password:   "password1",
			Height:     170.0,
			Weight:     65.0,
			IsStaff:    false,
			IsActive:   true,
			IsVerified: true,
			DateJoined: time.Now(),
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
			Password:   "password2",
			Height:     175.0,
			Weight:     70.0,
			IsStaff:    false,
			IsActive:   true,
			IsVerified: false,
			DateJoined: time.Now(),
		},
	}

	for _, user := range users {
		err := userRepo.Create(user)
		assert.NoError(t, err)
	}

	t.Run("successful get verified users", func(t *testing.T) {
		verifiedUsers, err := userRepo.GetVerifiedUsers()
		assert.NoError(t, err)
		assert.Len(t, verifiedUsers, 1)

		// Verify all returned users are verified
		for _, user := range verifiedUsers {
			assert.True(t, user.IsVerified)
		}

		assert.Equal(t, "user1", verifiedUsers[0].Username)
	})
}

func TestUserRepositoryGORM_GetUnverifiedUsers(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create multiple test users
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
			Password:   "password1",
			Height:     170.0,
			Weight:     65.0,
			IsStaff:    false,
			IsActive:   true,
			IsVerified: true,
			DateJoined: time.Now(),
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
			Password:   "password2",
			Height:     175.0,
			Weight:     70.0,
			IsStaff:    false,
			IsActive:   true,
			IsVerified: false,
			DateJoined: time.Now(),
		},
		{
			Base: model.Base{
				ID:        "user3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Username:   "user3",
			Email:      "user3@example.com",
			FirstName:  "User",
			LastName:   "Three",
			Password:   "password3",
			Height:     180.0,
			Weight:     75.0,
			IsStaff:    false,
			IsActive:   true,
			IsVerified: false,
			DateJoined: time.Now(),
		},
	}

	for _, user := range users {
		err := userRepo.Create(user)
		assert.NoError(t, err)
	}

	t.Run("successful get unverified users", func(t *testing.T) {
		unverifiedUsers, err := userRepo.GetUnverifiedUsers()
		assert.NoError(t, err)
		assert.Len(t, unverifiedUsers, 2)

		// Verify all returned users are unverified
		for _, user := range unverifiedUsers {
			assert.False(t, user.IsVerified)
		}

		// Verify correct users are returned
		usernames := make([]string, len(unverifiedUsers))
		for i, user := range unverifiedUsers {
			usernames[i] = user.Username
		}
		assert.Contains(t, usernames, "user2")
		assert.Contains(t, usernames, "user3")
	})
}

func TestUserRepositoryGORM_EmailExists(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create a test user first
	testUser := &userModel.User{
		Base: model.Base{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:   "testuser",
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Password:   "password",
		Height:     175.0,
		Weight:     70.0,
		IsStaff:    false,
		IsActive:   true,
		IsVerified: false,
		DateJoined: time.Now(),
	}

	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		email       string
		expected    bool
		expectError bool
	}{
		{
			name:        "email exists",
			email:       "test@example.com",
			expected:    true,
			expectError: false,
		},
		{
			name:        "email does not exist",
			email:       "nonexistent@example.com",
			expected:    false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := userRepo.EmailExists(tt.email)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, exists)
			}
		})
	}
}

func TestUserRepositoryGORM_UsernameExists(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repo.NewUserRepositoryGORM(db)

	// Create a test user first
	testUser := &userModel.User{
		Base: model.Base{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:   "testuser",
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Password:   "password",
		Height:     175.0,
		Weight:     70.0,
		IsStaff:    false,
		IsActive:   true,
		IsVerified: false,
		DateJoined: time.Now(),
	}

	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		username    string
		expected    bool
		expectError bool
	}{
		{
			name:        "username exists",
			username:    "testuser",
			expected:    true,
			expectError: false,
		},
		{
			name:        "username does not exist",
			username:    "nonexistentuser",
			expected:    false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := userRepo.UsernameExists(tt.username)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, exists)
			}
		})
	}
}
