package chat_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopi.com/internal/data/chat/model/gorm"
	"gopi.com/internal/data/chat/repo"
	chatModel "gopi.com/internal/domain/chat/model"
	"gopi.com/internal/domain/model"
	"gorm.io/driver/sqlite"
	gormLib "gorm.io/gorm"
)

// Setup in-memory database for testing
func setupTestDB(t *testing.T) *gormLib.DB {
	db, err := gormLib.Open(sqlite.Open(":memory:"), &gormLib.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&gorm.Group{}, &gorm.Message{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestGormGroupRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormGroupRepository(db)

	tests := []struct {
		name        string
		group       *chatModel.Group
		expectedErr error
	}{
		{
			name: "successful group creation",
			group: &chatModel.Group{
				Base: model.Base{
					ID:        "test-group-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:      "Test Group",
				Image:     "test-image.jpg",
				CreatorID: "creator123",
				MemberIDs: []string{"creator123", "member1", "member2"},
				Slug:      "test-group-abc123",
			},
			expectedErr: nil,
		},
		{
			name: "group with empty name",
			group: &chatModel.Group{
				Base: model.Base{
					ID:        "test-group-id-2",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:      "",
				CreatorID: "creator123",
				MemberIDs: []string{"creator123"},
				Slug:      "empty-group",
			},
			expectedErr: nil, // SQLite allows empty strings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.group)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.group.ID)
			}

			// Verify the group was created in the database
			var dbGroup gorm.Group
			result := db.First(&dbGroup, "id = ?", tt.group.ID)
			if tt.expectedErr == nil {
				assert.NoError(t, result.Error)
				assert.Equal(t, tt.group.Name, dbGroup.Name)
				assert.Equal(t, tt.group.CreatorID, dbGroup.CreatorID)
				assert.Equal(t, tt.group.Slug, dbGroup.Slug)
			}
		})
	}
}

func TestGormGroupRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormGroupRepository(db)

	// Create a test group first
	testGroup := &chatModel.Group{
		Base: model.Base{
			ID:        "test-group-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:      "Test Group",
		Image:     "test-image.jpg",
		CreatorID: "creator123",
		MemberIDs: []string{"creator123", "member1"},
		Slug:      "test-group-abc123",
	}
	err := repo.Create(testGroup)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		groupID     string
		expectedErr error
	}{
		{
			name:        "successful retrieval",
			groupID:     "test-group-id",
			expectedErr: nil,
		},
		{
			name:        "group not found",
			groupID:     "nonexistent-id",
			expectedErr: gormLib.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := repo.GetByID(tt.groupID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, group)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, group)
				assert.Equal(t, tt.groupID, group.ID)
				assert.Equal(t, "Test Group", group.Name)
				assert.Equal(t, "creator123", group.CreatorID)
			}
		})
	}
}

func TestGormGroupRepository_GetBySlug(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormGroupRepository(db)

	// Create test groups
	groups := []*chatModel.Group{
		{
			Base: model.Base{
				ID:        "group1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:      "First Group",
			CreatorID: "creator123",
			MemberIDs: []string{"creator123"},
			Slug:      "first-group-abc123",
		},
		{
			Base: model.Base{
				ID:        "group2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:      "Second Group",
			CreatorID: "creator456",
			MemberIDs: []string{"creator456", "member1"},
			Slug:      "second-group-def456",
		},
	}

	for _, group := range groups {
		err := repo.Create(group)
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
			slug:        "first-group-abc123",
			expectedErr: nil,
			expectedID:  "group1",
		},
		{
			name:        "successful retrieval by second slug",
			slug:        "second-group-def456",
			expectedErr: nil,
			expectedID:  "group2",
		},
		{
			name:        "group not found",
			slug:        "nonexistent-slug",
			expectedErr: gormLib.ErrRecordNotFound,
			expectedID:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := repo.GetBySlug(tt.slug)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, group)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, group)
				assert.Equal(t, tt.expectedID, group.ID)
				assert.Equal(t, tt.slug, group.Slug)
			}
		})
	}
}

func TestGormGroupRepository_GetByMemberID(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormGroupRepository(db)

	// Create test groups with different members
	groups := []*chatModel.Group{
		{
			Base: model.Base{
				ID:        "group1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:      "Group 1",
			CreatorID: "creator1",
			MemberIDs: []string{"creator1", "user123", "member1"},
			Slug:      "group-1-abc123",
		},
		{
			Base: model.Base{
				ID:        "group2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:      "Group 2",
			CreatorID: "creator2",
			MemberIDs: []string{"creator2", "user123", "member2"},
			Slug:      "group-2-def456",
		},
		{
			Base: model.Base{
				ID:        "group3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:      "Group 3",
			CreatorID: "creator3",
			MemberIDs: []string{"creator3", "member3"},
			Slug:      "group-3-ghi789",
		},
	}

	for _, group := range groups {
		err := repo.Create(group)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		memberID      string
		expectedCount int
		expectedErr   error
	}{
		{
			name:          "user is member of multiple groups",
			memberID:      "user123",
			expectedCount: 2,
			expectedErr:   nil,
		},
		{
			name:          "user is member of one group",
			memberID:      "member1",
			expectedCount: 1,
			expectedErr:   nil,
		},
		{
			name:          "user is not member of any group",
			memberID:      "lonely-user",
			expectedCount: 0,
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups, err := repo.GetByMemberID(tt.memberID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, groups, tt.expectedCount)

				// Verify that the returned groups contain the member
				for _, group := range groups {
					assert.Contains(t, group.MemberIDs, tt.memberID)
				}
			}
		})
	}
}

func TestGormGroupRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormGroupRepository(db)

	// Create a test group first
	testGroup := &chatModel.Group{
		Base: model.Base{
			ID:        "test-group-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:      "Original Name",
		Image:     "original-image.jpg",
		CreatorID: "creator123",
		MemberIDs: []string{"creator123", "member1"},
		Slug:      "original-slug-abc123",
	}
	err := repo.Create(testGroup)
	assert.NoError(t, err)

	// Update the group
	testGroup.Name = "Updated Name"
	testGroup.Image = "updated-image.jpg"
	testGroup.MemberIDs = []string{"creator123", "member1", "member2"}
	testGroup.UpdatedAt = time.Now()

	err = repo.Update(testGroup)
	assert.NoError(t, err)

	// Verify the update
	updatedGroup, err := repo.GetByID("test-group-id")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedGroup.Name)
	assert.Equal(t, "updated-image.jpg", updatedGroup.Image)
	assert.Len(t, updatedGroup.MemberIDs, 3)
	assert.Contains(t, updatedGroup.MemberIDs, "member2")
}

func TestGormGroupRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormGroupRepository(db)

	// Create a test group first
	testGroup := &chatModel.Group{
		Base: model.Base{
			ID:        "test-group-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:      "Test Group",
		CreatorID: "creator123",
		MemberIDs: []string{"creator123"},
		Slug:      "test-group-abc123",
	}
	err := repo.Create(testGroup)
	assert.NoError(t, err)

	// Delete the group
	err = repo.Delete("test-group-id")
	assert.NoError(t, err)

	// Verify the group was deleted
	_, err = repo.GetByID("test-group-id")
	assert.Error(t, err)
	assert.Equal(t, gormLib.ErrRecordNotFound, err)
}

func TestGormGroupRepository_AddMember(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormGroupRepository(db)

	// Create a test group first
	testGroup := &chatModel.Group{
		Base: model.Base{
			ID:        "test-group-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:      "Test Group",
		CreatorID: "creator123",
		MemberIDs: []string{"creator123", "member1"},
		Slug:      "test-group-abc123",
	}
	err := repo.Create(testGroup)
	assert.NoError(t, err)

	// Add a new member
	err = repo.AddMember("test-group-id", "new-member")
	assert.NoError(t, err)

	// Verify the member was added
	updatedGroup, err := repo.GetByID("test-group-id")
	assert.NoError(t, err)
	assert.Len(t, updatedGroup.MemberIDs, 3)
	assert.Contains(t, updatedGroup.MemberIDs, "new-member")
}

func TestGormGroupRepository_RemoveMember(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormGroupRepository(db)

	// Create a test group first
	testGroup := &chatModel.Group{
		Base: model.Base{
			ID:        "test-group-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:      "Test Group",
		CreatorID: "creator123",
		MemberIDs: []string{"creator123", "member1", "member2"},
		Slug:      "test-group-abc123",
	}
	err := repo.Create(testGroup)
	assert.NoError(t, err)

	// Remove a member
	err = repo.RemoveMember("test-group-id", "member1")
	assert.NoError(t, err)

	// Verify the member was removed
	updatedGroup, err := repo.GetByID("test-group-id")
	assert.NoError(t, err)
	assert.Len(t, updatedGroup.MemberIDs, 2)
	assert.NotContains(t, updatedGroup.MemberIDs, "member1")
	assert.Contains(t, updatedGroup.MemberIDs, "creator123")
	assert.Contains(t, updatedGroup.MemberIDs, "member2")
}

func TestGormGroupRepository_SearchByName(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormGroupRepository(db)

	// Create test groups
	groups := []*chatModel.Group{
		{
			Base: model.Base{
				ID:        "group1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:      "Test Group One",
			CreatorID: "creator1",
			MemberIDs: []string{"creator1"},
			Slug:      "test-group-one-abc123",
		},
		{
			Base: model.Base{
				ID:        "group2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:      "Test Group Two",
			CreatorID: "creator2",
			MemberIDs: []string{"creator2"},
			Slug:      "test-group-two-def456",
		},
		{
			Base: model.Base{
				ID:        "group3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:      "Another Group",
			CreatorID: "creator3",
			MemberIDs: []string{"creator3"},
			Slug:      "another-group-ghi789",
		},
	}

	for _, group := range groups {
		err := repo.Create(group)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		query         string
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "search for 'test' groups",
			query:         "test",
			limit:         10,
			offset:        0,
			expectedCount: 2,
		},
		{
			name:          "search for 'group' (should match all)",
			query:         "group",
			limit:         10,
			offset:        0,
			expectedCount: 3,
		},
		{
			name:          "search for non-existent term",
			query:         "nonexistent",
			limit:         10,
			offset:        0,
			expectedCount: 0,
		},
		{
			name:          "search with limit",
			query:         "group",
			limit:         1,
			offset:        0,
			expectedCount: 1,
		},
		{
			name:          "search with offset",
			query:         "group",
			limit:         10,
			offset:        1,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups, err := repo.SearchByName(tt.query, tt.limit, tt.offset)
			assert.NoError(t, err)
			assert.Len(t, groups, tt.expectedCount)

			// Verify search results match the query (case-insensitive)
			for _, group := range groups {
				assert.Regexp(t, regexp.MustCompile(`(?i)`+tt.query), group.Name)
			}
		})
	}
}

// Test Message Repository
func TestGormMessageRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormMessageRepository(db)

	testMessage := &chatModel.Message{
		Base: model.Base{
			ID:        "test-message-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		SenderID: "sender123",
		Content:  "Hello, this is a test message!",
		GroupID:  "group123",
	}

	err := repo.Create(testMessage)
	assert.NoError(t, err)
	assert.NotEmpty(t, testMessage.ID)

	// Verify the message was created in the database
	var dbMessage gorm.Message
	result := db.First(&dbMessage, "id = ?", testMessage.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, testMessage.Content, dbMessage.Content)
	assert.Equal(t, testMessage.SenderID, dbMessage.SenderID)
	assert.Equal(t, testMessage.GroupID, dbMessage.GroupID)
}

func TestGormMessageRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormMessageRepository(db)

	// Create a test message first
	testMessage := &chatModel.Message{
		Base: model.Base{
			ID:        "test-message-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		SenderID: "sender123",
		Content:  "Test message content",
		GroupID:  "group123",
	}
	err := repo.Create(testMessage)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		messageID   string
		expectedErr error
	}{
		{
			name:        "successful retrieval",
			messageID:   "test-message-id",
			expectedErr: nil,
		},
		{
			name:        "message not found",
			messageID:   "nonexistent-id",
			expectedErr: gormLib.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, err := repo.GetByID(tt.messageID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, message)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, message)
				assert.Equal(t, tt.messageID, message.ID)
				assert.Equal(t, "Test message content", message.Content)
			}
		})
	}
}

func TestGormMessageRepository_GetByGroupID(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormMessageRepository(db)

	// Create test messages for the same group
	messages := []*chatModel.Message{
		{
			Base: model.Base{
				ID:        "msg1",
				CreatedAt: time.Now().Add(-time.Hour),
				UpdatedAt: time.Now().Add(-time.Hour),
			},
			SenderID: "user1",
			Content:  "First message",
			GroupID:  "group123",
		},
		{
			Base: model.Base{
				ID:        "msg2",
				CreatedAt: time.Now().Add(-time.Minute * 30),
				UpdatedAt: time.Now().Add(-time.Minute * 30),
			},
			SenderID: "user2",
			Content:  "Second message",
			GroupID:  "group123",
		},
		{
			Base: model.Base{
				ID:        "msg3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			SenderID: "user3",
			Content:  "Third message",
			GroupID:  "group123",
		},
		{
			Base: model.Base{
				ID:        "msg4",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			SenderID: "user4",
			Content:  "Message for different group",
			GroupID:  "group456",
		},
	}

	for _, msg := range messages {
		err := repo.Create(msg)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		groupID       string
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "get all messages for group",
			groupID:       "group123",
			limit:         10,
			offset:        0,
			expectedCount: 3,
		},
		{
			name:          "get messages with limit",
			groupID:       "group123",
			limit:         2,
			offset:        0,
			expectedCount: 2,
		},
		{
			name:          "get messages with offset",
			groupID:       "group123",
			limit:         10,
			offset:        1,
			expectedCount: 2,
		},
		{
			name:          "get messages for different group",
			groupID:       "group456",
			limit:         10,
			offset:        0,
			expectedCount: 1,
		},
		{
			name:          "get messages for non-existent group",
			groupID:       "nonexistent",
			limit:         10,
			offset:        0,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := repo.GetByGroupID(tt.groupID, tt.limit, tt.offset)
			assert.NoError(t, err)
			assert.Len(t, messages, tt.expectedCount)

			// Verify all returned messages belong to the correct group
			for _, msg := range messages {
				assert.Equal(t, tt.groupID, msg.GroupID)
			}

			// Verify messages are ordered by creation time (most recent first)
			if len(messages) > 1 {
				for i := 0; i < len(messages)-1; i++ {
					assert.True(t, messages[i].CreatedAt.After(messages[i+1].CreatedAt) ||
						messages[i].CreatedAt.Equal(messages[i+1].CreatedAt))
				}
			}
		})
	}
}

func TestGormMessageRepository_GetBySenderID(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormMessageRepository(db)

	// Create test messages from different senders
	messages := []*chatModel.Message{
		{
			Base: model.Base{
				ID:        "msg1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			SenderID: "user123",
			Content:  "Message from user123",
			GroupID:  "group1",
		},
		{
			Base: model.Base{
				ID:        "msg2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			SenderID: "user123",
			Content:  "Another message from user123",
			GroupID:  "group2",
		},
		{
			Base: model.Base{
				ID:        "msg3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			SenderID: "user456",
			Content:  "Message from user456",
			GroupID:  "group1",
		},
	}

	for _, msg := range messages {
		err := repo.Create(msg)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		senderID      string
		expectedCount int
	}{
		{
			name:          "get messages from user123",
			senderID:      "user123",
			expectedCount: 2,
		},
		{
			name:          "get messages from user456",
			senderID:      "user456",
			expectedCount: 1,
		},
		{
			name:          "get messages from user with no messages",
			senderID:      "user789",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := repo.GetBySenderID(tt.senderID)
			assert.NoError(t, err)
			assert.Len(t, messages, tt.expectedCount)

			// Verify all returned messages have the correct sender
			for _, msg := range messages {
				assert.Equal(t, tt.senderID, msg.SenderID)
			}
		})
	}
}

func TestGormMessageRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormMessageRepository(db)

	// Create a test message first
	testMessage := &chatModel.Message{
		Base: model.Base{
			ID:        "test-message-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		SenderID: "sender123",
		Content:  "Original content",
		GroupID:  "group123",
	}
	err := repo.Create(testMessage)
	assert.NoError(t, err)

	// Update the message
	testMessage.Content = "Updated content"
	testMessage.UpdatedAt = time.Now()

	err = repo.Update(testMessage)
	assert.NoError(t, err)

	// Verify the update
	updatedMessage, err := repo.GetByID("test-message-id")
	assert.NoError(t, err)
	assert.Equal(t, "Updated content", updatedMessage.Content)
	assert.True(t, updatedMessage.UpdatedAt.After(testMessage.CreatedAt))
}

func TestGormMessageRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormMessageRepository(db)

	// Create a test message first
	testMessage := &chatModel.Message{
		Base: model.Base{
			ID:        "test-message-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		SenderID: "sender123",
		Content:  "Test message",
		GroupID:  "group123",
	}
	err := repo.Create(testMessage)
	assert.NoError(t, err)

	// Delete the message
	err = repo.Delete("test-message-id")
	assert.NoError(t, err)

	// Verify the message was deleted
	_, err = repo.GetByID("test-message-id")
	assert.Error(t, err)
	assert.Equal(t, gormLib.ErrRecordNotFound, err)
}
