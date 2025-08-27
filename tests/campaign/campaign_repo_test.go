package campaign_test

import (
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
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the schema - include all related models for proper relationship handling
	err = db.AutoMigrate(&gormmodel.Campaign{}, &gormmodel.CampaignRunner{}, &gormmodel.SponsorCampaign{}, &gormmodel.CampaignMember{}, &gormmodel.CampaignSponsor{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestGormCampaignRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCampaignRepository(db)

	tests := []struct {
		name        string
		campaign    *campaignModel.Campaign
		expectedErr error
	}{
		{
			name: "successful campaign creation",
			campaign: &campaignModel.Campaign{
				Base: model.Base{
					ID:        "test-campaign-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:              "Test Campaign",
				Description:       "A test campaign",
				Condition:         "Good condition",
				Mode:              campaignModel.CampaignModeFree,
				Goal:              "Test goal",
				Activity:          campaignModel.ActivityWalking,
				Location:          "Test location",
				TargetAmount:      1000.0,
				TargetAmountPerKm: 5.0,
				DistanceToCover:   10.0,
				StartDuration:     "2024-01-01",
				EndDuration:       "2024-01-31",
				OwnerID:           "owner123",
				AcceptTac:         true,
				MoneyRaised:       0,
				DistanceCovered:   0,
				Members:           []interface{}{},
				Sponsors:          []interface{}{},
				Slug:              "test-campaign-slug",
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.campaign)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.campaign.ID)
			}

			// Verify the campaign was created in the database
			var dbCampaign gormmodel.Campaign
			result := db.First(&dbCampaign, "id = ?", tt.campaign.ID)
			if tt.expectedErr == nil {
				assert.NoError(t, result.Error)
				assert.Equal(t, tt.campaign.Name, dbCampaign.Name)
				assert.Equal(t, tt.campaign.OwnerID, dbCampaign.OwnerID)
				assert.Equal(t, tt.campaign.Slug, dbCampaign.Slug)
			}
		})
	}
}

func TestGormCampaignRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCampaignRepository(db)

	// Create a test campaign first
	testCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "test-campaign-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:    "Test Campaign",
		OwnerID: "owner123",
		Slug:    "test-campaign-slug",
	}
	err := repo.Create(testCampaign)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		campaignID  string
		expectedErr error
	}{
		{
			name:        "successful retrieval",
			campaignID:  "test-campaign-id",
			expectedErr: nil,
		},
		{
			name:        "campaign not found",
			campaignID:  "nonexistent-id",
			expectedErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaign, err := repo.GetByID(tt.campaignID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, campaign)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, campaign)
				assert.Equal(t, tt.campaignID, campaign.ID)
				assert.Equal(t, "Test Campaign", campaign.Name)
			}
		})
	}
}

func TestGormCampaignRepository_GetBySlug(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCampaignRepository(db)

	// Create test campaigns
	campaigns := []*campaignModel.Campaign{
		{
			Base: model.Base{
				ID:        "campaign1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "First Campaign",
			Slug:    "first-campaign-slug",
			OwnerID: "owner123",
		},
		{
			Base: model.Base{
				ID:        "campaign2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Second Campaign",
			Slug:    "second-campaign-slug",
			OwnerID: "owner456",
		},
	}

	for _, campaign := range campaigns {
		err := repo.Create(campaign)
		assert.NoError(t, err)
	}

	tests := []struct {
		name        string
		slug        string
		expectedErr error
		expectedID  string
	}{
		{
			name:        "successful retrieval by slug",
			slug:        "first-campaign-slug",
			expectedErr: nil,
			expectedID:  "campaign1",
		},
		{
			name:        "successful retrieval by second slug",
			slug:        "second-campaign-slug",
			expectedErr: nil,
			expectedID:  "campaign2",
		},
		{
			name:        "campaign not found",
			slug:        "nonexistent-slug",
			expectedErr: gorm.ErrRecordNotFound,
			expectedID:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaign, err := repo.GetBySlug(tt.slug)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, campaign)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, campaign)
				assert.Equal(t, tt.expectedID, campaign.ID)
				assert.Equal(t, tt.slug, campaign.Slug)
			}
		})
	}
}

func TestGormCampaignRepository_GetByOwnerID(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCampaignRepository(db)

	// Create test campaigns with different owners
	campaigns := []*campaignModel.Campaign{
		{
			Base: model.Base{
				ID:        "campaign1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Campaign 1",
			OwnerID: "owner123",
			Slug:    "campaign-1-slug",
		},
		{
			Base: model.Base{
				ID:        "campaign2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Campaign 2",
			OwnerID: "owner123",
			Slug:    "campaign-2-slug",
		},
		{
			Base: model.Base{
				ID:        "campaign3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Campaign 3",
			OwnerID: "owner456",
			Slug:    "campaign-3-slug",
		},
	}

	for _, campaign := range campaigns {
		err := repo.Create(campaign)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		ownerID       string
		expectedCount int
		expectedErr   error
	}{
		{
			name:          "owner has multiple campaigns",
			ownerID:       "owner123",
			expectedCount: 2,
			expectedErr:   nil,
		},
		{
			name:          "owner has one campaign",
			ownerID:       "owner456",
			expectedCount: 1,
			expectedErr:   nil,
		},
		{
			name:          "owner has no campaigns",
			ownerID:       "owner789",
			expectedCount: 0,
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaigns, err := repo.GetByOwnerID(tt.ownerID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, campaigns, tt.expectedCount)

				// Verify that all returned campaigns belong to the correct owner
				for _, campaign := range campaigns {
					assert.Equal(t, tt.ownerID, campaign.OwnerID)
				}
			}
		})
	}
}

func TestGormCampaignRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCampaignRepository(db)

	// Create a test campaign first
	testCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "test-campaign-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "Original Name",
		Description: "Original description",
		OwnerID:     "owner123",
		Slug:        "original-slug",
		MoneyRaised: 100.0,
	}
	err := repo.Create(testCampaign)
	assert.NoError(t, err)

	// Update the campaign
	testCampaign.Name = "Updated Name"
	testCampaign.Description = "Updated description"
	testCampaign.MoneyRaised = 200.0
	testCampaign.UpdatedAt = time.Now()

	err = repo.Update(testCampaign)
	assert.NoError(t, err)

	// Verify the update
	updatedCampaign, err := repo.GetByID("test-campaign-id")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedCampaign.Name)
	assert.Equal(t, "Updated description", updatedCampaign.Description)
	assert.Equal(t, 200.0, updatedCampaign.MoneyRaised)
}

func TestGormCampaignRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCampaignRepository(db)

	// Create a test campaign first
	testCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "test-campaign-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:    "Test Campaign",
		OwnerID: "owner123",
		Slug:    "test-campaign-slug",
	}
	err := repo.Create(testCampaign)
	assert.NoError(t, err)

	// Delete the campaign
	err = repo.Delete("test-campaign-id")
	assert.NoError(t, err)

	// Verify the campaign was deleted
	_, err = repo.GetByID("test-campaign-id")
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestGormCampaignRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCampaignRepository(db)

	// Create test campaigns
	campaigns := []*campaignModel.Campaign{
		{
			Base: model.Base{
				ID:        "campaign1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Campaign 1",
			OwnerID: "owner1",
			Slug:    "campaign-1-slug",
		},
		{
			Base: model.Base{
				ID:        "campaign2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Campaign 2",
			OwnerID: "owner2",
			Slug:    "campaign-2-slug",
		},
		{
			Base: model.Base{
				ID:        "campaign3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:    "Campaign 3",
			OwnerID: "owner3",
			Slug:    "campaign-3-slug",
		},
	}

	for _, campaign := range campaigns {
		err := repo.Create(campaign)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "list all campaigns",
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
			offset:        1,
			expectedCount: 2,
		},
		{
			name:          "list with limit and offset",
			limit:         1,
			offset:        1,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaigns, err := repo.List(tt.limit, tt.offset)
			assert.NoError(t, err)
			assert.Len(t, campaigns, tt.expectedCount)
		})
	}
}

func TestGormCampaignRepository_Search(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCampaignRepository(db)

	// Create test campaigns
	campaigns := []*campaignModel.Campaign{
		{
			Base: model.Base{
				ID:        "campaign1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:        "Golang Tutorial",
			Description: "Learn Golang programming",
			OwnerID:     "owner1",
			Slug:        "golang-tutorial-slug",
		},
		{
			Base: model.Base{
				ID:        "campaign2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:        "Python Guide",
			Description: "Python programming basics",
			OwnerID:     "owner2",
			Slug:        "python-guide-slug",
		},
		{
			Base: model.Base{
				ID:        "campaign3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:        "JavaScript Tips",
			Description: "JS best practices",
			OwnerID:     "owner3",
			Slug:        "javascript-tips-slug",
		},
	}

	for _, campaign := range campaigns {
		err := repo.Create(campaign)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		query         string
		limit         int
		offset        int
		expectedCount int
		expectedIDs   []string
	}{
		{
			name:          "search for 'golang'",
			query:         "golang",
			limit:         10,
			offset:        0,
			expectedCount: 1,
			expectedIDs:   []string{"campaign1"},
		},
		{
			name:          "search for 'programming'",
			query:         "programming",
			limit:         10,
			offset:        0,
			expectedCount: 2,
			expectedIDs:   []string{"campaign1", "campaign2"},
		},
		{
			name:          "search for 'javascript'",
			query:         "javascript",
			limit:         10,
			offset:        0,
			expectedCount: 1,
			expectedIDs:   []string{"campaign3"},
		},
		{
			name:          "search for non-existent term",
			query:         "nonexistent",
			limit:         10,
			offset:        0,
			expectedCount: 0,
			expectedIDs:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaigns, err := repo.Search(tt.query, tt.limit, tt.offset)
			assert.NoError(t, err)
			assert.Len(t, campaigns, tt.expectedCount)

			// Verify the returned campaigns match expected IDs
			if len(tt.expectedIDs) > 0 {
				actualIDs := make([]string, len(campaigns))
				for i, campaign := range campaigns {
					actualIDs[i] = campaign.ID
				}
				assert.ElementsMatch(t, tt.expectedIDs, actualIDs)
			}
		})
	}
}

func TestGormCampaignRepository_AddMember(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCampaignRepository(db)

	// Create a test campaign first
	testCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "test-campaign-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:    "Test Campaign",
		OwnerID: "owner123",
		Slug:    "test-campaign-slug",
		Members: []interface{}{},
	}
	err := repo.Create(testCampaign)
	assert.NoError(t, err)

	// Add a member
	err = repo.AddMember("test-campaign-id", "user123")
	assert.NoError(t, err)

	// Verify the member was added
	updatedCampaign, err := repo.GetByID("test-campaign-id")
	assert.NoError(t, err)
	assert.Contains(t, updatedCampaign.Members, "user123")
}

func TestGormCampaignRepository_RemoveMember(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCampaignRepository(db)

	// Create a test campaign with a member
	testCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "test-campaign-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:    "Test Campaign",
		OwnerID: "owner123",
		Slug:    "test-campaign-slug",
		Members: []interface{}{"user123"},
	}
	err := repo.Create(testCampaign)
	assert.NoError(t, err)

	// Remove the member
	err = repo.RemoveMember("test-campaign-id", "user123")
	assert.NoError(t, err)

	// Verify the member was removed
	updatedCampaign, err := repo.GetByID("test-campaign-id")
	assert.NoError(t, err)
	assert.NotContains(t, updatedCampaign.Members, "user123")
}

func TestGormCampaignRepository_IsMember(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCampaignRepository(db)

	// Create a test campaign with members
	testCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "test-campaign-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:    "Test Campaign",
		OwnerID: "owner123",
		Slug:    "test-campaign-slug",
		Members: []interface{}{"user123", "user456"},
	}
	err := repo.Create(testCampaign)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		campaignID     string
		userID         string
		expectedResult bool
	}{
		{
			name:           "user is a member",
			campaignID:     "test-campaign-id",
			userID:         "user123",
			expectedResult: true,
		},
		{
			name:           "user is not a member",
			campaignID:     "test-campaign-id",
			userID:         "user789",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.IsMember(tt.campaignID, tt.userID)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestGormCampaignRepository_AddSponsor(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCampaignRepository(db)

	// Create a test campaign first
	testCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "test-campaign-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:     "Test Campaign",
		OwnerID:  "owner123",
		Slug:     "test-campaign-slug",
		Sponsors: []interface{}{},
	}
	err := repo.Create(testCampaign)
	assert.NoError(t, err)

	// Add a sponsor
	err = repo.AddSponsor("test-campaign-id", "user123")
	assert.NoError(t, err)

	// Verify the sponsor was added
	updatedCampaign, err := repo.GetByID("test-campaign-id")
	assert.NoError(t, err)
	assert.Contains(t, updatedCampaign.Sponsors, "user123")
}

func TestGormCampaignRepository_RemoveSponsor(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCampaignRepository(db)

	// Create a test campaign with a sponsor
	testCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "test-campaign-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:     "Test Campaign",
		OwnerID:  "owner123",
		Slug:     "test-campaign-slug",
		Sponsors: []interface{}{"user123"},
	}
	err := repo.Create(testCampaign)
	assert.NoError(t, err)

	// Remove the sponsor
	err = repo.RemoveSponsor("test-campaign-id", "user123")
	assert.NoError(t, err)

	// Verify the sponsor was removed
	updatedCampaign, err := repo.GetByID("test-campaign-id")
	assert.NoError(t, err)
	assert.NotContains(t, updatedCampaign.Sponsors, "user123")
}

func TestGormCampaignRepository_IsSponsor(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCampaignRepository(db)

	// Create a test campaign with sponsors
	testCampaign := &campaignModel.Campaign{
		Base: model.Base{
			ID:        "test-campaign-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:     "Test Campaign",
		OwnerID:  "owner123",
		Slug:     "test-campaign-slug",
		Sponsors: []interface{}{"user123", "user456"},
	}
	err := repo.Create(testCampaign)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		campaignID     string
		userID         string
		expectedResult bool
	}{
		{
			name:           "user is a sponsor",
			campaignID:     "test-campaign-id",
			userID:         "user123",
			expectedResult: true,
		},
		{
			name:           "user is not a sponsor",
			campaignID:     "test-campaign-id",
			userID:         "user789",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.IsSponsor(tt.campaignID, tt.userID)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
