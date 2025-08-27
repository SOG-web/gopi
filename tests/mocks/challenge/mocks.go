package challenge

import (
	"github.com/stretchr/testify/mock"
	challengeModel "gopi.com/internal/domain/challenge/model"
)

// MockChallengeRepository implements the ChallengeRepository interface for testing
type MockChallengeRepository struct {
	mock.Mock
}

func (m *MockChallengeRepository) Create(challenge *challengeModel.Challenge) error {
	args := m.Called(challenge)
	return args.Error(0)
}

func (m *MockChallengeRepository) GetByID(id string) (*challengeModel.Challenge, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*challengeModel.Challenge), args.Error(1)
}

func (m *MockChallengeRepository) GetBySlug(slug string) (*challengeModel.Challenge, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*challengeModel.Challenge), args.Error(1)
}

func (m *MockChallengeRepository) GetByOwnerID(ownerID string) ([]*challengeModel.Challenge, error) {
	args := m.Called(ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challengeModel.Challenge), args.Error(1)
}

func (m *MockChallengeRepository) Update(challenge *challengeModel.Challenge) error {
	args := m.Called(challenge)
	return args.Error(0)
}

func (m *MockChallengeRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockChallengeRepository) List(limit, offset int) ([]*challengeModel.Challenge, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challengeModel.Challenge), args.Error(1)
}

// MockCauseRepository implements the CauseRepository interface for testing
type MockCauseRepository struct {
	mock.Mock
}

func (m *MockCauseRepository) Create(cause *challengeModel.Cause) error {
	args := m.Called(cause)
	return args.Error(0)
}

func (m *MockCauseRepository) GetByID(id string) (*challengeModel.Cause, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*challengeModel.Cause), args.Error(1)
}

func (m *MockCauseRepository) GetBySlug(slug string) (*challengeModel.Cause, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*challengeModel.Cause), args.Error(1)
}

func (m *MockCauseRepository) GetByChallengeID(challengeID string) ([]*challengeModel.Cause, error) {
	args := m.Called(challengeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challengeModel.Cause), args.Error(1)
}

func (m *MockCauseRepository) GetByOwnerID(ownerID string) ([]*challengeModel.Cause, error) {
	args := m.Called(ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challengeModel.Cause), args.Error(1)
}

func (m *MockCauseRepository) Update(cause *challengeModel.Cause) error {
	args := m.Called(cause)
	return args.Error(0)
}

func (m *MockCauseRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockCauseRunnerRepository implements the CauseRunnerRepository interface for testing
type MockCauseRunnerRepository struct {
	mock.Mock
}

func (m *MockCauseRunnerRepository) Create(runner *challengeModel.CauseRunner) error {
	args := m.Called(runner)
	return args.Error(0)
}

func (m *MockCauseRunnerRepository) GetByID(id string) (*challengeModel.CauseRunner, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*challengeModel.CauseRunner), args.Error(1)
}

func (m *MockCauseRunnerRepository) GetByCauseID(causeID string) ([]*challengeModel.CauseRunner, error) {
	args := m.Called(causeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challengeModel.CauseRunner), args.Error(1)
}

func (m *MockCauseRunnerRepository) GetByOwnerID(ownerID string) ([]*challengeModel.CauseRunner, error) {
	args := m.Called(ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challengeModel.CauseRunner), args.Error(1)
}

func (m *MockCauseRunnerRepository) Update(runner *challengeModel.CauseRunner) error {
	args := m.Called(runner)
	return args.Error(0)
}

func (m *MockCauseRunnerRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCauseRunnerRepository) GetLeaderboard() ([]*challengeModel.CauseRunner, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challengeModel.CauseRunner), args.Error(1)
}

// MockSponsorChallengeRepository implements the SponsorChallengeRepository interface for testing
type MockSponsorChallengeRepository struct {
	mock.Mock
}

func (m *MockSponsorChallengeRepository) Create(sponsor *challengeModel.SponsorChallenge) error {
	args := m.Called(sponsor)
	return args.Error(0)
}

func (m *MockSponsorChallengeRepository) GetByID(id string) (*challengeModel.SponsorChallenge, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*challengeModel.SponsorChallenge), args.Error(1)
}

func (m *MockSponsorChallengeRepository) GetByChallengeID(challengeID string) ([]*challengeModel.SponsorChallenge, error) {
	args := m.Called(challengeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challengeModel.SponsorChallenge), args.Error(1)
}

func (m *MockSponsorChallengeRepository) GetBySponsorID(sponsorID string) ([]*challengeModel.SponsorChallenge, error) {
	args := m.Called(sponsorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challengeModel.SponsorChallenge), args.Error(1)
}

func (m *MockSponsorChallengeRepository) Update(sponsor *challengeModel.SponsorChallenge) error {
	args := m.Called(sponsor)
	return args.Error(0)
}

func (m *MockSponsorChallengeRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockSponsorCauseRepository implements the SponsorCauseRepository interface for testing
type MockSponsorCauseRepository struct {
	mock.Mock
}

func (m *MockSponsorCauseRepository) Create(sponsor *challengeModel.SponsorCause) error {
	args := m.Called(sponsor)
	return args.Error(0)
}

func (m *MockSponsorCauseRepository) GetByID(id string) (*challengeModel.SponsorCause, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*challengeModel.SponsorCause), args.Error(1)
}

func (m *MockSponsorCauseRepository) GetByCauseID(causeID string) ([]*challengeModel.SponsorCause, error) {
	args := m.Called(causeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challengeModel.SponsorCause), args.Error(1)
}

func (m *MockSponsorCauseRepository) GetBySponsorID(sponsorID string) ([]*challengeModel.SponsorCause, error) {
	args := m.Called(sponsorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challengeModel.SponsorCause), args.Error(1)
}

func (m *MockSponsorCauseRepository) Update(sponsor *challengeModel.SponsorCause) error {
	args := m.Called(sponsor)
	return args.Error(0)
}

func (m *MockSponsorCauseRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockCauseBuyerRepository implements the CauseBuyerRepository interface for testing
type MockCauseBuyerRepository struct {
	mock.Mock
}

func (m *MockCauseBuyerRepository) Create(buyer *challengeModel.CauseBuyer) error {
	args := m.Called(buyer)
	return args.Error(0)
}

func (m *MockCauseBuyerRepository) GetByID(id string) (*challengeModel.CauseBuyer, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*challengeModel.CauseBuyer), args.Error(1)
}

func (m *MockCauseBuyerRepository) GetByCauseID(causeID string) ([]*challengeModel.CauseBuyer, error) {
	args := m.Called(causeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challengeModel.CauseBuyer), args.Error(1)
}

func (m *MockCauseBuyerRepository) GetByBuyerID(buyerID string) ([]*challengeModel.CauseBuyer, error) {
	args := m.Called(buyerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challengeModel.CauseBuyer), args.Error(1)
}

func (m *MockCauseBuyerRepository) Update(buyer *challengeModel.CauseBuyer) error {
	args := m.Called(buyer)
	return args.Error(0)
}

func (m *MockCauseBuyerRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
