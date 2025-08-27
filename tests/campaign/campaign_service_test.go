package campaign_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	campaign "gopi.com/internal/app/campaign"
	campaignModel "gopi.com/internal/domain/campaign/model"
	"gopi.com/internal/domain/model"
	campaignMocks "gopi.com/tests/mocks/campaign"
)

func TestCampaignService_CreateCampaign(t *testing.T) {
	tests := []struct {
		name              string
		ownerID           string
		ownerUsername     string
		nameArg           string
		description       string
		condition         string
		goal              string
		location          string
		mode              campaignModel.CampaignMode
		activity          campaignModel.Activity
		targetAmount      float64
		targetAmountPerKm float64
		distanceToCover   float64
		startDuration     string
		endDuration       string
		expectedErr       error
		mockSetup         func(*campaignMocks.MockCampaignRepository)
	}{
		{
			name:              "successful campaign creation",
			ownerID:           "owner123",
			ownerUsername:     "testuser",
			nameArg:           "Test Campaign",
			description:       "A test campaign",
			condition:         "Good condition",
			goal:              "Test goal",
			location:          "Test location",
			mode:              campaignModel.CampaignModeFree,
			activity:          campaignModel.ActivityWalking,
			targetAmount:      1000.0,
			targetAmountPerKm: 5.0,
			distanceToCover:   10.0,
			startDuration:     "2024-01-01",
			endDuration:       "2024-01-31",
			expectedErr:       nil,
			mockSetup: func(mockRepo *campaignMocks.MockCampaignRepository) {
				mockRepo.On("Create", mock.MatchedBy(func(c *campaignModel.Campaign) bool {
					return c.OwnerID == "owner123" &&
						c.Name == "Test Campaign" &&
						c.Mode == campaignModel.CampaignModeFree &&
						c.Activity == campaignModel.ActivityWalking &&
						c.TargetAmount == 1000.0
				})).Return(nil)
			},
		},
		{
			name:              "repository error",
			ownerID:           "owner123",
			ownerUsername:     "testuser",
			nameArg:           "Test Campaign",
			description:       "A test campaign",
			condition:         "Good condition",
			goal:              "Test goal",
			location:          "Test location",
			mode:              campaignModel.CampaignModeFree,
			activity:          campaignModel.ActivityWalking,
			targetAmount:      1000.0,
			targetAmountPerKm: 5.0,
			distanceToCover:   10.0,
			startDuration:     "2024-01-01",
			endDuration:       "2024-01-31",
			expectedErr:       errors.New("repository error"),
			mockSetup: func(mockRepo *campaignMocks.MockCampaignRepository) {
				mockRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
			mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
			mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockCampaignRepo)
			}

			service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

			result, err := service.CreateCampaign(
				tt.ownerID, tt.ownerUsername, tt.nameArg, tt.description, tt.condition, tt.goal, tt.location,
				tt.mode, tt.activity, tt.targetAmount, tt.targetAmountPerKm, tt.distanceToCover,
				tt.startDuration, tt.endDuration,
			)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.ownerID, result.OwnerID)
				assert.Equal(t, tt.nameArg, result.Name)
				assert.Equal(t, tt.mode, result.Mode)
				assert.Equal(t, tt.activity, result.Activity)
			}

			mockCampaignRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignService_GetCampaignByID(t *testing.T) {
	tests := []struct {
		name        string
		campaignID  string
		expectedErr error
		mockSetup   func(*campaignMocks.MockCampaignRepository)
	}{
		{
			name:        "successful retrieval",
			campaignID:  "campaign123",
			expectedErr: nil,
			mockSetup: func(mockRepo *campaignMocks.MockCampaignRepository) {
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:    "Test Campaign",
					OwnerID: "owner123",
				}
				mockRepo.On("GetByID", "campaign123").Return(expectedCampaign, nil)
			},
		},
		{
			name:        "campaign not found",
			campaignID:  "nonexistent",
			expectedErr: errors.New("campaign not found"),
			mockSetup: func(mockRepo *campaignMocks.MockCampaignRepository) {
				mockRepo.On("GetByID", "nonexistent").Return(nil, errors.New("campaign not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
			mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
			mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockCampaignRepo)
			}

			service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

			result, err := service.GetCampaignByID(tt.campaignID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.campaignID, result.ID)
			}

			mockCampaignRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignService_GetCampaignBySlug(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	expectedCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "campaign123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:    "Test Campaign",
		Slug:    "test-campaign",
		OwnerID: "owner123",
	}

	mockCampaignRepo.On("GetBySlug", "test-campaign").Return(expectedCampaign, nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	result, err := service.GetCampaignBySlug("test-campaign")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-campaign", result.Slug)
	assert.Equal(t, "Test Campaign", result.Name)

	mockCampaignRepo.AssertExpectations(t)
}

func TestCampaignService_GetCampaignsByOwner(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	expectedCampaigns := []*campaignModel.Campaign{
		{
			Base: model.Base{
				ID:        "campaign1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Campaign 1",
			OwnerID: "owner123",
		},
		{
			Base: model.Base{
				ID:        "campaign2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Campaign 2",
			OwnerID: "owner123",
		},
	}

	mockCampaignRepo.On("GetByOwnerID", "owner123").Return(expectedCampaigns, nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	result, err := service.GetCampaignsByOwner("owner123")

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "campaign1", result[0].ID)
	assert.Equal(t, "campaign2", result[1].ID)

	mockCampaignRepo.AssertExpectations(t)
}

func TestCampaignService_JoinCampaign(t *testing.T) {
	tests := []struct {
		name        string
		campaignID  string
		userID      string
		expectedErr error
		mockSetup   func(*campaignMocks.MockCampaignRepository)
	}{
		{
			name:        "successful join",
			campaignID:  "campaign123",
			userID:      "user123",
			expectedErr: nil,
			mockSetup: func(mockRepo *campaignMocks.MockCampaignRepository) {
				mockRepo.On("IsMember", "campaign123", "user123").Return(false, nil)
				mockRepo.On("AddMember", "campaign123", "user123").Return(nil)
			},
		},
		{
			name:        "already a member",
			campaignID:  "campaign123",
			userID:      "user123",
			expectedErr: errors.New("user is already a member of this campaign"),
			mockSetup: func(mockRepo *campaignMocks.MockCampaignRepository) {
				mockRepo.On("IsMember", "campaign123", "user123").Return(true, nil)
			},
		},
		{
			name:        "repository error on IsMember",
			campaignID:  "campaign123",
			userID:      "user123",
			expectedErr: errors.New("repository error"),
			mockSetup: func(mockRepo *campaignMocks.MockCampaignRepository) {
				mockRepo.On("IsMember", "campaign123", "user123").Return(false, errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
			mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
			mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockCampaignRepo)
			}

			service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

			err := service.JoinCampaign(tt.campaignID, tt.userID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockCampaignRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignService_RecordActivity(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	existingCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "campaign123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:            "Test Campaign",
		DistanceCovered: 5.0,
	}

	mockCampaignRepo.On("GetByID", "campaign123").Return(existingCampaign, nil)
	mockRunnerRepo.On("Create", mock.MatchedBy(func(r *campaignModel.CampaignRunner) bool {
		return r.CampaignID == "campaign123" &&
			r.OwnerID == "user123" &&
			r.DistanceCovered == 10.0 &&
			r.Duration == "30:00" &&
			r.Activity == "Walking"
	})).Return(nil)
	mockCampaignRepo.On("Update", mock.MatchedBy(func(c *campaignModel.Campaign) bool {
		return c.ID == "campaign123" && c.DistanceCovered == 15.0
	})).Return(nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	err := service.RecordActivity("campaign123", "user123", 10.0, "30:00", "Walking")

	assert.NoError(t, err)

	mockCampaignRepo.AssertExpectations(t)
	mockRunnerRepo.AssertExpectations(t)
}

func TestCampaignService_SponsorCampaign(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	existingCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "campaign123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "Test Campaign",
		MoneyRaised: 100.0,
	}

	sponsors := []interface{}{"sponsor1", "sponsor2"}

	mockCampaignRepo.On("GetByID", "campaign123").Return(existingCampaign, nil)
	mockSponsorRepo.On("Create", mock.MatchedBy(func(s *campaignModel.SponsorCampaign) bool {
		return s.CampaignID == "campaign123" &&
			len(s.Sponsors) == 2 &&
			s.Distance == 10.0 &&
			s.AmountPerKm == 5.0 &&
			s.TotalAmount == 50.0
	})).Return(nil)
	mockCampaignRepo.On("Update", mock.MatchedBy(func(c *campaignModel.Campaign) bool {
		return c.ID == "campaign123" && c.MoneyRaised == 150.0
	})).Return(nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	err := service.SponsorCampaign("campaign123", sponsors, 10.0, 5.0)

	assert.NoError(t, err)

	mockCampaignRepo.AssertExpectations(t)
	mockSponsorRepo.AssertExpectations(t)
}

func TestCampaignService_ListCampaigns(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	expectedCampaigns := []*campaignModel.Campaign{
		{
			Base: model.Base{
				ID:        "campaign1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Campaign 1",
			OwnerID: "owner1",
		},
		{
			Base: model.Base{
				ID:        "campaign2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Campaign 2",
			OwnerID: "owner2",
		},
	}

	mockCampaignRepo.On("List", 10, 0).Return(expectedCampaigns, nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	result, err := service.ListCampaigns(10, 0)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "campaign1", result[0].ID)
	assert.Equal(t, "campaign2", result[1].ID)

	mockCampaignRepo.AssertExpectations(t)
}

func TestCampaignService_GetCampaignsByNonOwner(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	allCampaigns := []*campaignModel.Campaign{
		{
			Base: model.Base{
				ID:        "campaign1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Campaign 1",
			OwnerID: "owner123", // This is the user we're excluding
		},
		{
			Base: model.Base{
				ID:        "campaign2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Campaign 2",
			OwnerID: "owner456", // This should be included
		},
		{
			Base: model.Base{
				ID:        "campaign3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Campaign 3",
			OwnerID: "owner789", // This should be included
		},
	}

	mockCampaignRepo.On("List", 100, 0).Return(allCampaigns, nil) // Service multiplies limit by 10

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	result, err := service.GetCampaignsByNonOwner("owner123", 10, 0)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "campaign2", result[0].ID)
	assert.Equal(t, "campaign3", result[1].ID)

	mockCampaignRepo.AssertExpectations(t)
}

func TestCampaignService_IsMember(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	mockCampaignRepo.On("IsMember", "campaign123", "user123").Return(true, nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	result, err := service.IsMember("campaign123", "user123")

	assert.NoError(t, err)
	assert.True(t, result)

	mockCampaignRepo.AssertExpectations(t)
}

func TestCampaignService_AddMember(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	mockCampaignRepo.On("AddMember", "campaign123", "user123").Return(nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	err := service.AddMember("campaign123", "user123")

	assert.NoError(t, err)

	mockCampaignRepo.AssertExpectations(t)
}

func TestCampaignService_RemoveMember(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	mockCampaignRepo.On("RemoveMember", "campaign123", "user123").Return(nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	err := service.RemoveMember("campaign123", "user123")

	assert.NoError(t, err)

	mockCampaignRepo.AssertExpectations(t)
}

func TestCampaignService_UpdateCampaign(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	campaignToUpdate := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "campaign123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name: "Updated Campaign",
	}

	mockCampaignRepo.On("Update", campaignToUpdate).Return(nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	err := service.UpdateCampaign(campaignToUpdate)

	assert.NoError(t, err)

	mockCampaignRepo.AssertExpectations(t)
}

func TestCampaignService_DeleteCampaign(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	mockCampaignRepo.On("Delete", "campaign123").Return(nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	err := service.DeleteCampaign("campaign123")

	assert.NoError(t, err)

	mockCampaignRepo.AssertExpectations(t)
}

func TestCampaignService_GetLeaderboard(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	expectedCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "campaign123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name: "Test Campaign",
		Slug: "test-campaign",
	}

	expectedRunners := []*campaignModel.CampaignRunner{
		{
			Base: model.Base{
				ID:        "runner1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			CampaignID:      "campaign123",
			DistanceCovered: 10.0,
			OwnerID:         "user1",
		},
	}

	mockCampaignRepo.On("GetBySlug", "test-campaign").Return(expectedCampaign, nil)
	mockRunnerRepo.On("GetByCampaignID", "campaign123").Return(expectedRunners, nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	result, err := service.GetLeaderboard("test-campaign")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "runner1", result[0].ID)

	mockCampaignRepo.AssertExpectations(t)
	mockRunnerRepo.AssertExpectations(t)
}

func TestCampaignService_ParticipateCampaign(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	expectedCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "campaign123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name: "Test Campaign",
		Slug: "test-campaign",
	}

	mockCampaignRepo.On("GetBySlug", "test-campaign").Return(expectedCampaign, nil)
	mockCampaignRepo.On("IsMember", "campaign123", "user123").Return(false, nil)
	mockCampaignRepo.On("AddMember", "campaign123", "user123").Return(nil)
	mockRunnerRepo.On("Create", mock.MatchedBy(func(r *campaignModel.CampaignRunner) bool {
		return r.CampaignID == "campaign123" &&
			r.OwnerID == "user123" &&
			r.Activity == "Walking"
	})).Return(nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	result, err := service.ParticipateCampaign("test-campaign", "user123", "Walking")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "campaign123", result.CampaignID)
	assert.Equal(t, "user123", result.OwnerID)

	mockCampaignRepo.AssertExpectations(t)
	mockRunnerRepo.AssertExpectations(t)
}

func TestCampaignService_FinishActivity(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	existingRunner := &campaignModel.CampaignRunner{
		Base: model.Base{
			ID:        "runner123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		CampaignID:      "campaign123",
		DistanceCovered: 5.0,
		MoneyRaised:     10.0,
	}

	existingCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "campaign123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:            "Test Campaign",
		DistanceCovered: 10.0,
		MoneyRaised:     20.0,
	}

	mockRunnerRepo.On("GetByID", "runner123").Return(existingRunner, nil)
	mockRunnerRepo.On("Update", mock.MatchedBy(func(r *campaignModel.CampaignRunner) bool {
		return r.ID == "runner123" &&
			r.DistanceCovered == 15.0 &&
			r.Duration == "45:00" &&
			r.MoneyRaised == 25.0
	})).Return(nil)
	mockCampaignRepo.On("GetByID", "campaign123").Return(existingCampaign, nil)
	mockCampaignRepo.On("Update", mock.MatchedBy(func(c *campaignModel.Campaign) bool {
		return c.ID == "campaign123" &&
			c.DistanceCovered == 20.0 &&
			c.MoneyRaised == 35.0
	})).Return(nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	err := service.FinishActivity("runner123", 10.0, "45:00", 15.0)

	assert.NoError(t, err)

	mockRunnerRepo.AssertExpectations(t)
	mockCampaignRepo.AssertExpectations(t)
}

func TestCampaignService_GetRunnersByUser(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	expectedRunners := []*campaignModel.CampaignRunner{
		{
			Base: model.Base{
				ID:        "runner1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			CampaignID: "campaign123",
			OwnerID:    "user123",
		},
	}

	mockRunnerRepo.On("GetByOwnerID", "user123").Return(expectedRunners, nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	result, err := service.GetRunnersByUser("user123")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "runner1", result[0].ID)

	mockRunnerRepo.AssertExpectations(t)
}

func TestCampaignService_GetRunnerByID(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	expectedRunner := &campaignModel.CampaignRunner{
		Base: model.Base{
			ID:        "runner123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		CampaignID: "campaign123",
		OwnerID:    "user123",
	}

	mockRunnerRepo.On("GetByID", "runner123").Return(expectedRunner, nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	result, err := service.GetRunnerByID("runner123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "runner123", result.ID)

	mockRunnerRepo.AssertExpectations(t)
}

func TestCampaignService_UpdateRunner(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	runnerToUpdate := &campaignModel.CampaignRunner{
		Base: model.Base{
			ID:        "runner123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		CampaignID:      "campaign123",
		DistanceCovered: 10.0,
	}

	mockRunnerRepo.On("Update", runnerToUpdate).Return(nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	err := service.UpdateRunner(runnerToUpdate)

	assert.NoError(t, err)

	mockRunnerRepo.AssertExpectations(t)
}

func TestCampaignService_DeleteRunner(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	mockRunnerRepo.On("Delete", "runner123").Return(nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	err := service.DeleteRunner("runner123")

	assert.NoError(t, err)

	mockRunnerRepo.AssertExpectations(t)
}

func TestCampaignService_CreateSponsorCampaign(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	existingCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "campaign123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "Test Campaign",
		MoneyRaised: 100.0,
	}

	sponsors := []interface{}{"sponsor1", "sponsor2"}

	mockCampaignRepo.On("GetByID", "campaign123").Return(existingCampaign, nil)
	mockSponsorRepo.On("Create", mock.MatchedBy(func(s *campaignModel.SponsorCampaign) bool {
		return s.CampaignID == "campaign123" &&
			len(s.Sponsors) == 2 &&
			s.Distance == 10.0 &&
			s.AmountPerKm == 5.0 &&
			s.TotalAmount == 50.0
	})).Return(nil)
	mockCampaignRepo.On("Update", mock.MatchedBy(func(c *campaignModel.Campaign) bool {
		return c.ID == "campaign123" && c.MoneyRaised == 150.0
	})).Return(nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	result, err := service.CreateSponsorCampaign("campaign123", sponsors, 10.0, 5.0, "brand.jpg", "video.mp4")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "campaign123", result.CampaignID)
	assert.Equal(t, 50.0, result.TotalAmount)

	mockCampaignRepo.AssertExpectations(t)
	mockSponsorRepo.AssertExpectations(t)
}

func TestCampaignService_GetSponsorCampaignByID(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	expectedSponsor := &campaignModel.SponsorCampaign{
		Base: model.Base{
			ID:        "sponsor123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		CampaignID:  "campaign123",
		TotalAmount: 100.0,
	}

	mockSponsorRepo.On("GetByID", "sponsor123").Return(expectedSponsor, nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	result, err := service.GetSponsorCampaignByID("sponsor123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "sponsor123", result.ID)

	mockSponsorRepo.AssertExpectations(t)
}

func TestCampaignService_GetSponsorCampaignsByCampaign(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	expectedSponsors := []*campaignModel.SponsorCampaign{
		{
			Base: model.Base{
				ID:        "sponsor1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			CampaignID:  "campaign123",
			TotalAmount: 50.0,
		},
	}

	mockSponsorRepo.On("GetByCampaignID", "campaign123").Return(expectedSponsors, nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	result, err := service.GetSponsorCampaignsByCampaign("campaign123")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "sponsor1", result[0].ID)

	mockSponsorRepo.AssertExpectations(t)
}

func TestCampaignService_UpdateSponsorCampaign(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	sponsorToUpdate := &campaignModel.SponsorCampaign{
		Base: model.Base{
			ID:        "sponsor123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		CampaignID:  "campaign123",
		TotalAmount: 150.0,
	}

	mockSponsorRepo.On("Update", sponsorToUpdate).Return(nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	err := service.UpdateSponsorCampaign(sponsorToUpdate)

	assert.NoError(t, err)

	mockSponsorRepo.AssertExpectations(t)
}

func TestCampaignService_DeleteSponsorCampaign(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	mockSponsorRepo.On("Delete", "sponsor123").Return(nil)

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	err := service.DeleteSponsorCampaign("sponsor123")

	assert.NoError(t, err)

	mockSponsorRepo.AssertExpectations(t)
}

func TestCampaignService_SearchCampaigns(t *testing.T) {
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	allCampaigns := []*campaignModel.Campaign{
		{
			Base: model.Base{
				ID:        "campaign1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:        "Golang Tutorial",
			Description: "Learn Golang programming",
		},
		{
			Base: model.Base{
				ID:        "campaign2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:        "Python Guide",
			Description: "Python programming basics",
		},
	}

	mockCampaignRepo.On("List", 20, 0).Return(allCampaigns, nil) // Service multiplies limit by 2

	service := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)

	result, err := service.SearchCampaigns("python", 10, 0)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "campaign2", result[0].ID)
	assert.Contains(t, result[0].Description, "Python")

	mockCampaignRepo.AssertExpectations(t)
}
