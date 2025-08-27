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
func setupCampaignTest(t *testing.T) (*gin.Engine, *campaignMocks.MockCampaignRepository, *campaignMocks.MockCampaignRunnerRepository, *campaignMocks.MockSponsorCampaignRepository, *userMocks.MockUserRepository) {
	gin.SetMode(gin.TestMode)

	// Create mock repositories
	mockCampaignRepo := new(campaignMocks.MockCampaignRepository)
	mockRunnerRepo := new(campaignMocks.MockCampaignRunnerRepository)
	mockSponsorRepo := new(campaignMocks.MockSponsorCampaignRepository)
	mockUserRepo := new(userMocks.MockUserRepository)

	// Create services with mock repositories
	campaignService := campaign.NewCampaignService(mockCampaignRepo, mockRunnerRepo, mockSponsorRepo)
	userService := user.NewUserService(mockUserRepo, nil) // We'll mock email service separately if needed

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

	protected.POST("", campaignHandler.CreateCampaign)
	protected.GET("", campaignHandler.GetCampaigns)
	protected.PUT("/:slug", campaignHandler.UpdateCampaign)
	protected.DELETE("/:slug", campaignHandler.DeleteCampaign)
	protected.POST("/:slug/join", campaignHandler.JoinCampaign)
	protected.POST("/:slug/sponsor", campaignHandler.SponsorCampaign)
	protected.POST("/:slug/participate", campaignHandler.ParticipateCampaign)

	// Public routes (no auth required)
	router.GET("/campaigns/:slug", campaignHandler.GetCampaignBySlug)

	return router, mockCampaignRepo, mockRunnerRepo, mockSponsorRepo, mockUserRepo
}

func TestCampaignHandler_CreateCampaign(t *testing.T) {
	router, mockCampaignRepo, _, _, mockUserRepo := setupCampaignTest(t)

	tests := []struct {
		name           string
		requestBody    dto.CreateCampaignRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful campaign creation",
			requestBody: dto.CreateCampaignRequest{
				Name:              "Test Campaign",
				Description:       "A test campaign",
				Condition:         "Good condition",
				Mode:              "Free",
				Goal:              "Test goal",
				Activity:          "Walking",
				Location:          "Test location",
				TargetAmount:      1000.0,
				TargetAmountPerKm: 5.0,
				DistanceToCover:   10.0,
				StartDuration:     "2024-01-01",
				EndDuration:       "2024-01-31",
			},
			expectedStatus: http.StatusCreated,
			mockSetup: func() {
				// Mock user lookup
				testUser := &userModel.User{
					Base: model.Base{
						ID:        "test-user-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username: "testuser",
					Email:    "test@example.com",
				}
				mockUserRepo.On("GetByID", "test-user-id").Return(testUser, nil)

				// Mock campaign creation
				mockCampaignRepo.On("Create", mock.MatchedBy(func(c *campaignModel.Campaign) bool {
					return c.OwnerID == "test-user-id" &&
						c.Name == "Test Campaign" &&
						c.Mode == campaignModel.CampaignModeFree &&
						c.Activity == campaignModel.ActivityWalking
				})).Return(nil)
			},
		},
		{
			name: "invalid request body",
			requestBody: dto.CreateCampaignRequest{
				Name: "", // Invalid: empty name
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "user not authenticated",
			requestBody: dto.CreateCampaignRequest{
				Name: "Test Campaign",
			},
			expectedStatus: http.StatusUnauthorized,
			mockSetup:      func() {},
		},
		{
			name: "repository error",
			requestBody: dto.CreateCampaignRequest{
				Name:              "Test Campaign",
				Description:       "A test campaign",
				Condition:         "Good condition",
				Mode:              "Free",
				Goal:              "Test goal",
				Activity:          "Walking",
				Location:          "Test location",
				TargetAmount:      1000.0,
				TargetAmountPerKm: 5.0,
				DistanceToCover:   10.0,
				StartDuration:     "2024-01-01",
				EndDuration:       "2024-01-31",
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				// Mock user lookup
				testUser := &userModel.User{
					Base: model.Base{
						ID:        "test-user-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username: "testuser",
					Email:    "test@example.com",
				}
				mockUserRepo.On("GetByID", "test-user-id").Return(testUser, nil)

				// Mock campaign creation error
				mockCampaignRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "user not authenticated" {
				// Test without auth middleware
				router := gin.New()
				router.Use(gin.Recovery())
				campaignHandler := &handler.CampaignHandler{}

				router.POST("/campaigns", campaignHandler.CreateCampaign)

				req, _ := http.NewRequest(http.MethodPost, "/campaigns", bytes.NewBuffer([]byte(`{"name":"test"}`)))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.expectedStatus, w.Code)
			} else {
				// Clear previous mock expectations
				mockCampaignRepo.ExpectedCalls = nil
				mockUserRepo.ExpectedCalls = nil

				tt.mockSetup()

				requestBody, _ := json.Marshal(tt.requestBody)
				req, _ := http.NewRequest(http.MethodPost, "/campaigns", bytes.NewBuffer(requestBody))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.expectedStatus, w.Code)
				mockCampaignRepo.AssertExpectations(t)
				mockUserRepo.AssertExpectations(t)
			}
		})
	}
}

func TestCampaignHandler_GetCampaignBySlug(t *testing.T) {
	router, mockCampaignRepo, _, _, mockUserRepo := setupCampaignTest(t)

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

	expectedOwner := &userModel.User{
		Base: model.Base{
			ID:        "owner123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username: "testowner",
		Email:    "owner@example.com",
	}

	mockCampaignRepo.On("GetBySlug", "test-campaign").Return(expectedCampaign, nil)
	mockUserRepo.On("GetByID", "owner123").Return(expectedOwner, nil)

	req, _ := http.NewRequest(http.MethodGet, "/campaigns/test-campaign", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockCampaignRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestCampaignHandler_GetCampaigns(t *testing.T) {
	router, mockCampaignRepo, _, _, mockUserRepo := setupCampaignTest(t)

	expectedCampaigns := []*campaignModel.Campaign{
		{
			Base: model.Base{
				ID:        "campaign1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Campaign 1",
			OwnerID: "owner1",
			Slug:    "campaign-1",
		},
		{
			Base: model.Base{
				ID:        "campaign2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Campaign 2",
			OwnerID: "owner2",
			Slug:    "campaign-2",
		},
	}

	mockCampaignRepo.On("List", 10, 0).Return(expectedCampaigns, nil)

	// Mock user lookups for each campaign owner
	owner1 := &userModel.User{
		Base: model.Base{
			ID:        "owner1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username: "owner1",
		Email:    "owner1@example.com",
	}
	owner2 := &userModel.User{
		Base: model.Base{
			ID:        "owner2",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username: "owner2",
		Email:    "owner2@example.com",
	}
	mockUserRepo.On("GetByID", "owner1").Return(owner1, nil)
	mockUserRepo.On("GetByID", "owner2").Return(owner2, nil)

	req, _ := http.NewRequest(http.MethodGet, "/campaigns", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockCampaignRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestCampaignHandler_UpdateCampaign(t *testing.T) {
	router, mockCampaignRepo, _, _, mockUserRepo := setupCampaignTest(t)

	tests := []struct {
		name           string
		campaignSlug   string
		requestBody    dto.UpdateCampaignRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:         "successful campaign update",
			campaignSlug: "original-campaign",
			requestBody: dto.UpdateCampaignRequest{
				Name:        "Updated Campaign",
				Description: "Updated description",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				existingCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:        "Original Campaign",
					Description: "Original description",
					OwnerID:     "test-user-id",
					Slug:        "original-campaign",
				}
				mockCampaignRepo.On("GetBySlug", "original-campaign").Return(existingCampaign, nil)
				mockCampaignRepo.On("Update", mock.MatchedBy(func(c *campaignModel.Campaign) bool {
					return c.ID == "campaign123" && c.Name == "Updated Campaign"
				})).Return(nil)

				// Mock user lookup for response
				owner := &userModel.User{
					Base: model.Base{
						ID:        "test-user-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username: "testuser",
					Email:    "test@example.com",
				}
				mockUserRepo.On("GetByID", "test-user-id").Return(owner, nil)
			},
		},
		{
			name:         "campaign not found",
			campaignSlug: "nonexistent",
			requestBody: dto.UpdateCampaignRequest{
				Name: "Updated Name",
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockCampaignRepo.On("GetBySlug", "nonexistent").Return(nil, errors.New("campaign not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous mock expectations
			mockCampaignRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/campaigns/"+tt.campaignSlug, bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCampaignRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignHandler_DeleteCampaign(t *testing.T) {
	router, mockCampaignRepo, _, _, mockUserRepo := setupCampaignTest(t)

	tests := []struct {
		name           string
		campaignSlug   string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful campaign deletion",
			campaignSlug:   "test-campaign",
			expectedStatus: http.StatusNoContent,
			mockSetup: func() {
				existingCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:    "Test Campaign",
					OwnerID: "test-user-id",
					Slug:    "test-campaign",
				}
				mockCampaignRepo.On("GetBySlug", "test-campaign").Return(existingCampaign, nil)
				mockCampaignRepo.On("Delete", "campaign123").Return(nil)
			},
		},
		{
			name:           "campaign not found",
			campaignSlug:   "nonexistent",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockCampaignRepo.On("GetBySlug", "nonexistent").Return(nil, errors.New("campaign not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous mock expectations
			mockCampaignRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodDelete, "/campaigns/"+tt.campaignSlug, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCampaignRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}
