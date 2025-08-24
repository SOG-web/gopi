package model

import (
	"time"

	"gopi.com/internal/domain/model"
)

type CampaignMode string

const (
	CampaignModeFree CampaignMode = "Free"
	CampaignModePaid CampaignMode = "Paid"
)

type Activity string

const (
	ActivityWalking Activity = "Walking"
	ActivityRunning Activity = "Running"
	ActivityCycling Activity = "Cycling"
)

type Campaign struct {
	model.Base
	Name               string       `json:"name"`                 // name
	Description        string       `json:"description"`          // description
	Condition          string       `json:"condition"`            // condition
	Mode               CampaignMode `json:"mode"`                 // mode
	Goal               string       `json:"goal"`                 // goal
	Activity           Activity     `json:"activity"`             // activity
	AcceptTac          bool         `json:"accept_tac"`           // accept_tac
	Location           string       `json:"location"`             // location
	MoneyRaised        float64      `json:"money_raised"`         // money_raised
	TargetAmount       float64      `json:"target_amount"`        // target_amount
	TargetAmountPerKm  float64      `json:"target_amount_per_km"` // target_amount_per_km
	DistanceToCover    float64      `json:"distance_to_cover"`    // distance_to_cover
	DistanceCovered    float64      `json:"distance_covered"`     // distance_covered
	StartDuration      string       `json:"start_duration"`       // start_duration
	EndDuration        string       `json:"end_duration"`         // end_duration
	OwnerID            string       `json:"owner_id"`             // owner
	Slug               string       `json:"slug"`                 // slug
	WorkoutImg         string       `json:"workout_img"`          // workout_img
	// ManyToMany relationships - actual objects like Django
	Members  []interface{} `json:"members"`  // members (User objects)
	Sponsors []interface{} `json:"sponsors"` // sponsors (User objects)
}

type CampaignRunner struct {
	model.Base
	CampaignID      string    `json:"campaign_id"`      // campaign
	DistanceCovered float64   `json:"distance_covered"` // distance_covered
	Duration        string    `json:"duration"`         // duration
	MoneyRaised     float64   `json:"money_raised"`     // money_raised
	CoverImage      string    `json:"cover_image"`      // cover_image
	Activity        string    `json:"activity"`         // activity
	OwnerID         string    `json:"owner_id"`         // owner
	DateJoined      time.Time `json:"date_joined"`      // date_joined
}

type SponsorCampaign struct {
	model.Base
	CampaignID    string  `json:"campaign_id"`    // campaign
	Distance      float64 `json:"distance"`       // distance
	AmountPerKm   float64 `json:"amount_per_km"`  // amount_per_km
	TotalAmount   float64 `json:"total_amount"`   // total_amount
	BrandImg      string  `json:"brand_img"`      // brand_img
	VideoUrl      string  `json:"video_url"`      // video_url
	// ManyToMany relationship - actual objects like Django
	Sponsors []interface{} `json:"sponsors"` // sponsor (User objects via ManyToManyField)
}

// CalculateTotalAmount calculates the total amount based on distance and amount per km
func (sc *SponsorCampaign) CalculateTotalAmount() {
	sc.TotalAmount = sc.Distance * sc.AmountPerKm
}
