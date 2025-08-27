package user

import (
	"github.com/stretchr/testify/mock"
	"gopi.com/internal/app/user"
	userModel "gopi.com/internal/domain/user/model"
)

// MockUserRepository implements the UserRepository interface for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *userModel.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id string) (*userModel.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*userModel.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*userModel.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *userModel.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(limit, offset int) ([]*userModel.User, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmailAndPassword(email, password string) (*userModel.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserRepository) GetByOTP(email, otp string) (*userModel.User, error) {
	args := m.Called(email, otp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(id, newPassword string) error {
	args := m.Called(id, newPassword)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateOTP(id, otp string) error {
	args := m.Called(id, otp)
	return args.Error(0)
}

func (m *MockUserRepository) MarkAsVerified(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) GetAllUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserRepository) GetStaffUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserRepository) GetVerifiedUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserRepository) GetUnverifiedUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserRepository) EmailExists(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) UsernameExists(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

// MockEmailService implements the EmailService interface for testing
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendEmail(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

func (m *MockEmailService) SendPasswordResetEmail(to, resetToken string) error {
	args := m.Called(to, resetToken)
	return args.Error(0)
}

func (m *MockEmailService) SendWelcomeEmail(to, username string) error {
	args := m.Called(to, username)
	return args.Error(0)
}

// MockUserService implements the UserService interface for testing
type MockUserService struct {
	mock.Mock
	*user.UserService // Embed concrete type for compatibility
}

func (m *MockUserService) RegisterUser(username, email, firstName, lastName, password string, height, weight float64) (*userModel.User, error) {
	args := m.Called(username, email, firstName, lastName, password, height, weight)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserService) LoginUser(email, password string) (*userModel.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(id string) (*userModel.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(email string) (*userModel.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserService) GetUserByUsername(username string) (*userModel.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(user *userModel.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) ChangePassword(id, oldPassword, newPassword string) error {
	args := m.Called(id, oldPassword, newPassword)
	return args.Error(0)
}

func (m *MockUserService) ResetPassword(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockUserService) VerifyOTP(email, otp string) (*userModel.User, error) {
	args := m.Called(email, otp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserService) SendPasswordResetEmail(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockUserService) GetAllUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserService) GetStaffUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserService) GetVerifiedUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserService) GetUnverifiedUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
}
