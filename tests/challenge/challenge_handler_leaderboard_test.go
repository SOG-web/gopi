package challenge_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gopi.com/api/http/handler"
	"gopi.com/internal/app/challenge"
	"gopi.com/internal/app/user"
	challengeModel "gopi.com/internal/domain/challenge/model"
	userModel "gopi.com/internal/domain/user/model"
	"gopi.com/internal/domain/model"
	challengeMocks "gopi.com/tests/mocks/challenge"
	userMocks "gopi.com/tests/mocks/user"
)

// Test setup helper for leaderboard tests
func setupLeaderboardTest(t *testing.T) (*gin.Engine, *challengeMocks.MockChallengeRepository, *challengeMocks.MockCauseRepository, *challengeMocks.MockCauseRunnerRepository, *challengeMocks.MockSponsorChallengeRepository, *challengeMocks.MockSponsorCauseRepository, *challengeMocks.MockCauseBuyerRepository, *userMocks.MockUserRepository) {
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

	// Setup leaderboard route
	api := router.Group("/api/v1")

	// Leaderboard route
	api.GET("/leaderboard", challengeHandler.GetLeaderboard)

	return router, mockChallengeRepo, mockCauseRepo, mockCauseRunnerRepo, mockSponsorChallengeRepo, mockSponsorCauseRepo, mockCauseBuyerRepo, mockUserRepo
}

func TestChallengeHandler_GetLeaderboard(t *testing.T) {
	router, _, _, mockCauseRunnerRepo, _, _, _, mockUserRepo := setupLeaderboardTest(t)

	tests := []struct {
		name           string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful leaderboard retrieval",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedRunners := []*challengeModel.CauseRunner{
					{
						Base: model.Base{
							ID:        "runner1",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						CauseID:         "cause1",
						OwnerID:         "user1",
						DistanceToCover: 10.5,
						DistanceCovered: 8.2,
						Duration:        "45:30",
						MoneyRaised:     25.0,
						Activity:        "Walking",
						CoverImage:      "image1.jpg",
					},
					{
						Base: model.Base{
							ID:        "runner2",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						CauseID:         "cause2",
						OwnerID:         "user2",
						DistanceToCover: 15.0,
						DistanceCovered: 12.5,
						Duration:        "60:00",
						MoneyRaised:     40.0,
						Activity:        "Running",
						CoverImage:      "image2.jpg",
					},
				}

				mockCauseRunnerRepo.On("GetLeaderboard").Return(expectedRunners, nil)

				// Mock user lookups for runner owners
				user1 := &userModel.User{
					Base: model.Base{
						ID: "user1",
					},
					Username: "runner1",
				}
				user2 := &userModel.User{
					Base: model.Base{
						ID: "user2",
					},
					Username: "runner2",
				}

				mockUserRepo.On("GetByID", "user1").Return(user1, nil)
				mockUserRepo.On("GetByID", "user2").Return(user2, nil)
			},
		},
		{
			name:           "successful empty leaderboard",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				mockCauseRunnerRepo.On("GetLeaderboard").Return([]*challengeModel.CauseRunner{}, nil)
			},
		},
		{
			name:           "repository error on get leaderboard",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				mockCauseRunnerRepo.On("GetLeaderboard").Return(nil, assert.AnError)
			},
		},
		{
			name:           "user lookup failure - skip runner",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				expectedRunners := []*challengeModel.CauseRunner{
					{
						Base: model.Base{
							ID:        "runner1",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						CauseID:         "cause1",
						OwnerID:         "user1",
						DistanceToCover: 10.5,
						DistanceCovered: 8.2,
						Duration:        "45:30",
						MoneyRaised:     25.0,
						Activity:        "Walking",
					},
					{
						Base: model.Base{
							ID:        "runner2",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						CauseID:         "cause2",
						OwnerID:         "user2",
						DistanceToCover: 15.0,
						DistanceCovered: 12.5,
						Duration:        "60:00",
						MoneyRaised:     40.0,
						Activity:        "Running",
					},
				}

				mockCauseRunnerRepo.On("GetLeaderboard").Return(expectedRunners, nil)

				// Mock first user lookup to succeed
				user1 := &userModel.User{
					Base: model.Base{
						ID: "user1",
					},
					Username: "runner1",
				}
				mockUserRepo.On("GetByID", "user1").Return(user1, nil)

				// Mock second user lookup to fail (user not found)
				mockUserRepo.On("GetByID", "user2").Return(nil, assert.AnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCauseRunnerRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil

			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodGet, "/api/v1/leaderboard", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCauseRunnerRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}
