package model

import (
	"time"

	"gopi.com/internal/domain/model"
)

type ChallengeMode string

const (
	ChallengeModeF    ChallengeMode = "Free"
	ChallengeModePaid ChallengeMode = "Paid"
)

type Activity string

const (
	ActivityWalking Activity = "Walking"
	ActivityRunning Activity = "Running"
	ActivityCycling Activity = "Cycling"
)

type Challenge struct {
	model.Base
	OwnerID            string        `json:"owner_id"`            // owner
	Name               string        `json:"name"`                // name
	Description        string        `json:"description"`         // description
	Mode               ChallengeMode `json:"mode"`                // mode
	Condition          string        `json:"condition"`           // condition
	Goal               string        `json:"goal"`                // goal
	Location           string        `json:"location"`            // location
	DistanceToCover    float64       `json:"distance_to_cover"`   // distance_to_cover
	TargetAmount       float64       `json:"target_amount"`       // target_amount
	TargetAmountPerKm  float64       `json:"target_amount_per_km"` // target_amount_per_km
	StartDuration      string        `json:"start_duration"`      // start_duration
	EndDuration        string        `json:"end_duration"`        // end_duration
	NoOfWinner         int           `json:"no_of_winner"`        // no_of_winner
	WinningPrice       []interface{} `json:"winning_price"`       // winning_price (JSONField)
	CausePrice         []interface{} `json:"cause_price"`         // cause_price (JSONField)
	CoverImage         string        `json:"cover_image"`         // cover_image
	VideoUrl           string        `json:"video_url"`           // video_url
	Slug               string        `json:"slug"`                // slug
	// ManyToMany relationships - actual objects like Django
	Members  []interface{} `json:"members"`  // members (User objects)
	Sponsors []interface{} `json:"sponsors"` // sponsors (User objects)
}

type Cause struct {
	model.Base
	ChallengeID        string   `json:"challenge_id"`        // challenge
	Name               string   `json:"name"`                // name
	Problem            string   `json:"problem"`             // problem
	Solution           string   `json:"solution"`            // solution
	ProductDescription string   `json:"product_description"` // product_description
	Activity           Activity `json:"activity"`            // activity
	Location           string   `json:"location"`            // location
	Description        string   `json:"description"`         // description
	IsCommercial       bool     `json:"is_commercial"`       // is_commercial
	WhoIdeaImpact      string   `json:"who_idea_impact"`     // who_idea_impact
	BuyerUser          string   `json:"buyer_user"`          // buyer_user
	DistanceCovered    float64  `json:"distance_covered"`    // distance_covered
	AmountPerPiece     float64  `json:"amount_per_piece"`    // amount_per_piece
	Duration           string   `json:"duration"`            // duration
	FundCause          bool     `json:"fund_cause"`          // fund_cause
	FundAmount         float64  `json:"fund_amount"`         // fund_amount
	WillingAmount      float64  `json:"willing_amount"`      // willing_amount
	UnitPrice          float64  `json:"unit_price"`          // unit_price
	CostToLaunch       string   `json:"cost_to_launch"`      // cost_to_launch
	BenefitDesc        string   `json:"benefit_desc"`        // benefit_desc
	OwnerID            string   `json:"owner_id"`            // owner
	WorkoutImg         string   `json:"workout_img"`         // workout_img
	VideoUrl           string   `json:"video_url"`           // video_url
	Slug               string   `json:"slug"`                // slug
	// ManyToMany relationships - actual objects like Django
	Members  []interface{} `json:"members"`  // members (User objects)
	Sponsors []interface{} `json:"sponsors"` // sponsors (User objects)
}

type CauseRunner struct {
	model.Base
	CauseID         string  `json:"cause_id"`         // cause
	DistanceToCover float64 `json:"distance_to_cover"` // distance_to_cover
	DistanceCovered float64 `json:"distance_covered"` // distance_covered
	Duration        string  `json:"duration"`         // duration
	MoneyRaised     float64 `json:"money_raised"`     // money_raised
	CoverImage      string  `json:"cover_image"`      // cover_image
	Activity        string  `json:"activity"`         // activity
	OwnerID         string  `json:"owner_id"`         // owner
	DateJoined      time.Time `json:"date_joined"`    // date_joined
}

type SponsorChallenge struct {
	model.Base
	SponsorID   string  `json:"sponsor_id"`   // sponsor (OneToOneField)
	ChallengeID string  `json:"challenge_id"` // challenge
	Distance    float64 `json:"distance"`     // distance
	AmountPerKm float64 `json:"amount_per_km"` // amount_per_km
	TotalAmount float64 `json:"total_amount"` // total_amount
	BrandImg    string  `json:"brand_img"`    // brand_img
	VideoUrl    string  `json:"video_url"`    // video_url
}

type SponsorCause struct {
	model.Base
	SponsorID   string  `json:"sponsor_id"`   // sponsor (ForeignKey)
	CauseID     string  `json:"cause_id"`     // cause
	Distance    float64 `json:"distance"`     // distance
	AmountPerKm float64 `json:"amount_per_km"` // amount_per_km
	TotalAmount float64 `json:"total_amount"` // total_amount
	BrandImg    string  `json:"brand_img"`    // brand_img
	VideoUrl    string  `json:"video_url"`    // video_url
}

type CauseBuyer struct {
	model.Base
	BuyerID    string    `json:"buyer_id"`    // buyer (SET_NULL)
	CauseID    string    `json:"cause_id"`    // cause (OneToOneField)
	Amount     float64   `json:"amount"`      // amount (DecimalField)
	DateBought time.Time `json:"date_bought"` // date_bought
}

// CalculateTotalAmount calculates the total amount for sponsor challenge
func (sc *SponsorChallenge) CalculateTotalAmount() {
	sc.TotalAmount = sc.AmountPerKm * sc.Distance
}

// CalculateTotalAmount calculates the total amount for sponsor cause
func (sc *SponsorCause) CalculateTotalAmount() {
	sc.TotalAmount = sc.AmountPerKm * sc.Distance
}

// GetRunnerList replicates Django's get_runner_list class method
func GetRunnerList(runners []*CauseRunner) []*CauseRunner {
	// Sort by distance covered descending
	sortedRunners := make([]*CauseRunner, len(runners))
	copy(sortedRunners, runners)
	
	// Simple sort implementation - in production use sort.Slice
	for i := 0; i < len(sortedRunners)-1; i++ {
		for j := i + 1; j < len(sortedRunners); j++ {
			if sortedRunners[i].DistanceCovered < sortedRunners[j].DistanceCovered {
				sortedRunners[i], sortedRunners[j] = sortedRunners[j], sortedRunners[i]
			}
		}
	}
	
	// Filter unique owners (similar to Django logic)
	unique := []*CauseRunner{}
	seen := make(map[string]bool)
	
	for _, runner := range sortedRunners {
		if runner.Duration != "" && !seen[runner.OwnerID] {
			unique = append(unique, runner)
			seen[runner.OwnerID] = true
		}
	}
	
	return unique
}
