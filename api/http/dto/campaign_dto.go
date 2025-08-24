package dto

import (
	"time"
)

// Campaign DTOs
type CreateCampaignRequest struct {
	Name              string  `json:"name" binding:"required,max=100"`
	Description       string  `json:"description,omitempty"`
	Condition         string  `json:"condition,omitempty"`
	Mode              string  `json:"mode,omitempty" binding:"omitempty,oneof=Free Paid"`
	Goal              string  `json:"goal,omitempty"`
	Activity          string  `json:"activity,omitempty" binding:"omitempty,oneof=Walking Running Cycling"`
	Location          string  `json:"location,omitempty"`
	TargetAmount      float64 `json:"target_amount,omitempty"`
	TargetAmountPerKm float64 `json:"target_amount_per_km,omitempty"`
	DistanceToCover   float64 `json:"distance_to_cover,omitempty"`
	StartDuration     string  `json:"start_duration,omitempty"`
	EndDuration       string  `json:"end_duration,omitempty"`
	WorkoutImg        string  `json:"workout_img,omitempty"`
}

type UpdateCampaignRequest struct {
	Name              string   `json:"name,omitempty" binding:"omitempty,max=100"`
	Description       string   `json:"description,omitempty"`
	Condition         string   `json:"condition,omitempty"`
	Mode              string   `json:"mode,omitempty" binding:"omitempty,oneof=Free Paid"`
	Goal              string   `json:"goal,omitempty"`
	Activity          string   `json:"activity,omitempty" binding:"omitempty,oneof=Walking Running Cycling"`
	Location          string   `json:"location,omitempty"`
	TargetAmount      *float64 `json:"target_amount,omitempty"`
	TargetAmountPerKm *float64 `json:"target_amount_per_km,omitempty"`
	DistanceToCover   *float64 `json:"distance_to_cover,omitempty"`
	StartDuration     string   `json:"start_duration,omitempty"`
	EndDuration       string   `json:"end_duration,omitempty"`
	WorkoutImg        string   `json:"workout_img,omitempty"`
	AcceptTac         *bool    `json:"accept_tac,omitempty"`
}

type CampaignResponse struct {
	ID                string                `json:"id"`
	Name              string                `json:"name"`
	Description       string                `json:"description"`
	Condition         string                `json:"condition"`
	Mode              string                `json:"mode"`
	Goal              string                `json:"goal"`
	Activity          string                `json:"activity"`
	AcceptTac         bool                  `json:"accept_tac"`
	Location          string                `json:"location"`
	MoneyRaised       float64               `json:"money_raised"`
	TargetAmount      float64               `json:"target_amount"`
	TargetAmountPerKm float64               `json:"target_amount_per_km"`
	DistanceToCover   float64               `json:"distance_to_cover"`
	DistanceCovered   float64               `json:"distance_covered"`
	StartDuration     string                `json:"start_duration"`
	EndDuration       string                `json:"end_duration"`
	Members           []interface{}         `json:"members"`
	Sponsors          []interface{}         `json:"sponsors"`
	Owner             CampaignOwnerInfo     `json:"owner"`
	Slug              string                `json:"slug"`
	WorkoutImg        string                `json:"workout_img"`
	DateCreated       time.Time             `json:"date_created"`
	DateUpdated       time.Time             `json:"date_updated"`
}

type CampaignOwnerInfo struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Username string `json:"username"`
}

type CampaignSponsorInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	BrandImg string `json:"brand_img"`
	VideoUrl string `json:"video_url"`
}

type CampaignListResponse struct {
	Campaigns []CampaignResponse `json:"campaigns"`
	Total     int                `json:"total"`
	Page      int                `json:"page"`
	Limit     int                `json:"limit"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// Campaign participation DTOs
type ParticipateCampaignRequest struct {
	Activity string `json:"activity" binding:"required"`
}

type ParticipateCampaignResponse struct {
	Message  string `json:"message"`
	Success  bool   `json:"success"`
	PostID   string `json:"post_id"`
	RunnerID string `json:"runner_id"`
}

// FinishActivityRequest represents the request body for finishing a campaign activity
type FinishActivityRequest struct {
	DistanceCovered float64 `json:"distance_covered" binding:"required"`
	Duration        string  `json:"duration" binding:"required"`
	MoneyRaised     float64 `json:"money_raised"`
}

// CampaignRunnerResponse represents a campaign runner response
type CampaignRunnerResponse struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	RunnerID        string    `json:"runner_id"`
	Username        string    `json:"username"`
	DistanceCovered float64   `json:"distance_covered"`
	Duration        string    `json:"duration"`
	MoneyRaised     float64   `json:"money_raised"`
	CampaignID      string    `json:"campaign_id"`
	CoverImage      string    `json:"cover_image"`
	Activity        string    `json:"activity"`
	DateJoined      time.Time `json:"date_joined"`
}

// Sponsor Campaign DTOs
type SponsorCampaignRequest struct {
	Distance    float64 `json:"distance" binding:"required,gt=0"`
	AmountPerKm float64 `json:"amount_per_km" binding:"required,gt=0"`
	BrandImg    string  `json:"brand_img,omitempty"`
	VideoUrl    string  `json:"video_url,omitempty"`
}

type SponsorCampaignResponse struct {
	ID          string      `json:"id"`
	Distance    float64     `json:"distance"`
	Campaign    string      `json:"campaign"`
	Sponsor     string      `json:"sponsor"`
	AmountPerKm float64     `json:"amount_per_km"`
	TotalAmount float64     `json:"total_amount"`
	BrandImg    string      `json:"brand_img"`
	VideoUrl    string      `json:"video_url"`
	DateCreated time.Time   `json:"date_created"`
	Paystack    interface{} `json:"paystack,omitempty"`
}

// Campaign leaderboard DTOs
type CampaignLeaderboardEntry struct {
	UserID          string  `json:"user_id"`
	Username        string  `json:"username"`
	FullName        string  `json:"full_name"`
	DistanceCovered float64 `json:"distance_covered"`
	MoneyRaised     float64 `json:"money_raised"`
	Duration        string  `json:"duration"`
	Activity        string  `json:"activity"`
	CoverImage      string  `json:"cover_image,omitempty"`
}

type CampaignLeaderboardResponse struct {
	CampaignSlug string                     `json:"campaign_slug"`
	Leaderboard  []CampaignLeaderboardEntry `json:"leaderboard"`
}

// Join Campaign Response
type JoinCampaignResponse struct {
	Response   string `json:"response"`
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
}

// Search Campaigns Request
type SearchCampaignsRequest struct {
	Query string `form:"q" binding:"required"`
	Page  int    `form:"page,default=1" binding:"min=1"`
	Limit int    `form:"limit,default=10" binding:"min=1,max=100"`
}

// Admin DTOs for Campaign Runner
type CreateCampaignRunnerRequest struct {
	CampaignID      string  `json:"campaign_id" binding:"required"`
	UserID          string  `json:"user_id" binding:"required"`
	Activity        string  `json:"activity" binding:"required"`
	DistanceCovered float64 `json:"distance_covered,omitempty"`
	Duration        string  `json:"duration,omitempty"`
	MoneyRaised     float64 `json:"money_raised,omitempty"`
}

type UpdateCampaignRunnerRequest struct {
	Activity        string   `json:"activity,omitempty"`
	DistanceCovered *float64 `json:"distance_covered,omitempty"`
	Duration        string   `json:"duration,omitempty"`
	MoneyRaised     *float64 `json:"money_raised,omitempty"`
	CoverImage      string   `json:"cover_image,omitempty"`
}

type CampaignRunnerListResponse struct {
	Runners []CampaignRunnerResponse `json:"runners"`
	Total   int                      `json:"total"`
	Page    int                      `json:"page"`
	Limit   int                      `json:"limit"`
}

// Admin DTOs for Sponsor Campaign
type CreateSponsorCampaignRequest struct {
	CampaignID  string   `json:"campaign_id" binding:"required"`
	SponsorID   string   `json:"sponsor_id" binding:"required"`
	Distance    float64  `json:"distance" binding:"required,gt=0"`
	AmountPerKm float64  `json:"amount_per_km" binding:"required,gt=0"`
	BrandImg    string   `json:"brand_img,omitempty"`
	VideoUrl    string   `json:"video_url,omitempty"`
	SponsorIDs  []string `json:"sponsor_ids,omitempty"`
}

type SponsorCampaignListResponse struct {
	Sponsors []SponsorCampaignResponse `json:"sponsors"`
	Total    int                       `json:"total"`
	Page     int                       `json:"page"`
	Limit    int                       `json:"limit"`
}

// Update DTOs
type UpdateSponsorCampaignRequest struct {
	Distance    *float64 `json:"distance,omitempty" binding:"omitempty,gt=0"`
	AmountPerKm *float64 `json:"amount_per_km,omitempty" binding:"omitempty,gt=0"`
	BrandImg    string   `json:"brand_img,omitempty"`
	VideoUrl    string   `json:"video_url,omitempty"`
}
