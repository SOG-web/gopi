package campaign_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopi.com/api/http/dto"
	"gopi.com/api/http/handler"
	"gopi.com/internal/app/campaign"
	"gopi.com/internal/app/user"
	campaignModel "gopi.com/internal/domain/campaign/model"
	"gopi.com/internal/domain/model"
	userModel "gopi.com/internal/domain/user/model"
	campaignMocks "gopi.com/tests/mocks/campaign"
	userMocks "gopi.com/tests/mocks/user"
)

// Test setup helper
func setupCampaignActivityTest(t *testing.T) (*gin.Engine, *campaignMocks.MockCampaignRepository, *campaignMocks.MockCampaignRunnerRepository, *campaignMocks.MockSponsorCampaignRepository, *userMocks.MockUserRepository) {
	gin.SetMode(gin.TestMode)

	// Create mock repositories
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)
	mockUserRepo := new(userMocks.MockUserRepository)

	// Create services with mock repositories
	campaignService := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)
	userService := user.NewUserService(mockUserRepo, nil)

	// Create handler with services
	campaignHandler := handler.NewCampaignHandler(campaignService, userService)

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup routes
	protected := router.Group("/campaigns")
	protected.Use(func(c *gin.Context) {
		// Mock auth middleware - set user_id in context
		c.Set("user_id", "test-user-id")
		c.Next()
	})

	// Activity routes
	protected.GET("/:slug/finish_campaign/:runner_id", campaignHandler.GetFinishCampaignDetails)
	protected.PUT("/:slug/finish_campaign/:runner_id", campaignHandler.FinishCampaignRun)

	return router, mockCampaignRepo, mockRunnerRepo, mockSponsorRepo, mockUserRepo
}

func TestCampaignHandler_GetFinishCampaignDetails(t *testing.T) {
	router, mockCampaignRepo, mockRunnerRepo, _, mockUserRepo := setupCampaignActivityTest(t)

	tests := []struct {
		name           string
		slug           string
		runnerID       string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful finish details retrieval",
			slug:           "test-campaign-slug",
			runnerID:       "runner123",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:    "Test Campaign",
					Slug:    "test-campaign-slug",
					OwnerID: "campaign-owner",
				}

				expectedRunner := &campaignModel.CampaignRunner{
					Base: model.Base{
						ID:        "runner123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:      "campaign123",
					DistanceCovered: 5.0,
					Duration:        "25:00",
					MoneyRaised:     10.0,
					Activity:        "Running",
					OwnerID:         "test-user-id",
					DateJoined:      time.Now(),
					CoverImage:      "runner.jpg",
				}

				runnerOwner := &userModel.User{
					Base: model.Base{
						ID:        "test-user-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:  "testuser",
					FirstName: "Test",
					LastName:  "User",
				}

				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockRunnerRepo.On("GetByID", "runner123").Return(expectedRunner, nil)
				mockUserRepo.On("GetByID", "test-user-id").Return(runnerOwner, nil)
			},
		},
		{
			name:           "campaign not found",
			slug:           "nonexistent-slug",
			runnerID:       "runner123",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockCampaignRepo.On("GetBySlug", "nonexistent-slug").Return(nil, errors.New("campaign not found"))
			},
		},
		{
			name:           "runner not found",
			slug:           "test-campaign-slug",
			runnerID:       "nonexistent-runner",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:    "Test Campaign",
					Slug:    "test-campaign-slug",
					OwnerID: "campaign-owner",
				}

				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockRunnerRepo.On("GetByID", "nonexistent-runner").Return(nil, errors.New("runner not found"))
			},
		},
		{
			name:           "unauthorized - different runner owner",
			slug:           "test-campaign-slug",
			runnerID:       "runner456",
			expectedStatus: http.StatusForbidden,
			mockSetup: func() {
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:    "Test Campaign",
					Slug:    "test-campaign-slug",
					OwnerID: "campaign-owner",
				}

				expectedRunner := &campaignModel.CampaignRunner{
					Base: model.Base{
						ID:        "runner456",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:      "campaign123",
					DistanceCovered: 3.0,
					Duration:        "15:00",
					MoneyRaised:     5.0,
					Activity:        "Walking",
					OwnerID:         "different-user", // Different from test-user-id
					DateJoined:      time.Now(),
				}

				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockRunnerRepo.On("GetByID", "runner456").Return(expectedRunner, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCampaignRepo.ExpectedCalls = nil
			mockRunnerRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodGet, "/campaigns/"+tt.slug+"/finish_campaign/"+tt.runnerID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCampaignRepo.AssertExpectations(t)
			mockRunnerRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignHandler_FinishCampaignRun(t *testing.T) {
	router, mockCampaignRepo, mockRunnerRepo, _, mockUserRepo := setupCampaignActivityTest(t)

	tests := []struct {
		name           string
		slug           string
		runnerID       string
		requestBody    dto.FinishActivityRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:     "successful activity finish",
			slug:     "test-campaign-slug",
			runnerID: "runner123",
			requestBody: dto.FinishActivityRequest{
				DistanceCovered: 10.5,
				Duration:        "35:20",
				MoneyRaised:     15.0,
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:            "Test Campaign",
					Slug:            "test-campaign-slug",
					OwnerID:         "campaign-owner",
					DistanceCovered: 25.0,
					MoneyRaised:     50.0,
				}

				expectedRunner := &campaignModel.CampaignRunner{
					Base: model.Base{
						ID:        "runner123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:      "campaign123",
					DistanceCovered: 5.0,
					Duration:        "25:00",
					MoneyRaised:     10.0,
					Activity:        "Running",
					OwnerID:         "test-user-id",
					DateJoined:      time.Now(),
				}

				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockCampaignRepo.On("GetByID", "campaign123").Return(expectedCampaign, nil)
				mockRunnerRepo.On("GetByID", "runner123").Return(expectedRunner, nil)
				mockRunnerRepo.On("Update", mock.MatchedBy(func(r *campaignModel.CampaignRunner) bool {
					return r.ID == "runner123" &&
						r.DistanceCovered == 10.5 &&
						r.Duration == "35:20" &&
						r.MoneyRaised == 15.0
				})).Return(nil)
				mockCampaignRepo.On("Update", mock.MatchedBy(func(c *campaignModel.Campaign) bool {
					return c.ID == "campaign123" &&
						c.DistanceCovered == 35.5 &&
						c.MoneyRaised == 65.0
				})).Return(nil)

				// Mock user service for response
				runnerOwner := &userModel.User{
					Base: model.Base{
						ID:        "test-user-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:  "testuser",
					FirstName: "Test",
					LastName:  "User",
				}
				mockUserRepo.On("GetByID", "test-user-id").Return(runnerOwner, nil)
			},
		},
		{
			name:     "finish with zero money raised",
			slug:     "test-campaign-slug",
			runnerID: "runner123",
			requestBody: dto.FinishActivityRequest{
				DistanceCovered: 8.0,
				Duration:        "28:15",
				MoneyRaised:     0,
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:            "Test Campaign",
					Slug:            "test-campaign-slug",
					OwnerID:         "campaign-owner",
					DistanceCovered: 15.0,
					MoneyRaised:     30.0,
				}

				expectedRunner := &campaignModel.CampaignRunner{
					Base: model.Base{
						ID:        "runner123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:      "campaign123",
					DistanceCovered: 3.0,
					Duration:        "12:00",
					MoneyRaised:     0,
					Activity:        "Walking",
					OwnerID:         "test-user-id",
					DateJoined:      time.Now(),
				}

				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockCampaignRepo.On("GetByID", "campaign123").Return(expectedCampaign, nil)
				mockRunnerRepo.On("GetByID", "runner123").Return(expectedRunner, nil)
				mockRunnerRepo.On("Update", mock.MatchedBy(func(r *campaignModel.CampaignRunner) bool {
					return r.ID == "runner123" &&
						r.DistanceCovered == 8.0 &&
						r.Duration == "28:15" &&
						r.MoneyRaised == 0
				})).Return(nil)
				mockCampaignRepo.On("Update", mock.MatchedBy(func(c *campaignModel.Campaign) bool {
					return c.ID == "campaign123" &&
						c.DistanceCovered == 23.0 &&
						c.MoneyRaised == 30.0 // No change since MoneyRaised is 0
				})).Return(nil)

				// Mock user service for response
				runnerOwner := &userModel.User{
					Base: model.Base{
						ID:        "test-user-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:  "testuser",
					FirstName: "Test",
					LastName:  "User",
				}
				mockUserRepo.On("GetByID", "test-user-id").Return(runnerOwner, nil)
			},
		},
		{
			name:     "invalid request - missing distance",
			slug:     "test-campaign-slug",
			runnerID: "runner123",
			requestBody: dto.FinishActivityRequest{
				DistanceCovered: 0,
				Duration:        "30:00",
				MoneyRaised:     10.0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// No mocks needed - validation happens before service calls
			},
		},
		{
			name:     "invalid request - missing duration",
			slug:     "test-campaign-slug",
			runnerID: "runner123",
			requestBody: dto.FinishActivityRequest{
				DistanceCovered: 5.0,
				Duration:        "",
				MoneyRaised:     10.0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// No mocks needed - validation happens before service calls
			},
		},
		{
			name:     "campaign not found",
			slug:     "nonexistent-slug",
			runnerID: "runner123",
			requestBody: dto.FinishActivityRequest{
				DistanceCovered: 5.0,
				Duration:        "20:00",
				MoneyRaised:     8.0,
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockCampaignRepo.On("GetBySlug", "nonexistent-slug").Return(nil, errors.New("campaign not found"))
			},
		},
		{
			name:     "runner not found",
			slug:     "test-campaign-slug",
			runnerID: "nonexistent-runner",
			requestBody: dto.FinishActivityRequest{
				DistanceCovered: 5.0,
				Duration:        "20:00",
				MoneyRaised:     8.0,
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:    "Test Campaign",
					Slug:    "test-campaign-slug",
					OwnerID: "campaign-owner",
				}

				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockRunnerRepo.On("GetByID", "nonexistent-runner").Return(nil, errors.New("runner not found"))
			},
		},
		{
			name:     "unauthorized - different runner owner",
			slug:     "test-campaign-slug",
			runnerID: "runner456",
			requestBody: dto.FinishActivityRequest{
				DistanceCovered: 5.0,
				Duration:        "20:00",
				MoneyRaised:     8.0,
			},
			expectedStatus: http.StatusForbidden,
			mockSetup: func() {
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:    "Test Campaign",
					Slug:    "test-campaign-slug",
					OwnerID: "campaign-owner",
				}

				expectedRunner := &campaignModel.CampaignRunner{
					Base: model.Base{
						ID:        "runner456",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:      "campaign123",
					DistanceCovered: 3.0,
					Duration:        "15:00",
					MoneyRaised:     5.0,
					Activity:        "Walking",
					OwnerID:         "different-user", // Different from test-user-id
					DateJoined:      time.Now(),
				}

				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockRunnerRepo.On("GetByID", "runner456").Return(expectedRunner, nil)
			},
		},
		{
			name:     "repository error on runner update",
			slug:     "test-campaign-slug",
			runnerID: "runner123",
			requestBody: dto.FinishActivityRequest{
				DistanceCovered: 12.0,
				Duration:        "40:00",
				MoneyRaised:     20.0,
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:    "Test Campaign",
					Slug:    "test-campaign-slug",
					OwnerID: "campaign-owner",
				}

				expectedRunner := &campaignModel.CampaignRunner{
					Base: model.Base{
						ID:        "runner123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:      "campaign123",
					DistanceCovered: 5.0,
					Duration:        "25:00",
					MoneyRaised:     10.0,
					Activity:        "Running",
					OwnerID:         "test-user-id",
					DateJoined:      time.Now(),
				}

				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockRunnerRepo.On("GetByID", "runner123").Return(expectedRunner, nil)
				mockRunnerRepo.On("Update", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCampaignRepo.ExpectedCalls = nil
			mockRunnerRepo.ExpectedCalls = nil

			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/campaigns/"+tt.slug+"/finish_campaign/"+tt.runnerID, bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCampaignRepo.AssertExpectations(t)
			mockRunnerRepo.AssertExpectations(t)
		})
	}
}
