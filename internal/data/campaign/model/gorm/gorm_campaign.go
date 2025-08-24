package gorm

import (
	"encoding/json"
	"time"

	userGorm "gopi.com/internal/data/user/model/gorm"
	campaignModel "gopi.com/internal/domain/campaign/model"
	"gopi.com/internal/domain/model"
	"gopi.com/internal/lib/id"
	"gorm.io/gorm"
)

type Campaign struct {
	ID                string `gorm:"type:varchar(255);primary_key"`
	Name              string `gorm:"not null;index"`
	Description       string `gorm:"type:text"`
	Condition         string `gorm:"type:text"`
	Mode              string `gorm:"type:varchar(50);index"`
	Goal              string
	Activity          string `gorm:"type:varchar(50);index"`
	AcceptTac         bool    `gorm:"default:false"`
	Location          string  `gorm:"index"`
	MoneyRaised       float64 `gorm:"default:0;index"`
	TargetAmount      float64 `gorm:"default:0"`
	TargetAmountPerKm float64 `gorm:"default:0"`
	DistanceToCover   float64 `gorm:"default:0"`
	DistanceCovered   float64 `gorm:"default:0;index"`
	StartDuration     string
	EndDuration       string
	OwnerID           string `gorm:"not null;index"`
	Slug              string `gorm:"unique;not null;index"`
	WorkoutImg        string
	CreatedAt         time.Time `gorm:"index;column:date_created"`
	UpdatedAt         time.Time `gorm:"column:date_updated"`
	
	// Database relationships - proper foreign keys and many-to-many
	Owner             userGorm.UserGORM    `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE"`
	Members           []CampaignMember     `gorm:"foreignKey:CampaignID;constraint:OnDelete:CASCADE"`
	Sponsors          []CampaignSponsor    `gorm:"foreignKey:CampaignID;constraint:OnDelete:CASCADE"`
	CampaignRunners   []CampaignRunner     `gorm:"foreignKey:CampaignID;constraint:OnDelete:CASCADE"`
	SponsorCampaigns  []SponsorCampaign    `gorm:"foreignKey:CampaignID;constraint:OnDelete:CASCADE"`
}

func (Campaign) TableName() string {
	return "campaign_campaign"
}

// Junction table for Campaign Members (many-to-many)
type CampaignMember struct {
	ID         string `gorm:"type:varchar(255);primary_key"`
	CampaignID string `gorm:"not null;index;uniqueIndex:idx_campaign_member"`
	UserID     string `gorm:"not null;index;uniqueIndex:idx_campaign_member"`
	JoinedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	
	// Foreign key relationships
	Campaign   Campaign          `gorm:"foreignKey:CampaignID;constraint:OnDelete:CASCADE"`
	User       userGorm.UserGORM `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (CampaignMember) TableName() string {
	return "campaign_members"
}

// Junction table for Campaign Sponsors (many-to-many)
type CampaignSponsor struct {
	ID         string `gorm:"type:varchar(255);primary_key"`
	CampaignID string `gorm:"not null;index;uniqueIndex:idx_campaign_sponsor"`
	UserID     string `gorm:"not null;index;uniqueIndex:idx_campaign_sponsor"`
	SponsoredAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	
	// Foreign key relationships
	Campaign   Campaign          `gorm:"foreignKey:CampaignID;constraint:OnDelete:CASCADE"`
	User       userGorm.UserGORM `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (CampaignSponsor) TableName() string {
	return "campaign_sponsors"
}

type CampaignRunner struct {
	ID              string `gorm:"type:varchar(255);primary_key"`
	CampaignID      string `gorm:"not null;index"`
	DistanceCovered float64 `gorm:"default:0;index"`
	Duration        string
	MoneyRaised     float64 `gorm:"default:0;index"`
	CoverImage      string
	Activity        string `gorm:"type:varchar(50);index"`
	OwnerID         string `gorm:"not null;index"`
	DateJoined      time.Time `gorm:"index;column:date_joined"`
	CreatedAt       time.Time `gorm:"index"`
	UpdatedAt       time.Time `gorm:"column:date_updated"`
	
	// Database relationships
	Campaign        Campaign          `gorm:"foreignKey:CampaignID;constraint:OnDelete:CASCADE"`
	Owner           userGorm.UserGORM `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE"`
}

func (CampaignRunner) TableName() string {
	return "campaign_campaignrunner"
}

type SponsorCampaign struct {
	ID          string `gorm:"type:varchar(255);primary_key"`
	Sponsors    string `gorm:"type:text"` // JSON array stored as text for multiple sponsors
	CampaignID  string `gorm:"not null;index"`
	Distance    float64 `gorm:"default:0;index"`
	AmountPerKm float64 `gorm:"default:0"`
	TotalAmount float64 `gorm:"default:0;index"`
	BrandImg    string
	VideoUrl    string
	CreatedAt   time.Time `gorm:"index;column:date_created"`
	UpdatedAt   time.Time
	
	// Database relationships
	Campaign    Campaign `gorm:"foreignKey:CampaignID;constraint:OnDelete:CASCADE"`
}

func (SponsorCampaign) TableName() string {
	return "campaign_sponsorcampaign"
}

func (c *Campaign) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID = id.New()
	}
	return
}

func (cm *CampaignMember) BeforeCreate(tx *gorm.DB) (err error) {
	if cm.ID == "" {
		cm.ID = id.New()
	}
	return
}

func (cs *CampaignSponsor) BeforeCreate(tx *gorm.DB) (err error) {
	if cs.ID == "" {
		cs.ID = id.New()
	}
	return
}

func (cr *CampaignRunner) BeforeCreate(tx *gorm.DB) (err error) {
	if cr.ID == "" {
		cr.ID = id.New()
	}
	return
}

func (sc *SponsorCampaign) BeforeCreate(tx *gorm.DB) (err error) {
	if sc.ID == "" {
		sc.ID = id.New()
	}
	return
}

// Convert from domain Campaign to GORM Campaign
func FromDomainCampaign(c *campaignModel.Campaign) *Campaign {
	return &Campaign{
		ID:                c.ID,
		Name:              c.Name,
		Description:       c.Description,
		Condition:         c.Condition,
		Mode:              string(c.Mode),
		Goal:              c.Goal,
		Activity:          string(c.Activity),
		AcceptTac:         c.AcceptTac,
		Location:          c.Location,
		MoneyRaised:       c.MoneyRaised,
		TargetAmount:      c.TargetAmount,
		TargetAmountPerKm: c.TargetAmountPerKm,
		DistanceToCover:   c.DistanceToCover,
		DistanceCovered:   c.DistanceCovered,
		StartDuration:     c.StartDuration,
		EndDuration:       c.EndDuration,
		OwnerID:           c.OwnerID,
		Slug:              c.Slug,
		WorkoutImg:        c.WorkoutImg,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
		// Note: Members and Sponsors relationships are handled separately
	}
}

// Convert from GORM Campaign to domain Campaign
func ToDomainCampaign(c *Campaign) *campaignModel.Campaign {
	// Convert GORM member and sponsor relationships to interface{} slices
	var members []interface{}
	var sponsors []interface{}

	// Convert Members relationship to interface{} slice
	for _, member := range c.Members {
		members = append(members, member)
	}
	
	// Convert Sponsors relationship to interface{} slice  
	for _, sponsor := range c.Sponsors {
		sponsors = append(sponsors, sponsor)
	}

	return &campaignModel.Campaign{
		Base: model.Base{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		},
		Name:               c.Name,
		Description:        c.Description,
		Condition:          c.Condition,
		Mode:               campaignModel.CampaignMode(c.Mode),
		Goal:               c.Goal,
		Activity:           campaignModel.Activity(c.Activity),
		AcceptTac:          c.AcceptTac,
		Location:           c.Location,
		MoneyRaised:        c.MoneyRaised,
		TargetAmount:       c.TargetAmount,
		TargetAmountPerKm:  c.TargetAmountPerKm,
		DistanceToCover:    c.DistanceToCover,
		DistanceCovered:    c.DistanceCovered,
		StartDuration:      c.StartDuration,
		EndDuration:        c.EndDuration,
		Members:            members,
		Sponsors:           sponsors,
		OwnerID:            c.OwnerID,
		Slug:               c.Slug,
		WorkoutImg:         c.WorkoutImg,
	}
}

// Convert from domain CampaignRunner to GORM CampaignRunner
func FromDomainCampaignRunner(cr *campaignModel.CampaignRunner) *CampaignRunner {
	return &CampaignRunner{
		ID:              cr.ID,
		CampaignID:      cr.CampaignID,
		DistanceCovered: cr.DistanceCovered,
		Duration:        cr.Duration,
		MoneyRaised:     cr.MoneyRaised,
		CoverImage:      cr.CoverImage,
		Activity:        cr.Activity,
		OwnerID:         cr.OwnerID,
		DateJoined:      cr.DateJoined,
		CreatedAt:       cr.CreatedAt,
		UpdatedAt:       cr.UpdatedAt,
	}
}

// Convert from GORM CampaignRunner to domain CampaignRunner
func ToDomainCampaignRunner(cr *CampaignRunner) *campaignModel.CampaignRunner {
	return &campaignModel.CampaignRunner{
		Base: model.Base{
			ID:        cr.ID,
			CreatedAt: cr.CreatedAt,
			UpdatedAt: cr.UpdatedAt,
		},
		CampaignID:      cr.CampaignID,
		DistanceCovered: cr.DistanceCovered,
		Duration:        cr.Duration,
		MoneyRaised:     cr.MoneyRaised,
		CoverImage:      cr.CoverImage,
		Activity:        cr.Activity,
		OwnerID:         cr.OwnerID,
		DateJoined:      cr.DateJoined,
	}
}

// Convert from domain SponsorCampaign to GORM SponsorCampaign
func FromDomainSponsorCampaign(sc *campaignModel.SponsorCampaign) *SponsorCampaign {
	sponsors := ""
	// Convert Sponsors slice to JSON string
	if len(sc.Sponsors) > 0 {
		if data, err := json.Marshal(sc.Sponsors); err == nil {
			sponsors = string(data)
		}
	}

	return &SponsorCampaign{
		ID:          sc.ID,
		Sponsors:    sponsors,
		CampaignID:  sc.CampaignID,
		Distance:    sc.Distance,
		AmountPerKm: sc.AmountPerKm,
		TotalAmount: sc.TotalAmount,
		BrandImg:    sc.BrandImg,
		VideoUrl:    sc.VideoUrl,
		CreatedAt:   sc.CreatedAt,
		UpdatedAt:   sc.UpdatedAt,
	}
}

// Convert from GORM SponsorCampaign to domain SponsorCampaign
func ToDomainSponsorCampaign(sc *SponsorCampaign) *campaignModel.SponsorCampaign {
	var sponsors []interface{}

	// Parse JSON string back to interface{} slice
	if sc.Sponsors != "" {
		json.Unmarshal([]byte(sc.Sponsors), &sponsors)
	}

	return &campaignModel.SponsorCampaign{
		Base: model.Base{
			ID:        sc.ID,
			CreatedAt: sc.CreatedAt,
			UpdatedAt: sc.UpdatedAt,
		},
		Sponsors:    sponsors,
		CampaignID:  sc.CampaignID,
		Distance:    sc.Distance,
		AmountPerKm: sc.AmountPerKm,
		TotalAmount: sc.TotalAmount,
		BrandImg:    sc.BrandImg,
		VideoUrl:    sc.VideoUrl,
	}
}
