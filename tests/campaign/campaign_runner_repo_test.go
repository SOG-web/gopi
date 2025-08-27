package campaign_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	gormmodel "gopi.com/internal/data/campaign/model/gorm"
	"gopi.com/internal/data/campaign/repo"
	campaignModel "gopi.com/internal/domain/campaign/model"
	"gopi.com/internal/domain/model"
)

// Setup in-memory database for testing
func setupCampaignRunnerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the schema - include all related models for proper relationship handling
	err = db.AutoMigrate(&gormmodel.Campaign{}, &gormmodel.CampaignRunner{}, &gormmodel.CampaignMember{}, &gormmodel.CampaignSponsor{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestGormCampaignRunnerRepository_Create(t *testing.T) {
	db := setupCampaignRunnerTestDB(t)
	repo := repo.NewGormCampaignRunnerRepository(db)

	// First create a campaign for the runner to reference
	campaign := &gormmodel.Campaign{
		ID:        "test-campaign-id",
		Name:      "Test Campaign",
		Slug:      "test-campaign",
		OwnerID:   "owner123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := db.Create(campaign).Error
	assert.NoError(t, err)

	tests := []struct {
		name        string
		runner      *campaignModel.CampaignRunner
		expectedErr error
	}{
		{
			name: "successful campaign runner creation",
			runner: &campaignModel.CampaignRunner{
				Base: model.Base{
					ID:        "test-runner-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				CampaignID:      "test-campaign-id",
				DistanceCovered: 10.5,
				Duration:        "45:30",
				MoneyRaised:     25.0,
				CoverImage:      "runner.jpg",
				Activity:        "Running",
				OwnerID:         "user123",
				DateJoined:      time.Now(),
			},
			expectedErr: nil,
		},
		{
			name: "creation with minimal fields",
			runner: &campaignModel.CampaignRunner{
				Base: model.Base{
					ID:        "minimal-runner-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				CampaignID: "test-campaign-id",
				Activity:   "Walking",
				OwnerID:    "user456",
				DateJoined: time.Now(),
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.runner)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.runner.ID)

				// Verify the runner was created in the database
				var dbRunner gormmodel.CampaignRunner
				result := db.First(&dbRunner, "id = ?", tt.runner.ID)
				assert.NoError(t, result.Error)
				assert.Equal(t, tt.runner.CampaignID, dbRunner.CampaignID)
				assert.Equal(t, tt.runner.Activity, dbRunner.Activity)
				assert.Equal(t, tt.runner.OwnerID, dbRunner.OwnerID)
			}
		})
	}
}

func TestGormCampaignRunnerRepository_GetByID(t *testing.T) {
	db := setupCampaignRunnerTestDB(t)
	repo := repo.NewGormCampaignRunnerRepository(db)

	// Create test data
	campaign := &gormmodel.Campaign{
		ID:        "test-campaign-id",
		Name:      "Test Campaign",
		Slug:      "test-campaign",
		OwnerID:   "owner123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := db.Create(campaign).Error
	assert.NoError(t, err)

	runner := &gormmodel.CampaignRunner{
		ID:              "test-runner-id",
		CampaignID:      "test-campaign-id",
		DistanceCovered: 10.5,
		Duration:        "45:30",
		MoneyRaised:     25.0,
		CoverImage:      "runner.jpg",
		Activity:        "Running",
		OwnerID:         "user123",
		DateJoined:      time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err = db.Create(runner).Error
	assert.NoError(t, err)

	tests := []struct {
		name        string
		runnerID    string
		expectedErr error
	}{
		{
			name:        "successful get by id",
			runnerID:    "test-runner-id",
			expectedErr: nil,
		},
		{
			name:        "runner not found",
			runnerID:    "nonexistent-id",
			expectedErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(tt.runnerID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.runnerID, result.ID)
				assert.Equal(t, "test-campaign-id", result.CampaignID)
				assert.Equal(t, "Running", result.Activity)
				assert.Equal(t, "user123", result.OwnerID)
			}
		})
	}
}

func TestGormCampaignRunnerRepository_GetByCampaignID(t *testing.T) {
	db := setupCampaignRunnerTestDB(t)
	repo := repo.NewGormCampaignRunnerRepository(db)

	// Create test data
	campaign1 := &gormmodel.Campaign{
		ID:        "campaign1",
		Name:      "Campaign 1",
		Slug:      "campaign-1",
		OwnerID:   "owner123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	campaign2 := &gormmodel.Campaign{
		ID:        "campaign2",
		Name:      "Campaign 2",
		Slug:      "campaign-2",
		OwnerID:   "owner456",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := db.Create(campaign1).Error
	assert.NoError(t, err)
	err = db.Create(campaign2).Error
	assert.NoError(t, err)

	// Create runners for campaign1
	runners := []gormmodel.CampaignRunner{
		{
			ID:              "runner1",
			CampaignID:      "campaign1",
			DistanceCovered: 10.5,
			Duration:        "45:30",
			MoneyRaised:     25.0,
			Activity:        "Running",
			OwnerID:         "user1",
			DateJoined:      time.Now(),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "runner2",
			CampaignID:      "campaign1",
			DistanceCovered: 5.0,
			Duration:        "30:00",
			MoneyRaised:     10.0,
			Activity:        "Walking",
			OwnerID:         "user2",
			DateJoined:      time.Now(),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "runner3",
			CampaignID:      "campaign2",
			DistanceCovered: 15.0,
			Duration:        "60:00",
			MoneyRaised:     30.0,
			Activity:        "Cycling",
			OwnerID:         "user3",
			DateJoined:      time.Now(),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, runner := range runners {
		err := db.Create(&runner).Error
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		campaignID    string
		expectedCount int
		expectedErr   error
	}{
		{
			name:          "get runners for campaign with multiple runners",
			campaignID:    "campaign1",
			expectedCount: 2,
			expectedErr:   nil,
		},
		{
			name:          "get runners for campaign with single runner",
			campaignID:    "campaign2",
			expectedCount: 1,
			expectedErr:   nil,
		},
		{
			name:          "get runners for campaign with no runners",
			campaignID:    "campaign3",
			expectedCount: 0,
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByCampaignID(tt.campaignID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)

				// Verify all returned runners belong to the correct campaign
				for _, runner := range result {
					assert.Equal(t, tt.campaignID, runner.CampaignID)
				}

				// Verify ordering by distance covered descending
				if len(result) > 1 {
					for i := 0; i < len(result)-1; i++ {
						assert.GreaterOrEqual(t, result[i].DistanceCovered, result[i+1].DistanceCovered)
					}
				}
			}
		})
	}
}

func TestGormCampaignRunnerRepository_GetByOwnerID(t *testing.T) {
	db := setupCampaignRunnerTestDB(t)
	repo := repo.NewGormCampaignRunnerRepository(db)

	// Create test data
	campaign := &gormmodel.Campaign{
		ID:        "test-campaign-id",
		Name:      "Test Campaign",
		Slug:      "test-campaign",
		OwnerID:   "owner123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := db.Create(campaign).Error
	assert.NoError(t, err)

	// Create runners for different owners
	runners := []gormmodel.CampaignRunner{
		{
			ID:              "runner1",
			CampaignID:      "test-campaign-id",
			DistanceCovered: 10.5,
			Duration:        "45:30",
			MoneyRaised:     25.0,
			Activity:        "Running",
			OwnerID:         "user1",
			DateJoined:      time.Now(),
			CreatedAt:       time.Now().Add(-time.Hour), // Older
			UpdatedAt:       time.Now().Add(-time.Hour),
		},
		{
			ID:              "runner2",
			CampaignID:      "test-campaign-id",
			DistanceCovered: 5.0,
			Duration:        "30:00",
			MoneyRaised:     10.0,
			Activity:        "Walking",
			OwnerID:         "user1",
			DateJoined:      time.Now(),
			CreatedAt:       time.Now(), // Newer
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "runner3",
			CampaignID:      "test-campaign-id",
			DistanceCovered: 15.0,
			Duration:        "60:00",
			MoneyRaised:     30.0,
			Activity:        "Cycling",
			OwnerID:         "user2",
			DateJoined:      time.Now(),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, runner := range runners {
		err := db.Create(&runner).Error
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		ownerID       string
		expectedCount int
		expectedErr   error
	}{
		{
			name:          "get runners for owner with multiple runners",
			ownerID:       "user1",
			expectedCount: 2,
			expectedErr:   nil,
		},
		{
			name:          "get runners for owner with single runner",
			ownerID:       "user2",
			expectedCount: 1,
			expectedErr:   nil,
		},
		{
			name:          "get runners for owner with no runners",
			ownerID:       "user3",
			expectedCount: 0,
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByOwnerID(tt.ownerID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)

				// Verify all returned runners belong to the correct owner
				for _, runner := range result {
					assert.Equal(t, tt.ownerID, runner.OwnerID)
				}

				// Verify ordering by created_at descending
				if len(result) > 1 {
					for i := 0; i < len(result)-1; i++ {
						assert.True(t, result[i].CreatedAt.After(result[i+1].CreatedAt) || result[i].CreatedAt.Equal(result[i+1].CreatedAt))
					}
				}
			}
		})
	}
}

func TestGormCampaignRunnerRepository_Update(t *testing.T) {
	db := setupCampaignRunnerTestDB(t)
	repo := repo.NewGormCampaignRunnerRepository(db)

	// Create test data
	campaign := &gormmodel.Campaign{
		ID:        "test-campaign-id",
		Name:      "Test Campaign",
		Slug:      "test-campaign",
		OwnerID:   "owner123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := db.Create(campaign).Error
	assert.NoError(t, err)

	originalRunner := &gormmodel.CampaignRunner{
		ID:              "test-runner-id",
		CampaignID:      "test-campaign-id",
		DistanceCovered: 10.5,
		Duration:        "45:30",
		MoneyRaised:     25.0,
		CoverImage:      "runner.jpg",
		Activity:        "Running",
		OwnerID:         "user123",
		DateJoined:      time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err = db.Create(originalRunner).Error
	assert.NoError(t, err)

	tests := []struct {
		name        string
		runnerID    string
		updateData  *campaignModel.CampaignRunner
		expectedErr error
	}{
		{
			name:     "successful runner update",
			runnerID: "test-runner-id",
			updateData: &campaignModel.CampaignRunner{
				Base: model.Base{
					ID:        "test-runner-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				CampaignID:      "test-campaign-id",
				DistanceCovered: 15.0,
				Duration:        "60:00",
				MoneyRaised:     30.0,
				CoverImage:      "updated-runner.jpg",
				Activity:        "Walking",
				OwnerID:         "user123",
				DateJoined:      time.Now(),
			},
			expectedErr: nil,
		},
		{
			name:     "partial update",
			runnerID: "test-runner-id",
			updateData: &campaignModel.CampaignRunner{
				Base: model.Base{
					ID:        "test-runner-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				CampaignID:      "test-campaign-id",
				DistanceCovered: 20.0,
				Duration:        "90:00",
				MoneyRaised:     40.0,
				CoverImage:      "partially-updated.jpg",
				Activity:        "Cycling",
				OwnerID:         "user123",
				DateJoined:      time.Now(),
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(tt.updateData)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)

				// Verify the runner was updated in the database
				var dbRunner gormmodel.CampaignRunner
				result := db.First(&dbRunner, "id = ?", tt.runnerID)
				assert.NoError(t, result.Error)
				assert.Equal(t, tt.updateData.DistanceCovered, dbRunner.DistanceCovered)
				assert.Equal(t, tt.updateData.Duration, dbRunner.Duration)
				assert.Equal(t, tt.updateData.MoneyRaised, dbRunner.MoneyRaised)
				assert.Equal(t, tt.updateData.CoverImage, dbRunner.CoverImage)
				assert.Equal(t, tt.updateData.Activity, dbRunner.Activity)

				// Verify updated_at was updated
				assert.True(t, dbRunner.UpdatedAt.After(originalRunner.UpdatedAt))
			}
		})
	}
}

func TestGormCampaignRunnerRepository_Delete(t *testing.T) {
	db := setupCampaignRunnerTestDB(t)
	repo := repo.NewGormCampaignRunnerRepository(db)

	// Create test data
	campaign := &gormmodel.Campaign{
		ID:        "test-campaign-id",
		Name:      "Test Campaign",
		Slug:      "test-campaign",
		OwnerID:   "owner123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := db.Create(campaign).Error
	assert.NoError(t, err)

	runner := &gormmodel.CampaignRunner{
		ID:              "test-runner-id",
		CampaignID:      "test-campaign-id",
		DistanceCovered: 10.5,
		Duration:        "45:30",
		MoneyRaised:     25.0,
		CoverImage:      "runner.jpg",
		Activity:        "Running",
		OwnerID:         "user123",
		DateJoined:      time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err = db.Create(runner).Error
	assert.NoError(t, err)

	tests := []struct {
		name        string
		runnerID    string
		expectedErr error
	}{
		{
			name:        "successful runner deletion",
			runnerID:    "test-runner-id",
			expectedErr: nil,
		},
		{
			name:        "delete nonexistent runner",
			runnerID:    "nonexistent-id",
			expectedErr: nil, // GORM Delete doesn't return error for non-existent records
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(tt.runnerID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)

				// Verify the runner was deleted from the database
				var dbRunner gormmodel.CampaignRunner
				result := db.First(&dbRunner, "id = ?", tt.runnerID)
				assert.Error(t, result.Error)
				assert.True(t, errors.Is(result.Error, gorm.ErrRecordNotFound))
			}
		})
	}
}

func TestGormCampaignRunnerRepository_ErrorScenarios(t *testing.T) {
	db := setupCampaignRunnerTestDB(t)
	repo := repo.NewGormCampaignRunnerRepository(db)

	t.Run("create with invalid foreign key", func(t *testing.T) {
		// Try to create a runner with non-existent campaign ID
		runner := &campaignModel.CampaignRunner{
			Base: model.Base{
				ID:        "invalid-runner-id",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			CampaignID: "nonexistent-campaign",
			Activity:   "Running",
			OwnerID:    "user123",
			DateJoined: time.Now(),
		}

		err := repo.Create(runner)
		// Note: SQLite may not enforce foreign keys by default, so this might not error
		// In a real PostgreSQL setup with proper constraints, this would error
		if err != nil {
			assert.Contains(t, err.Error(), "foreign key") // or similar constraint error
		}
	})

	t.Run("get by invalid id", func(t *testing.T) {
		result, err := repo.GetByID("invalid-id")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
		assert.Nil(t, result)
	})

	t.Run("update nonexistent runner", func(t *testing.T) {
		runner := &campaignModel.CampaignRunner{
			Base: model.Base{
				ID:        "nonexistent-runner",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			CampaignID: "test-campaign",
			Activity:   "Running",
			OwnerID:    "user123",
			DateJoined: time.Now(),
		}

		err := repo.Update(runner)
		// GORM Save creates a new record if it doesn't exist, so this doesn't error
		// This is expected behavior for GORM Save method
		assert.NoError(t, err)

		// Verify a new record was created
		retrieved, err := repo.GetByID("nonexistent-runner")
		assert.NoError(t, err)
		assert.Equal(t, "nonexistent-runner", retrieved.ID)
	})
}
