package challenge_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopi.com/api/http/handler"
	"gopi.com/internal/app/challenge"
	"gopi.com/internal/app/user"
	challengeModel "gopi.com/internal/domain/challenge/model"
	"gopi.com/internal/domain/model"
	challengeMocks "gopi.com/tests/mocks/challenge"
	userMocks "gopi.com/tests/mocks/user"
)

// Test setup helper for membership tests
func setupMembershipTest(t *testing.T) (*gin.Engine, *challengeMocks.MockChallengeRepository, *challengeMocks.MockCauseRepository, *challengeMocks.MockCauseRunnerRepository, *challengeMocks.MockSponsorChallengeRepository, *challengeMocks.MockSponsorCauseRepository, *challengeMocks.MockCauseBuyerRepository, *userMocks.MockUserRepository) {
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
		challenges.POST("/:id/join", challengeHandler.JoinChallenge)
	}

	return router, mockChallengeRepo, mockCauseRepo, mockCauseRunnerRepo, mockSponsorChallengeRepo, mockSponsorCauseRepo, mockCauseBuyerRepo, mockUserRepo
}

func TestChallengeHandler_JoinChallenge(t *testing.T) {
	router, mockChallengeRepo, _, _, _, _, _, _ := setupMembershipTest(t)

	tests := []struct {
		name           string
		challengeID    string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful challenge join",
			challengeID:    "challenge123",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedChallenge := &challengeModel.Challenge{
					Base: model.Base{
						ID:        "challenge123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:    "Test Challenge",
					Slug:    "test-challenge",
					OwnerID: "owner123",
				}

				mockChallengeRepo.On("GetByID", "challenge123").Return(expectedChallenge, nil)
				mockChallengeRepo.On("Update", mock.MatchedBy(func(c *challengeModel.Challenge) bool {
					return c.ID == "challenge123"
				})).Return(nil)
			},
		},
		{
			name:           "challenge not found",
			challengeID:    "nonexistent",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				mockChallengeRepo.On("GetByID", "nonexistent").Return(nil, errors.New("challenge not found"))
			},
		},
		{
			name:           "user not authenticated",
			challengeID:    "challenge123",
			expectedStatus: http.StatusUnauthorized,
			mockSetup: func() {
				// This test will use a modified router without authentication
			},
		},
		{
			name:           "repository error on update",
			challengeID:    "challenge123",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				expectedChallenge := &challengeModel.Challenge{
					Base: model.Base{
						ID:        "challenge123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:    "Test Challenge",
					Slug:    "test-challenge",
					OwnerID: "owner123",
				}

				mockChallengeRepo.On("GetByID", "challenge123").Return(expectedChallenge, nil)
				mockChallengeRepo.On("Update", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockChallengeRepo.ExpectedCalls = nil

			tt.mockSetup()

			var testRouter *gin.Engine
			if tt.name == "user not authenticated" {
				// Create a separate router without authentication for this test
				testRouter = createUnauthenticatedMembershipRouter(t)
			} else {
				testRouter = router
			}

			req, _ := http.NewRequest(http.MethodPost, "/api/v1/challenges/"+tt.challengeID+"/join", nil)
			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockChallengeRepo.AssertExpectations(t)
		})
	}
}

// createUnauthenticatedMembershipRouter creates a router without authentication middleware for testing unauthenticated scenarios
func createUnauthenticatedMembershipRouter(t *testing.T) *gin.Engine {
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
		challenges.POST("/:id/join", challengeHandler.JoinChallenge)
	}

	return router
}
