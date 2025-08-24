package repo

import "gopi.com/internal/domain/user/model"

type UserRepository interface {
	// Basic CRUD operations
	Create(user *model.User) error
	GetByID(id string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	Update(user *model.User) error
	Delete(id string) error
	List(limit, offset int) ([]*model.User, error)

	// Authentication specific operations
	GetByEmailAndPassword(email, password string) (*model.User, error)
	GetByOTP(email, otp string) (*model.User, error)
	UpdatePassword(id, newPassword string) error
	UpdateOTP(id, otp string) error
	MarkAsVerified(id string) error
	UpdateLastLogin(id string) error

	// Admin operations
	GetAllUsers() ([]*model.User, error)
	GetStaffUsers() ([]*model.User, error)
	GetVerifiedUsers() ([]*model.User, error)
	GetUnverifiedUsers() ([]*model.User, error)

	// Validation helpers
	EmailExists(email string) (bool, error)
	UsernameExists(username string) (bool, error)
}
