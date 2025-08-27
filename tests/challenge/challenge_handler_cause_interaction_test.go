package challenge_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopi.com/api/http/dto"
	"gopi.com/api/http/handler"
	"gopi.com/internal/app/challenge"
	"gopi.com/internal/app/user"
	challengeModel "gopi.com/internal/domain/challenge/model"
	"gopi.com/internal/domain/model"
	challengeMocks "gopi.com/tests/mocks/challenge"
	userMocks "gopi.com/tests/mocks/user"
)

// Test setup helper for cause interaction tests
func setupCauseInteractionTest(t *testing.T) (*gin.Engine, *challengeMocks.MockChallengeRepository, *challengeMocks.MockCauseRepository, *challengeMocks.MockCauseRunnerRepository, *challengeMocks.MockSponsorChallengeRepository, *challengeMocks.MockSponsorCauseRepository, *challengeMocks.MockCauseBuyerRepository, *userMocks.MockUserRepository) {
	gin.SetMode(gin.TestMode)

	// Create mock repositories
	mockChallengeRepo := new(challengeMocks.MockChallengeRepository)
	mockCauseRepo := new(challengeMocks.MockCauseRepository)
	mockCauseRunnerRepo := new(challengeMocks.MockCauseRunnerRepository)
	mockSponsorChallengeRepo := new(challengeMocks.MockSponsorChallengeRepository)
	mockSponsorCauseRepo := new(challengeMocks.MockSponsorCauseRepository)
	mockCauseBuyerRepo := new(challengeMocks.MockCauseBuyerRepository)
	mockUserRepo := new(userMocks.MockUserRepository)

	// Create services with mock repositories
	challengeService := challenge.NewChallengeService(mockChallengeRepo, mockCauseRepo, mockCauseRunnerRepo, mockSponsorChallengeRepo, mockSponsorCauseRepo, mockCauseBuyerRepo)
	userService := user.NewUserService(mockUserRepo, nil)

	// Create handler with services
	challengeHandler := handler.NewChallengeHandler(challengeService, userService)

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup auth middleware mock
	router.Use(func(c *gin.Context) {
		// Mock auth middleware - set user_id in context
		c.Set("user_id", "test-user-id")
		c.Next()
	})

	// Setup cause interaction routes
	api := router.Group("/api/v1")

	// Cause interaction routes
	api.POST("/causes/:id/join", challengeHandler.JoinCause)
	api.POST("/causes/activity", challengeHandler.RecordCauseActivity)
	api.POST("/causes/sponsor", challengeHandler.SponsorCause)
	api.POST("/causes/buy", challengeHandler.BuyCause)

	return router, mockChallengeRepo, mockCauseRepo, mockCauseRunnerRepo, mockSponsorChallengeRepo, mockSponsorCauseRepo, mockCauseBuyerRepo, mockUserRepo
}

func TestChallengeHandler_JoinCause(t *testing.T) {
	router, _, mockCauseRepo, _, _, _, _, _ := setupCauseInteractionTest(t)

	tests := []struct {
		name           string
		causeID        string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful cause join",
			causeID:        "cause123",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedCause := &challengeModel.Cause{
					Base: model.Base{
						ID: "cause123",
					},
					Name:    "Test Cause",
					OwnerID: "owner123",
				}

				mockCauseRepo.On("GetByID", "cause123").Return(expectedCause, nil)
				mockCauseRepo.On("Update", mock.MatchedBy(func(c *challengeModel.Cause) bool {
					return c.ID == "cause123"
				})).Return(nil)
			},
		},
		{
			name:           "cause not found",
			causeID:        "nonexistent",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				mockCauseRepo.On("GetByID", "nonexistent").Return(nil, errors.New("cause not found"))
			},
		},
		{
			name:           "user not authenticated",
			causeID:        "cause123",
			expectedStatus: http.StatusUnauthorized,
			mockSetup: func() {
				// This test will use a modified router without authentication
			},
		},
		{
			name:           "repository error on update",
			causeID:        "cause123",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				expectedCause := &challengeModel.Cause{
					Base: model.Base{
						ID: "cause123",
					},
					Name:    "Test Cause",
					OwnerID: "owner123",
				}

				mockCauseRepo.On("GetByID", "cause123").Return(expectedCause, nil)
				mockCauseRepo.On("Update", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCauseRepo.ExpectedCalls = nil

			tt.mockSetup()

			var testRouter *gin.Engine
			if tt.name == "user not authenticated" {
				// Create a separate router without authentication for this test
				testRouter = createUnauthenticatedCauseInteractionRouter(t)
			} else {
				testRouter = router
			}

			req, _ := http.NewRequest(http.MethodPost, "/api/v1/causes/"+tt.causeID+"/join", nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCauseRepo.AssertExpectations(t)
		})
	}
}

func TestChallengeHandler_RecordCauseActivity(t *testing.T) {
	router, _, mockCauseRepo, mockCauseRunnerRepo, _, _, _, _ := setupCauseInteractionTest(t)

	tests := []struct {
		name           string
		requestBody    dto.RecordActivityRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful activity recording",
			requestBody: dto.RecordActivityRequest{
				CauseID:         "cause123",
				DistanceToCover: 10.5,
				DistanceCovered: 8.2,
				Duration:        "45:30",
				Activity:        "Walking",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				// Mock cause retrieval
				mockCauseRepo.On("GetByID", "cause123").Return(&challengeModel.Cause{
					Base: model.Base{ID: "cause123"},
					Name: "Test Cause",
					DistanceCovered: 5.0, // existing distance
				}, nil)

				// Mock cause runner creation
				mockCauseRunnerRepo.On("Create", mock.MatchedBy(func(r *challengeModel.CauseRunner) bool {
					return r.CauseID == "cause123" &&
						r.OwnerID == "test-user-id" &&
						r.DistanceToCover == 10.5 &&
						r.DistanceCovered == 8.2 &&
						r.Duration == "45:30" &&
						r.Activity == "Walking"
				})).Return(nil)

				// Mock cause update after activity recording
				mockCauseRepo.On("Update", mock.MatchedBy(func(c *challengeModel.Cause) bool {
					return c.ID == "cause123" && c.DistanceCovered == 13.2 // 5.0 + 8.2
				})).Return(nil)
			},
		},
		{
			name: "invalid request body - missing cause_id",
			requestBody: dto.RecordActivityRequest{
				DistanceToCover: 10.5,
				DistanceCovered: 8.2,
				Duration:        "45:30",
				Activity:        "Walking",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "invalid request body - zero distance_to_cover",
			requestBody: dto.RecordActivityRequest{
				CauseID:         "cause123",
				DistanceToCover: 0,
				DistanceCovered: 8.2,
				Duration:        "45:30",
				Activity:        "Walking",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "invalid request body - zero distance_covered",
			requestBody: dto.RecordActivityRequest{
				CauseID:         "cause123",
				DistanceToCover: 10.5,
				DistanceCovered: 0,
				Duration:        "45:30",
				Activity:        "Walking",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "user not authenticated",
			requestBody: dto.RecordActivityRequest{
				CauseID:         "cause123",
				DistanceToCover: 10.5,
				DistanceCovered: 8.2,
				Duration:        "45:30",
				Activity:        "Walking",
			},
			expectedStatus: http.StatusUnauthorized,
			mockSetup: func() {
				// This test will use a modified router without authentication
			},
		},
		{
			name: "cause not found",
			requestBody: dto.RecordActivityRequest{
				CauseID:         "nonexistent",
				DistanceToCover: 10.5,
				DistanceCovered: 8.2,
				Duration:        "45:30",
				Activity:        "Walking",
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				// Mock cause runner creation (succeeds)
				mockCauseRunnerRepo.On("Create", mock.MatchedBy(func(r *challengeModel.CauseRunner) bool {
					return r.CauseID == "nonexistent" &&
						r.OwnerID == "test-user-id"
				})).Return(nil)

				// Mock cause lookup for update (fails because cause doesn't exist)
				mockCauseRepo.On("GetByID", "nonexistent").Return(nil, errors.New("cause not found"))
			},
		},
		{
			name: "repository error on create activity",
			requestBody: dto.RecordActivityRequest{
				CauseID:         "cause123",
				DistanceToCover: 10.5,
				DistanceCovered: 8.2,
				Duration:        "45:30",
				Activity:        "Walking",
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				mockCauseRepo.On("GetByID", "cause123").Return(&challengeModel.Cause{
					Base: model.Base{ID: "cause123"},
					Name: "Test Cause",
					DistanceCovered: 5.0,
				}, nil)

				mockCauseRunnerRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
				// Note: Update should not be called since Create fails
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCauseRepo.ExpectedCalls = nil
			mockCauseRunnerRepo.ExpectedCalls = nil

			tt.mockSetup()

			var testRouter *gin.Engine
			if tt.name == "user not authenticated" {
				// Create a separate router without authentication for this test
				testRouter = createUnauthenticatedCauseInteractionRouter(t)
			} else {
				testRouter = router
			}

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/causes/activity", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCauseRepo.AssertExpectations(t)
			mockCauseRunnerRepo.AssertExpectations(t)
		})
	}
}

func TestChallengeHandler_SponsorCause(t *testing.T) {
	router, _, _, _, _, mockSponsorCauseRepo, _, _ := setupCauseInteractionTest(t)

	tests := []struct {
		name           string
		requestBody    dto.SponsorCauseRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful cause sponsorship",
			requestBody: dto.SponsorCauseRequest{
				CauseID:    "cause123",
				Distance:   10.5,
				AmountPerKm: 5.0,
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				mockSponsorCauseRepo.On("Create", mock.MatchedBy(func(s *challengeModel.SponsorCause) bool {
					return s.SponsorID == "test-user-id" &&
						s.CauseID == "cause123" &&
						s.Distance == 10.5 &&
						s.AmountPerKm == 5.0
				})).Return(nil)
			},
		},
		{
			name: "invalid request body - missing cause_id",
			requestBody: dto.SponsorCauseRequest{
				Distance:   10.5,
				AmountPerKm: 5.0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "invalid request body - zero distance",
			requestBody: dto.SponsorCauseRequest{
				CauseID:    "cause123",
				Distance:   0,
				AmountPerKm: 5.0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "invalid request body - zero amount_per_km",
			requestBody: dto.SponsorCauseRequest{
				CauseID:    "cause123",
				Distance:   10.5,
				AmountPerKm: 0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "user not authenticated",
			requestBody: dto.SponsorCauseRequest{
				CauseID:    "cause123",
				Distance:   10.5,
				AmountPerKm: 5.0,
			},
			expectedStatus: http.StatusUnauthorized,
			mockSetup: func() {
				// This test will use a modified router without authentication
			},
		},
		{
			name: "repository error on create sponsorship",
			requestBody: dto.SponsorCauseRequest{
				CauseID:    "cause123",
				Distance:   10.5,
				AmountPerKm: 5.0,
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				mockSponsorCauseRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockSponsorCauseRepo.ExpectedCalls = nil

			tt.mockSetup()

			var testRouter *gin.Engine
			if tt.name == "user not authenticated" {
				// Create a separate router without authentication for this test
				testRouter = createUnauthenticatedCauseInteractionRouter(t)
			} else {
				testRouter = router
			}

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/causes/sponsor", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockSponsorCauseRepo.AssertExpectations(t)
		})
	}
}

func TestChallengeHandler_BuyCause(t *testing.T) {
	router, _, _, _, _, _, mockCauseBuyerRepo, _ := setupCauseInteractionTest(t)

	tests := []struct {
		name           string
		requestBody    dto.BuyCauseRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful cause purchase",
			requestBody: dto.BuyCauseRequest{
				CauseID: "cause123",
				Amount:  50.0,
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				mockCauseBuyerRepo.On("Create", mock.MatchedBy(func(b *challengeModel.CauseBuyer) bool {
					return b.BuyerID == "test-user-id" &&
						b.CauseID == "cause123" &&
						b.Amount == 50.0
				})).Return(nil)
			},
		},
		{
			name: "invalid request body - missing cause_id",
			requestBody: dto.BuyCauseRequest{
				Amount: 50.0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "invalid request body - zero amount",
			requestBody: dto.BuyCauseRequest{
				CauseID: "cause123",
				Amount:  0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "user not authenticated",
			requestBody: dto.BuyCauseRequest{
				CauseID: "cause123",
				Amount:  50.0,
			},
			expectedStatus: http.StatusUnauthorized,
			mockSetup: func() {
				// This test will use a modified router without authentication
			},
		},
		{
			name: "repository error on create purchase",
			requestBody: dto.BuyCauseRequest{
				CauseID: "cause123",
				Amount:  50.0,
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				mockCauseBuyerRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCauseBuyerRepo.ExpectedCalls = nil

			tt.mockSetup()

			var testRouter *gin.Engine
			if tt.name == "user not authenticated" {
				// Create a separate router without authentication for this test
				testRouter = createUnauthenticatedCauseInteractionRouter(t)
			} else {
				testRouter = router
			}

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/causes/buy", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCauseBuyerRepo.AssertExpectations(t)
		})
	}
}

// createUnauthenticatedCauseInteractionRouter creates a router without authentication middleware for testing unauthenticated scenarios
func createUnauthenticatedCauseInteractionRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Create mock repositories
	mockChallengeRepo := new(challengeMocks.MockChallengeRepository)
	mockCauseRepo := new(challengeMocks.MockCauseRepository)
	mockCauseRunnerRepo := new(challengeMocks.MockCauseRunnerRepository)
	mockSponsorChallengeRepo := new(challengeMocks.MockSponsorChallengeRepository)
	mockSponsorCauseRepo := new(challengeMocks.MockSponsorCauseRepository)
	mockCauseBuyerRepo := new(challengeMocks.MockCauseBuyerRepository)
	mockUserRepo := new(userMocks.MockUserRepository)

	// Create services with mock repositories
	challengeService := challenge.NewChallengeService(mockChallengeRepo, mockCauseRepo, mockCauseRunnerRepo, mockSponsorChallengeRepo, mockSponsorCauseRepo, mockCauseBuyerRepo)
	userService := user.NewUserService(mockUserRepo, nil)

	// Create handler with services
	challengeHandler := handler.NewChallengeHandler(challengeService, userService)

	// Setup router WITHOUT auth middleware
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup cause interaction routes
	api := router.Group("/api/v1")

	// Cause interaction routes
	api.POST("/causes/:id/join", challengeHandler.JoinCause)
	api.POST("/causes/activity", challengeHandler.RecordCauseActivity)
	api.POST("/causes/sponsor", challengeHandler.SponsorCause)
	api.POST("/causes/buy", challengeHandler.BuyCause)

	return router
}
