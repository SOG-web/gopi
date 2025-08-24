package challenge

import (
	"time"

	challengeModel "gopi.com/internal/domain/challenge/model"
	"gopi.com/internal/domain/challenge/repo"
	"gopi.com/internal/domain/model"
	"gopi.com/internal/lib/id"
)

type ChallengeService struct {
	challengeRepo     repo.ChallengeRepository
	causeRepo         repo.CauseRepository
	causeRunnerRepo   repo.CauseRunnerRepository
	sponsorRepo       repo.SponsorChallengeRepository
	sponsorCauseRepo  repo.SponsorCauseRepository
	causeBuyerRepo    repo.CauseBuyerRepository
}

func NewChallengeService(
	challengeRepo repo.ChallengeRepository,
	causeRepo repo.CauseRepository,
	causeRunnerRepo repo.CauseRunnerRepository,
	sponsorRepo repo.SponsorChallengeRepository,
	sponsorCauseRepo repo.SponsorCauseRepository,
	causeBuyerRepo repo.CauseBuyerRepository,
) *ChallengeService {
	return &ChallengeService{
		challengeRepo:    challengeRepo,
		causeRepo:        causeRepo,
		causeRunnerRepo:  causeRunnerRepo,
		sponsorRepo:      sponsorRepo,
		sponsorCauseRepo: sponsorCauseRepo,
		causeBuyerRepo:   causeBuyerRepo,
	}
}

func (s *ChallengeService) CreateChallenge(
	ownerID, name, description, condition, goal, location string,
	mode challengeModel.ChallengeMode,
	distanceToCover, targetAmount, targetAmountPerKm float64,
	startDuration, endDuration string,
	noOfWinner int,
) (*challengeModel.Challenge, error) {
	challenge := &challengeModel.Challenge{
		Base: model.Base{
			ID:        id.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		OwnerID:           ownerID,
		Name:              name,
		Description:       description,
		Mode:              mode,
		Condition:         condition,
		Goal:              goal,
		Location:          location,
		DistanceToCover:   distanceToCover,
		TargetAmount:      targetAmount,
		TargetAmountPerKm: targetAmountPerKm,
		StartDuration:     startDuration,
		EndDuration:       endDuration,
		NoOfWinner:        noOfWinner,
		WinningPrice:      []interface{}{},
		CausePrice:        []interface{}{},
		Members:           []interface{}{}, // Initialize empty
		Sponsors:          []interface{}{}, // Initialize empty
		Slug:              generateSlug(name),
	}

	err := s.challengeRepo.Create(challenge)
	if err != nil {
		return nil, err
	}

	return challenge, nil
}

func (s *ChallengeService) CreateCause(
	challengeID, ownerID, name, problem, solution, productDescription string,
	activity challengeModel.Activity,
	location, description string,
	isCommercial bool,
	amountPerPiece, fundAmount, willingAmount, unitPrice float64,
) (*challengeModel.Cause, error) {
	cause := &challengeModel.Cause{
		Base: model.Base{
			ID:        id.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ChallengeID:        challengeID,
		Name:               name,
		Problem:            problem,
		Solution:           solution,
		ProductDescription: productDescription,
		Activity:           activity,
		Location:           location,
		Description:        description,
		IsCommercial:       isCommercial,
		OwnerID:            ownerID,
		AmountPerPiece:     amountPerPiece,
		FundAmount:         fundAmount,
		WillingAmount:      willingAmount,
		UnitPrice:          unitPrice,
		Members:            []interface{}{}, // Initialize empty
		Sponsors:           []interface{}{}, // Initialize empty
		Slug:               generateSlug(name),
	}

	err := s.causeRepo.Create(cause)
	if err != nil {
		return nil, err
	}

	return cause, nil
}

func (s *ChallengeService) GetChallengeByID(id string) (*challengeModel.Challenge, error) {
	return s.challengeRepo.GetByID(id)
}

func (s *ChallengeService) GetChallengeBySlug(slug string) (*challengeModel.Challenge, error) {
	return s.challengeRepo.GetBySlug(slug)
}

func (s *ChallengeService) GetCauseByID(id string) (*challengeModel.Cause, error) {
	return s.causeRepo.GetByID(id)
}

func (s *ChallengeService) GetCausesByChallenge(challengeID string) ([]*challengeModel.Cause, error) {
	return s.causeRepo.GetByChallengeID(challengeID)
}

func (s *ChallengeService) JoinChallenge(challengeID, userID string) error {
	// In the new implementation, we should use the junction table
	// For now, we'll implement a basic version that works with the GORM junction tables
	challenge, err := s.challengeRepo.GetByID(challengeID)
	if err != nil {
		return err
	}

	// In production, you'd create an entry in the ChallengeMember junction table
	// For now, we'll just update the challenge's updated timestamp
	challenge.UpdatedAt = time.Now()
	return s.challengeRepo.Update(challenge)
}

func (s *ChallengeService) JoinCause(causeID, userID string) error {
	// In the new implementation, we should use the junction table  
	// For now, we'll implement a basic version that works with the GORM junction tables
	cause, err := s.causeRepo.GetByID(causeID)
	if err != nil {
		return err
	}

	// In production, you'd create an entry in the CauseMember junction table
	// For now, we'll just update the cause's updated timestamp
	cause.UpdatedAt = time.Now()
	return s.causeRepo.Update(cause)
}

func (s *ChallengeService) RecordCauseActivity(causeID, userID string, distanceToCover, distanceCovered float64, duration, activity string) error {
	causeRunner := &challengeModel.CauseRunner{
		Base: model.Base{
			ID:        id.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		CauseID:         causeID,
		DistanceToCover: distanceToCover,
		DistanceCovered: distanceCovered,
		Duration:        duration,
		Activity:        activity,
		OwnerID:         userID,
	}

	err := s.causeRunnerRepo.Create(causeRunner)
	if err != nil {
		return err
	}

	// Update cause's total distance covered
	cause, err := s.causeRepo.GetByID(causeID)
	if err != nil {
		return err
	}

	cause.DistanceCovered += distanceCovered
	cause.UpdatedAt = time.Now()

	return s.causeRepo.Update(cause)
}

func (s *ChallengeService) SponsorChallenge(challengeID, sponsorID string, distance, amountPerKm float64) error {
	sponsor := &challengeModel.SponsorChallenge{
		Base: model.Base{
			ID:        id.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		SponsorID:   sponsorID,
		ChallengeID: challengeID,
		Distance:    distance,
		AmountPerKm: amountPerKm,
	}

	// Calculate total amount
	sponsor.CalculateTotalAmount()

	return s.sponsorRepo.Create(sponsor)
}

func (s *ChallengeService) SponsorCause(causeID, sponsorID string, distance, amountPerKm float64) error {
	sponsor := &challengeModel.SponsorCause{
		Base: model.Base{
			ID:        id.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		SponsorID:   sponsorID,
		CauseID:     causeID,
		Distance:    distance,
		AmountPerKm: amountPerKm,
	}

	// Calculate total amount
	sponsor.CalculateTotalAmount()

	return s.sponsorCauseRepo.Create(sponsor)
}

func (s *ChallengeService) BuyCause(causeID, buyerID string, amount float64) error {
	buyer := &challengeModel.CauseBuyer{
		Base: model.Base{
			ID:        id.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		BuyerID: buyerID,
		CauseID: causeID,
		Amount:  amount,
	}

	return s.causeBuyerRepo.Create(buyer)
}

func (s *ChallengeService) GetLeaderboard() ([]*challengeModel.CauseRunner, error) {
	return s.causeRunnerRepo.GetLeaderboard()
}

func (s *ChallengeService) ListChallenges(limit, offset int) ([]*challengeModel.Challenge, error) {
	return s.challengeRepo.List(limit, offset)
}

// Helper function to generate slug from name
func generateSlug(name string) string {
	// Simplified implementation - in real app you'd use a proper slug library
	return name + "-" + id.New()[:8]
}
