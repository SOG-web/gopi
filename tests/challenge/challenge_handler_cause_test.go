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

// Test setup helper for cause tests
func setupCauseTest(t *testing.T) (*gin.Engine, *challengeMocks.MockChallengeRepository, *challengeMocks.MockCauseRepository, *challengeMocks.MockCauseRunnerRepository, *challengeMocks.MockSponsorChallengeRepository, *challengeMocks.MockSponsorCauseRepository, *challengeMocks.MockCauseBuyerRepository, *userMocks.MockUserRepository) {
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

	// Setup cause routes
	api := router.Group("/api/v1")

	// Cause routes
	api.POST("/causes", challengeHandler.CreateCause)
	api.GET("/challenges/:challenge_id/causes", challengeHandler.GetCausesByChallenge)
	api.GET("/causes/:id", challengeHandler.GetCauseByID)

	return router, mockChallengeRepo, mockCauseRepo, mockCauseRunnerRepo, mockSponsorChallengeRepo, mockSponsorCauseRepo, mockCauseBuyerRepo, mockUserRepo
}

func TestChallengeHandler_CreateCause(t *testing.T) {
	router, _, mockCauseRepo, _, _, _, _, mockUserRepo := setupCauseTest(t)

	tests := []struct {
		name           string
		requestBody    dto.CreateCauseRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful cause creation",
			requestBody: dto.CreateCauseRequest{
				ChallengeID:        "challenge123",
				Name:               "Test Cause",
				Problem:            "Environmental pollution",
				Solution:           "Plant trees",
				ProductDescription: "Tree planting initiative",
				Activity:           "Walking",
				Location:           "Park",
				Description:        "Help save the environment",
				IsCommercial:       false,
				AmountPerPiece:     10.0,
				FundAmount:         100.0,
				WillingAmount:      50.0,
				UnitPrice:          5.0,
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

				// Mock cause creation
				mockCauseRepo.On("Create", mock.MatchedBy(func(c *challengeModel.Cause) bool {
					return c.Name == "Test Cause" &&
						c.ChallengeID == "challenge123" &&
						c.OwnerID == "test-user-id" &&
						c.Activity == challengeModel.ActivityWalking
				})).Return(nil)
			},
		},
		{
			name: "invalid request body - missing challenge_id",
			requestBody: dto.CreateCauseRequest{
				Name: "Test Cause",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "invalid request body - missing name",
			requestBody: dto.CreateCauseRequest{
				ChallengeID: "challenge123",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "invalid request body - invalid activity",
			requestBody: dto.CreateCauseRequest{
				ChallengeID: "challenge123",
				Name:        "Test Cause",
				Activity:    "InvalidActivity",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "user not authenticated",
			requestBody: dto.CreateCauseRequest{
				ChallengeID: "challenge123",
				Name:        "Test Cause",
			},
			expectedStatus: http.StatusUnauthorized,
			mockSetup: func() {
				// This test will use a modified router without authentication
			},
		},
		{
			name: "cause creation repository error",
			requestBody: dto.CreateCauseRequest{
				ChallengeID: "challenge123",
				Name:        "Test Cause",
				Activity:    "Walking",
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				// Mock user lookup
				expectedUser := &userModel.User{
					Base: model.Base{
						ID: "test-user-id",
					},
					Username: "testuser",
				}
				mockUserRepo.On("GetByID", "test-user-id").Return(expectedUser, nil)

				// Mock cause creation failure
				mockCauseRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCauseRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			var testRouter *gin.Engine
			if tt.name == "user not authenticated" {
				// Create a separate router without authentication for this test
				testRouter = createUnauthenticatedCauseRouter(t)
			} else {
				testRouter = router
			}

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/causes", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCauseRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestChallengeHandler_GetCausesByChallenge(t *testing.T) {
	router, _, mockCauseRepo, _, _, _, _, mockUserRepo := setupCauseTest(t)

	tests := []struct {
		name           string
		challengeID    string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful get causes by challenge",
			challengeID:    "challenge123",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedCauses := []*challengeModel.Cause{
					{
						Base: model.Base{
							ID:        "cause1",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						ChallengeID: "challenge123",
						OwnerID:     "user1",
						Name:        "Cause 1",
						Activity:    challengeModel.ActivityWalking,
					},
					{
						Base: model.Base{
							ID:        "cause2",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						ChallengeID: "challenge123",
						OwnerID:     "user2",
						Name:        "Cause 2",
						Activity:    challengeModel.ActivityRunning,
					},
				}

				mockCauseRepo.On("GetByChallengeID", "challenge123").Return(expectedCauses, nil)

				// Mock user lookups for cause owners
				user1 := &userModel.User{
					Base: model.Base{
						ID: "user1",
					},
					Username: "user1",
				}
				user2 := &userModel.User{
					Base: model.Base{
						ID: "user2",
					},
					Username: "user2",
				}

				mockUserRepo.On("GetByID", "user1").Return(user1, nil)
				mockUserRepo.On("GetByID", "user2").Return(user2, nil)
			},
		},
		{
			name:           "repository error on get causes",
			challengeID:    "challenge123",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				mockCauseRepo.On("GetByChallengeID", "challenge123").Return(nil, errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCauseRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodGet, "/api/v1/challenges/"+tt.challengeID+"/causes", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCauseRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestChallengeHandler_GetCauseByID(t *testing.T) {
	router, _, mockCauseRepo, _, _, _, _, mockUserRepo := setupCauseTest(t)

	tests := []struct {
		name           string
		causeID        string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful get cause by id",
			causeID:        "cause123",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedCause := &challengeModel.Cause{
					Base: model.Base{
						ID:        "cause123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					ChallengeID: "challenge123",
					OwnerID:     "user123",
					Name:        "Test Cause",
					Activity:    challengeModel.ActivityWalking,
				}

				mockCauseRepo.On("GetByID", "cause123").Return(expectedCause, nil)

				// Mock owner lookup for response
				expectedUser := &userModel.User{
					Base: model.Base{
						ID: "user123",
					},
					Username: "testuser",
				}
				mockUserRepo.On("GetByID", "user123").Return(expectedUser, nil)
			},
		},
		{
			name:           "cause not found",
			causeID:        "nonexistent",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockCauseRepo.On("GetByID", "nonexistent").Return(nil, errors.New("cause not found"))
			},
		},
		{
			name:           "repository error on get cause",
			causeID:        "cause123",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockCauseRepo.On("GetByID", "cause123").Return(nil, errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCauseRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodGet, "/api/v1/causes/"+tt.causeID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCauseRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

// createUnauthenticatedCauseRouter creates a router without authentication middleware for testing unauthenticated scenarios
func createUnauthenticatedCauseRouter(t *testing.T) *gin.Engine {
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

	// Setup cause routes
	api := router.Group("/api/v1")

	// Cause routes
	api.POST("/causes", challengeHandler.CreateCause)

	return router
}
