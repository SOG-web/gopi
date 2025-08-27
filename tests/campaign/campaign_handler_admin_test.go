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

// Test setup helper for admin campaign tests
func setupCampaignAdminTest(t *testing.T) (*gin.Engine, *campaignMocks.MockCampaignRepository, *campaignMocks.MockCampaignRunnerRepository, *campaignMocks.MockSponsorCampaignRepository, *userMocks.MockUserRepository) {
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
	campaignAdminHandler := handler.NewCampaignAdminHandler(campaignService, userService)

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup admin routes (assuming admin middleware is already handled)
	adminCampaigns := router.Group("/admin")
	adminCampaigns.Use(func(c *gin.Context) {
		// Mock admin auth middleware - set admin user_id in context
		c.Set("user_id", "admin-user-id")
		c.Set("is_admin", true)
		c.Next()
	})

	// Campaign Runner admin routes
	adminCampaigns.POST("/campaign-runners", campaignAdminHandler.CreateCampaignRunner)
	adminCampaigns.GET("/campaign-runners", campaignAdminHandler.GetCampaignRunners)
	adminCampaigns.GET("/campaign-runners/:id", campaignAdminHandler.GetCampaignRunnerByID)
	adminCampaigns.PUT("/campaign-runners/:id", campaignAdminHandler.UpdateCampaignRunner)
	adminCampaigns.DELETE("/campaign-runners/:id", campaignAdminHandler.DeleteCampaignRunner)

	// Sponsor Campaign admin routes
	adminCampaigns.POST("/sponsor-campaigns", campaignAdminHandler.CreateSponsorCampaign)
	adminCampaigns.GET("/sponsor-campaigns", campaignAdminHandler.GetSponsorCampaigns)
	adminCampaigns.GET("/sponsor-campaigns/:id", campaignAdminHandler.GetSponsorCampaignByID)
	adminCampaigns.PUT("/sponsor-campaigns/:id", campaignAdminHandler.UpdateSponsorCampaign)
	adminCampaigns.DELETE("/sponsor-campaigns/:id", campaignAdminHandler.DeleteSponsorCampaign)

	return router, mockCampaignRepo, mockRunnerRepo, mockSponsorRepo, mockUserRepo
}

func TestCampaignAdminHandler_CreateCampaignRunner(t *testing.T) {
	router, mockCampaignRepo, mockRunnerRepo, _, mockUserRepo := setupCampaignAdminTest(t)

	tests := []struct {
		name           string
		requestBody    dto.CreateCampaignRunnerRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful campaign runner creation",
			requestBody: dto.CreateCampaignRunnerRequest{
				CampaignID:      "campaign123",
				UserID:          "user123",
				Activity:        "Running",
				DistanceCovered: 10.5,
				Duration:        "45:30",
				MoneyRaised:     25.0,
			},
			expectedStatus: http.StatusCreated,
			mockSetup: func() {
				// Mock campaign lookup
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:    "Test Campaign",
					Slug:    "test-campaign",
					OwnerID: "campaign-owner",
				}
				mockCampaignRepo.On("GetByID", "campaign123").Return(expectedCampaign, nil)
				mockCampaignRepo.On("GetBySlug", "test-campaign").Return(expectedCampaign, nil)
				mockCampaignRepo.On("IsMember", "campaign123", "user123").Return(false, nil)
				mockCampaignRepo.On("AddMember", "campaign123", "user123").Return(nil)

				// Mock user lookup
				expectedUser := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:  "testuser",
					FirstName: "Test",
					LastName:  "User",
				}
				mockUserRepo.On("GetByID", "user123").Return(expectedUser, nil)

				// Mock runner creation (only basic fields are set by ParticipateCampaign)
				mockRunnerRepo.On("Create", mock.MatchedBy(func(r *campaignModel.CampaignRunner) bool {
					return r.CampaignID == "campaign123" &&
						r.OwnerID == "user123" &&
						r.Activity == "Running" &&
						r.DistanceCovered == 0 &&
						r.Duration == "" &&
						r.MoneyRaised == 0
				})).Return(nil)

				// Mock FinishActivity call (updates runner with additional details)
				mockRunnerRepo.On("Update", mock.MatchedBy(func(r *campaignModel.CampaignRunner) bool {
					return r.DistanceCovered == 10.5 &&
						r.Duration == "45:30" &&
						r.MoneyRaised == 25.0
				})).Return(nil)

				// Mock campaign update for FinishActivity
				mockCampaignRepo.On("Update", mock.MatchedBy(func(c *campaignModel.Campaign) bool {
					return c.DistanceCovered == 10.5 &&
						c.MoneyRaised == 25.0
				})).Return(nil)

				// Mock getting runner for FinishActivity
				mockRunnerRepo.On("GetByID", mock.Anything).Return(&campaignModel.CampaignRunner{
					Base: model.Base{
						ID:        "temp-id", // Will be overridden by actual ID
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:      "campaign123",
					OwnerID:         "user123",
					Activity:        "Running",
					DistanceCovered: 0,
					Duration:        "",
					MoneyRaised:     0,
					DateJoined:      time.Now(),
				}, nil).Once()

				// Mock getting updated runner for response
				expectedRunner := &campaignModel.CampaignRunner{
					Base: model.Base{
						ID:        "temp-id", // Will be overridden by actual ID
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:      "campaign123",
					OwnerID:         "user123",
					Activity:        "Running",
					DistanceCovered: 10.5,
					Duration:        "45:30",
					MoneyRaised:     25.0,
					DateJoined:      time.Now(),
				}
				mockRunnerRepo.On("GetByID", mock.Anything).Return(expectedRunner, nil)
			},
		},
		{
			name: "invalid request body - missing campaign_id",
			requestBody: dto.CreateCampaignRunnerRequest{
				UserID:   "user123",
				Activity: "Running",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "invalid request body - missing user_id",
			requestBody: dto.CreateCampaignRunnerRequest{
				CampaignID: "campaign123",
				Activity:   "Running",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "invalid request body - missing activity",
			requestBody: dto.CreateCampaignRunnerRequest{
				CampaignID: "campaign123",
				UserID:     "user123",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "campaign not found",
			requestBody: dto.CreateCampaignRunnerRequest{
				CampaignID: "nonexistent",
				UserID:     "user123",
				Activity:   "Running",
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockCampaignRepo.On("GetByID", "nonexistent").Return(nil, errors.New("campaign not found"))
			},
		},
		{
			name: "user not found",
			requestBody: dto.CreateCampaignRunnerRequest{
				CampaignID: "campaign123",
				UserID:     "nonexistent",
				Activity:   "Running",
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				// Mock campaign lookup success
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name: "Test Campaign",
				}
				mockCampaignRepo.On("GetByID", "campaign123").Return(expectedCampaign, nil)

				// Mock user lookup failure
				mockUserRepo.On("GetByID", "nonexistent").Return(nil, errors.New("user not found"))
			},
		},
		{
			name: "repository error on creation",
			requestBody: dto.CreateCampaignRunnerRequest{
				CampaignID: "campaign123",
				UserID:     "user123",
				Activity:   "Running",
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				// Mock successful lookups
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name: "Test Campaign",
					Slug: "test-campaign",
				}
				mockCampaignRepo.On("GetByID", "campaign123").Return(expectedCampaign, nil)
				mockCampaignRepo.On("GetBySlug", "test-campaign").Return(expectedCampaign, nil)
				mockCampaignRepo.On("IsMember", "campaign123", "user123").Return(false, nil)
				mockCampaignRepo.On("AddMember", "campaign123", "user123").Return(nil)

				expectedUser := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username: "testuser",
				}
				mockUserRepo.On("GetByID", "user123").Return(expectedUser, nil)

				// Mock creation failure
				mockRunnerRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
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

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/admin/campaign-runners", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCampaignRepo.AssertExpectations(t)
			mockRunnerRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignAdminHandler_GetCampaignRunners(t *testing.T) {
	router, mockCampaignRepo, mockRunnerRepo, _, mockUserRepo := setupCampaignAdminTest(t)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful get campaign runners with default pagination",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedRunners := []*campaignModel.CampaignRunner{
					{
						Base: model.Base{
							ID:        "runner1",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						CampaignID:      "campaign123",
						OwnerID:         "user1",
						Activity:        "Running",
						DistanceCovered: 10.5,
						Duration:        "45:30",
						MoneyRaised:     25.0,
						DateJoined:      time.Now(),
					},
					{
						Base: model.Base{
							ID:        "runner2",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						CampaignID:      "campaign123",
						OwnerID:         "user2",
						Activity:        "Walking",
						DistanceCovered: 5.0,
						Duration:        "30:00",
						MoneyRaised:     10.0,
						DateJoined:      time.Now(),
					},
				}

				// Mock campaigns list
				expectedCampaigns := []*campaignModel.Campaign{
					{
						Base: model.Base{
							ID:        "campaign123",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Name: "Test Campaign",
						Slug: "test-campaign",
					},
				}
				mockCampaignRepo.On("List", 10, 0).Return(expectedCampaigns, nil)

				// Mock GetBySlug for leaderboard
				mockCampaignRepo.On("GetBySlug", "test-campaign").Return(expectedCampaigns[0], nil)

				// Mock GetByCampaignID for leaderboard (runners)
				mockRunnerRepo.On("GetByCampaignID", "campaign123").Return(expectedRunners, nil)

				// Mock user lookups for each runner
				user1 := &userModel.User{
					Base: model.Base{
						ID:        "user1",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:  "user1",
					FirstName: "User",
					LastName:  "One",
				}
				user2 := &userModel.User{
					Base: model.Base{
						ID:        "user2",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:  "user2",
					FirstName: "User",
					LastName:  "Two",
				}
				mockUserRepo.On("GetByID", "user1").Return(user1, nil)
				mockUserRepo.On("GetByID", "user2").Return(user2, nil)
			},
		},
		{
			name:           "get campaign runners with custom pagination",
			queryParams:    "?page=2&limit=5",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedRunners := []*campaignModel.CampaignRunner{
					{
						Base: model.Base{
							ID:        "runner3",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						CampaignID:      "campaign123",
						OwnerID:         "user3",
						Activity:        "Cycling",
						DistanceCovered: 15.0,
						Duration:        "60:00",
						MoneyRaised:     30.0,
						DateJoined:      time.Now(),
					},
				}

				// Mock campaigns list
				expectedCampaigns := []*campaignModel.Campaign{
					{
						Base: model.Base{
							ID:        "campaign123",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Name: "Test Campaign",
						Slug: "test-campaign",
					},
				}
				mockCampaignRepo.On("List", 10, 0).Return(expectedCampaigns, nil) // Handler uses default pagination

				// Mock GetBySlug for leaderboard
				mockCampaignRepo.On("GetBySlug", "test-campaign").Return(expectedCampaigns[0], nil)

				// Mock GetByCampaignID for leaderboard (runners)
				mockRunnerRepo.On("GetByCampaignID", "campaign123").Return(expectedRunners, nil)

				user3 := &userModel.User{
					Base: model.Base{
						ID:        "user3",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:  "user3",
					FirstName: "User",
					LastName:  "Three",
				}
				mockUserRepo.On("GetByID", "user3").Return(user3, nil)
			},
		},
		{
			name:           "repository error",
			queryParams:    "",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				// Mock campaign list to return error
				mockCampaignRepo.On("List", 10, 0).Return(nil, errors.New("repository error"))
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

			req, _ := http.NewRequest(http.MethodGet, "/admin/campaign-runners"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCampaignRepo.AssertExpectations(t)
			mockRunnerRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignAdminHandler_GetCampaignRunnerByID(t *testing.T) {
	router, _, mockRunnerRepo, _, mockUserRepo := setupCampaignAdminTest(t)

	tests := []struct {
		name           string
		runnerID       string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful get campaign runner by id",
			runnerID:       "runner123",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedRunner := &campaignModel.CampaignRunner{
					Base: model.Base{
						ID:        "runner123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:      "campaign123",
					OwnerID:         "user123",
					Activity:        "Running",
					DistanceCovered: 10.5,
					Duration:        "45:30",
					MoneyRaised:     25.0,
					DateJoined:      time.Now(),
				}

				mockRunnerRepo.On("GetByID", "runner123").Return(expectedRunner, nil)

				expectedUser := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:  "testuser",
					FirstName: "Test",
					LastName:  "User",
				}
				mockUserRepo.On("GetByID", "user123").Return(expectedUser, nil)
			},
		},
		{
			name:           "runner not found",
			runnerID:       "nonexistent",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockRunnerRepo.On("GetByID", "nonexistent").Return(nil, errors.New("runner not found"))
			},
		},
		{
			name:           "repository error",
			runnerID:       "runner123",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				mockRunnerRepo.On("GetByID", "runner123").Return(nil, errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockRunnerRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodGet, "/admin/campaign-runners/"+tt.runnerID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRunnerRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignAdminHandler_UpdateCampaignRunner(t *testing.T) {
	router, _, mockRunnerRepo, _, mockUserRepo := setupCampaignAdminTest(t)

	tests := []struct {
		name           string
		runnerID       string
		requestBody    dto.UpdateCampaignRunnerRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:     "successful campaign runner update",
			runnerID: "runner123",
			requestBody: dto.UpdateCampaignRunnerRequest{
				Activity:        "Walking",
				DistanceCovered: floatPtr(15.0),
				Duration:        "60:00",
				MoneyRaised:     floatPtr(30.0),
				CoverImage:      "new-cover.jpg",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				existingRunner := &campaignModel.CampaignRunner{
					Base: model.Base{
						ID:        "runner123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:      "campaign123",
					OwnerID:         "user123",
					Activity:        "Running",
					DistanceCovered: 10.5,
					Duration:        "45:30",
					MoneyRaised:     25.0,
					CoverImage:      "old-cover.jpg",
				}

				mockRunnerRepo.On("GetByID", "runner123").Return(existingRunner, nil)
				mockRunnerRepo.On("Update", mock.MatchedBy(func(r *campaignModel.CampaignRunner) bool {
					return r.ID == "runner123" &&
						r.Activity == "Walking" &&
						r.DistanceCovered == 15.0 &&
						r.Duration == "60:00" &&
						r.MoneyRaised == 30.0 &&
						r.CoverImage == "new-cover.jpg"
				})).Return(nil)

				// Mock GetByID for response
				updatedRunner := &campaignModel.CampaignRunner{
					Base: model.Base{
						ID:        "runner123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:      "campaign123",
					OwnerID:         "user123",
					Activity:        "Walking",
					DistanceCovered: 15.0,
					Duration:        "60:00",
					MoneyRaised:     30.0,
					CoverImage:      "new-cover.jpg",
				}
				mockRunnerRepo.On("GetByID", "runner123").Return(updatedRunner, nil)

				// Mock user lookup for response
				expectedUser := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:  "testuser",
					FirstName: "Test",
					LastName:  "User",
				}
				mockUserRepo.On("GetByID", "user123").Return(expectedUser, nil)
			},
		},
		{
			name:     "partial update - only activity",
			runnerID: "runner123",
			requestBody: dto.UpdateCampaignRunnerRequest{
				Activity: "Cycling",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				existingRunner := &campaignModel.CampaignRunner{
					Base: model.Base{
						ID:        "runner123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:      "campaign123",
					OwnerID:         "user123",
					Activity:        "Running",
					DistanceCovered: 10.5,
					Duration:        "45:30",
					MoneyRaised:     25.0,
				}

				mockRunnerRepo.On("GetByID", "runner123").Return(existingRunner, nil)
				mockRunnerRepo.On("Update", mock.MatchedBy(func(r *campaignModel.CampaignRunner) bool {
					return r.ID == "runner123" &&
						r.Activity == "Cycling" &&
						r.DistanceCovered == 10.5 && // unchanged
						r.Duration == "45:30" && // unchanged
						r.MoneyRaised == 25.0 // unchanged
				})).Return(nil)

				// Mock user lookup for response
				expectedUser := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:  "testuser",
					FirstName: "Test",
					LastName:  "User",
				}
				mockUserRepo.On("GetByID", "user123").Return(expectedUser, nil)
			},
		},
		{
			name:     "runner not found",
			runnerID: "nonexistent",
			requestBody: dto.UpdateCampaignRunnerRequest{
				Activity: "Walking",
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockRunnerRepo.On("GetByID", "nonexistent").Return(nil, errors.New("runner not found"))
			},
		},
		{
			name:     "repository error on update",
			runnerID: "runner123",
			requestBody: dto.UpdateCampaignRunnerRequest{
				Activity: "Walking",
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				existingRunner := &campaignModel.CampaignRunner{
					Base: model.Base{
						ID:        "runner123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID: "campaign123",
					OwnerID:    "user123",
					Activity:   "Running",
				}

				mockRunnerRepo.On("GetByID", "runner123").Return(existingRunner, nil)
				mockRunnerRepo.On("Update", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockRunnerRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/admin/campaign-runners/"+tt.runnerID, bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRunnerRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignAdminHandler_DeleteCampaignRunner(t *testing.T) {
	router, _, mockRunnerRepo, _, _ := setupCampaignAdminTest(t)

	tests := []struct {
		name           string
		runnerID       string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful campaign runner deletion",
			runnerID:       "runner123",
			expectedStatus: http.StatusNoContent,
			mockSetup: func() {
				existingRunner := &campaignModel.CampaignRunner{
					Base: model.Base{
						ID:        "runner123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID: "campaign123",
					OwnerID:    "user123",
					Activity:   "Running",
				}

				mockRunnerRepo.On("GetByID", "runner123").Return(existingRunner, nil)
				mockRunnerRepo.On("Delete", "runner123").Return(nil)
			},
		},
		{
			name:           "runner not found",
			runnerID:       "nonexistent",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockRunnerRepo.On("GetByID", "nonexistent").Return(nil, errors.New("runner not found"))
			},
		},
		{
			name:           "repository error on deletion",
			runnerID:       "runner123",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				existingRunner := &campaignModel.CampaignRunner{
					Base: model.Base{
						ID:        "runner123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID: "campaign123",
					OwnerID:    "user123",
					Activity:   "Running",
				}

				mockRunnerRepo.On("GetByID", "runner123").Return(existingRunner, nil)
				mockRunnerRepo.On("Delete", "runner123").Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockRunnerRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodDelete, "/admin/campaign-runners/"+tt.runnerID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRunnerRepo.AssertExpectations(t)
		})
	}
}

// Sponsor Campaign Admin Tests

func TestCampaignAdminHandler_CreateSponsorCampaign(t *testing.T) {
	router, mockCampaignRepo, _, mockSponsorRepo, mockUserRepo := setupCampaignAdminTest(t)

	tests := []struct {
		name           string
		requestBody    dto.CreateSponsorCampaignRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful sponsor campaign creation",
			requestBody: dto.CreateSponsorCampaignRequest{
				CampaignID:  "campaign123",
				SponsorID:   "sponsor123",
				Distance:    10.0,
				AmountPerKm: 5.0,
				BrandImg:    "brand.jpg",
				VideoUrl:    "video.mp4",
			},
			expectedStatus: http.StatusCreated,
			mockSetup: func() {
				// Mock campaign lookup
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:    "Test Campaign",
					Slug:    "test-campaign",
					OwnerID: "campaign-owner",
				}
				mockCampaignRepo.On("GetByID", "campaign123").Return(expectedCampaign, nil)

				// Mock sponsor user lookup
				expectedSponsor := &userModel.User{
					Base: model.Base{
						ID:        "sponsor123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username:  "sponsoruser",
					FirstName: "Sponsor",
					LastName:  "User",
				}
				mockUserRepo.On("GetByID", "sponsor123").Return(expectedSponsor, nil)

				// Mock sponsor campaign creation
				mockSponsorRepo.On("Create", mock.MatchedBy(func(s *campaignModel.SponsorCampaign) bool {
					return s.CampaignID == "campaign123" &&
						s.Distance == 10.0 &&
						s.AmountPerKm == 5.0 &&
						s.TotalAmount == 50.0 &&
						s.BrandImg == "brand.jpg" &&
						s.VideoUrl == "video.mp4"
				})).Return(nil)

				// Mock second GetByID call in CreateSponsorCampaign for updating campaign
				mockCampaignRepo.On("GetByID", "campaign123").Return(expectedCampaign, nil)

				// Mock campaign update in CreateSponsorCampaign
				mockCampaignRepo.On("Update", mock.MatchedBy(func(c *campaignModel.Campaign) bool {
					return c.ID == "campaign123" && c.MoneyRaised == 50.0
				})).Return(nil)

				// Mock AddSponsor call
				mockCampaignRepo.On("AddSponsor", "campaign123", "sponsor123").Return(nil)

				// The handler uses the sponsor object returned from CreateSponsorCampaign directly
			},
		},
		{
			name: "sponsor campaign creation without optional fields",
			requestBody: dto.CreateSponsorCampaignRequest{
				CampaignID:  "campaign123",
				SponsorID:   "sponsor123",
				Distance:    5.0,
				AmountPerKm: 2.0,
			},
			expectedStatus: http.StatusCreated,
			mockSetup: func() {
				// Mock campaign lookup
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name: "Test Campaign",
				}
				mockCampaignRepo.On("GetByID", "campaign123").Return(expectedCampaign, nil)

				// Mock sponsor user lookup
				expectedSponsor := &userModel.User{
					Base: model.Base{
						ID:        "sponsor123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username: "sponsoruser",
				}
				mockUserRepo.On("GetByID", "sponsor123").Return(expectedSponsor, nil)

				// Mock sponsor campaign creation
				mockSponsorRepo.On("Create", mock.MatchedBy(func(s *campaignModel.SponsorCampaign) bool {
					return s.CampaignID == "campaign123" &&
						s.Distance == 5.0 &&
						s.AmountPerKm == 2.0 &&
						s.TotalAmount == 10.0 &&
						s.BrandImg == "" &&
						s.VideoUrl == ""
				})).Return(nil)

				// Mock second GetByID call in CreateSponsorCampaign for updating campaign
				mockCampaignRepo.On("GetByID", "campaign123").Return(expectedCampaign, nil)

				// Mock campaign update in CreateSponsorCampaign
				mockCampaignRepo.On("Update", mock.MatchedBy(func(c *campaignModel.Campaign) bool {
					return c.ID == "campaign123" && c.MoneyRaised == 10.0
				})).Return(nil)

				// Mock AddSponsor call
				mockCampaignRepo.On("AddSponsor", "campaign123", "sponsor123").Return(nil)
			},
		},
		{
			name: "invalid request body - missing campaign_id",
			requestBody: dto.CreateSponsorCampaignRequest{
				SponsorID:   "sponsor123",
				Distance:    10.0,
				AmountPerKm: 5.0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "invalid request body - zero distance",
			requestBody: dto.CreateSponsorCampaignRequest{
				CampaignID:  "campaign123",
				SponsorID:   "sponsor123",
				Distance:    0,
				AmountPerKm: 5.0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "invalid request body - zero amount per km",
			requestBody: dto.CreateSponsorCampaignRequest{
				CampaignID:  "campaign123",
				SponsorID:   "sponsor123",
				Distance:    10.0,
				AmountPerKm: 0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "campaign not found",
			requestBody: dto.CreateSponsorCampaignRequest{
				CampaignID:  "nonexistent",
				SponsorID:   "sponsor123",
				Distance:    10.0,
				AmountPerKm: 5.0,
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockCampaignRepo.On("GetByID", "nonexistent").Return(nil, errors.New("campaign not found"))
			},
		},
		{
			name: "sponsor not found",
			requestBody: dto.CreateSponsorCampaignRequest{
				CampaignID:  "campaign123",
				SponsorID:   "nonexistent",
				Distance:    10.0,
				AmountPerKm: 5.0,
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				// Mock campaign lookup success
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name: "Test Campaign",
				}
				mockCampaignRepo.On("GetByID", "campaign123").Return(expectedCampaign, nil)

				// Mock sponsor lookup failure
				mockUserRepo.On("GetByID", "nonexistent").Return(nil, errors.New("sponsor not found"))
			},
		},
		{
			name: "repository error on creation",
			requestBody: dto.CreateSponsorCampaignRequest{
				CampaignID:  "campaign123",
				SponsorID:   "sponsor123",
				Distance:    10.0,
				AmountPerKm: 5.0,
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				// Mock successful lookups
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name: "Test Campaign",
				}
				mockCampaignRepo.On("GetByID", "campaign123").Return(expectedCampaign, nil)

				expectedSponsor := &userModel.User{
					Base: model.Base{
						ID:        "sponsor123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username: "sponsoruser",
				}
				mockUserRepo.On("GetByID", "sponsor123").Return(expectedSponsor, nil)

				// Mock creation failure
				mockSponsorRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCampaignRepo.ExpectedCalls = nil
			mockSponsorRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/admin/sponsor-campaigns", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCampaignRepo.AssertExpectations(t)
			mockSponsorRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignAdminHandler_GetSponsorCampaigns(t *testing.T) {
	router, mockCampaignRepo, _, mockSponsorRepo, mockUserRepo := setupCampaignAdminTest(t)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful get sponsor campaigns with default pagination",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedSponsorCampaigns := []*campaignModel.SponsorCampaign{
					{
						Base: model.Base{
							ID:        "sponsor_campaign1",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						CampaignID:  "campaign123",
						Distance:    10.0,
						AmountPerKm: 5.0,
						TotalAmount: 50.0,
						BrandImg:    "brand1.jpg",
						VideoUrl:    "video1.mp4",
					},
					{
						Base: model.Base{
							ID:        "sponsor_campaign2",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						CampaignID:  "campaign456",
						Distance:    5.0,
						AmountPerKm: 2.0,
						TotalAmount: 10.0,
						BrandImg:    "brand2.jpg",
						VideoUrl:    "",
					},
				}

				// Mock campaigns list
				expectedCampaigns := []*campaignModel.Campaign{
					{
						Base: model.Base{
							ID:        "campaign123",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Name: "Campaign 1",
						Slug: "campaign-1",
					},
					{
						Base: model.Base{
							ID:        "campaign456",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Name: "Campaign 2",
						Slug: "campaign-2",
					},
				}
				mockCampaignRepo.On("List", 10, 0).Return(expectedCampaigns, nil)

				// Mock GetByCampaignID for sponsor campaigns
				mockSponsorRepo.On("GetByCampaignID", "campaign123").Return([]*campaignModel.SponsorCampaign{expectedSponsorCampaigns[0]}, nil)
				mockSponsorRepo.On("GetByCampaignID", "campaign456").Return([]*campaignModel.SponsorCampaign{expectedSponsorCampaigns[1]}, nil)
			},
		},
		{
			name:           "get sponsor campaigns with custom pagination",
			queryParams:    "?page=2&limit=5",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedSponsorCampaigns := []*campaignModel.SponsorCampaign{
					{
						Base: model.Base{
							ID:        "sponsor_campaign3",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						CampaignID:  "campaign789",
						Distance:    15.0,
						AmountPerKm: 3.0,
						TotalAmount: 45.0,
						BrandImg:    "brand3.jpg",
						VideoUrl:    "video3.mp4",
					},
				}

				// Mock campaigns list (handler uses default pagination regardless of query params)
				expectedCampaigns := []*campaignModel.Campaign{
					{
						Base: model.Base{
							ID:        "campaign789",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Name: "Campaign 3",
						Slug: "campaign-3",
					},
				}
				mockCampaignRepo.On("List", 10, 0).Return(expectedCampaigns, nil)

				// Mock GetByCampaignID for sponsor campaigns
				mockSponsorRepo.On("GetByCampaignID", "campaign789").Return(expectedSponsorCampaigns, nil)
			},
		},
		{
			name:           "repository error",
			queryParams:    "",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				mockCampaignRepo.On("List", 10, 0).Return(nil, errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCampaignRepo.ExpectedCalls = nil
			mockSponsorRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodGet, "/admin/sponsor-campaigns"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCampaignRepo.AssertExpectations(t)
			mockSponsorRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignAdminHandler_GetSponsorCampaignByID(t *testing.T) {
	router, mockCampaignRepo, _, mockSponsorRepo, _ := setupCampaignAdminTest(t)

	tests := []struct {
		name           string
		sponsorID      string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful get sponsor campaign by id",
			sponsorID:      "sponsor_campaign123",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedSponsorCampaign := &campaignModel.SponsorCampaign{
					Base: model.Base{
						ID:        "sponsor_campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:  "campaign123",
					Distance:    10.0,
					AmountPerKm: 5.0,
					TotalAmount: 50.0,
					BrandImg:    "brand.jpg",
					VideoUrl:    "video.mp4",
				}

				mockSponsorRepo.On("GetByID", "sponsor_campaign123").Return(expectedSponsorCampaign, nil)

				// Mock campaign lookup for response
				expectedCampaign := &campaignModel.Campaign{
					Base: model.Base{
						ID:        "campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name: "Test Campaign",
					Slug: "test-campaign",
				}
				mockCampaignRepo.On("GetByID", "campaign123").Return(expectedCampaign, nil)
			},
		},
		{
			name:           "sponsor campaign not found",
			sponsorID:      "nonexistent",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockSponsorRepo.On("GetByID", "nonexistent").Return(nil, errors.New("sponsor campaign not found"))
			},
		},
		{
			name:           "repository error",
			sponsorID:      "sponsor_campaign123",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				mockSponsorRepo.On("GetByID", "sponsor_campaign123").Return(nil, errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCampaignRepo.ExpectedCalls = nil
			mockSponsorRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodGet, "/admin/sponsor-campaigns/"+tt.sponsorID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCampaignRepo.AssertExpectations(t)
			mockSponsorRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignAdminHandler_UpdateSponsorCampaign(t *testing.T) {
	router, _, _, mockSponsorRepo, _ := setupCampaignAdminTest(t)

	tests := []struct {
		name           string
		sponsorID      string
		requestBody    dto.UpdateSponsorCampaignRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:      "successful sponsor campaign update",
			sponsorID: "sponsor_campaign123",
			requestBody: dto.UpdateSponsorCampaignRequest{
				Distance:    floatPtr(15.0),
				AmountPerKm: floatPtr(6.0),
				BrandImg:    "new-brand.jpg",
				VideoUrl:    "new-video.mp4",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				existingSponsorCampaign := &campaignModel.SponsorCampaign{
					Base: model.Base{
						ID:        "sponsor_campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:  "campaign123",
					Distance:    10.0,
					AmountPerKm: 5.0,
					TotalAmount: 50.0,
					BrandImg:    "old-brand.jpg",
					VideoUrl:    "old-video.mp4",
				}

				mockSponsorRepo.On("GetByID", "sponsor_campaign123").Return(existingSponsorCampaign, nil)
				mockSponsorRepo.On("Update", mock.MatchedBy(func(s *campaignModel.SponsorCampaign) bool {
					return s.ID == "sponsor_campaign123" &&
						s.Distance == 15.0 &&
						s.AmountPerKm == 6.0 &&
						s.TotalAmount == 90.0 &&
						s.BrandImg == "new-brand.jpg" &&
						s.VideoUrl == "new-video.mp4"
				})).Return(nil)

				// Mock GetByID for response
				updatedSponsorCampaign := &campaignModel.SponsorCampaign{
					Base: model.Base{
						ID:        "sponsor_campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:  "campaign123",
					Distance:    15.0,
					AmountPerKm: 6.0,
					TotalAmount: 90.0,
					BrandImg:    "new-brand.jpg",
					VideoUrl:    "new-video.mp4",
				}
				mockSponsorRepo.On("GetByID", "sponsor_campaign123").Return(updatedSponsorCampaign, nil)
			},
		},
		{
			name:      "partial update - only distance",
			sponsorID: "sponsor_campaign123",
			requestBody: dto.UpdateSponsorCampaignRequest{
				Distance: floatPtr(12.0),
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				existingSponsorCampaign := &campaignModel.SponsorCampaign{
					Base: model.Base{
						ID:        "sponsor_campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:  "campaign123",
					Distance:    10.0,
					AmountPerKm: 5.0,
					TotalAmount: 50.0,
					BrandImg:    "brand.jpg",
					VideoUrl:    "video.mp4",
				}

				mockSponsorRepo.On("GetByID", "sponsor_campaign123").Return(existingSponsorCampaign, nil)
				mockSponsorRepo.On("Update", mock.MatchedBy(func(s *campaignModel.SponsorCampaign) bool {
					return s.ID == "sponsor_campaign123" &&
						s.Distance == 12.0 &&
						s.AmountPerKm == 5.0 && // unchanged
						s.TotalAmount == 60.0 && // recalculated
						s.BrandImg == "brand.jpg" && // unchanged
						s.VideoUrl == "video.mp4" // unchanged
				})).Return(nil)
			},
		},
		{
			name:      "invalid update - zero distance",
			sponsorID: "sponsor_campaign123",
			requestBody: dto.UpdateSponsorCampaignRequest{
				Distance: floatPtr(0),
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				existingSponsorCampaign := &campaignModel.SponsorCampaign{
					Base: model.Base{
						ID:        "sponsor_campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:  "campaign123",
					Distance:    10.0,
					AmountPerKm: 5.0,
					TotalAmount: 50.0,
				}

				mockSponsorRepo.On("GetByID", "sponsor_campaign123").Return(existingSponsorCampaign, nil)
			},
		},
		{
			name:      "invalid update - zero amount per km",
			sponsorID: "sponsor_campaign123",
			requestBody: dto.UpdateSponsorCampaignRequest{
				AmountPerKm: floatPtr(0),
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				existingSponsorCampaign := &campaignModel.SponsorCampaign{
					Base: model.Base{
						ID:        "sponsor_campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:  "campaign123",
					Distance:    10.0,
					AmountPerKm: 5.0,
					TotalAmount: 50.0,
				}

				mockSponsorRepo.On("GetByID", "sponsor_campaign123").Return(existingSponsorCampaign, nil)
			},
		},
		{
			name:      "sponsor campaign not found",
			sponsorID: "nonexistent",
			requestBody: dto.UpdateSponsorCampaignRequest{
				Distance: floatPtr(15.0),
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockSponsorRepo.On("GetByID", "nonexistent").Return(nil, errors.New("sponsor campaign not found"))
			},
		},
		{
			name:      "repository error on update",
			sponsorID: "sponsor_campaign123",
			requestBody: dto.UpdateSponsorCampaignRequest{
				Distance: floatPtr(15.0),
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				existingSponsorCampaign := &campaignModel.SponsorCampaign{
					Base: model.Base{
						ID:        "sponsor_campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:  "campaign123",
					Distance:    10.0,
					AmountPerKm: 5.0,
					TotalAmount: 50.0,
				}

				mockSponsorRepo.On("GetByID", "sponsor_campaign123").Return(existingSponsorCampaign, nil)
				mockSponsorRepo.On("Update", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockSponsorRepo.ExpectedCalls = nil

			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/admin/sponsor-campaigns/"+tt.sponsorID, bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockSponsorRepo.AssertExpectations(t)
		})
	}
}

func TestCampaignAdminHandler_DeleteSponsorCampaign(t *testing.T) {
	router, _, _, mockSponsorRepo, _ := setupCampaignAdminTest(t)

	tests := []struct {
		name           string
		sponsorID      string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful sponsor campaign deletion",
			sponsorID:      "sponsor_campaign123",
			expectedStatus: http.StatusNoContent,
			mockSetup: func() {
				existingSponsorCampaign := &campaignModel.SponsorCampaign{
					Base: model.Base{
						ID:        "sponsor_campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:  "campaign123",
					Distance:    10.0,
					AmountPerKm: 5.0,
					TotalAmount: 50.0,
				}

				mockSponsorRepo.On("GetByID", "sponsor_campaign123").Return(existingSponsorCampaign, nil)
				mockSponsorRepo.On("Delete", "sponsor_campaign123").Return(nil)
			},
		},
		{
			name:           "sponsor campaign not found",
			sponsorID:      "nonexistent",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockSponsorRepo.On("GetByID", "nonexistent").Return(nil, errors.New("sponsor campaign not found"))
			},
		},
		{
			name:           "repository error on deletion",
			sponsorID:      "sponsor_campaign123",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				existingSponsorCampaign := &campaignModel.SponsorCampaign{
					Base: model.Base{
						ID:        "sponsor_campaign123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					CampaignID:  "campaign123",
					Distance:    10.0,
					AmountPerKm: 5.0,
					TotalAmount: 50.0,
				}

				mockSponsorRepo.On("GetByID", "sponsor_campaign123").Return(existingSponsorCampaign, nil)
				mockSponsorRepo.On("Delete", "sponsor_campaign123").Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockSponsorRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodDelete, "/admin/sponsor-campaigns/"+tt.sponsorID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockSponsorRepo.AssertExpectations(t)
		})
	}
}

// Helper function to create float64 pointer
func floatPtr(f float64) *float64 {
	return &f
}
