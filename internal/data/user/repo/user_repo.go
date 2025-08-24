package repo

import (
	"time"

	userGORM "gopi.com/internal/data/user/model/gorm"
	userModel "gopi.com/internal/domain/user/model"
	"gopi.com/internal/domain/user/repo"
	"gorm.io/gorm"
)

// UserRepositoryGORM implements UserRepository using GORM
type UserRepositoryGORM struct {
	db *gorm.DB
}

func NewUserRepositoryGORM(db *gorm.DB) repo.UserRepository {
	return &UserRepositoryGORM{db: db}
}

// NewGormUserRepository creates a new UserRepositoryGORM (compatibility function)
func NewGormUserRepository(db *gorm.DB) repo.UserRepository {
	return NewUserRepositoryGORM(db)
}

func (r *UserRepositoryGORM) Create(user *userModel.User) error {
	userGORMModel := userGORM.UserModelToGORM(user)
	return r.db.Create(userGORMModel).Error
}

func (r *UserRepositoryGORM) GetByID(id string) (*userModel.User, error) {
	var userGORMModel userGORM.UserGORM
	err := r.db.Where("id = ?", id).First(&userGORMModel).Error
	if err != nil {
		return nil, err
	}
	return userGORMModel.ToUserModel(), nil
}

func (r *UserRepositoryGORM) GetByEmail(email string) (*userModel.User, error) {
	var userGORMModel userGORM.UserGORM
	err := r.db.Where("email = ?", email).First(&userGORMModel).Error
	if err != nil {
		return nil, err
	}
	return userGORMModel.ToUserModel(), nil
}

func (r *UserRepositoryGORM) GetByUsername(username string) (*userModel.User, error) {
	var userGORMModel userGORM.UserGORM
	err := r.db.Where("username = ?", username).First(&userGORMModel).Error
	if err != nil {
		return nil, err
	}
	return userGORMModel.ToUserModel(), nil
}

func (r *UserRepositoryGORM) Update(user *userModel.User) error {
	userGORMModel := userGORM.UserModelToGORM(user)
	return r.db.Save(userGORMModel).Error
}

func (r *UserRepositoryGORM) Delete(id string) error {
	return r.db.Delete(&userGORM.UserGORM{}, "id = ?", id).Error
}

func (r *UserRepositoryGORM) List(limit, offset int) ([]*userModel.User, error) {
	var usersGORM []userGORM.UserGORM
	err := r.db.Limit(limit).Offset(offset).Find(&usersGORM).Error
	if err != nil {
		return nil, err
	}

	users := make([]*userModel.User, len(usersGORM))
	for i, userGORMModel := range usersGORM {
		users[i] = userGORMModel.ToUserModel()
	}
	return users, nil
}

func (r *UserRepositoryGORM) GetByEmailAndPassword(email, password string) (*userModel.User, error) {
	var userGORMModel userGORM.UserGORM
	err := r.db.Where("email = ? AND password = ?", email, password).First(&userGORMModel).Error
	if err != nil {
		return nil, err
	}
	return userGORMModel.ToUserModel(), nil
}

func (r *UserRepositoryGORM) GetByOTP(email, otp string) (*userModel.User, error) {
	var userGORMModel userGORM.UserGORM
	err := r.db.Where("email = ? AND otp = ?", email, otp).First(&userGORMModel).Error
	if err != nil {
		return nil, err
	}
	return userGORMModel.ToUserModel(), nil
}

func (r *UserRepositoryGORM) UpdatePassword(id, newPassword string) error {
	return r.db.Model(&userGORM.UserGORM{}).Where("id = ?", id).Update("password", newPassword).Error
}

func (r *UserRepositoryGORM) UpdateOTP(id, otp string) error {
	return r.db.Model(&userGORM.UserGORM{}).Where("id = ?", id).Update("otp", otp).Error
}

func (r *UserRepositoryGORM) MarkAsVerified(id string) error {
	now := time.Now()
	return r.db.Model(&userGORM.UserGORM{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_verified": true,
		"otp":         nil,
		"updated_at":  now,
	}).Error
}

func (r *UserRepositoryGORM) UpdateLastLogin(id string) error {
	now := time.Now()
	return r.db.Model(&userGORM.UserGORM{}).Where("id = ?", id).Update("last_login", now).Error
}

func (r *UserRepositoryGORM) GetAllUsers() ([]*userModel.User, error) {
	var usersGORM []userGORM.UserGORM
	err := r.db.Order("date_joined DESC").Find(&usersGORM).Error
	if err != nil {
		return nil, err
	}

	users := make([]*userModel.User, len(usersGORM))
	for i, userGORMModel := range usersGORM {
		users[i] = userGORMModel.ToUserModel()
	}
	return users, nil
}

func (r *UserRepositoryGORM) GetStaffUsers() ([]*userModel.User, error) {
	var usersGORM []userGORM.UserGORM
	err := r.db.Where("is_staff = ?", true).Find(&usersGORM).Error
	if err != nil {
		return nil, err
	}

	users := make([]*userModel.User, len(usersGORM))
	for i, userGORMModel := range usersGORM {
		users[i] = userGORMModel.ToUserModel()
	}
	return users, nil
}

func (r *UserRepositoryGORM) GetVerifiedUsers() ([]*userModel.User, error) {
	var usersGORM []userGORM.UserGORM
	err := r.db.Where("is_verified = ?", true).Find(&usersGORM).Error
	if err != nil {
		return nil, err
	}

	users := make([]*userModel.User, len(usersGORM))
	for i, userGORMModel := range usersGORM {
		users[i] = userGORMModel.ToUserModel()
	}
	return users, nil
}

func (r *UserRepositoryGORM) GetUnverifiedUsers() ([]*userModel.User, error) {
	var usersGORM []userGORM.UserGORM
	err := r.db.Where("is_verified = ?", false).Find(&usersGORM).Error
	if err != nil {
		return nil, err
	}

	users := make([]*userModel.User, len(usersGORM))
	for i, userGORMModel := range usersGORM {
		users[i] = userGORMModel.ToUserModel()
	}
	return users, nil
}

func (r *UserRepositoryGORM) EmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&userGORM.UserGORM{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserRepositoryGORM) UsernameExists(username string) (bool, error) {
	var count int64
	err := r.db.Model(&userGORM.UserGORM{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}
