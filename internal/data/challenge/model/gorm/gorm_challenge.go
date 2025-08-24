package gorm

import (
	"time"

	userGorm "gopi.com/internal/data/user/model/gorm"
	challengeModel "gopi.com/internal/domain/challenge/model"
	"gopi.com/internal/domain/model"
	"gopi.com/internal/lib/id"
	"gorm.io/gorm"
)

type Challenge struct {
	ID                string `gorm:"type:varchar(255);primary_key"`
	OwnerID           string `gorm:"not null;index"`
	Name              string `gorm:"not null"`
	Description       string `gorm:"type:text"`
	Mode              string `gorm:"type:varchar(50);index"`
	Condition         string `gorm:"type:text"`
	Goal              string
	Location          string `gorm:"index"`
	DistanceToCover   float64 `gorm:"default:0"`
	TargetAmount      float64 `gorm:"default:0"`
	TargetAmountPerKm float64 `gorm:"default:0"`
	StartDuration     string
	EndDuration       string
	NoOfWinner        int    `gorm:"default:3"`
	WinningPrice      string `gorm:"type:json"` // JSON stored as text
	CausePrice        string `gorm:"type:json"` // JSON stored as text
	CoverImage        string
	VideoUrl          string
	Slug              string `gorm:"unique;index"`
	CreatedAt         time.Time `gorm:"index"`
	UpdatedAt         time.Time
	
	// Database relationships
	Owner userGorm.UserGORM `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE"`
}

func (Challenge) TableName() string {
	return "challenge_challenge"
}

// Junction table for Challenge Members (many-to-many)
type ChallengeMember struct {
	ID          string `gorm:"type:varchar(255);primary_key"`
	ChallengeID string `gorm:"not null;index;uniqueIndex:idx_challenge_member"`
	UserID      string `gorm:"not null;index;uniqueIndex:idx_challenge_member"`
	JoinedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	
	// Foreign key relationships
	Challenge Challenge         `gorm:"foreignKey:ChallengeID;constraint:OnDelete:CASCADE"`
	User      userGorm.UserGORM `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (ChallengeMember) TableName() string {
	return "challenge_members"
}

// Junction table for Challenge Sponsors (many-to-many)
type ChallengeSponsor struct {
	ID          string `gorm:"type:varchar(255);primary_key"`
	ChallengeID string `gorm:"not null;index;uniqueIndex:idx_challenge_sponsor"`
	UserID      string `gorm:"not null;index;uniqueIndex:idx_challenge_sponsor"`
	SponsoredAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	
	// Foreign key relationships
	Challenge Challenge         `gorm:"foreignKey:ChallengeID;constraint:OnDelete:CASCADE"`
	User      userGorm.UserGORM `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (ChallengeSponsor) TableName() string {
	return "challenge_sponsors"
}

type Cause struct {
	ID                 string `gorm:"type:varchar(255);primary_key"`
	ChallengeID        string `gorm:"not null;index"`
	Name               string `gorm:"not null"`
	Problem            string `gorm:"type:text"`
	Solution           string `gorm:"type:text"`
	ProductDescription string `gorm:"type:text"`
	Activity           string `gorm:"type:varchar(50);index"`
	Location           string
	Description        string
	IsCommercial       bool `gorm:"default:false"`
	WhoIdeaImpact      string
	BuyerUser          string
	DistanceCovered    float64 `gorm:"default:0;index"`
	AmountPerPiece     float64 `gorm:"default:0"`
	Duration           string
	FundCause          bool    `gorm:"default:false"`
	FundAmount         float64 `gorm:"default:0"`
	WillingAmount      float64 `gorm:"default:0"`
	UnitPrice          float64 `gorm:"default:0"`
	CostToLaunch       string
	BenefitDesc        string
	OwnerID            string `gorm:"not null;index"`
	WorkoutImg         string
	VideoUrl           string
	Slug               string `gorm:"unique;index"`
	CreatedAt          time.Time `gorm:"index"`
	UpdatedAt          time.Time
	
	// Database relationships
	Challenge Challenge         `gorm:"foreignKey:ChallengeID;constraint:OnDelete:CASCADE"`
	Owner     userGorm.UserGORM `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE"`
}

func (Cause) TableName() string {
	return "challenge_cause"
}

// Junction table for Cause Members (many-to-many)
type CauseMember struct {
	ID       string `gorm:"type:varchar(255);primary_key"`
	CauseID  string `gorm:"not null;index;uniqueIndex:idx_cause_member"`
	UserID   string `gorm:"not null;index;uniqueIndex:idx_cause_member"`
	JoinedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	
	// Foreign key relationships
	Cause Cause             `gorm:"foreignKey:CauseID;constraint:OnDelete:CASCADE"`
	User  userGorm.UserGORM `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (CauseMember) TableName() string {
	return "cause_members"
}

// Junction table for Cause Sponsors (many-to-many)
type CauseSponsorMember struct {
	ID          string `gorm:"type:varchar(255);primary_key"`
	CauseID     string `gorm:"not null;index;uniqueIndex:idx_cause_sponsor"`
	UserID      string `gorm:"not null;index;uniqueIndex:idx_cause_sponsor"`
	SponsoredAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	
	// Foreign key relationships
	Cause Cause             `gorm:"foreignKey:CauseID;constraint:OnDelete:CASCADE"`
	User  userGorm.UserGORM `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (CauseSponsorMember) TableName() string {
	return "cause_sponsor_members"
}

type CauseRunner struct {
	ID              string `gorm:"type:varchar(255);primary_key"`
	CauseID         string `gorm:"not null;index"`
	DistanceToCover float64 `gorm:"default:0"`
	DistanceCovered float64 `gorm:"default:0;index"`
	Duration        string
	MoneyRaised     float64 `gorm:"default:0;index"`
	CoverImage      string
	Activity        string `gorm:"type:varchar(50);index"`
	OwnerID         string `gorm:"not null;index"`
	DateJoined      time.Time `gorm:"index;column:date_joined;autoCreateTime"`
	CreatedAt       time.Time `gorm:"index;column:date_joined"`
	UpdatedAt       time.Time `gorm:"column:date_updated"`
	
	// Database relationships
	Cause Cause             `gorm:"foreignKey:CauseID;constraint:OnDelete:CASCADE"`
	Owner userGorm.UserGORM `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE"`
}

func (CauseRunner) TableName() string {
	return "challenge_causerunner"
}

type SponsorChallenge struct {
	ID          string `gorm:"type:varchar(255);primary_key"`
	SponsorID   string `gorm:"not null;index;unique"` // OneToOneField in Django
	ChallengeID string `gorm:"not null;index"`
	Distance    float64 `gorm:"default:0"`
	AmountPerKm float64 `gorm:"default:0"`
	TotalAmount float64 `gorm:"default:0"`
	BrandImg    string
	VideoUrl    string
	CreatedAt   time.Time `gorm:"index;column:date_created"`
	UpdatedAt   time.Time `gorm:"column:date_updated"`
	
	// Database relationships - OneToOne with User for sponsor
	Sponsor   userGorm.UserGORM `gorm:"foreignKey:SponsorID;constraint:OnDelete:CASCADE"`
	Challenge Challenge         `gorm:"foreignKey:ChallengeID;constraint:OnDelete:CASCADE"`
}

func (SponsorChallenge) TableName() string {
	return "challenge_sponsorchallenge"
}

type SponsorCause struct {
	ID          string `gorm:"type:varchar(255);primary_key"`
	SponsorID   string `gorm:"not null;index"` // ForeignKey in Django (multiple sponsors per cause)
	CauseID     string `gorm:"not null;index"`
	Distance    float64 `gorm:"default:0"`
	AmountPerKm float64 `gorm:"default:0"`
	TotalAmount float64 `gorm:"default:0"`
	BrandImg    string
	VideoUrl    string
	CreatedAt   time.Time `gorm:"index;column:date_created"`
	UpdatedAt   time.Time `gorm:"column:date_updated"`
	
	// Database relationships - ForeignKey allows multiple sponsors per cause
	Sponsor userGorm.UserGORM `gorm:"foreignKey:SponsorID;constraint:OnDelete:CASCADE"`
	Cause   Cause             `gorm:"foreignKey:CauseID;constraint:OnDelete:CASCADE"`
}

func (SponsorCause) TableName() string {
	return "challenge_sponsorcause"
}

type CauseBuyer struct {
	ID         string `gorm:"type:varchar(255);primary_key"`
	BuyerID    string `gorm:"index"` // SET_NULL in Django, so nullable
	CauseID    string `gorm:"not null;unique"` // OneToOneField in Django
	Amount     float64 `gorm:"type:decimal(19,2);default:0"` // DecimalField in Django
	DateBought time.Time `gorm:"index;column:date_bought;autoCreateTime"` // date_bought in Django
	CreatedAt  time.Time `gorm:"index;column:date_bought"` // date_bought in Django
	UpdatedAt  time.Time

	// Database relationships - OneToOne with Cause, CASCADE with User
	Buyer userGorm.UserGORM `gorm:"foreignKey:BuyerID;constraint:OnDelete:CASCADE"`
	Cause Cause             `gorm:"foreignKey:CauseID;constraint:OnDelete:CASCADE"`
}

func (CauseBuyer) TableName() string {
	return "challenge_causebuyer"
}

// GORM hooks
func (c *Challenge) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID = id.New()
	}
	return
}

func (c *Cause) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID = id.New()
	}
	return
}

func (cr *CauseRunner) BeforeCreate(tx *gorm.DB) (err error) {
	if cr.ID == "" {
		cr.ID = id.New()
	}
	return
}

func (sc *SponsorChallenge) BeforeCreate(tx *gorm.DB) (err error) {
	if sc.ID == "" {
		sc.ID = id.New()
	}
	return
}

func (sc *SponsorCause) BeforeCreate(tx *gorm.DB) (err error) {
	if sc.ID == "" {
		sc.ID = id.New()
	}
	return
}

func (cb *CauseBuyer) BeforeCreate(tx *gorm.DB) (err error) {
	if cb.ID == "" {
		cb.ID = id.New()
	}
	return
}

// Auto-calculate total amount for sponsor challenge
func (sc *SponsorChallenge) BeforeSave(tx *gorm.DB) (err error) {
	if sc.TotalAmount == 0 {
		sc.TotalAmount = sc.AmountPerKm * sc.Distance
	}
	return
}

// Auto-calculate total amount for sponsor cause
func (sc *SponsorCause) BeforeSave(tx *gorm.DB) (err error) {
	if sc.TotalAmount == 0 {
		sc.TotalAmount = sc.AmountPerKm * sc.Distance
	}
	return
}

// Conversion functions
func FromDomainChallenge(c *challengeModel.Challenge) *Challenge {
	// Convert JSON fields to strings
	winningPrice := ""
	causePrice := ""

	return &Challenge{
		ID:                c.ID,
		OwnerID:           c.OwnerID,
		Name:              c.Name,
		Description:       c.Description,
		Mode:              string(c.Mode),
		Condition:         c.Condition,
		Goal:              c.Goal,
		Location:          c.Location,
		DistanceToCover:   c.DistanceToCover,
		TargetAmount:      c.TargetAmount,
		TargetAmountPerKm: c.TargetAmountPerKm,
		StartDuration:     c.StartDuration,
		EndDuration:       c.EndDuration,
		NoOfWinner:        c.NoOfWinner,
		WinningPrice:      winningPrice,
		CausePrice:        causePrice,
		CoverImage:        c.CoverImage,
		VideoUrl:          c.VideoUrl,
		Slug:              c.Slug,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}

func ToDomainChallenge(c *Challenge) *challengeModel.Challenge {
	// In real implementation, you'd properly unmarshal JSON fields
	var winningPrice []interface{}
	var causePrice []interface{}

	return &challengeModel.Challenge{
		Base: model.Base{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		},
		OwnerID:           c.OwnerID,
		Name:              c.Name,
		Description:       c.Description,
		Mode:              challengeModel.ChallengeMode(c.Mode),
		Condition:         c.Condition,
		Goal:              c.Goal,
		Location:          c.Location,
		DistanceToCover:   c.DistanceToCover,
		TargetAmount:      c.TargetAmount,
		TargetAmountPerKm: c.TargetAmountPerKm,
		StartDuration:     c.StartDuration,
		EndDuration:       c.EndDuration,
		NoOfWinner:        c.NoOfWinner,
		WinningPrice:      winningPrice,
		CausePrice:        causePrice,
		CoverImage:        c.CoverImage,
		Members:           []interface{}{}, // Will be populated by repository when needed
		Sponsors:          []interface{}{}, // Will be populated by repository when needed
		VideoUrl:          c.VideoUrl,
		Slug:              c.Slug,
	}
}

func FromDomainCause(c *challengeModel.Cause) *Cause {
	return &Cause{
		ID:                 c.ID,
		ChallengeID:        c.ChallengeID,
		Name:               c.Name,
		Problem:            c.Problem,
		Solution:           c.Solution,
		ProductDescription: c.ProductDescription,
		Activity:           string(c.Activity),
		Location:           c.Location,
		Description:        c.Description,
		IsCommercial:       c.IsCommercial,
		WhoIdeaImpact:      c.WhoIdeaImpact,
		BuyerUser:          c.BuyerUser,
		DistanceCovered:    c.DistanceCovered,
		AmountPerPiece:     c.AmountPerPiece,
		Duration:           c.Duration,
		FundCause:          c.FundCause,
		FundAmount:         c.FundAmount,
		WillingAmount:      c.WillingAmount,
		UnitPrice:          c.UnitPrice,
		CostToLaunch:       c.CostToLaunch,
		BenefitDesc:        c.BenefitDesc,
		OwnerID:            c.OwnerID,
		WorkoutImg:         c.WorkoutImg,
		VideoUrl:           c.VideoUrl,
		Slug:               c.Slug,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
	}
}

func ToDomainCause(c *Cause) *challengeModel.Cause {
	return &challengeModel.Cause{
		Base: model.Base{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		},
		ChallengeID:        c.ChallengeID,
		Name:               c.Name,
		Problem:            c.Problem,
		Solution:           c.Solution,
		ProductDescription: c.ProductDescription,
		Activity:           challengeModel.Activity(c.Activity),
		Location:           c.Location,
		Description:        c.Description,
		IsCommercial:       c.IsCommercial,
		WhoIdeaImpact:      c.WhoIdeaImpact,
		BuyerUser:          c.BuyerUser,
		DistanceCovered:    c.DistanceCovered,
		AmountPerPiece:     c.AmountPerPiece,
		Duration:           c.Duration,
		FundCause:          c.FundCause,
		FundAmount:         c.FundAmount,
		WillingAmount:      c.WillingAmount,
		UnitPrice:          c.UnitPrice,
		CostToLaunch:       c.CostToLaunch,
		BenefitDesc:        c.BenefitDesc,
		OwnerID:            c.OwnerID,
		Members:            []interface{}{}, // Will be populated by repository when needed
		Sponsors:           []interface{}{}, // Will be populated by repository when needed
		WorkoutImg:         c.WorkoutImg,
		VideoUrl:           c.VideoUrl,
		Slug:               c.Slug,
	}
}

// Additional conversion functions for other models
func FromDomainCauseRunner(cr *challengeModel.CauseRunner) *CauseRunner {
	return &CauseRunner{
		ID:              cr.ID,
		CauseID:         cr.CauseID,
		DistanceToCover: cr.DistanceToCover,
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

func ToDomainCauseRunner(cr *CauseRunner) *challengeModel.CauseRunner {
	return &challengeModel.CauseRunner{
		Base: model.Base{
			ID:        cr.ID,
			CreatedAt: cr.CreatedAt,
			UpdatedAt: cr.UpdatedAt,
		},
		CauseID:         cr.CauseID,
		DistanceToCover: cr.DistanceToCover,
		DistanceCovered: cr.DistanceCovered,
		Duration:        cr.Duration,
		MoneyRaised:     cr.MoneyRaised,
		CoverImage:      cr.CoverImage,
		Activity:        cr.Activity,
		OwnerID:         cr.OwnerID,
		DateJoined:      cr.DateJoined,
	}
}

func FromDomainSponsorChallenge(sc *challengeModel.SponsorChallenge) *SponsorChallenge {
	return &SponsorChallenge{
		ID:          sc.ID,
		SponsorID:   sc.SponsorID,
		ChallengeID: sc.ChallengeID,
		Distance:    sc.Distance,
		AmountPerKm: sc.AmountPerKm,
		TotalAmount: sc.TotalAmount,
		BrandImg:    sc.BrandImg,
		VideoUrl:    sc.VideoUrl,
		CreatedAt:   sc.CreatedAt,
		UpdatedAt:   sc.UpdatedAt,
	}
}

func ToDomainSponsorChallenge(sc *SponsorChallenge) *challengeModel.SponsorChallenge {
	return &challengeModel.SponsorChallenge{
		Base: model.Base{
			ID:        sc.ID,
			CreatedAt: sc.CreatedAt,
			UpdatedAt: sc.UpdatedAt,
		},
		SponsorID:   sc.SponsorID,
		ChallengeID: sc.ChallengeID,
		Distance:    sc.Distance,
		AmountPerKm: sc.AmountPerKm,
		TotalAmount: sc.TotalAmount,
		BrandImg:    sc.BrandImg,
		VideoUrl:    sc.VideoUrl,
	}
}

func FromDomainSponsorCause(sc *challengeModel.SponsorCause) *SponsorCause {
	return &SponsorCause{
		ID:          sc.ID,
		SponsorID:   sc.SponsorID,
		CauseID:     sc.CauseID,
		Distance:    sc.Distance,
		AmountPerKm: sc.AmountPerKm,
		TotalAmount: sc.TotalAmount,
		BrandImg:    sc.BrandImg,
		VideoUrl:    sc.VideoUrl,
		CreatedAt:   sc.CreatedAt,
		UpdatedAt:   sc.UpdatedAt,
	}
}

func ToDomainSponsorCause(sc *SponsorCause) *challengeModel.SponsorCause {
	return &challengeModel.SponsorCause{
		Base: model.Base{
			ID:        sc.ID,
			CreatedAt: sc.CreatedAt,
			UpdatedAt: sc.UpdatedAt,
		},
		SponsorID:   sc.SponsorID,
		CauseID:     sc.CauseID,
		Distance:    sc.Distance,
		AmountPerKm: sc.AmountPerKm,
		TotalAmount: sc.TotalAmount,
		BrandImg:    sc.BrandImg,
		VideoUrl:    sc.VideoUrl,
	}
}

func FromDomainCauseBuyer(cb *challengeModel.CauseBuyer) *CauseBuyer {
	return &CauseBuyer{
		ID:         cb.ID,
		BuyerID:    cb.BuyerID,
		CauseID:    cb.CauseID,
		Amount:     cb.Amount,
		DateBought: cb.DateBought,
		CreatedAt:  cb.CreatedAt,
		UpdatedAt:  cb.UpdatedAt,
	}
}

func ToDomainCauseBuyer(cb *CauseBuyer) *challengeModel.CauseBuyer {
	return &challengeModel.CauseBuyer{
		Base: model.Base{
			ID:        cb.ID,
			CreatedAt: cb.CreatedAt,
			UpdatedAt: cb.UpdatedAt,
		},
		BuyerID:    cb.BuyerID,
		CauseID:    cb.CauseID,
		Amount:     cb.Amount,
		DateBought: cb.DateBought,
	}
}
