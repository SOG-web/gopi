package challenge_test

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
	"gopi.com/internal/app/challenge"
	"gopi.com/internal/app/user"
	challengeModel "gopi.com/internal/domain/challenge/model"
	"gopi.com/internal/domain/model"
	userModel "gopi.com/internal/domain/user/model"
	challengeMocks "gopi.com/tests/mocks/challenge"
	userMocks "gopi.com/tests/mocks/user"
)

// Test setup helper for challenge tests
func setupChallengeTest(t *testing.T) (*gin.Engine, *challengeMocks.MockChallengeRepository, *challengeMocks.MockCauseRepository, *challengeMocks.MockCauseRunnerRepository, *challengeMocks.MockSponsorChallengeRepository, *challengeMocks.MockSponsorCauseRepository, *challengeMocks.MockCauseBuyerRepository, *userMocks.MockUserRepository) {
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
		challenges.GET("", challengeHandler.GetChallenges)
		challenges.GET("/slug/:slug", challengeHandler.GetChallengeBySlug)
		challenges.GET("/id/:id", challengeHandler.GetChallengeByID)
		challenges.POST("", challengeHandler.CreateChallenge)
	}

	return router, mockChallengeRepo, mockCauseRepo, mockCauseRunnerRepo, mockSponsorChallengeRepo, mockSponsorCauseRepo, mockCauseBuyerRepo, mockUserRepo
}

// createUnauthenticatedRouter creates a router without authentication middleware for testing unauthenticated scenarios
func createUnauthenticatedRouter(t *testing.T) *gin.Engine {
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
		challenges.POST("", challengeHandler.CreateChallenge)
	}

	return router
}

func TestChallengeHandler_CreateChallenge(t *testing.T) {
	router, mockChallengeRepo, _, _, _, _, _, mockUserRepo := setupChallengeTest(t)

	tests := []struct {
		name           string
		requestBody    dto.CreateChallengeRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful challenge creation",
			requestBody: dto.CreateChallengeRequest{
				Name:            "Test Challenge",
				Description:     "A test challenge",
				Mode:            "Free",
				Condition:       "Complete 5km run",
				Goal:            "Fitness",
				Location:        "Park",
				DistanceToCover: 5.0,
				TargetAmount:    100.0,
				StartDuration:   "2024-01-01",
				EndDuration:     "2024-01-31",
				NoOfWinner:      3,
			},
			expectedStatus: http.StatusCreated,
			mockSetup: func() {
				// Mock user lookup
				expectedUser := &userModel.User{
					Base: model.Base{
						ID:        "test-user-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username: "testuser",
				}
				mockUserRepo.On("GetByID", "test-user-id").Return(expectedUser, nil)

				// Mock challenge creation
				mockChallengeRepo.On("Create", mock.MatchedBy(func(c *challengeModel.Challenge) bool {
					return c.Name == "Test Challenge" &&
						c.Description == "A test challenge" &&
						c.Mode == "Free" &&
						c.OwnerID == "test-user-id"
				})).Return(nil)

				// The service returns the challenge directly, so no GetByID call is needed
			},
		},
		{
			name: "invalid request body - missing name",
			requestBody: dto.CreateChallengeRequest{
				Description: "A test challenge",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "user not authenticated",
			requestBody: dto.CreateChallengeRequest{
				Name: "Test Challenge",
			},
			expectedStatus: http.StatusUnauthorized,
			mockSetup: func() {
				// This test will use a modified router without user authentication
			},
		},
		{
			name: "challenge creation repository error",
			requestBody: dto.CreateChallengeRequest{
				Name:        "Test Challenge",
				Description: "A test challenge",
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				// Mock user lookup
				expectedUser := &userModel.User{
					Base: model.Base{
						ID:        "test-user-id",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username: "testuser",
				}
				mockUserRepo.On("GetByID", "test-user-id").Return(expectedUser, nil)

				// Mock challenge creation failure
				mockChallengeRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockChallengeRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			var testRouter *gin.Engine
			if tt.name == "user not authenticated" {
				// Create a separate router without authentication for this test
				testRouter = createUnauthenticatedRouter(t)
			} else {
				testRouter = router
			}

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/challenges", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockChallengeRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestChallengeHandler_GetChallenges(t *testing.T) {
	router, mockChallengeRepo, _, _, _, _, _, mockUserRepo := setupChallengeTest(t)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful get challenges with default pagination",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedChallenges := []*challengeModel.Challenge{
					{
						Base: model.Base{
							ID:        "challenge1",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Name:    "Challenge 1",
						Slug:    "challenge-1",
						OwnerID: "user1",
					},
					{
						Base: model.Base{
							ID:        "challenge2",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Name:    "Challenge 2",
						Slug:    "challenge-2",
						OwnerID: "user2",
					},
				}

				mockChallengeRepo.On("List", 10, 0).Return(expectedChallenges, nil)

				// Mock user lookups for challenge owners
				user1 := &userModel.User{
					Base: model.Base{
						ID:        "user1",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username: "user1",
				}
				user2 := &userModel.User{
					Base: model.Base{
						ID:        "user2",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username: "user2",
				}

				mockUserRepo.On("GetByID", "user1").Return(user1, nil)
				mockUserRepo.On("GetByID", "user2").Return(user2, nil)
			},
		},
		{
			name:           "get challenges with custom pagination",
			queryParams:    "?page=2&limit=5",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedChallenges := []*challengeModel.Challenge{
					{
						Base: model.Base{
							ID:        "challenge3",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Name:    "Challenge 3",
						Slug:    "challenge-3",
						OwnerID: "user3",
					},
				}

				// For page=2&limit=5: offset = (2-1)*5 = 5
				mockChallengeRepo.On("List", 5, 5).Return(expectedChallenges, nil)

				// Mock user lookup for challenge owner
				user3 := &userModel.User{
					Base: model.Base{
						ID:        "user3",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username: "user3",
				}

				mockUserRepo.On("GetByID", "user3").Return(user3, nil)
			},
		},
		{
			name:           "repository error on get challenges",
			queryParams:    "",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				mockChallengeRepo.On("List", 10, 0).Return(nil, errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockChallengeRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodGet, "/api/v1/challenges"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockChallengeRepo.AssertExpectations(t)
		})
	}
}

func TestChallengeHandler_GetChallengeByID(t *testing.T) {
	router, mockChallengeRepo, _, _, _, _, _, mockUserRepo := setupChallengeTest(t)

	tests := []struct {
		name           string
		challengeID    string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful get challenge by id",
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
					OwnerID: "user123",
				}

				mockChallengeRepo.On("GetByID", "challenge123").Return(expectedChallenge, nil)

				// Mock owner lookup for response
				expectedUser := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username: "testuser",
				}
				mockUserRepo.On("GetByID", "user123").Return(expectedUser, nil)
			},
		},
		{
			name:           "challenge not found",
			challengeID:    "nonexistent",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockChallengeRepo.On("GetByID", "nonexistent").Return(nil, errors.New("challenge not found"))
			},
		},
		{
			name:           "repository error on get challenge",
			challengeID:    "challenge123",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockChallengeRepo.On("GetByID", "challenge123").Return(nil, errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockChallengeRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodGet, "/api/v1/challenges/id/"+tt.challengeID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockChallengeRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestChallengeHandler_GetChallengeBySlug(t *testing.T) {
	router, mockChallengeRepo, _, _, _, _, _, mockUserRepo := setupChallengeTest(t)

	tests := []struct {
		name           string
		slug           string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful get challenge by slug",
			slug:           "test-challenge",
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
					OwnerID: "user123",
				}

				mockChallengeRepo.On("GetBySlug", "test-challenge").Return(expectedChallenge, nil)

				// Mock owner lookup for response
				expectedUser := &userModel.User{
					Base: model.Base{
						ID:        "user123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Username: "testuser",
				}
				mockUserRepo.On("GetByID", "user123").Return(expectedUser, nil)
			},
		},
		{
			name:           "challenge not found by slug",
			slug:           "nonexistent",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockChallengeRepo.On("GetBySlug", "nonexistent").Return(nil, errors.New("challenge not found"))
			},
		},
		{
			name:           "repository error on get challenge by slug",
			slug:           "test-challenge",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockChallengeRepo.On("GetBySlug", "test-challenge").Return(nil, errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockChallengeRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodGet, "/api/v1/challenges/slug/"+tt.slug, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockChallengeRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}
