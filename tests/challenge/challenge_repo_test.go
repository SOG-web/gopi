package challenge_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	gormmodel "gopi.com/internal/data/challenge/model/gorm"
	"gopi.com/internal/data/challenge/repo"
	challengeModel "gopi.com/internal/domain/challenge/model"
	"gopi.com/internal/domain/model"
)

// Setup in-memory database for testing
func setupChallengeTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the schema - include all related models for proper relationship handling
	err = db.AutoMigrate(
		&gormmodel.Challenge{},
		&gormmodel.Cause{},
		&gormmodel.CauseRunner{},
		&gormmodel.SponsorChallenge{},
		&gormmodel.SponsorCause{},
		&gormmodel.CauseBuyer{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestGormChallengeRepository_Create(t *testing.T) {
	db := setupChallengeTestDB(t)
	repo := repo.NewGormChallengeRepository(db)

	tests := []struct {
		name        string
		challenge   *challengeModel.Challenge
		expectedErr error
	}{
		{
			name: "successful challenge creation",
			challenge: &challengeModel.Challenge{
				Base: model.Base{
					ID:        "test-challenge-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				OwnerID:         "owner123",
				Name:            "Test Challenge",
				Description:     "Test Description",
				Mode:            challengeModel.ChallengeModeF,
				Condition:       "Complete 5km run",
				Goal:            "Fitness",
				Location:        "Park",
				DistanceToCover: 5.0,
				TargetAmount:    100.0,
				StartDuration:   "2024-01-01",
				EndDuration:     "2024-01-31",
				NoOfWinner:      3,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.challenge)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.challenge.ID)

				// Verify the challenge was created in the database
				var count int64
				db.Model(&gormmodel.Challenge{}).Where("id = ?", tt.challenge.ID).Count(&count)
				assert.Equal(t, int64(1), count)
			}
		})
	}
}

func TestGormChallengeRepository_GetByID(t *testing.T) {
	db := setupChallengeTestDB(t)
	repo := repo.NewGormChallengeRepository(db)

	// Create a test challenge first
	testChallenge := &challengeModel.Challenge{
		Base: model.Base{
			ID:        "test-challenge-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		OwnerID:         "owner123",
		Name:            "Test Challenge",
		Description:     "Test Description",
		Mode:            challengeModel.ChallengeModeF,
		Condition:       "Complete 5km run",
		Goal:            "Fitness",
		Location:        "Park",
		DistanceToCover: 5.0,
		TargetAmount:    100.0,
		StartDuration:   "2024-01-01",
		EndDuration:     "2024-01-31",
		NoOfWinner:      3,
	}
	err := repo.Create(testChallenge)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		id          string
		expectedErr error
	}{
		{
			name:        "successful challenge retrieval",
			id:          "test-challenge-id",
			expectedErr: nil,
		},
		{
			name:        "challenge not found",
			id:          "nonexistent",
			expectedErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(tt.id)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.id, result.ID)
				assert.Equal(t, "Test Challenge", result.Name)
			}
		})
	}
}

func TestGormChallengeRepository_List(t *testing.T) {
	db := setupChallengeTestDB(t)
	repo := repo.NewGormChallengeRepository(db)

	// Create test challenges with manually set unique slugs to avoid generation conflicts
	challengeData := []struct {
		id   string
		name string
		slug string
	}{
		{"challenge-1", "Running Challenge", "running-challenge-test-1"},
		{"challenge-2", "Walking Marathon", "walking-marathon-test-2"},
		{"challenge-3", "Cycling Adventure", "cycling-adventure-test-3"},
	}

	for _, data := range challengeData {
		challenge := &challengeModel.Challenge{
			Base: model.Base{
				ID:        data.id,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			OwnerID:         "owner123",
			Name:            data.name,
			Description:     "Test Description",
			Mode:            challengeModel.ChallengeModeF,
			Condition:       "Complete 5km run",
			Goal:            "Fitness",
			Location:        "Park",
			DistanceToCover: 5.0,
			TargetAmount:    100.0,
			StartDuration:   "2024-01-01",
			EndDuration:     "2024-01-31",
			NoOfWinner:      3,
			Slug:            data.slug, // Set slug manually to avoid conflicts
		}
		err := repo.Create(challenge)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "list all challenges",
			limit:         10,
			offset:        0,
			expectedCount: 3,
		},
		{
			name:          "list with limit",
			limit:         2,
			offset:        0,
			expectedCount: 2,
		},
		{
			name:          "list with offset",
			limit:         10,
			offset:        2,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.List(tt.limit, tt.offset)

			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)
		})
	}
}

func TestGormCauseRepository_Create(t *testing.T) {
	db := setupChallengeTestDB(t)
	repo := repo.NewGormCauseRepository(db)

	tests := []struct {
		name        string
		cause       *challengeModel.Cause
		expectedErr error
	}{
		{
			name: "successful cause creation",
			cause: &challengeModel.Cause{
				Base: model.Base{
					ID:        "test-cause-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				ChallengeID:        "challenge123",
				OwnerID:            "owner123",
				Name:               "Test Cause",
				Problem:            "Environmental pollution",
				Solution:           "Plant trees",
				ProductDescription: "Tree planting initiative",
				Activity:           challengeModel.ActivityWalking,
				Location:           "Park",
				Description:        "Help save the environment",
				IsCommercial:       false,
				AmountPerPiece:     10.0,
				FundAmount:         100.0,
				WillingAmount:      50.0,
				UnitPrice:          5.0,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.cause)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.cause.ID)

				// Verify the cause was created in the database
				var count int64
				db.Model(&gormmodel.Cause{}).Where("id = ?", tt.cause.ID).Count(&count)
				assert.Equal(t, int64(1), count)
			}
		})
	}
}

func TestGormCauseRepository_GetByChallengeID(t *testing.T) {
	db := setupChallengeTestDB(t)
	repo := repo.NewGormCauseRepository(db)

	// Create test causes with manually set unique slugs to avoid generation conflicts
	causeData := []struct {
		id   string
		name string
		slug string
	}{
		{"cause-1", "Tree Planting Initiative", "tree-planting-initiative-test-1"},
		{"cause-2", "Clean Water Project", "clean-water-project-test-2"},
	}

	for _, data := range causeData {
		cause := &challengeModel.Cause{
			Base: model.Base{
				ID:        data.id,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			ChallengeID:        "challenge123",
			OwnerID:            "owner123",
			Name:               data.name,
			Problem:            "Environmental pollution",
			Solution:           "Plant trees",
			ProductDescription: "Tree planting initiative",
			Activity:           challengeModel.ActivityWalking,
			Location:           "Park",
			Description:        "Help save the environment",
			IsCommercial:       false,
			AmountPerPiece:     10.0,
			FundAmount:         100.0,
			WillingAmount:      50.0,
			UnitPrice:          5.0,
			Slug:               data.slug, // Set slug manually to avoid conflicts
		}
		err := repo.Create(cause)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		challengeID   string
		expectedCount int
	}{
		{
			name:          "get causes by challenge ID",
			challengeID:   "challenge123",
			expectedCount: 2,
		},
		{
			name:          "get causes by non-existent challenge ID",
			challengeID:   "nonexistent",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByChallengeID(tt.challengeID)

			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)
		})
	}
}

func TestGormCauseRunnerRepository_GetLeaderboard(t *testing.T) {
	db := setupChallengeTestDB(t)
	repo := repo.NewGormCauseRunnerRepository(db)

	// Create test cause runners with different distances
	runners := []*challengeModel.CauseRunner{
		{
			Base: model.Base{
				ID:        "runner1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			CauseID:         "cause1",
			OwnerID:         "user1",
			DistanceToCover: 10.0,
			DistanceCovered: 8.5,
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
			DistanceCovered: 12.0,
			Duration:        "60:00",
			MoneyRaised:     40.0,
			Activity:        "Running",
		},
		{
			Base: model.Base{
				ID:        "runner3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			CauseID:         "cause3",
			OwnerID:         "user3",
			DistanceToCover: 8.0,
			DistanceCovered: 6.0,
			Duration:        "30:00",
			MoneyRaised:     15.0,
			Activity:        "Cycling",
		},
	}

	for _, runner := range runners {
		err := repo.Create(runner)
		assert.NoError(t, err)
	}

	t.Run("get leaderboard ordered by distance", func(t *testing.T) {
		result, err := repo.GetLeaderboard()

		assert.NoError(t, err)
		assert.Len(t, result, 3)

		// Verify ordering by distance_covered DESC
		assert.Equal(t, "runner2", result[0].ID) // 12.0 km
		assert.Equal(t, "runner1", result[1].ID) // 8.5 km
		assert.Equal(t, "runner3", result[2].ID) // 6.0 km
	})
}

func TestGormSponsorCauseRepository_Create(t *testing.T) {
	db := setupChallengeTestDB(t)
	repo := repo.NewGormSponsorCauseRepository(db)

	tests := []struct {
		name        string
		sponsor     *challengeModel.SponsorCause
		expectedErr error
	}{
		{
			name: "successful sponsorship creation",
			sponsor: &challengeModel.SponsorCause{
				Base: model.Base{
					ID:        "test-sponsor-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				SponsorID:   "sponsor123",
				CauseID:     "cause123",
				Distance:    10.5,
				AmountPerKm: 5.0,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.sponsor)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.sponsor.ID)

				// Verify the sponsorship was created in the database
				var count int64
				db.Model(&gormmodel.SponsorCause{}).Where("id = ?", tt.sponsor.ID).Count(&count)
				assert.Equal(t, int64(1), count)
			}
		})
	}
}

func TestGormCauseBuyerRepository_Create(t *testing.T) {
	db := setupChallengeTestDB(t)
	repo := repo.NewGormCauseBuyerRepository(db)

	tests := []struct {
		name        string
		buyer       *challengeModel.CauseBuyer
		expectedErr error
	}{
		{
			name: "successful purchase creation",
			buyer: &challengeModel.CauseBuyer{
				Base: model.Base{
					ID:        "test-buyer-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				BuyerID: "buyer123",
				CauseID: "cause123",
				Amount:  50.0,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.buyer)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.buyer.ID)

				// Verify the purchase was created in the database
				var count int64
				db.Model(&gormmodel.CauseBuyer{}).Where("id = ?", tt.buyer.ID).Count(&count)
				assert.Equal(t, int64(1), count)
			}
		})
	}
}
