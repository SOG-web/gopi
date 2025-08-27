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
	challengeMocks "gopi.com/tests/mocks/challenge"
	userMocks "gopi.com/tests/mocks/user"
)

// Test setup helper for sponsorship tests
func setupSponsorshipTest(t *testing.T) (*gin.Engine, *challengeMocks.MockChallengeRepository, *challengeMocks.MockCauseRepository, *challengeMocks.MockCauseRunnerRepository, *challengeMocks.MockSponsorChallengeRepository, *challengeMocks.MockSponsorCauseRepository, *challengeMocks.MockCauseBuyerRepository, *userMocks.MockUserRepository) {
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

	// Setup challenge routes
	api := router.Group("/api/v1")

	// Challenge routes
	challenges := api.Group("/challenges")
	{
		challenges.POST("/sponsor", challengeHandler.SponsorChallenge)
	}

	return router, mockChallengeRepo, mockCauseRepo, mockCauseRunnerRepo, mockSponsorChallengeRepo, mockSponsorCauseRepo, mockCauseBuyerRepo, mockUserRepo
}

func TestChallengeHandler_SponsorChallenge(t *testing.T) {
	router, _, _, _, mockSponsorChallengeRepo, _, _, _ := setupSponsorshipTest(t)

	tests := []struct {
		name           string
		requestBody    dto.SponsorChallengeRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful challenge sponsorship",
			requestBody: dto.SponsorChallengeRequest{
				ChallengeID: "challenge123",
				Distance:    10.5,
				AmountPerKm: 5.0,
				BrandImg:    "brand.png",
				VideoUrl:    "video.mp4",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				mockSponsorChallengeRepo.On("Create", mock.MatchedBy(func(s *challengeModel.SponsorChallenge) bool {
					return s.SponsorID == "test-user-id" &&
						s.ChallengeID == "challenge123" &&
						s.Distance == 10.5 &&
						s.AmountPerKm == 5.0
				})).Return(nil)
			},
		},
		{
			name: "invalid request body - missing challenge_id",
			requestBody: dto.SponsorChallengeRequest{
				Distance:    10.5,
				AmountPerKm: 5.0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "invalid request body - zero distance",
			requestBody: dto.SponsorChallengeRequest{
				ChallengeID: "challenge123",
				Distance:    0,
				AmountPerKm: 5.0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "invalid request body - zero amount_per_km",
			requestBody: dto.SponsorChallengeRequest{
				ChallengeID: "challenge123",
				Distance:    10.5,
				AmountPerKm: 0,
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "user not authenticated",
			requestBody: dto.SponsorChallengeRequest{
				ChallengeID: "challenge123",
				Distance:    10.5,
				AmountPerKm: 5.0,
			},
			expectedStatus: http.StatusUnauthorized,
			mockSetup: func() {
				// This test will use a modified router without authentication
			},
		},
		{
			name: "repository error on create sponsorship",
			requestBody: dto.SponsorChallengeRequest{
				ChallengeID: "challenge123",
				Distance:    10.5,
				AmountPerKm: 5.0,
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				mockSponsorChallengeRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockSponsorChallengeRepo.ExpectedCalls = nil

			tt.mockSetup()

			var testRouter *gin.Engine
			if tt.name == "user not authenticated" {
				// Create a separate router without authentication for this test
				testRouter = createUnauthenticatedSponsorshipRouter(t)
			} else {
				testRouter = router
			}

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/challenges/sponsor", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockSponsorChallengeRepo.AssertExpectations(t)
		})
	}
}

// createUnauthenticatedSponsorshipRouter creates a router without authentication middleware for testing unauthenticated scenarios
func createUnauthenticatedSponsorshipRouter(t *testing.T) *gin.Engine {
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

	// Setup challenge routes
	api := router.Group("/api/v1")

	// Challenge routes
	challenges := api.Group("/challenges")
	{
		challenges.POST("/sponsor", challengeHandler.SponsorChallenge)
	}

	return router
}
