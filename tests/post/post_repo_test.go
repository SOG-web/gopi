package post_test

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	postGorm "gopi.com/internal/data/post/model/gorm"
	"gopi.com/internal/data/post/repo"
	"gopi.com/internal/domain/model"
	postModel "gopi.com/internal/domain/post/model"
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
	err = db.AutoMigrate(&postGorm.Post{}, &postGorm.Comment{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestGormPostRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormPostRepository(db)

	tests := []struct {
		name        string
		post        *postModel.Post
		expectedErr error
	}{
		{
			name: "successful post creation",
			post: &postModel.Post{
				Base: model.Base{
					ID:        "test-post-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Title:         "Test Post",
				Slug:          "test-post",
				Content:       "This is a test post content",
				AuthorID:      "author123",
				CoverImageURL: "test-image.jpg",
				IsPublished:   true,
				PublishedAt:   &time.Time{},
			},
			expectedErr: nil,
		},
		{
			name: "unpublished post creation",
			post: &postModel.Post{
				Base: model.Base{
					ID:        "unpublished-post-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Title:         "Unpublished Post",
				Slug:          "unpublished-post",
				Content:       "This is an unpublished post",
				AuthorID:      "author456",
				CoverImageURL: "",
				IsPublished:   false,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.post)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.post.ID)

				// Verify the post was created in the database
				var dbPost postGorm.Post
				result := db.First(&dbPost, "id = ?", tt.post.ID)
				assert.NoError(t, result.Error)
				assert.Equal(t, tt.post.Title, dbPost.Title)
				assert.Equal(t, tt.post.AuthorID, dbPost.AuthorID)
				assert.Equal(t, tt.post.Slug, dbPost.Slug)
			}
		})
	}
}

func TestGormPostRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormPostRepository(db)

	// Create a test post first
	testPost := &postModel.Post{
		Base: model.Base{
			ID:        "test-post-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Title:         "Test Post",
		Slug:          "test-post",
		Content:       "Test content",
		AuthorID:      "author123",
		CoverImageURL: "test-image.jpg",
		IsPublished:   true,
		PublishedAt:   &time.Time{},
	}
	err := repo.Create(testPost)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		postID      string
		expectedErr error
	}{
		{
			name:        "successful retrieval",
			postID:      "test-post-id",
			expectedErr: nil,
		},
		{
			name:        "post not found",
			postID:      "nonexistent-id",
			expectedErr: gormLib.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post, err := repo.GetByID(tt.postID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, post)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, post)
				assert.Equal(t, tt.postID, post.ID)
				assert.Equal(t, "Test Post", post.Title)
			}
		})
	}
}

func TestGormPostRepository_GetBySlug(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormPostRepository(db)

	// Create test posts
	posts := []*postModel.Post{
		{
			Base: model.Base{
				ID:        "post1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Title:         "First Post",
			Slug:          "first-post",
			Content:       "First post content",
			AuthorID:      "author123",
			CoverImageURL: "first-image.jpg",
			IsPublished:   true,
			PublishedAt:   &time.Time{},
		},
		{
			Base: model.Base{
				ID:        "post2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Title:         "Second Post",
			Slug:          "second-post",
			Content:       "Second post content",
			AuthorID:      "author456",
			CoverImageURL: "second-image.jpg",
			IsPublished:   false,
		},
	}

	for _, post := range posts {
		err := repo.Create(post)
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
			slug:        "first-post",
			expectedErr: nil,
			expectedID:  "post1",
		},
		{
			name:        "successful retrieval by second slug",
			slug:        "second-post",
			expectedErr: nil,
			expectedID:  "post2",
		},
		{
			name:        "post not found",
			slug:        "nonexistent-slug",
			expectedErr: gormLib.ErrRecordNotFound,
			expectedID:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post, err := repo.GetBySlug(tt.slug)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, post)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, post)
				assert.Equal(t, tt.expectedID, post.ID)
				assert.Equal(t, tt.slug, post.Slug)
			}
		})
	}
}

func TestGormPostRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormPostRepository(db)

	// Create a test post first
	testPost := &postModel.Post{
		Base: model.Base{
			ID:        "test-post-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Title:         "Original Title",
		Slug:          "original-title",
		Content:       "Original content",
		AuthorID:      "author123",
		CoverImageURL: "original-image.jpg",
		IsPublished:   false,
	}
	err := repo.Create(testPost)
	assert.NoError(t, err)

	// Update the post
	testPost.Title = "Updated Title"
	testPost.Content = "Updated content"
	testPost.CoverImageURL = "updated-image.jpg"
	testPost.IsPublished = true
	now := time.Now()
	testPost.PublishedAt = &now
	testPost.UpdatedAt = now

	err = repo.Update(testPost)
	assert.NoError(t, err)

	// Verify the update
	updatedPost, err := repo.GetByID("test-post-id")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", updatedPost.Title)
	assert.Equal(t, "Updated content", updatedPost.Content)
	assert.Equal(t, "updated-image.jpg", updatedPost.CoverImageURL)
	assert.True(t, updatedPost.IsPublished)
	assert.NotNil(t, updatedPost.PublishedAt)
}

func TestGormPostRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormPostRepository(db)

	// Create a test post first
	testPost := &postModel.Post{
		Base: model.Base{
			ID:        "test-post-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Title:         "Test Post",
		Slug:          "test-post",
		Content:       "Test content",
		AuthorID:      "author123",
		CoverImageURL: "test-image.jpg",
		IsPublished:   true,
		PublishedAt:   &time.Time{},
	}
	err := repo.Create(testPost)
	assert.NoError(t, err)

	// Delete the post
	err = repo.Delete("test-post-id")
	assert.NoError(t, err)

	// Verify the post was deleted
	_, err = repo.GetByID("test-post-id")
	assert.Error(t, err)
	assert.Equal(t, gormLib.ErrRecordNotFound, err)
}

func TestGormPostRepository_ListPublished(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormPostRepository(db)

	// Create test posts with different publication status
	posts := []*postModel.Post{
		{
			Base: model.Base{
				ID:        "published1",
				CreatedAt: time.Now().Add(-time.Hour * 2),
				UpdatedAt: time.Now().Add(-time.Hour * 2),
			},
			Title:         "Published Post 1",
			Slug:          "published-post-1",
			Content:       "Published content 1",
			AuthorID:      "author123",
			CoverImageURL: "image1.jpg",
			IsPublished:   true,
			PublishedAt:   &time.Time{},
		},
		{
			Base: model.Base{
				ID:        "published2",
				CreatedAt: time.Now().Add(-time.Hour),
				UpdatedAt: time.Now().Add(-time.Hour),
			},
			Title:         "Published Post 2",
			Slug:          "published-post-2",
			Content:       "Published content 2",
			AuthorID:      "author456",
			CoverImageURL: "image2.jpg",
			IsPublished:   true,
			PublishedAt:   &time.Time{},
		},
		{
			Base: model.Base{
				ID:        "unpublished",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Title:         "Unpublished Post",
			Slug:          "unpublished-post",
			Content:       "Unpublished content",
			AuthorID:      "author123",
			CoverImageURL: "image3.jpg",
			IsPublished:   false,
		},
	}

	for _, post := range posts {
		err := repo.Create(post)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "list all published posts",
			limit:         10,
			offset:        0,
			expectedCount: 2,
		},
		{
			name:          "list published posts with limit",
			limit:         1,
			offset:        0,
			expectedCount: 1,
		},
		{
			name:          "list published posts with offset",
			limit:         10,
			offset:        1,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts, err := repo.ListPublished(tt.limit, tt.offset)
			assert.NoError(t, err)
			assert.Len(t, posts, tt.expectedCount)

			// Verify all returned posts are published
			for _, post := range posts {
				assert.True(t, post.IsPublished)
				assert.NotNil(t, post.PublishedAt)
			}
		})
	}
}

func TestGormPostRepository_ListByAuthor(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormPostRepository(db)

	// Create test posts by different authors
	posts := []*postModel.Post{
		{
			Base: model.Base{
				ID:        "post1",
				CreatedAt: time.Now().Add(-time.Hour * 2),
				UpdatedAt: time.Now().Add(-time.Hour * 2),
			},
			Title:         "Post 1",
			Slug:          "post-1",
			Content:       "Content 1",
			AuthorID:      "author123",
			CoverImageURL: "image1.jpg",
			IsPublished:   true,
			PublishedAt:   &time.Time{},
		},
		{
			Base: model.Base{
				ID:        "post2",
				CreatedAt: time.Now().Add(-time.Hour),
				UpdatedAt: time.Now().Add(-time.Hour),
			},
			Title:         "Post 2",
			Slug:          "post-2",
			Content:       "Content 2",
			AuthorID:      "author123",
			CoverImageURL: "image2.jpg",
			IsPublished:   false,
		},
		{
			Base: model.Base{
				ID:        "post3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Title:         "Post 3",
			Slug:          "post-3",
			Content:       "Content 3",
			AuthorID:      "author456",
			CoverImageURL: "image3.jpg",
			IsPublished:   true,
			PublishedAt:   &time.Time{},
		},
	}

	for _, post := range posts {
		err := repo.Create(post)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		authorID      string
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "list posts by author123",
			authorID:      "author123",
			limit:         10,
			offset:        0,
			expectedCount: 2,
		},
		{
			name:          "list posts by author456",
			authorID:      "author456",
			limit:         10,
			offset:        0,
			expectedCount: 1,
		},
		{
			name:          "list posts by non-existent author",
			authorID:      "nonexistent",
			limit:         10,
			offset:        0,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts, err := repo.ListByAuthor(tt.authorID, tt.limit, tt.offset)
			assert.NoError(t, err)
			assert.Len(t, posts, tt.expectedCount)

			// Verify all returned posts have the correct author
			for _, post := range posts {
				assert.Equal(t, tt.authorID, post.AuthorID)
			}
		})
	}
}

func TestGormPostRepository_SearchPublished(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormPostRepository(db)

	// Create test posts
	posts := []*postModel.Post{
		{
			Base: model.Base{
				ID:        "post1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Title:         "Golang Tutorial",
			Slug:          "golang-tutorial",
			Content:       "Learn Golang basics",
			AuthorID:      "author123",
			CoverImageURL: "golang.jpg",
			IsPublished:   true,
			PublishedAt:   &time.Time{},
		},
		{
			Base: model.Base{
				ID:        "post2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Title:         "Python Guide",
			Slug:          "python-guide",
			Content:       "Python programming",
			AuthorID:      "author456",
			CoverImageURL: "python.jpg",
			IsPublished:   true,
			PublishedAt:   &time.Time{},
		},
		{
			Base: model.Base{
				ID:        "post3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Title:         "JavaScript Tips",
			Slug:          "javascript-tips",
			Content:       "JS best practices",
			AuthorID:      "author123",
			CoverImageURL: "js.jpg",
			IsPublished:   false, // Unpublished
		},
	}

	for _, post := range posts {
		err := repo.Create(post)
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
			name:          "search for 'golang' posts",
			query:         "golang",
			limit:         10,
			offset:        0,
			expectedCount: 1,
		},
		{
			name:          "search for 'programming' posts (should match python)",
			query:         "programming",
			limit:         10,
			offset:        0,
			expectedCount: 1,
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
			query:         "golang",
			limit:         1,
			offset:        0,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts, err := repo.SearchPublished(tt.query, tt.limit, tt.offset)
			assert.NoError(t, err)
			assert.Len(t, posts, tt.expectedCount)

			// Verify search results match the query (case-insensitive)
			for _, post := range posts {
				assert.True(t, post.IsPublished) // Should only return published posts
				assert.Regexp(t, regexp.MustCompile(`(?i)`+tt.query), post.Title+" "+post.Content)
			}
		})
	}
}

// Test Comment Repository
func TestGormCommentRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCommentRepository(db)

	testComment := &postModel.Comment{
		Base: model.Base{
			ID:        "test-comment-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		AuthorID:   "author123",
		Content:    "This is a test comment!",
		TargetType: "post",
		TargetID:   "post123",
		ParentID:   nil,
		IsDeleted:  false,
	}

	err := repo.Create(testComment)
	assert.NoError(t, err)
	assert.NotEmpty(t, testComment.ID)

	// Verify the comment was created in the database
	var dbComment postGorm.Comment
	result := db.First(&dbComment, "id = ?", testComment.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, testComment.Content, dbComment.Content)
	assert.Equal(t, testComment.AuthorID, dbComment.AuthorID)
	assert.Equal(t, testComment.TargetType, dbComment.TargetType)
	assert.Equal(t, testComment.TargetID, dbComment.TargetID)
}

func TestGormCommentRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCommentRepository(db)

	// Create a test comment first
	testComment := &postModel.Comment{
		Base: model.Base{
			ID:        "test-comment-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		AuthorID:   "author123",
		Content:    "Test comment content",
		TargetType: "post",
		TargetID:   "post123",
		ParentID:   nil,
		IsDeleted:  false,
	}
	err := repo.Create(testComment)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		commentID   string
		expectedErr error
	}{
		{
			name:        "successful retrieval",
			commentID:   "test-comment-id",
			expectedErr: nil,
		},
		{
			name:        "comment not found",
			commentID:   "nonexistent-id",
			expectedErr: gormLib.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comment, err := repo.GetByID(tt.commentID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, comment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, comment)
				assert.Equal(t, tt.commentID, comment.ID)
				assert.Equal(t, "Test comment content", comment.Content)
			}
		})
	}
}

func TestGormCommentRepository_ListByTarget(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCommentRepository(db)

	// Create test comments for different targets
	comments := []*postModel.Comment{
		{
			Base: model.Base{
				ID:        "comment1",
				CreatedAt: time.Now().Add(-time.Hour),
				UpdatedAt: time.Now().Add(-time.Hour),
			},
			AuthorID:   "user1",
			Content:    "First comment",
			TargetType: "post",
			TargetID:   "post123",
			ParentID:   nil,
			IsDeleted:  false,
		},
		{
			Base: model.Base{
				ID:        "comment2",
				CreatedAt: time.Now().Add(-time.Minute * 30),
				UpdatedAt: time.Now().Add(-time.Minute * 30),
			},
			AuthorID:   "user2",
			Content:    "Second comment",
			TargetType: "post",
			TargetID:   "post123",
			ParentID:   nil,
			IsDeleted:  false,
		},
		{
			Base: model.Base{
				ID:        "comment3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			AuthorID:   "user3",
			Content:    "Comment for different post",
			TargetType: "post",
			TargetID:   "post456",
			ParentID:   nil,
			IsDeleted:  false,
		},
	}

	for _, comment := range comments {
		err := repo.Create(comment)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		targetType    string
		targetID      string
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "list all comments for post123",
			targetType:    "post",
			targetID:      "post123",
			limit:         10,
			offset:        0,
			expectedCount: 2,
		},
		{
			name:          "list comments with limit",
			targetType:    "post",
			targetID:      "post123",
			limit:         1,
			offset:        0,
			expectedCount: 1,
		},
		{
			name:          "list comments for different post",
			targetType:    "post",
			targetID:      "post456",
			limit:         10,
			offset:        0,
			expectedCount: 1,
		},
		{
			name:          "list comments for non-existent target",
			targetType:    "post",
			targetID:      "nonexistent",
			limit:         10,
			offset:        0,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comments, err := repo.ListByTarget(tt.targetType, tt.targetID, tt.limit, tt.offset)
			assert.NoError(t, err)
			assert.Len(t, comments, tt.expectedCount)

			// Verify all returned comments belong to the correct target
			for _, comment := range comments {
				assert.Equal(t, tt.targetType, comment.TargetType)
				assert.Equal(t, tt.targetID, comment.TargetID)
			}
		})
	}
}

func TestGormCommentRepository_ListByAuthor(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCommentRepository(db)

	// Create test comments from different authors
	comments := []*postModel.Comment{
		{
			Base: model.Base{
				ID:        "comment1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			AuthorID:   "user123",
			Content:    "Comment from user123",
			TargetType: "post",
			TargetID:   "post1",
			ParentID:   nil,
			IsDeleted:  false,
		},
		{
			Base: model.Base{
				ID:        "comment2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			AuthorID:   "user123",
			Content:    "Another comment from user123",
			TargetType: "challenge",
			TargetID:   "challenge1",
			ParentID:   nil,
			IsDeleted:  false,
		},
		{
			Base: model.Base{
				ID:        "comment3",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			AuthorID:   "user456",
			Content:    "Comment from user456",
			TargetType: "post",
			TargetID:   "post2",
			ParentID:   nil,
			IsDeleted:  false,
		},
	}

	for _, comment := range comments {
		err := repo.Create(comment)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		authorID      string
		expectedCount int
	}{
		{
			name:          "list comments from user123",
			authorID:      "user123",
			expectedCount: 2,
		},
		{
			name:          "list comments from user456",
			authorID:      "user456",
			expectedCount: 1,
		},
		{
			name:          "list comments from user with no comments",
			authorID:      "user789",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comments, err := repo.ListByAuthor(tt.authorID, 10, 0)
			assert.NoError(t, err)
			assert.Len(t, comments, tt.expectedCount)

			// Verify all returned comments have the correct author
			for _, comment := range comments {
				assert.Equal(t, tt.authorID, comment.AuthorID)
			}
		})
	}
}

// MISSING TESTS FOR GORM COMMENT REPOSITORY
// These tests cover the 3 missing methods that are critical for comment functionality

func TestGormCommentRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCommentRepository(db)

	// Create a test comment first
	originalComment := &postModel.Comment{
		Base: model.Base{
			ID:        "test-comment-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		AuthorID:   "author123",
		Content:    "Original comment content",
		TargetType: "post",
		TargetID:   "post123",
		ParentID:   nil,
		IsDeleted:  false,
	}

	err := repo.Create(originalComment)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		commentID   string
		updateData  *postModel.Comment
		expectedErr error
	}{
		{
			name:      "successful comment update",
			commentID: "test-comment-id",
			updateData: &postModel.Comment{
				Base: model.Base{
					ID:        "test-comment-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				AuthorID:   "author123",
				Content:    "Updated comment content",
				TargetType: "post",
				TargetID:   "post123",
				ParentID:   nil,
				IsDeleted:  false,
			},
			expectedErr: nil,
		},
		{
			name:      "partial update - only content",
			commentID: "test-comment-id",
			updateData: &postModel.Comment{
				Base: model.Base{
					ID:        "test-comment-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				AuthorID:   "author123",
				Content:    "Partially updated content",
				TargetType: "post",
				TargetID:   "post123",
				ParentID:   nil,
				IsDeleted:  false,
			},
			expectedErr: nil,
		},
		{
			name:      "update with reply structure",
			commentID: "test-comment-id",
			updateData: &postModel.Comment{
				Base: model.Base{
					ID:        "test-comment-id",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				AuthorID:   "author123",
				Content:    "Reply comment content",
				TargetType: "post",
				TargetID:   "post123",
				ParentID:   stringPtr("parent-comment-id"),
				IsDeleted:  false,
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

				// Verify the comment was updated in the database
				var dbComment postGorm.Comment
				result := db.First(&dbComment, "id = ?", tt.commentID)
				assert.NoError(t, result.Error)
				assert.Equal(t, tt.updateData.Content, dbComment.Content)
				assert.Equal(t, tt.updateData.AuthorID, dbComment.AuthorID)
				assert.Equal(t, tt.updateData.TargetType, dbComment.TargetType)
				assert.Equal(t, tt.updateData.TargetID, dbComment.TargetID)
				assert.Equal(t, tt.updateData.IsDeleted, dbComment.IsDeleted)

				// Verify ParentID handling
				if tt.updateData.ParentID != nil {
					assert.NotNil(t, dbComment.ParentID)
					assert.Equal(t, *tt.updateData.ParentID, *dbComment.ParentID)
				} else {
					assert.Nil(t, dbComment.ParentID)
				}

				// Verify updated_at was updated
				assert.True(t, dbComment.UpdatedAt.After(originalComment.UpdatedAt))
			}
		})
	}
}

func TestGormCommentRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCommentRepository(db)

	// Create a test comment first
	testComment := &postModel.Comment{
		Base: model.Base{
			ID:        "test-comment-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		AuthorID:   "author123",
		Content:    "Comment to be deleted",
		TargetType: "post",
		TargetID:   "post123",
		ParentID:   nil,
		IsDeleted:  false,
	}

	err := repo.Create(testComment)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		commentID   string
		expectedErr error
	}{
		{
			name:        "successful comment deletion",
			commentID:   "test-comment-id",
			expectedErr: nil,
		},
		{
			name:        "delete nonexistent comment",
			commentID:   "nonexistent-id",
			expectedErr: nil, // GORM Delete doesn't return error for non-existent records
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(tt.commentID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)

				// Verify the comment was deleted from the database
				var dbComment postGorm.Comment
				result := db.First(&dbComment, "id = ?", tt.commentID)
				assert.Error(t, result.Error)
				assert.True(t, errors.Is(result.Error, gormLib.ErrRecordNotFound))
			}
		})
	}
}

func TestGormCommentRepository_ListReplies(t *testing.T) {
	db := setupTestDB(t)
	repo := repo.NewGormCommentRepository(db)

	// Create test data
	parentCommentID := "parent-comment-id"
	replyCommentID1 := "reply-comment-id-1"
	replyCommentID2 := "reply-comment-id-2"
	replyCommentID3 := "reply-comment-id-3"

	// Create parent comment
	parentComment := &postModel.Comment{
		Base: model.Base{
			ID:        parentCommentID,
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now().Add(-time.Hour),
		},
		AuthorID:   "author123",
		Content:    "Parent comment",
		TargetType: "post",
		TargetID:   "post123",
		ParentID:   nil,
		IsDeleted:  false,
	}
	err := repo.Create(parentComment)
	assert.NoError(t, err)

	// Create reply comments
	replyComments := []struct {
		id        string
		authorID  string
		content   string
		createdAt time.Time
	}{
		{replyCommentID1, "user1", "First reply", time.Now().Add(-time.Minute * 30)},
		{replyCommentID2, "user2", "Second reply", time.Now().Add(-time.Minute * 15)},
		{replyCommentID3, "user3", "Third reply", time.Now()},
	}

	for _, reply := range replyComments {
		replyComment := &postModel.Comment{
			Base: model.Base{
				ID:        reply.id,
				CreatedAt: reply.createdAt,
				UpdatedAt: reply.createdAt,
			},
			AuthorID:   reply.authorID,
			Content:    reply.content,
			TargetType: "post",
			TargetID:   "post123",
			ParentID:   &parentCommentID,
			IsDeleted:  false,
		}
		err := repo.Create(replyComment)
		assert.NoError(t, err)
	}

	// Create a comment that replies to a reply (nested reply)
	nestedReplyID := "nested-reply-id"
	nestedReply := &postModel.Comment{
		Base: model.Base{
			ID:        nestedReplyID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		AuthorID:   "user4",
		Content:    "Nested reply",
		TargetType: "post",
		TargetID:   "post123",
		ParentID:   &replyCommentID1,
		IsDeleted:  false,
	}
	err = repo.Create(nestedReply)
	assert.NoError(t, err)

	// Create a soft-deleted reply
	deletedReplyID := "deleted-reply-id"
	deletedReply := &postModel.Comment{
		Base: model.Base{
			ID:        deletedReplyID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		AuthorID:   "user5",
		Content:    "Deleted reply",
		TargetType: "post",
		TargetID:   "post123",
		ParentID:   &parentCommentID,
		IsDeleted:  true, // This should be filtered out
	}
	err = repo.Create(deletedReply)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		parentID      string
		limit         int
		offset        int
		expectedCount int
		expectedErr   error
	}{
		{
			name:          "list all replies to parent comment",
			parentID:      parentCommentID,
			limit:         10,
			offset:        0,
			expectedCount: 3, // Should get 3 replies (excluding nested reply and deleted reply)
			expectedErr:   nil,
		},
		{
			name:          "list replies with limit",
			parentID:      parentCommentID,
			limit:         2,
			offset:        0,
			expectedCount: 2,
			expectedErr:   nil,
		},
		{
			name:          "list replies with offset",
			parentID:      parentCommentID,
			limit:         10,
			offset:        2,
			expectedCount: 1,
			expectedErr:   nil,
		},
		{
			name:          "list replies for nested comment",
			parentID:      replyCommentID1,
			limit:         10,
			offset:        0,
			expectedCount: 1, // Should get the nested reply
			expectedErr:   nil,
		},
		{
			name:          "list replies for nonexistent parent",
			parentID:      "nonexistent-parent",
			limit:         10,
			offset:        0,
			expectedCount: 0,
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.ListReplies(tt.parentID, tt.limit, tt.offset)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)

				// Verify all returned comments have the correct parent ID
				for _, comment := range result {
					assert.NotNil(t, comment.ParentID)
					assert.Equal(t, tt.parentID, *comment.ParentID)
					assert.False(t, comment.IsDeleted) // Soft deleted comments should be filtered out
				}

				// Verify ordering by created_at ascending
				if len(result) > 1 {
					for i := 0; i < len(result)-1; i++ {
						assert.True(t, result[i].CreatedAt.Before(result[i+1].CreatedAt) ||
							result[i].CreatedAt.Equal(result[i+1].CreatedAt))
					}
				}
			}
		})
	}
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
