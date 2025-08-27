package campaign

import (
	"github.com/stretchr/testify/mock"
	campaignModel "gopi.com/internal/domain/campaign/model"
)

// MockCampaignRepository implements the CampaignRepository interface for testing
type MockCampaignRepository struct {
	mock.Mock
}

func (m *MockCampaignRepository) Create(campaign *campaignModel.Campaign) error {
	args := m.Called(campaign)
	return args.Error(0)
}

func (m *MockCampaignRepository) GetByID(id string) (*campaignModel.Campaign, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*campaignModel.Campaign), args.Error(1)
}

func (m *MockCampaignRepository) GetBySlug(slug string) (*campaignModel.Campaign, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*campaignModel.Campaign), args.Error(1)
}

func (m *MockCampaignRepository) GetByOwnerID(ownerID string) ([]*campaignModel.Campaign, error) {
	args := m.Called(ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*campaignModel.Campaign), args.Error(1)
}

func (m *MockCampaignRepository) Update(campaign *campaignModel.Campaign) error {
	args := m.Called(campaign)
	return args.Error(0)
}

func (m *MockCampaignRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCampaignRepository) List(limit, offset int) ([]*campaignModel.Campaign, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*campaignModel.Campaign), args.Error(1)
}

func (m *MockCampaignRepository) Search(query string, limit, offset int) ([]*campaignModel.Campaign, error) {
	args := m.Called(query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*campaignModel.Campaign), args.Error(1)
}

func (m *MockCampaignRepository) AddMember(campaignID, userID string) error {
	args := m.Called(campaignID, userID)
	return args.Error(0)
}

func (m *MockCampaignRepository) RemoveMember(campaignID, userID string) error {
	args := m.Called(campaignID, userID)
	return args.Error(0)
}

func (m *MockCampaignRepository) IsMember(campaignID, userID string) (bool, error) {
	args := m.Called(campaignID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCampaignRepository) AddSponsor(campaignID, userID string) error {
	args := m.Called(campaignID, userID)
	return args.Error(0)
}

func (m *MockCampaignRepository) RemoveSponsor(campaignID, userID string) error {
	args := m.Called(campaignID, userID)
	return args.Error(0)
}

func (m *MockCampaignRepository) IsSponsor(campaignID, userID string) (bool, error) {
	args := m.Called(campaignID, userID)
	return args.Bool(0), args.Error(1)
}

// MockCampaignRunnerRepository implements the CampaignRunnerRepository interface for testing
type MockCampaignRunnerRepository struct {
	mock.Mock
}

func (m *MockCampaignRunnerRepository) Create(runner *campaignModel.CampaignRunner) error {
	args := m.Called(runner)
	return args.Error(0)
}

func (m *MockCampaignRunnerRepository) GetByID(id string) (*campaignModel.CampaignRunner, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*campaignModel.CampaignRunner), args.Error(1)
}

func (m *MockCampaignRunnerRepository) GetByCampaignID(campaignID string) ([]*campaignModel.CampaignRunner, error) {
	args := m.Called(campaignID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*campaignModel.CampaignRunner), args.Error(1)
}

func (m *MockCampaignRunnerRepository) GetByOwnerID(ownerID string) ([]*campaignModel.CampaignRunner, error) {
	args := m.Called(ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*campaignModel.CampaignRunner), args.Error(1)
}

func (m *MockCampaignRunnerRepository) Update(runner *campaignModel.CampaignRunner) error {
	args := m.Called(runner)
	return args.Error(0)
}

func (m *MockCampaignRunnerRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockSponsorCampaignRepository implements the SponsorCampaignRepository interface for testing
type MockSponsorCampaignRepository struct {
	mock.Mock
}

func (m *MockSponsorCampaignRepository) Create(sponsor *campaignModel.SponsorCampaign) error {
	args := m.Called(sponsor)
	return args.Error(0)
}

func (m *MockSponsorCampaignRepository) GetByID(id string) (*campaignModel.SponsorCampaign, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*campaignModel.SponsorCampaign), args.Error(1)
}

func (m *MockSponsorCampaignRepository) GetByCampaignID(campaignID string) ([]*campaignModel.SponsorCampaign, error) {
	args := m.Called(campaignID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*campaignModel.SponsorCampaign), args.Error(1)
}

func (m *MockSponsorCampaignRepository) Update(sponsor *campaignModel.SponsorCampaign) error {
	args := m.Called(sponsor)
	return args.Error(0)
}

func (m *MockSponsorCampaignRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
