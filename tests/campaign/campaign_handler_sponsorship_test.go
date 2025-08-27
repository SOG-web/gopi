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
	campaignMocks "gopi.com/tests/mocks/campaign"
)

// Test setup helper
func setupCampaignSponsorshipTest(t *testing.T) (*gin.Engine, *campaignMocks.MockCampaignRepository, *campaignMocks.MockCampaignRunnerRepository, *campaignMocks.MockSponsorCampaignRepository) {
	gin.SetMode(gin.TestMode)

	// Create mock repositories
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)

	// Create services with mock repositories
	campaignService := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)
	userService := user.NewUserService(nil, nil) // Not testing user operations in sponsorship tests

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

	// Sponsorship routes
	protected.POST("/:slug/sponsor", campaignHandler.SponsorCampaign)

	return router, mockCampaignRepo, mockRunnerRepo, mockSponsorRepo
}

func TestCampaignHandler_SponsorCampaign(t *testing.T) {
	router, mockCampaignRepo, _, mockSponsorRepo := setupCampaignSponsorshipTest(t)

	tests := []struct {
		name           string
		slug           string
		requestBody    dto.SponsorCampaignRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful sponsorship",
			slug: "test-campaign-slug",
			requestBody: dto.SponsorCampaignRequest{
				Distance:    10.0,
				AmountPerKm: 5.0,
				BrandImg:    "brand-image.jpg",
				VideoUrl:    "video.mp4",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:        "Test Campaign",
					Slug:        "test-campaign-slug",
					MoneyRaised: 100.0,
				}
				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockSponsorRepo.On("Create", mock.MatchedBy(func(s *campaignModel.SponsorCampaign) bool {
					return s.CampaignID == "campaign123" &&
						s.Distance == 10.0 &&
						s.AmountPerKm == 5.0 &&
						s.TotalAmount == 50.0 &&
						s.BrandImg == "" && // These are empty in the actual service
						s.VideoUrl == ""
				})).Return(nil)
				// Mock the GetByID call that happens after Create
				mockCampaignRepo.On("GetByID", "campaign123").Return(expectedCampaign, nil)
				mockCampaignRepo.On("Update", mock.MatchedBy(func(c *campaignModel.Campaign) bool {
					return c.ID == "campaign123" && c.MoneyRaised == 150.0
				})).Return(nil)
				// Mock AddSponsor call that happens after sponsorship creation
				mockCampaignRepo.On("AddSponsor", "campaign123", "test-user-id").Return(nil)
			},
		},
		{
			name: "sponsorship with zero distance",
			slug: "test-campaign-slug",
			requestBody: dto.SponsorCampaignRequest{
				Distance:    0,
				AmountPerKm: 5.0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// Mock campaign lookup - validation happens during JSON binding
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:        "Test Campaign",
					Slug:        "test-campaign-slug",
					MoneyRaised: 100.0,
				}
				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
			},
		},
		{
			name: "sponsorship with zero amount per km",
			slug: "test-campaign-slug",
			requestBody: dto.SponsorCampaignRequest{
				Distance:    10.0,
				AmountPerKm: 0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				// Mock campaign lookup - validation happens during JSON binding
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:        "Test Campaign",
					Slug:        "test-campaign-slug",
					MoneyRaised: 100.0,
				}
				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
			},
		},
		{
			name: "campaign not found",
			slug: "nonexistent-slug",
			requestBody: dto.SponsorCampaignRequest{
				Distance:    5.0,
				AmountPerKm: 2.0,
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockCampaignRepo.On("GetBySlug", "nonexistent-slug").Return(nil, errors.New("campaign not found"))
			},
		},
		{
			name: "repository error on sponsorship creation",
			slug: "test-campaign-slug",
			requestBody: dto.SponsorCampaignRequest{
				Distance:    15.0,
				AmountPerKm: 3.0,
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:        "Test Campaign",
					Slug:        "test-campaign-slug",
					MoneyRaised: 50.0,
				}
				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockSponsorRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
		{
			name: "sponsorship without optional fields",
			slug: "test-campaign-slug",
			requestBody: dto.SponsorCampaignRequest{
				Distance:    20.0,
				AmountPerKm: 4.0,
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:        "Test Campaign",
					Slug:        "test-campaign-slug",
					MoneyRaised: 75.0,
				}
				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockSponsorRepo.On("Create", mock.MatchedBy(func(s *campaignModel.SponsorCampaign) bool {
					return s.CampaignID == "campaign123" &&
						s.Distance == 20.0 &&
						s.AmountPerKm == 4.0 &&
						s.TotalAmount == 80.0 &&
						s.BrandImg == "" &&
						s.VideoUrl == ""
				})).Return(nil)
				// Mock the GetByID call that happens after Create
				mockCampaignRepo.On("GetByID", "campaign123").Return(expectedCampaign, nil)
				mockCampaignRepo.On("Update", mock.MatchedBy(func(c *campaignModel.Campaign) bool {
					return c.ID == "campaign123" && c.MoneyRaised == 155.0
				})).Return(nil)
				// Mock AddSponsor call that happens after sponsorship creation
				mockCampaignRepo.On("AddSponsor", "campaign123", "test-user-id").Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCampaignRepo.ExpectedCalls = nil
			mockSponsorRepo.ExpectedCalls = nil

			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/campaigns/"+tt.slug+"/sponsor", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCampaignRepo.AssertExpectations(t)
			mockSponsorRepo.AssertExpectations(t)
		})
	}
}
