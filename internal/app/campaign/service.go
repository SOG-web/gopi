package campaign

import (
	"errors"
	"fmt"
	"strings"
	"time"

	campaignModel "gopi.com/internal/domain/campaign/model"
	"gopi.com/internal/domain/campaign/repo"
	"gopi.com/internal/domain/model"
	"gopi.com/internal/lib/id"
)

type CampaignService struct {
	campaignRepo       repo.CampaignRepository
	campaignRunnerRepo repo.CampaignRunnerRepository
	sponsorRepo        repo.SponsorCampaignRepository
}

func NewCampaignService(
	campaignRepo repo.CampaignRepository,
	campaignRunnerRepo repo.CampaignRunnerRepository,
	sponsorRepo repo.SponsorCampaignRepository,
) *CampaignService {
	return &CampaignService{
		campaignRepo:       campaignRepo,
		campaignRunnerRepo: campaignRunnerRepo,
		sponsorRepo:        sponsorRepo,
	}
}

func (s *CampaignService) CreateCampaign(
	ownerID, ownerUsername, name, description, condition, goal, location string,
	mode campaignModel.CampaignMode,
	activity campaignModel.Activity,
	targetAmount, targetAmountPerKm, distanceToCover float64,
	startDuration, endDuration string,
) (*campaignModel.Campaign, error) {
	campaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        id.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:              name,
		Description:       description,
		Condition:         condition,
		Mode:              mode,
		Goal:              goal,
		Activity:          activity,
		Location:          location,
		TargetAmount:      targetAmount,
		TargetAmountPerKm: targetAmountPerKm,
		DistanceToCover:   distanceToCover,
		StartDuration:     startDuration,
		EndDuration:       endDuration,
		OwnerID:           ownerID,
		AcceptTac:         false,
		MoneyRaised:       0,
		DistanceCovered:   0,
		Members:           []interface{}{},
		Sponsors:          []interface{}{},
		Slug:              generateSlug(ownerUsername, name),
	}

	err := s.campaignRepo.Create(campaign)
	if err != nil {
		return nil, err
	}

	return campaign, nil
}

func (s *CampaignService) GetCampaignByID(id string) (*campaignModel.Campaign, error) {
	return s.campaignRepo.GetByID(id)
}

func (s *CampaignService) GetCampaignBySlug(slug string) (*campaignModel.Campaign, error) {
	return s.campaignRepo.GetBySlug(slug)
}

func (s *CampaignService) GetCampaignsByOwner(ownerID string) ([]*campaignModel.Campaign, error) {
	return s.campaignRepo.GetByOwnerID(ownerID)
}

func (s *CampaignService) JoinCampaign(campaignID, userID string) error {
	// Check if user is already a member
	isMember, err := s.campaignRepo.IsMember(campaignID, userID)
	if err != nil {
		return err
	}

	if isMember {
		return errors.New("user is already a member of this campaign")
	}

	// Add user to members using repository method
	return s.campaignRepo.AddMember(campaignID, userID)
}

func (s *CampaignService) RecordActivity(campaignID, userID string, distance float64, duration, activity string) error {
	campaignRunner := &campaignModel.CampaignRunner{
		Base: model.Base{
			ID:        id.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		CampaignID:      campaignID,
		DistanceCovered: distance,
		Duration:        duration,
		Activity:        activity,
		OwnerID:         userID,
		DateJoined:      time.Now(),
	}

	err := s.campaignRunnerRepo.Create(campaignRunner)
	if err != nil {
		return err
	}

	// Update campaign's total distance covered
	campaign, err := s.campaignRepo.GetByID(campaignID)
	if err != nil {
		return err
	}

	campaign.DistanceCovered += distance
	campaign.UpdatedAt = time.Now()

	return s.campaignRepo.Update(campaign)
}

func (s *CampaignService) SponsorCampaign(campaignID string, sponsors []interface{}, distance, amountPerKm float64) error {
	sponsor := &campaignModel.SponsorCampaign{
		Base: model.Base{
			ID:        id.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		CampaignID:  campaignID,
		Sponsors:    sponsors,
		Distance:    distance,
		AmountPerKm: amountPerKm,
	}

	// Calculate total amount
	sponsor.CalculateTotalAmount()

	err := s.sponsorRepo.Create(sponsor)
	if err != nil {
		return err
	}

	// Update campaign's money raised
	campaign, err := s.campaignRepo.GetByID(campaignID)
	if err != nil {
		return err
	}

	campaign.MoneyRaised += sponsor.TotalAmount
	campaign.UpdatedAt = time.Now()

	return s.campaignRepo.Update(campaign)
}

func (s *CampaignService) ListCampaigns(limit, offset int) ([]*campaignModel.Campaign, error) {
	return s.campaignRepo.List(limit, offset)
}

func (s *CampaignService) GetCampaignsByNonOwner(excludeUserID string, limit, offset int) ([]*campaignModel.Campaign, error) {
	campaigns, err := s.campaignRepo.List(limit*10, 0) // Get more to filter
	if err != nil {
		return nil, err
	}

	var result []*campaignModel.Campaign
	count := 0
	for _, campaign := range campaigns {
		if campaign.OwnerID != excludeUserID {
			if count >= offset && len(result) < limit {
				result = append(result, campaign)
			}
			count++
		}
	}
	return result, nil
}

func (s *CampaignService) IsMember(campaignID, userID string) (bool, error) {
	return s.campaignRepo.IsMember(campaignID, userID)
}

func (s *CampaignService) AddMember(campaignID, userID string) error {
	return s.campaignRepo.AddMember(campaignID, userID)
}

func (s *CampaignService) RemoveMember(campaignID, userID string) error {
	return s.campaignRepo.RemoveMember(campaignID, userID)
}

func (s *CampaignService) IsSponsor(campaignID, userID string) (bool, error) {
	return s.campaignRepo.IsSponsor(campaignID, userID)
}

func (s *CampaignService) AddSponsor(campaignID, userID string) error {
	return s.campaignRepo.AddSponsor(campaignID, userID)
}

func (s *CampaignService) RemoveSponsor(campaignID, userID string) error {
	return s.campaignRepo.RemoveSponsor(campaignID, userID)
}

func (s *CampaignService) UpdateCampaign(campaign *campaignModel.Campaign) error {
	campaign.UpdatedAt = time.Now()
	return s.campaignRepo.Update(campaign)
}

func (s *CampaignService) DeleteCampaign(id string) error {
	return s.campaignRepo.Delete(id)
}

func (s *CampaignService) GetLeaderboard(campaignSlug string) ([]*campaignModel.CampaignRunner, error) {
	campaign, err := s.campaignRepo.GetBySlug(campaignSlug)
	if err != nil {
		return nil, err
	}

	return s.campaignRunnerRepo.GetByCampaignID(campaign.ID)
}

func (s *CampaignService) ParticipateCampaign(campaignSlug, userID, activity string) (*campaignModel.CampaignRunner, error) {
	campaign, err := s.campaignRepo.GetBySlug(campaignSlug)
	if err != nil {
		return nil, err
	}

	// Add user to members if not already
	if err := s.JoinCampaign(campaign.ID, userID); err != nil {
		// Ignore error if already a member
	}

	// Create campaign runner
	campaignRunner := &campaignModel.CampaignRunner{
		Base: model.Base{
			ID:        id.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		CampaignID: campaign.ID,
		Activity:   activity,
		OwnerID:    userID,
		DateJoined: time.Now(),
	}

	err = s.campaignRunnerRepo.Create(campaignRunner)
	if err != nil {
		return nil, err
	}

	return campaignRunner, nil
}

func (s *CampaignService) FinishActivity(runnerID string, distance float64, duration string, moneyRaised float64) error {
	runner, err := s.campaignRunnerRepo.GetByID(runnerID)
	if err != nil {
		return err
	}

	runner.DistanceCovered += distance
	runner.Duration = duration
	runner.MoneyRaised += moneyRaised
	runner.UpdatedAt = time.Now()

	err = s.campaignRunnerRepo.Update(runner)
	if err != nil {
		return err
	}

	// Update campaign totals
	campaign, err := s.campaignRepo.GetByID(runner.CampaignID)
	if err != nil {
		return err
	}

	campaign.DistanceCovered += distance
	campaign.MoneyRaised += moneyRaised
	campaign.UpdatedAt = time.Now()

	return s.campaignRepo.Update(campaign)
}

func (s *CampaignService) GetRunnersByUser(userID string) ([]*campaignModel.CampaignRunner, error) {
	return s.campaignRunnerRepo.GetByOwnerID(userID)
}

func (s *CampaignService) GetRunnerByID(runnerID string) (*campaignModel.CampaignRunner, error) {
	return s.campaignRunnerRepo.GetByID(runnerID)
}

func (s *CampaignService) UpdateRunner(runner *campaignModel.CampaignRunner) error {
	return s.campaignRunnerRepo.Update(runner)
}

func (s *CampaignService) DeleteRunner(runnerID string) error {
	return s.campaignRunnerRepo.Delete(runnerID)
}

func (s *CampaignService) CreateSponsorCampaign(campaignID string, sponsors []interface{}, distance, amountPerKm float64, brandImg, videoUrl string) (*campaignModel.SponsorCampaign, error) {
	sponsor := &campaignModel.SponsorCampaign{
		Base: model.Base{
			ID:        id.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		CampaignID:  campaignID,
		Sponsors:    sponsors,
		Distance:    distance,
		AmountPerKm: amountPerKm,
		BrandImg:    brandImg,
		VideoUrl:    videoUrl,
	}

	// Calculate total amount
	sponsor.CalculateTotalAmount()

	err := s.sponsorRepo.Create(sponsor)
	if err != nil {
		return nil, err
	}

	// Update campaign's money raised
	campaign, err := s.campaignRepo.GetByID(campaignID)
	if err != nil {
		return nil, err
	}

	campaign.MoneyRaised += sponsor.TotalAmount
	campaign.UpdatedAt = time.Now()
	s.campaignRepo.Update(campaign)

	return sponsor, nil
}

func (s *CampaignService) GetSponsorCampaignByID(sponsorID string) (*campaignModel.SponsorCampaign, error) {
	return s.sponsorRepo.GetByID(sponsorID)
}

func (s *CampaignService) GetSponsorCampaignsByCampaign(campaignID string) ([]*campaignModel.SponsorCampaign, error) {
	return s.sponsorRepo.GetByCampaignID(campaignID)
}

func (s *CampaignService) UpdateSponsorCampaign(sponsor *campaignModel.SponsorCampaign) error {
	return s.sponsorRepo.Update(sponsor)
}

func (s *CampaignService) DeleteSponsorCampaign(sponsorID string) error {
	return s.sponsorRepo.Delete(sponsorID)
}

func (s *CampaignService) SearchCampaigns(query string, limit, offset int) ([]*campaignModel.Campaign, error) {
	// This is a simplified implementation
	// In a real implementation, you'd use database full-text search
	campaigns, err := s.campaignRepo.List(limit*2, offset) // Get more to filter
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var result []*campaignModel.Campaign
	for _, campaign := range campaigns {
		if strings.Contains(strings.ToLower(campaign.Name), query) ||
			strings.Contains(strings.ToLower(campaign.Description), query) ||
			strings.Contains(strings.ToLower(campaign.Location), query) {
			result = append(result, campaign)
		}
	}

	// Limit results
	if len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}

// Helper function to generate slug from username and name (like Django)
func generateSlug(ownerUsername, name string) string {
	// Format timestamp like Django signal: "D H M S"
	now := time.Now()
	dateTime := fmt.Sprintf("%02d-%02d-%02d %02d %02d %02d",
		now.Month(), now.Day(), now.Year()%100,
		now.Hour(), now.Minute(), now.Second())

	// Slugify like Django: username-datetime-name
	slug := fmt.Sprintf("%s-%s-%s",
		slugify(ownerUsername),
		dateTime,
		slugify(name))

	return slug
}

// Simple slugify function
func slugify(text string) string {
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, " ", "-")
	// Remove special characters (simplified)
	var result strings.Builder
	for _, char := range text {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}
	return result.String()
}
