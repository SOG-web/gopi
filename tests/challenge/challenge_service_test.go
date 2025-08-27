package challenge_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopi.com/internal/app/challenge"
	challengeModel "gopi.com/internal/domain/challenge/model"
	"gopi.com/internal/domain/model"
	challengeMocks "gopi.com/tests/mocks/challenge"
)

func TestChallengeService_CreateChallenge(t *testing.T) {
	// Setup mock repositories
	mockChallengeRepo := new(challengeMocks.MockChallengeRepository)
	mockCauseRepo := new(challengeMocks.MockCauseRepository)
	mockCauseRunnerRepo := new(challengeMocks.MockCauseRunnerRepository)
	mockSponsorChallengeRepo := new(challengeMocks.MockSponsorChallengeRepository)
	mockSponsorCauseRepo := new(challengeMocks.MockSponsorCauseRepository)
	mockCauseBuyerRepo := new(challengeMocks.MockCauseBuyerRepository)

	// Create service with mocks
	service := challenge.NewChallengeService(mockChallengeRepo, mockCauseRepo, mockCauseRunnerRepo, mockSponsorChallengeRepo, mockSponsorCauseRepo, mockCauseBuyerRepo)

	tests := []struct {
		name        string
		ownerID     string
		nameInput   string
		expectedErr bool
		mockSetup   func()
	}{
		{
			name:        "successful challenge creation",
			ownerID:     "user123",
			nameInput:   "Test Challenge",
			expectedErr: false,
			mockSetup: func() {
				mockChallengeRepo.On("Create", mock.MatchedBy(func(c *challengeModel.Challenge) bool {
					return c.OwnerID == "user123" &&
						c.Name == "Test Challenge" &&
						c.Mode == challengeModel.ChallengeModeF
				})).Return(nil)
			},
		},
		{
			name:        "repository error on creation",
			ownerID:     "user123",
			nameInput:   "Test Challenge",
			expectedErr: true,
			mockSetup: func() {
				mockChallengeRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockChallengeRepo.ExpectedCalls = nil

			tt.mockSetup()

			result, err := service.CreateChallenge(
				tt.ownerID,
				tt.nameInput,
				"Test Description",
				"Complete 5km run",
				"Fitness",
				"Park",
				challengeModel.ChallengeModeF,
				5.0,
				100.0,
				2.0,
				"2024-01-01",
				"2024-01-31",
				3,
			)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.ownerID, result.OwnerID)
				assert.Equal(t, tt.nameInput, result.Name)
			}

			mockChallengeRepo.AssertExpectations(t)
		})
	}
}

func TestChallengeService_GetChallengeByID(t *testing.T) {
	// Setup mock repositories
	mockChallengeRepo := new(challengeMocks.MockChallengeRepository)
	mockCauseRepo := new(challengeMocks.MockCauseRepository)
	mockCauseRunnerRepo := new(challengeMocks.MockCauseRunnerRepository)
	mockSponsorChallengeRepo := new(challengeMocks.MockSponsorChallengeRepository)
	mockSponsorCauseRepo := new(challengeMocks.MockSponsorCauseRepository)
	mockCauseBuyerRepo := new(challengeMocks.MockCauseBuyerRepository)

	// Create service with mocks
	service := challenge.NewChallengeService(mockChallengeRepo, mockCauseRepo, mockCauseRunnerRepo, mockSponsorChallengeRepo, mockSponsorCauseRepo, mockCauseBuyerRepo)

	tests := []struct {
		name        string
		challengeID string
		expectedErr bool
		mockSetup   func()
	}{
		{
			name:        "successful challenge retrieval",
			challengeID: "challenge123",
			expectedErr: false,
			mockSetup: func() {
				expectedChallenge := &challengeModel.Challenge{
					Base: model.Base{
						ID: "challenge123",
					},
					Name:    "Test Challenge",
					OwnerID: "user123",
				}
				mockChallengeRepo.On("GetByID", "challenge123").Return(expectedChallenge, nil)
			},
		},
		{
			name:        "challenge not found",
			challengeID: "nonexistent",
			expectedErr: true,
			mockSetup: func() {
				mockChallengeRepo.On("GetByID", "nonexistent").Return(nil, errors.New("not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockChallengeRepo.ExpectedCalls = nil

			tt.mockSetup()

			result, err := service.GetChallengeByID(tt.challengeID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.challengeID, result.ID)
			}

			mockChallengeRepo.AssertExpectations(t)
		})
	}
}

func TestChallengeService_JoinChallenge(t *testing.T) {
	// Setup mock repositories
	mockChallengeRepo := new(challengeMocks.MockChallengeRepository)
	mockCauseRepo := new(challengeMocks.MockCauseRepository)
	mockCauseRunnerRepo := new(challengeMocks.MockCauseRunnerRepository)
	mockSponsorChallengeRepo := new(challengeMocks.MockSponsorChallengeRepository)
	mockSponsorCauseRepo := new(challengeMocks.MockSponsorCauseRepository)
	mockCauseBuyerRepo := new(challengeMocks.MockCauseBuyerRepository)

	// Create service with mocks
	service := challenge.NewChallengeService(mockChallengeRepo, mockCauseRepo, mockCauseRunnerRepo, mockSponsorChallengeRepo, mockSponsorCauseRepo, mockCauseBuyerRepo)

	tests := []struct {
		name        string
		challengeID string
		userID      string
		expectedErr bool
		mockSetup   func()
	}{
		{
			name:        "successful challenge join",
			challengeID: "challenge123",
			userID:      "user123",
			expectedErr: false,
			mockSetup: func() {
				expectedChallenge := &challengeModel.Challenge{
					Base: model.Base{
						ID: "challenge123",
					},
					Name:    "Test Challenge",
					OwnerID: "owner123",
				}
				mockChallengeRepo.On("GetByID", "challenge123").Return(expectedChallenge, nil)
				mockChallengeRepo.On("Update", mock.MatchedBy(func(c *challengeModel.Challenge) bool {
					return c.ID == "challenge123"
				})).Return(nil)
			},
		},
		{
			name:        "challenge not found",
			challengeID: "nonexistent",
			userID:      "user123",
			expectedErr: true,
			mockSetup: func() {
				mockChallengeRepo.On("GetByID", "nonexistent").Return(nil, errors.New("not found"))
			},
		},
		{
			name:        "repository error on update",
			challengeID: "challenge123",
			userID:      "user123",
			expectedErr: true,
			mockSetup: func() {
				expectedChallenge := &challengeModel.Challenge{
					Base: model.Base{
						ID: "challenge123",
					},
					Name:    "Test Challenge",
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

			err := service.JoinChallenge(tt.challengeID, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockChallengeRepo.AssertExpectations(t)
		})
	}
}

func TestChallengeService_ListChallenges(t *testing.T) {
	// Setup mock repositories
	mockChallengeRepo := new(challengeMocks.MockChallengeRepository)
	mockCauseRepo := new(challengeMocks.MockCauseRepository)
	mockCauseRunnerRepo := new(challengeMocks.MockCauseRunnerRepository)
	mockSponsorChallengeRepo := new(challengeMocks.MockSponsorChallengeRepository)
	mockSponsorCauseRepo := new(challengeMocks.MockSponsorCauseRepository)
	mockCauseBuyerRepo := new(challengeMocks.MockCauseBuyerRepository)

	// Create service with mocks
	service := challenge.NewChallengeService(mockChallengeRepo, mockCauseRepo, mockCauseRunnerRepo, mockSponsorChallengeRepo, mockSponsorCauseRepo, mockCauseBuyerRepo)

	tests := []struct {
		name        string
		limit       int
		offset      int
		expectedErr bool
		mockSetup   func()
	}{
		{
			name:        "successful challenge listing",
			limit:       10,
			offset:      0,
			expectedErr: false,
			mockSetup: func() {
				expectedChallenges := []*challengeModel.Challenge{
					{
						Base: model.Base{
							ID: "challenge1",
						},
						Name:    "Challenge 1",
						OwnerID: "user1",
					},
					{
						Base: model.Base{
							ID: "challenge2",
						},
						Name:    "Challenge 2",
						OwnerID: "user2",
					},
				}
				mockChallengeRepo.On("List", 10, 0).Return(expectedChallenges, nil)
			},
		},
		{
			name:        "repository error on listing",
			limit:       10,
			offset:      0,
			expectedErr: true,
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

			result, err := service.ListChallenges(tt.limit, tt.offset)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockChallengeRepo.AssertExpectations(t)
		})
	}
}

func TestChallengeService_GetLeaderboard(t *testing.T) {
	// Setup mock repositories
	mockChallengeRepo := new(challengeMocks.MockChallengeRepository)
	mockCauseRepo := new(challengeMocks.MockCauseRepository)
	mockCauseRunnerRepo := new(challengeMocks.MockCauseRunnerRepository)
	mockSponsorChallengeRepo := new(challengeMocks.MockSponsorChallengeRepository)
	mockSponsorCauseRepo := new(challengeMocks.MockSponsorCauseRepository)
	mockCauseBuyerRepo := new(challengeMocks.MockCauseBuyerRepository)

	// Create service with mocks
	service := challenge.NewChallengeService(mockChallengeRepo, mockCauseRepo, mockCauseRunnerRepo, mockSponsorChallengeRepo, mockSponsorCauseRepo, mockCauseBuyerRepo)

	tests := []struct {
		name        string
		expectedErr bool
		mockSetup   func()
	}{
		{
			name:        "successful leaderboard retrieval",
			expectedErr: false,
			mockSetup: func() {
				expectedRunners := []*challengeModel.CauseRunner{
					{
						Base: model.Base{
							ID: "runner1",
						},
						CauseID:         "cause1",
						OwnerID:         "user1",
						DistanceCovered: 10.5,
					},
				}
				mockCauseRunnerRepo.On("GetLeaderboard").Return(expectedRunners, nil)
			},
		},
		{
			name:        "repository error on leaderboard retrieval",
			expectedErr: true,
			mockSetup: func() {
				mockCauseRunnerRepo.On("GetLeaderboard").Return(nil, errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockCauseRunnerRepo.ExpectedCalls = nil

			tt.mockSetup()

			result, err := service.GetLeaderboard()

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockCauseRunnerRepo.AssertExpectations(t)
		})
	}
}
