package dto

import (
	"time"
)

// Challenge DTOs
type CreateChallengeRequest struct {
	Name              string  `json:"name" binding:"required,max=100"`
	Description       string  `json:"description,omitempty"`
	Mode              string  `json:"mode,omitempty" binding:"omitempty,oneof=Free Paid"`
	Condition         string  `json:"condition,omitempty"`
	Goal              string  `json:"goal,omitempty"`
	Location          string  `json:"location,omitempty"`
	DistanceToCover   float64 `json:"distance_to_cover,omitempty"`
	TargetAmount      float64 `json:"target_amount,omitempty"`
	TargetAmountPerKm float64 `json:"target_amount_per_km,omitempty"`
	StartDuration     string  `json:"start_duration,omitempty"`
	EndDuration       string  `json:"end_duration,omitempty"`
	NoOfWinner        int     `json:"no_of_winner,omitempty"`
	CoverImage        string  `json:"cover_image,omitempty"`
	VideoUrl          string  `json:"video_url,omitempty"`
}

type UpdateChallengeRequest struct {
	Name              string   `json:"name,omitempty" binding:"omitempty,max=100"`
	Description       string   `json:"description,omitempty"`
	Mode              string   `json:"mode,omitempty" binding:"omitempty,oneof=Free Paid"`
	Condition         string   `json:"condition,omitempty"`
	Goal              string   `json:"goal,omitempty"`
	Location          string   `json:"location,omitempty"`
	DistanceToCover   *float64 `json:"distance_to_cover,omitempty"`
	TargetAmount      *float64 `json:"target_amount,omitempty"`
	TargetAmountPerKm *float64 `json:"target_amount_per_km,omitempty"`
	StartDuration     string   `json:"start_duration,omitempty"`
	EndDuration       string   `json:"end_duration,omitempty"`
	NoOfWinner        *int     `json:"no_of_winner,omitempty"`
	CoverImage        string   `json:"cover_image,omitempty"`
	VideoUrl          string   `json:"video_url,omitempty"`
}

type ChallengeResponse struct {
	ID                string                `json:"id"`
	Name              string                `json:"name"`
	Description       string                `json:"description"`
	Mode              string                `json:"mode"`
	Condition         string                `json:"condition"`
	Goal              string                `json:"goal"`
	Location          string                `json:"location"`
	DistanceToCover   float64               `json:"distance_to_cover"`
	TargetAmount      float64               `json:"target_amount"`
	TargetAmountPerKm float64               `json:"target_amount_per_km"`
	StartDuration     string                `json:"start_duration"`
	EndDuration       string                `json:"end_duration"`
	NoOfWinner        int                   `json:"no_of_winner"`
	WinningPrice      []interface{}         `json:"winning_price"`
	CausePrice        []interface{}         `json:"cause_price"`
	CoverImage        string                `json:"cover_image"`
	VideoUrl          string                `json:"video_url"`
	Slug              string                `json:"slug"`
	Owner             ChallengeOwnerInfo    `json:"owner"`
	Members           []ChallengeMemberInfo `json:"members"`
	Sponsors          []ChallengeSponsorInfo `json:"sponsors"`
	DateCreated       time.Time             `json:"date_created"`
	DateUpdated       time.Time             `json:"date_updated"`
}

type ChallengeOwnerInfo struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Username string `json:"username"`
}

type ChallengeMemberInfo struct {
	ID       string    `json:"id"`
	FullName string    `json:"full_name"`
	Username string    `json:"username"`
	JoinedAt time.Time `json:"joined_at"`
}

type ChallengeSponsorInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Username    string    `json:"username"`
	SponsoredAt time.Time `json:"sponsored_at"`
}

type ChallengeListResponse struct {
	Challenges []ChallengeResponse `json:"challenges"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
}

// Cause DTOs
type CreateCauseRequest struct {
	ChallengeID        string  `json:"challenge_id" binding:"required"`
	Name               string  `json:"name" binding:"required,max=100"`
	Problem            string  `json:"problem,omitempty"`
	Solution           string  `json:"solution,omitempty"`
	ProductDescription string  `json:"product_description,omitempty"`
	Activity           string  `json:"activity,omitempty" binding:"omitempty,oneof=Walking Running Cycling"`
	Location           string  `json:"location,omitempty"`
	Description        string  `json:"description,omitempty"`
	IsCommercial       bool    `json:"is_commercial,omitempty"`
	WhoIdeaImpact      string  `json:"who_idea_impact,omitempty"`
	BuyerUser          string  `json:"buyer_user,omitempty"`
	AmountPerPiece     float64 `json:"amount_per_piece,omitempty"`
	FundCause          bool    `json:"fund_cause,omitempty"`
	FundAmount         float64 `json:"fund_amount,omitempty"`
	WillingAmount      float64 `json:"willing_amount,omitempty"`
	UnitPrice          float64 `json:"unit_price,omitempty"`
	CostToLaunch       string  `json:"cost_to_launch,omitempty"`
	BenefitDesc        string  `json:"benefit_desc,omitempty"`
	WorkoutImg         string  `json:"workout_img,omitempty"`
	VideoUrl           string  `json:"video_url,omitempty"`
}

type UpdateCauseRequest struct {
	Name               string   `json:"name,omitempty" binding:"omitempty,max=100"`
	Problem            string   `json:"problem,omitempty"`
	Solution           string   `json:"solution,omitempty"`
	ProductDescription string   `json:"product_description,omitempty"`
	Activity           string   `json:"activity,omitempty" binding:"omitempty,oneof=Walking Running Cycling"`
	Location           string   `json:"location,omitempty"`
	Description        string   `json:"description,omitempty"`
	IsCommercial       *bool    `json:"is_commercial,omitempty"`
	WhoIdeaImpact      string   `json:"who_idea_impact,omitempty"`
	BuyerUser          string   `json:"buyer_user,omitempty"`
	AmountPerPiece     *float64 `json:"amount_per_piece,omitempty"`
	FundCause          *bool    `json:"fund_cause,omitempty"`
	FundAmount         *float64 `json:"fund_amount,omitempty"`
	WillingAmount      *float64 `json:"willing_amount,omitempty"`
	UnitPrice          *float64 `json:"unit_price,omitempty"`
	CostToLaunch       string   `json:"cost_to_launch,omitempty"`
	BenefitDesc        string   `json:"benefit_desc,omitempty"`
	WorkoutImg         string   `json:"workout_img,omitempty"`
	VideoUrl           string   `json:"video_url,omitempty"`
}

type CauseResponse struct {
	ID                 string              `json:"id"`
	ChallengeID        string              `json:"challenge_id"`
	Name               string              `json:"name"`
	Problem            string              `json:"problem"`
	Solution           string              `json:"solution"`
	ProductDescription string              `json:"product_description"`
	Activity           string              `json:"activity"`
	Location           string              `json:"location"`
	Description        string              `json:"description"`
	IsCommercial       bool                `json:"is_commercial"`
	WhoIdeaImpact      string              `json:"who_idea_impact"`
	BuyerUser          string              `json:"buyer_user"`
	DistanceCovered    float64             `json:"distance_covered"`
	AmountPerPiece     float64             `json:"amount_per_piece"`
	Duration           string              `json:"duration"`
	FundCause          bool                `json:"fund_cause"`
	FundAmount         float64             `json:"fund_amount"`
	WillingAmount      float64             `json:"willing_amount"`
	UnitPrice          float64             `json:"unit_price"`
	CostToLaunch       string              `json:"cost_to_launch"`
	BenefitDesc        string              `json:"benefit_desc"`
	WorkoutImg         string              `json:"workout_img"`
	VideoUrl           string              `json:"video_url"`
	Slug               string              `json:"slug"`
	Owner              CauseOwnerInfo      `json:"owner"`
	Members            []CauseMemberInfo   `json:"members"`
	Sponsors           []CauseSponsorInfo  `json:"sponsors"`
	DateCreated        time.Time           `json:"date_created"`
	DateUpdated        time.Time           `json:"date_updated"`
}

type CauseOwnerInfo struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Username string `json:"username"`
}

type CauseMemberInfo struct {
	ID       string    `json:"id"`
	FullName string    `json:"full_name"`
	Username string    `json:"username"`
	JoinedAt time.Time `json:"joined_at"`
}

type CauseSponsorInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Username    string    `json:"username"`
	SponsoredAt time.Time `json:"sponsored_at"`
}

type CauseListResponse struct {
	Causes []CauseResponse `json:"causes"`
	Total  int             `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

// Cause participation DTOs
type ParticipateCauseRequest struct {
	Activity string `json:"activity" binding:"required"`
}

type ParticipateCauseResponse struct {
	Message  string `json:"message"`
	Success  bool   `json:"success"`
	PostID   string `json:"post_id"`
	RunnerID string `json:"runner_id"`
}

// Finish Cause DTOs
type FinishCauseRequest struct {
	DistanceCovered float64 `json:"distance_covered" binding:"required"`
	Duration        string  `json:"duration" binding:"required"`
	MoneyRaised     float64 `json:"money_raised"`
}

// Record Activity DTOs
type RecordActivityRequest struct {
	CauseID         string  `json:"cause_id" binding:"required"`
	DistanceToCover float64 `json:"distance_to_cover" binding:"required,gt=0"`
	DistanceCovered float64 `json:"distance_covered" binding:"required,gt=0"`
	Duration        string  `json:"duration" binding:"required"`
	Activity        string  `json:"activity" binding:"required"`
}

type CauseRunnerResponse struct {
	ID              string    `json:"id"`
	CauseID         string    `json:"cause_id"`
	UserID          string    `json:"user_id"`
	RunnerID        string    `json:"runner_id"`
	Username        string    `json:"username"`
	DistanceToCover float64   `json:"distance_to_cover"`
	DistanceCovered float64   `json:"distance_covered"`
	Duration        string    `json:"duration"`
	MoneyRaised     float64   `json:"money_raised"`
	CoverImage      string    `json:"cover_image"`
	Activity        string    `json:"activity"`
	DateJoined      time.Time `json:"date_joined"`
}

// Sponsor DTOs
type SponsorChallengeRequest struct {
	ChallengeID string  `json:"challenge_id" binding:"required"`
	Distance    float64 `json:"distance" binding:"required,gt=0"`
	AmountPerKm float64 `json:"amount_per_km" binding:"required,gt=0"`
	BrandImg    string  `json:"brand_img,omitempty"`
	VideoUrl    string  `json:"video_url,omitempty"`
}

type SponsorCauseRequest struct {
	CauseID     string  `json:"cause_id" binding:"required"`
	Distance    float64 `json:"distance" binding:"required,gt=0"`
	AmountPerKm float64 `json:"amount_per_km" binding:"required,gt=0"`
	BrandImg    string  `json:"brand_img,omitempty"`
	VideoUrl    string  `json:"video_url,omitempty"`
}

type SponsorChallengeResponse struct {
	ID          string    `json:"id"`
	SponsorID   string    `json:"sponsor_id"`
	ChallengeID string    `json:"challenge_id"`
	Distance    float64   `json:"distance"`
	AmountPerKm float64   `json:"amount_per_km"`
	TotalAmount float64   `json:"total_amount"`
	BrandImg    string    `json:"brand_img"`
	VideoUrl    string    `json:"video_url"`
	DateCreated time.Time `json:"date_created"`
}

type SponsorCauseResponse struct {
	ID          string    `json:"id"`
	SponsorID   string    `json:"sponsor_id"`
	CauseID     string    `json:"cause_id"`
	Distance    float64   `json:"distance"`
	AmountPerKm float64   `json:"amount_per_km"`
	TotalAmount float64   `json:"total_amount"`
	BrandImg    string    `json:"brand_img"`
	VideoUrl    string    `json:"video_url"`
	DateCreated time.Time `json:"date_created"`
}

// Buy Cause DTOs
type BuyCauseRequest struct {
	CauseID string  `json:"cause_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
}

type BuyCauseResponse struct {
	ID          string    `json:"id"`
	BuyerID     string    `json:"buyer_id"`
	CauseID     string    `json:"cause_id"`
	Amount      float64   `json:"amount"`
	DateCreated time.Time `json:"date_created"`
}

// Leaderboard DTOs
type ChallengeLeaderboardEntry struct {
	UserID          string  `json:"user_id"`
	Username        string  `json:"username"`
	FullName        string  `json:"full_name"`
	DistanceCovered float64 `json:"distance_covered"`
	MoneyRaised     float64 `json:"money_raised"`
	Duration        string  `json:"duration"`
	Activity        string  `json:"activity"`
	CoverImage      string  `json:"cover_image,omitempty"`
}

type ChallengeLeaderboardResponse struct {
	ChallengeSlug string                      `json:"challenge_slug"`
	Leaderboard   []ChallengeLeaderboardEntry `json:"leaderboard"`
}

// General Leaderboard Response
type LeaderboardResponse struct {
	Runners []CauseRunnerResponse `json:"runners"`
}

type CauseLeaderboardEntry struct {
	CauseID         string  `json:"cause_id"`
	CauseName       string  `json:"cause_name"`
	UserID          string  `json:"user_id"`
	Username        string  `json:"username"`
	FullName        string  `json:"full_name"`
	DistanceCovered float64 `json:"distance_covered"`
	MoneyRaised     float64 `json:"money_raised"`
	Duration        string  `json:"duration"`
	Activity        string  `json:"activity"`
	CoverImage      string  `json:"cover_image,omitempty"`
}

type CauseLeaderboardResponse struct {
	ChallengeSlug string                  `json:"challenge_slug"`
	Leaderboard   []CauseLeaderboardEntry `json:"leaderboard"`
}

// Join responses
type JoinChallengeResponse struct {
	Response   string `json:"response"`
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
}

type JoinCauseResponse struct {
	Response   string `json:"response"`
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
}

// Search Challenges Request
type SearchChallengesRequest struct {
	Query string `form:"q" binding:"required"`
	Page  int    `form:"page,default=1" binding:"min=1"`
	Limit int    `form:"limit,default=10" binding:"min=1,max=100"`
}
