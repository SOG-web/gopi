package campaign_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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
func setupCampaignLeaderboardTest(t *testing.T) (*gin.Engine, *campaignMocks.MockCampaignRepository, *campaignMocks.MockCampaignRunnerRepository, *campaignMocks.MockSponsorCampaignRepository, *userMocks.MockUserRepository) {
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

	// Leaderboard routes
	protected.GET("/:slug/leaderboard", campaignHandler.GetCampaignLeaderboard)

	return router, mockCampaignRepo, mockRunnerRepo, mockSponsorRepo, mockUserRepo
}

func TestCampaignHandler_GetCampaignLeaderboard(t *testing.T) {
	router, mockCampaignRepo, mockRunnerRepo, _, mockUserRepo := setupCampaignLeaderboardTest(t)

	tests := []struct {
		name           string
		slug           string
		expectedStatus int
		mockSetup      func()
		expectedCount  int
	}{
		{
			name:           "successful leaderboard retrieval",
			slug:           "test-campaign-slug",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
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

				expectedRunners := []*campaignModel.CampaignRunner{
					{
						Base: model.Base{
							ID:        "runner1",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						CampaignID:      "campaign123",
						DistanceCovered: 15.5,
						Duration:        "45:30",
						MoneyRaised:     25.0,
						Activity:        "Running",
						OwnerID:         "user1",
						DateJoined:      time.Now(),
						CoverImage:      "runner1.jpg",
					},
					{
						Base: model.Base{
							ID:        "runner2",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						CampaignID:      "campaign123",
						DistanceCovered: 12.0,
						Duration:        "38:15",
						MoneyRaised:     18.0,
						Activity:        "Walking",
						OwnerID:         "user2",
						DateJoined:      time.Now(),
						CoverImage:      "runner2.jpg",
					},
					{
						Base: model.Base{
							ID:        "runner3",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						CampaignID:      "campaign123",
						DistanceCovered: 8.5,
						Duration:        "28:45",
						MoneyRaised:     12.0,
						Activity:        "Cycling",
						OwnerID:         "user3",
						DateJoined:      time.Now(),
						CoverImage:      "",
					},
				}

				// Mock user lookups for runner details
				for _, runner := range expectedRunners {
					user := &userModel.User{
						Base: model.Base{
							ID:        runner.OwnerID,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Username:        runner.OwnerID + "_username",
						FirstName:       runner.OwnerID + "_first",
						LastName:        runner.OwnerID + "_last",
						ProfileImageURL: runner.CoverImage,
					}
					mockUserRepo.On("GetByID", runner.OwnerID).Return(user, nil)
				}

				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockRunnerRepo.On("GetByCampaignID", "campaign123").Return(expectedRunners, nil)
			},
		},
		{
			name:           "campaign not found",
			slug:           "nonexistent-slug",
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
			mockSetup: func() {
				mockCampaignRepo.On("GetBySlug", "nonexistent-slug").Return(nil, assert.AnError)
			},
		},
		{
			name:           "empty leaderboard",
			slug:           "empty-campaign-slug",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			mockSetup: func() {
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign456",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:    "Empty Campaign",
					Slug:    "empty-campaign-slug",
					OwnerID: "campaign-owner",
				}

				mockCampaignRepo.On("GetBySlug", "empty-campaign-slug").Return(expectedCampaign, nil)
				mockRunnerRepo.On("GetByCampaignID", "campaign456").Return([]*campaignModel.CampaignRunner{}, nil)
			},
		},
		{
			name:           "repository error on runners",
			slug:           "error-campaign-slug",
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
			mockSetup: func() {
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign789",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:    "Error Campaign",
					Slug:    "error-campaign-slug",
					OwnerID: "campaign-owner",
				}

				mockCampaignRepo.On("GetBySlug", "error-campaign-slug").Return(expectedCampaign, nil)
				mockRunnerRepo.On("GetByCampaignID", "campaign789").Return(nil, assert.AnError)
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

			req, _ := http.NewRequest(http.MethodGet, "/campaigns/"+tt.slug+"/leaderboard", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK && tt.expectedCount > 0 {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Check that leaderboard exists and has expected count
				leaderboard, exists := response["leaderboard"]
				assert.True(t, exists, "leaderboard should exist in response")

				leaderboardSlice, ok := leaderboard.([]interface{})
				assert.True(t, ok, "leaderboard should be an array")
				assert.Len(t, leaderboardSlice, tt.expectedCount, "leaderboard should have expected number of entries")
			}

			mockCampaignRepo.AssertExpectations(t)
			mockRunnerRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}
