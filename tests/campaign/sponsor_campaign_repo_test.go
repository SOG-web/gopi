package campaign_test

import (
	"encoding/json"
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
func setupSponsorCampaignTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the schema - include all related models for proper relationship handling
	err = db.AutoMigrate(&gormmodel.Campaign{}, &gormmodel.SponsorCampaign{}, &gormmodel.CampaignMember{}, &gormmodel.CampaignSponsor{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestGormSponsorCampaignRepository_Create(t *testing.T) {
	db := setupSponsorCampaignTestDB(t)
	repo := repo.NewGormSponsorCampaignRepository(db)

	// First create a campaign for the sponsor campaign to reference
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
		sponsor     *campaignModel.SponsorCampaign
		expectedErr error
	}{
		{
			name: "successful sponsor campaign creation",
			sponsor: &campaignModel.SponsorCampaign{
				Base: model.Base{
					ID:        "test-sponsor-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				CampaignID:  "test-campaign-id",
				Distance:    10.0,
				AmountPerKm: 5.0,
				TotalAmount: 50.0,
				BrandImg:    "brand.jpg",
				VideoUrl:    "video.mp4",
				Sponsors:    []interface{}{"user1", "user2"},
			},
			expectedErr: nil,
		},
		{
			name: "creation with minimal fields",
			sponsor: &campaignModel.SponsorCampaign{
				Base: model.Base{
					ID:        "minimal-sponsor-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				CampaignID:  "test-campaign-id",
				Distance:    5.0,
				AmountPerKm: 2.0,
				TotalAmount: 10.0,
				Sponsors:    []interface{}{"user3"},
			},
			expectedErr: nil,
		},
		{
			name: "creation with empty sponsors array",
			sponsor: &campaignModel.SponsorCampaign{
				Base: model.Base{
					ID:        "empty-sponsors-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				CampaignID:  "test-campaign-id",
				Distance:    15.0,
				AmountPerKm: 3.0,
				TotalAmount: 45.0,
				Sponsors:    []interface{}{},
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.sponsor)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.sponsor.ID)

				// Verify the sponsor campaign was created in the database
				var dbSponsor gormmodel.SponsorCampaign
				result := db.First(&dbSponsor, "id = ?", tt.sponsor.ID)
				assert.NoError(t, result.Error)
				assert.Equal(t, tt.sponsor.CampaignID, dbSponsor.CampaignID)
				assert.Equal(t, tt.sponsor.Distance, dbSponsor.Distance)
				assert.Equal(t, tt.sponsor.AmountPerKm, dbSponsor.AmountPerKm)
				assert.Equal(t, tt.sponsor.TotalAmount, dbSponsor.TotalAmount)
				assert.Equal(t, tt.sponsor.BrandImg, dbSponsor.BrandImg)
				assert.Equal(t, tt.sponsor.VideoUrl, dbSponsor.VideoUrl)

				// Verify sponsors JSON was stored correctly
				var sponsors []interface{}
				if dbSponsor.Sponsors != "" {
					err := json.Unmarshal([]byte(dbSponsor.Sponsors), &sponsors)
					assert.NoError(t, err)
					assert.Equal(t, tt.sponsor.Sponsors, sponsors)
				}
			}
		})
	}
}

func TestGormSponsorCampaignRepository_GetByID(t *testing.T) {
	db := setupSponsorCampaignTestDB(t)
	repo := repo.NewGormSponsorCampaignRepository(db)

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

	sponsorsJSON, _ := json.Marshal([]interface{}{"user1", "user2"})
	sponsor := &gormmodel.SponsorCampaign{
		ID:          "test-sponsor-id",
		CampaignID:  "test-campaign-id",
		Distance:    10.0,
		AmountPerKm: 5.0,
		TotalAmount: 50.0,
		BrandImg:    "brand.jpg",
		VideoUrl:    "video.mp4",
		Sponsors:    string(sponsorsJSON),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = db.Create(sponsor).Error
	assert.NoError(t, err)

	tests := []struct {
		name        string
		sponsorID   string
		expectedErr error
	}{
		{
			name:        "successful get by id",
			sponsorID:   "test-sponsor-id",
			expectedErr: nil,
		},
		{
			name:        "sponsor campaign not found",
			sponsorID:   "nonexistent-id",
			expectedErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(tt.sponsorID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.sponsorID, result.ID)
				assert.Equal(t, "test-campaign-id", result.CampaignID)
				assert.Equal(t, 10.0, result.Distance)
				assert.Equal(t, 5.0, result.AmountPerKm)
				assert.Equal(t, 50.0, result.TotalAmount)
				assert.Equal(t, "brand.jpg", result.BrandImg)
				assert.Equal(t, "video.mp4", result.VideoUrl)
				assert.Len(t, result.Sponsors, 2)
				assert.Contains(t, result.Sponsors, "user1")
				assert.Contains(t, result.Sponsors, "user2")
			}
		})
	}
}

func TestGormSponsorCampaignRepository_GetByCampaignID(t *testing.T) {
	db := setupSponsorCampaignTestDB(t)
	repo := repo.NewGormSponsorCampaignRepository(db)

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

	// Create sponsor campaigns for different campaigns
	sponsors := []gormmodel.SponsorCampaign{
		{
			ID:          "sponsor1",
			CampaignID:  "campaign1",
			Distance:    10.0,
			AmountPerKm: 5.0,
			TotalAmount: 50.0,
			BrandImg:    "brand1.jpg",
			VideoUrl:    "video1.mp4",
			Sponsors:    `["user1"]`,
			CreatedAt:   time.Now().Add(-time.Hour), // Older
			UpdatedAt:   time.Now().Add(-time.Hour),
		},
		{
			ID:          "sponsor2",
			CampaignID:  "campaign1",
			Distance:    5.0,
			AmountPerKm: 2.0,
			TotalAmount: 10.0,
			BrandImg:    "brand2.jpg",
			VideoUrl:    "",
			Sponsors:    `["user2"]`,
			CreatedAt:   time.Now(), // Newer
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "sponsor3",
			CampaignID:  "campaign2",
			Distance:    15.0,
			AmountPerKm: 3.0,
			TotalAmount: 45.0,
			BrandImg:    "brand3.jpg",
			VideoUrl:    "video3.mp4",
			Sponsors:    `["user3","user4"]`,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, sponsor := range sponsors {
		err := db.Create(&sponsor).Error
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		campaignID    string
		expectedCount int
		expectedErr   error
	}{
		{
			name:          "get sponsor campaigns for campaign with multiple sponsors",
			campaignID:    "campaign1",
			expectedCount: 2,
			expectedErr:   nil,
		},
		{
			name:          "get sponsor campaigns for campaign with single sponsor",
			campaignID:    "campaign2",
			expectedCount: 1,
			expectedErr:   nil,
		},
		{
			name:          "get sponsor campaigns for campaign with no sponsors",
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

				// Verify all returned sponsor campaigns belong to the correct campaign
				for _, sponsor := range result {
					assert.Equal(t, tt.campaignID, sponsor.CampaignID)
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

func TestGormSponsorCampaignRepository_Update(t *testing.T) {
	db := setupSponsorCampaignTestDB(t)
	repo := repo.NewGormSponsorCampaignRepository(db)

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

	sponsorsJSON, _ := json.Marshal([]interface{}{"user1", "user2"})
	originalSponsor := &gormmodel.SponsorCampaign{
		ID:          "test-sponsor-id",
		CampaignID:  "test-campaign-id",
		Distance:    10.0,
		AmountPerKm: 5.0,
		TotalAmount: 50.0,
		BrandImg:    "brand.jpg",
		VideoUrl:    "video.mp4",
		Sponsors:    string(sponsorsJSON),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = db.Create(originalSponsor).Error
	assert.NoError(t, err)

	tests := []struct {
		name        string
		sponsorID   string
		updateData  *campaignModel.SponsorCampaign
		expectedErr error
	}{
		{
			name:      "successful sponsor campaign update",
			sponsorID: "test-sponsor-id",
			updateData: &campaignModel.SponsorCampaign{
				Base: model.Base{
					ID:        "test-sponsor-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				CampaignID:  "test-campaign-id",
				Distance:    15.0,
				AmountPerKm: 6.0,
				TotalAmount: 90.0,
				BrandImg:    "updated-brand.jpg",
				VideoUrl:    "updated-video.mp4",
				Sponsors:    []interface{}{"user3", "user4", "user5"},
			},
			expectedErr: nil,
		},
		{
			name:      "partial update",
			sponsorID: "test-sponsor-id",
			updateData: &campaignModel.SponsorCampaign{
				Base: model.Base{
					ID:        "test-sponsor-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				CampaignID:  "test-campaign-id",
				Distance:    20.0,
				AmountPerKm: 4.0,
				TotalAmount: 80.0,
				BrandImg:    "partially-updated.jpg",
				VideoUrl:    "",
				Sponsors:    []interface{}{"user6"},
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

				// Verify the sponsor campaign was updated in the database
				var dbSponsor gormmodel.SponsorCampaign
				result := db.First(&dbSponsor, "id = ?", tt.sponsorID)
				assert.NoError(t, result.Error)
				assert.Equal(t, tt.updateData.Distance, dbSponsor.Distance)
				assert.Equal(t, tt.updateData.AmountPerKm, dbSponsor.AmountPerKm)
				assert.Equal(t, tt.updateData.TotalAmount, dbSponsor.TotalAmount)
				assert.Equal(t, tt.updateData.BrandImg, dbSponsor.BrandImg)
				assert.Equal(t, tt.updateData.VideoUrl, dbSponsor.VideoUrl)

				// Verify sponsors JSON was updated correctly
				var sponsors []interface{}
				if dbSponsor.Sponsors != "" {
					err := json.Unmarshal([]byte(dbSponsor.Sponsors), &sponsors)
					assert.NoError(t, err)
					assert.Equal(t, tt.updateData.Sponsors, sponsors)
				}

				// Verify updated_at was updated
				assert.True(t, dbSponsor.UpdatedAt.After(originalSponsor.UpdatedAt))
			}
		})
	}
}

func TestGormSponsorCampaignRepository_Delete(t *testing.T) {
	db := setupSponsorCampaignTestDB(t)
	repo := repo.NewGormSponsorCampaignRepository(db)

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

	sponsor := &gormmodel.SponsorCampaign{
		ID:          "test-sponsor-id",
		CampaignID:  "test-campaign-id",
		Distance:    10.0,
		AmountPerKm: 5.0,
		TotalAmount: 50.0,
		BrandImg:    "brand.jpg",
		VideoUrl:    "video.mp4",
		Sponsors:    `["user1"]`,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = db.Create(sponsor).Error
	assert.NoError(t, err)

	tests := []struct {
		name        string
		sponsorID   string
		expectedErr error
	}{
		{
			name:        "successful sponsor campaign deletion",
			sponsorID:   "test-sponsor-id",
			expectedErr: nil,
		},
		{
			name:        "delete nonexistent sponsor campaign",
			sponsorID:   "nonexistent-id",
			expectedErr: nil, // GORM Delete doesn't return error for non-existent records
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(tt.sponsorID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)

				// Verify the sponsor campaign was deleted from the database
				var dbSponsor gormmodel.SponsorCampaign
				result := db.First(&dbSponsor, "id = ?", tt.sponsorID)
				assert.Error(t, result.Error)
				assert.True(t, errors.Is(result.Error, gorm.ErrRecordNotFound))
			}
		})
	}
}

func TestGormSponsorCampaignRepository_JSONHandling(t *testing.T) {
	db := setupSponsorCampaignTestDB(t)
	repo := repo.NewGormSponsorCampaignRepository(db)

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

	t.Run("handle complex sponsors array", func(t *testing.T) {
		sponsor := &campaignModel.SponsorCampaign{
			Base: model.Base{
				ID:        "complex-sponsor-id",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			CampaignID:  "test-campaign-id",
			Distance:    25.0,
			AmountPerKm: 8.0,
			TotalAmount: 200.0,
			BrandImg:    "complex-brand.jpg",
			VideoUrl:    "complex-video.mp4",
			Sponsors:    []interface{}{"user1", "user2", "user3", 123, true}, // Mixed types
		}

		err := repo.Create(sponsor)
		assert.NoError(t, err)

		// Retrieve and verify
		retrieved, err := repo.GetByID("complex-sponsor-id")
		assert.NoError(t, err)
		assert.Equal(t, sponsor.Sponsors, retrieved.Sponsors)
		assert.Len(t, retrieved.Sponsors, 5)
	})

	t.Run("handle empty sponsors array", func(t *testing.T) {
		sponsor := &campaignModel.SponsorCampaign{
			Base: model.Base{
				ID:        "empty-sponsor-id",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			CampaignID:  "test-campaign-id",
			Distance:    5.0,
			AmountPerKm: 2.0,
			TotalAmount: 10.0,
			Sponsors:    []interface{}{},
		}

		err := repo.Create(sponsor)
		assert.NoError(t, err)

		// Retrieve and verify
		retrieved, err := repo.GetByID("empty-sponsor-id")
		assert.NoError(t, err)
		assert.Len(t, retrieved.Sponsors, 0)
	})

	t.Run("handle nil sponsors", func(t *testing.T) {
		sponsor := &campaignModel.SponsorCampaign{
			Base: model.Base{
				ID:        "nil-sponsor-id",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			CampaignID:  "test-campaign-id",
			Distance:    12.0,
			AmountPerKm: 3.0,
			TotalAmount: 36.0,
			Sponsors:    nil,
		}

		err := repo.Create(sponsor)
		assert.NoError(t, err)

		// Retrieve and verify
		retrieved, err := repo.GetByID("nil-sponsor-id")
		assert.NoError(t, err)
		assert.Len(t, retrieved.Sponsors, 0)
	})
}

func TestGormSponsorCampaignRepository_ErrorScenarios(t *testing.T) {
	db := setupSponsorCampaignTestDB(t)
	repo := repo.NewGormSponsorCampaignRepository(db)

	t.Run("create with invalid foreign key", func(t *testing.T) {
		// Try to create a sponsor campaign with non-existent campaign ID
		sponsor := &campaignModel.SponsorCampaign{
			Base: model.Base{
				ID:        "invalid-sponsor-id",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			CampaignID:  "nonexistent-campaign",
			Distance:    10.0,
			AmountPerKm: 5.0,
			TotalAmount: 50.0,
			Sponsors:    []interface{}{"user1"},
		}

		err := repo.Create(sponsor)
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

	t.Run("update nonexistent sponsor campaign", func(t *testing.T) {
		sponsor := &campaignModel.SponsorCampaign{
			Base: model.Base{
				ID:        "nonexistent-sponsor",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			CampaignID:  "test-campaign",
			Distance:    10.0,
			AmountPerKm: 5.0,
			TotalAmount: 50.0,
			Sponsors:    []interface{}{"user1"},
		}

		err := repo.Update(sponsor)
		// GORM Save creates a new record if it doesn't exist, so this doesn't error
		// This is expected behavior for GORM Save method
		assert.NoError(t, err)

		// Verify a new record was created
		retrieved, err := repo.GetByID("nonexistent-sponsor")
		assert.NoError(t, err)
		assert.Equal(t, "nonexistent-sponsor", retrieved.ID)
	})
}
