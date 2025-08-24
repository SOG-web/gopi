package gorm

import (
	"time"

	"gopi.com/internal/domain/model"
	userModel "gopi.com/internal/domain/user/model"
	"gopi.com/internal/lib/id"
	"gorm.io/gorm"
)

// UserGORM represents the GORM model for User
type UserGORM struct {
	ID          string     `gorm:"type:varchar(26);primaryKey"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
	Username    string     `gorm:"unique;not null;size:150"`
	Email       string     `gorm:"unique;not null;size:254"`
	FirstName   string     `gorm:"size:150"`
	LastName    string     `gorm:"size:150"`
	Password    string     `gorm:"not null;size:128"`
	Height      float64    `gorm:"not null"`
	Weight      float64    `gorm:"not null"`
	OTP         *string    `gorm:"size:6"`
	IsStaff     bool       `gorm:"default:false"`
	IsActive    bool       `gorm:"default:true"`
	IsSuperuser bool       `gorm:"default:false"`
	IsVerified  bool       `gorm:"default:false"`
	DateJoined  time.Time  `gorm:"autoCreateTime"`
	LastLogin   *time.Time `gorm:"type:timestamp"`
	ProfileImageURL string  `gorm:"size:512"`

	// Foreign key relationships - these will be handled in other models
	// CampaignMembers     []CampaignGORM     `gorm:"many2many:campaign_members;"`
	// CampaignSponsors    []CampaignGORM     `gorm:"many2many:campaign_sponsors;"`
	// CampaignOwned       []CampaignGORM     `gorm:"foreignKey:OwnerID"`
	// CampaignRuns        []CampaignRunnerGORM `gorm:"foreignKey:OwnerID"`
}

func (UserGORM) TableName() string {
	return "users"
}

// BeforeCreate hook to set ID if not provided
func (u *UserGORM) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = id.New()
	}
	return
}

// ToUserModel converts GORM model to domain model
func (u *UserGORM) ToUserModel() *userModel.User {
	var otp string
	if u.OTP != nil {
		otp = *u.OTP
	}

	return &userModel.User{
		Base: model.Base{
			ID:        u.ID,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		},
		Username:    u.Username,
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Password:    u.Password,
		Height:      u.Height,
		Weight:      u.Weight,
		OTP:         otp,
		IsStaff:     u.IsStaff,
		IsActive:    u.IsActive,
		IsSuperuser: u.IsSuperuser,
		IsVerified:  u.IsVerified,
		DateJoined:  u.DateJoined,
		LastLogin:   u.LastLogin,
		ProfileImageURL: u.ProfileImageURL,
	}
}

// FromUserModel converts domain model to GORM model
func UserModelToGORM(u *userModel.User) *UserGORM {
	var otp *string
	if u.OTP != "" {
		otp = &u.OTP
	}

	return &UserGORM{
		ID:          u.ID,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		Username:    u.Username,
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Password:    u.Password,
		Height:      u.Height,
		Weight:      u.Weight,
		OTP:         otp,
		IsStaff:     u.IsStaff,
		IsActive:    u.IsActive,
		IsSuperuser: u.IsSuperuser,
		IsVerified:  u.IsVerified,
		DateJoined:  u.DateJoined,
		LastLogin:   u.LastLogin,
		ProfileImageURL: u.ProfileImageURL,
	}
}
