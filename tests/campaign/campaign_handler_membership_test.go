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
func setupCampaignMembershipTest(t *testing.T) (*gin.Engine, *campaignMocks.MockCampaignRepository, *campaignMocks.MockCampaignRunnerRepository, *campaignMocks.MockSponsorCampaignRepository, *userMocks.MockUserRepository) {
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

	// Membership and participation routes
	protected.PUT("/:slug/join", campaignHandler.JoinCampaign)
	protected.POST("/:slug/participate", campaignHandler.ParticipateCampaign)
	protected.GET("/by_user", campaignHandler.GetCampaignsByUser)
	protected.GET("/by_others", campaignHandler.GetCampaignsByOthers)

	return router, mockCampaignRepo, mockRunnerRepo, mockSponsorRepo, mockUserRepo
}

func TestCampaignHandler_JoinCampaign(t *testing.T) {
	router, mockCampaignRepo, _, _, _ := setupCampaignMembershipTest(t)

	tests := []struct {
		name           string
		slug           string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful join",
			slug:           "test-campaign-slug",
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
					OwnerID: "other-user",
				}
				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockCampaignRepo.On("IsMember", "campaign123", "test-user-id").Return(false, nil)
				mockCampaignRepo.On("AddMember", "campaign123", "test-user-id").Return(nil)
			},
		},
		{
			name:           "already a member",
			slug:           "test-campaign-slug",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:    "Test Campaign",
					Slug:    "test-campaign-slug",
					OwnerID: "other-user",
				}
				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockCampaignRepo.On("IsMember", "campaign123", "test-user-id").Return(true, nil)
			},
		},
		{
			name:           "campaign not found",
			slug:           "nonexistent-slug",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockCampaignRepo.On("GetBySlug", "nonexistent-slug").Return(nil, errors.New("campaign not found"))
			},
		},
		{
			name:           "repository error on join",
			slug:           "test-campaign-slug",
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
					OwnerID: "other-user",
				}
				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockCampaignRepo.On("IsMember", "campaign123", "test-user-id").Return(false, nil)
				mockCampaignRepo.On("AddMember", "campaign123", "test-user-id").Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCampaignRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodPut, "/campaigns/"+tt.slug+"/join", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCampaignRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignHandler_ParticipateCampaign(t *testing.T) {
	router, mockCampaignRepo, mockRunnerRepo, mockSponsorRepo, _ := setupCampaignMembershipTest(t)

	tests := []struct {
		name           string
		slug           string
		requestBody    dto.ParticipateCampaignRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful participation",
			slug: "test-campaign-slug",
			requestBody: dto.ParticipateCampaignRequest{
				Activity: "Walking",
			},
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
					OwnerID: "other-user",
				}
				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockCampaignRepo.On("IsMember", "campaign123", "test-user-id").Return(false, nil)
				mockCampaignRepo.On("AddMember", "campaign123", "test-user-id").Return(nil)
				mockRunnerRepo.On("Create", mock.MatchedBy(func(r *campaignModel.CampaignRunner) bool {
					return r.CampaignID == "campaign123" &&
						r.OwnerID == "test-user-id" &&
						r.Activity == "Walking"
				})).Return(nil)
			},
		},
		{
			name: "participation with empty activity",
			slug: "test-campaign-slug",
			requestBody: dto.ParticipateCampaignRequest{
				Activity: "",
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
					Name:    "Test Campaign",
					Slug:    "test-campaign-slug",
					OwnerID: "other-user",
				}
				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
			},
		},
		{
			name: "campaign not found",
			slug: "nonexistent-slug",
			requestBody: dto.ParticipateCampaignRequest{
				Activity: "Running",
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockCampaignRepo.On("GetBySlug", "nonexistent-slug").Return(nil, errors.New("campaign not found"))
			},
		},
		{
			name: "repository error on participation",
			slug: "test-campaign-slug",
			requestBody: dto.ParticipateCampaignRequest{
				Activity: "Cycling",
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
					OwnerID: "other-user",
				}
				mockCampaignRepo.On("GetBySlug", "test-campaign-slug").Return(expectedCampaign, nil)
				mockCampaignRepo.On("IsMember", "campaign123", "test-user-id").Return(false, nil)
				mockCampaignRepo.On("AddMember", "campaign123", "test-user-id").Return(nil)
				mockRunnerRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCampaignRepo.ExpectedCalls = nil
			mockRunnerRepo.ExpectedCalls = nil
			mockSponsorRepo.ExpectedCalls = nil

			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/campaigns/"+tt.slug+"/participate", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCampaignRepo.AssertExpectations(t)
			mockRunnerRepo.AssertExpectations(t)
			mockSponsorRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignHandler_GetCampaignsByUser(t *testing.T) {
	router, mockCampaignRepo, _, _, mockUserRepo := setupCampaignMembershipTest(t)

	expectedCampaigns := []*campaignModel.Campaign{
		{
			Base: model.Base{
				ID:        "campaign1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "My Campaign 1",
			Slug:    "my-campaign-1",
			OwnerID: "test-user-id",
		},
		{
			Base: model.Base{
				ID:        "campaign2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "My Campaign 2",
			Slug:    "my-campaign-2",
			OwnerID: "test-user-id",
		},
	}

	mockCampaignRepo.On("GetByOwnerID", "test-user-id").Return(expectedCampaigns, nil)

	// Mock user lookup for campaign owner
	expectedUser := &userModel.User{
		Base: model.Base{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username: "testuser",
		Email:    "test@example.com",
	}
	mockUserRepo.On("GetByID", "test-user-id").Return(expectedUser, nil)

	req, _ := http.NewRequest(http.MethodGet, "/campaigns/by_user", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.CampaignListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Len(t, response.Campaigns, 2)

	mockCampaignRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestCampaignHandler_GetCampaignsByOthers(t *testing.T) {
	router, mockCampaignRepo, _, _, mockUserRepo := setupCampaignMembershipTest(t)

	// Mock campaigns from other users
	allCampaigns := []*campaignModel.Campaign{
		{
			Base: model.Base{
				ID:        "campaign1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Other Campaign 1",
			Slug:    "other-campaign-1",
			OwnerID: "other-user-1",
		},
		{
			Base: model.Base{
				ID:        "campaign2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Other Campaign 2",
			Slug:    "other-campaign-2",
			OwnerID: "other-user-2",
		},
		{
			Base: model.Base{
				ID:        "campaign3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "My Campaign",
			Slug:    "my-campaign",
			OwnerID: "test-user-id", // This should be excluded
		},
	}

	mockCampaignRepo.On("List", 100, 0).Return(allCampaigns, nil) // Gets more to filter (limit*10)

	// Mock user lookups for campaign owners
	user1 := &userModel.User{
		Base: model.Base{
			ID:        "other-user-1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username: "user1",
		Email:    "user1@example.com",
	}
	user2 := &userModel.User{
		Base: model.Base{
			ID:        "other-user-2",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username: "user2",
		Email:    "user2@example.com",
	}
	mockUserRepo.On("GetByID", "other-user-1").Return(user1, nil)
	mockUserRepo.On("GetByID", "other-user-2").Return(user2, nil)

	req, _ := http.NewRequest(http.MethodGet, "/campaigns/by_others", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.CampaignListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Len(t, response.Campaigns, 2) // Should exclude user's own campaign

	mockCampaignRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}
