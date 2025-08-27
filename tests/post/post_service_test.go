package post_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopi.com/internal/app/post"
	"gopi.com/internal/domain/model"
	postModel "gopi.com/internal/domain/post/model"
	postMocks "gopi.com/tests/mocks/post"
)

func TestPostService_CreatePost(t *testing.T) {
	tests := []struct {
		name          string
		authorID      string
		title         string
		content       string
		coverImageURL string
		publish       bool
		expectedErr   error
		mockSetup     func(*postMocks.MockPostRepository)
	}{
		{
			name:          "successful post creation with publish",
			authorID:      "author123",
			title:         "Test Post",
			content:       "This is a test post content",
			coverImageURL: "test-image.jpg",
			publish:       true,
			mockSetup: func(mockRepo *postMocks.MockPostRepository) {
				mockRepo.On("Create", mock.MatchedBy(func(p *postModel.Post) bool {
					return p.Title == "Test Post" &&
						p.AuthorID == "author123" &&
						p.Content == "This is a test post content" &&
						p.IsPublished == true &&
						p.PublishedAt != nil
				})).Return(nil)
			},
		},
		{
			name:          "successful post creation without publish",
			authorID:      "author123",
			title:         "Draft Post",
			content:       "This is a draft post",
			coverImageURL: "",
			publish:       false,
			mockSetup: func(mockRepo *postMocks.MockPostRepository) {
				mockRepo.On("Create", mock.MatchedBy(func(p *postModel.Post) bool {
					return p.Title == "Draft Post" &&
						p.AuthorID == "author123" &&
						p.IsPublished == false &&
						p.PublishedAt == nil
				})).Return(nil)
			},
		},
		{
			name:          "title too long",
			authorID:      "author123",
			title:         "This is a very very very very long title that exceeds the 200 character limit and should cause validation to fail when creating a post through the service and is definitely over 200 characters now with additional text to make it longer",
			content:       "Valid content",
			coverImageURL: "",
			publish:       false,
			expectedErr:   errors.New("title must be 1-200 characters"),
			mockSetup:     func(mockRepo *postMocks.MockPostRepository) {},
		},
		{
			name:          "empty title",
			authorID:      "author123",
			title:         "",
			content:       "Valid content",
			coverImageURL: "",
			publish:       false,
			expectedErr:   errors.New("title must be 1-200 characters"),
			mockSetup:     func(mockRepo *postMocks.MockPostRepository) {},
		},
		{
			name:          "empty content",
			authorID:      "author123",
			title:         "Valid Title",
			content:       "",
			coverImageURL: "",
			publish:       false,
			expectedErr:   errors.New("content is required"),
			mockSetup:     func(mockRepo *postMocks.MockPostRepository) {},
		},
		{
			name:          "repository error",
			authorID:      "author123",
			title:         "Test Post",
			content:       "Test content",
			coverImageURL: "",
			publish:       false,
			expectedErr:   errors.New("repository error"),
			mockSetup: func(mockRepo *postMocks.MockPostRepository) {
				mockRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := new(postMocks.MockPostRepository)
			mockCommentRepo := new(postMocks.MockCommentRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockPostRepo)
			}

			postService := post.NewPostService(mockPostRepo, mockCommentRepo)

			result, err := postService.CreatePost(tt.authorID, tt.title, tt.content, tt.coverImageURL, tt.publish)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.title, result.Title)
				assert.Equal(t, tt.authorID, result.AuthorID)
				assert.Equal(t, tt.content, result.Content)
				assert.Equal(t, tt.coverImageURL, result.CoverImageURL)
				assert.Equal(t, tt.publish, result.IsPublished)
				if tt.publish {
					assert.NotNil(t, result.PublishedAt)
				} else {
					assert.Nil(t, result.PublishedAt)
				}
			}

			mockPostRepo.AssertExpectations(t)
		})
	}
}

func TestPostService_GetPostByID(t *testing.T) {
	tests := []struct {
		name        string
		postID      string
		expectedErr error
		mockSetup   func(*postMocks.MockPostRepository)
	}{
		{
			name:   "successful retrieval",
			postID: "post123",
			mockSetup: func(mockRepo *postMocks.MockPostRepository) {
				expectedPost := &postModel.Post{
					Base: model.Base{
						ID:        "post123",
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
				mockRepo.On("GetByID", "post123").Return(expectedPost, nil)
			},
		},
		{
			name:        "post not found",
			postID:      "nonexistent",
			expectedErr: errors.New("post not found"),
			mockSetup: func(mockRepo *postMocks.MockPostRepository) {
				mockRepo.On("GetByID", "nonexistent").Return(nil, errors.New("post not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := new(postMocks.MockPostRepository)
			mockCommentRepo := new(postMocks.MockCommentRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockPostRepo)
			}

			postService := post.NewPostService(mockPostRepo, mockCommentRepo)

			result, err := postService.GetPostByID(tt.postID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.postID, result.ID)
			}

			mockPostRepo.AssertExpectations(t)
		})
	}
}

func TestPostService_UpdatePost(t *testing.T) {
	tests := []struct {
		name          string
		postID        string
		title         string
		content       string
		coverImageURL string
		publish       bool
		expectedErr   error
		mockSetup     func(*postMocks.MockPostRepository)
	}{
		{
			name:          "successful post update",
			postID:        "post123",
			title:         "Updated Title",
			content:       "Updated content",
			coverImageURL: "updated-image.jpg",
			publish:       false,
			mockSetup: func(mockRepo *postMocks.MockPostRepository) {
				existingPost := &postModel.Post{
					Base: model.Base{
						ID:        "post123",
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
				mockRepo.On("GetByID", "post123").Return(existingPost, nil)
				mockRepo.On("Update", mock.MatchedBy(func(p *postModel.Post) bool {
					return p.ID == "post123" &&
						p.Title == "Updated Title" &&
						p.Content == "Updated content" &&
						p.CoverImageURL == "updated-image.jpg" &&
						p.IsPublished == false
				})).Return(nil)
			},
		},
		{
			name:          "post not found",
			postID:        "nonexistent",
			title:         "Updated Title",
			content:       "Updated content",
			coverImageURL: "",
			publish:       false,
			expectedErr:   errors.New("post not found"),
			mockSetup: func(mockRepo *postMocks.MockPostRepository) {
				mockRepo.On("GetByID", "nonexistent").Return(nil, errors.New("post not found"))
			},
		},
		{
			name:          "title too long",
			postID:        "post123",
			title:         "This is a very very very very long title that exceeds the 200 character limit and should cause validation to fail when creating a post through the service and is definitely over 200 characters now with additional text to make it longer",
			content:       "Valid content",
			coverImageURL: "",
			publish:       false,
			expectedErr:   errors.New("title must be at most 200 characters"),
			mockSetup: func(mockRepo *postMocks.MockPostRepository) {
				existingPost := &postModel.Post{
					Base: model.Base{
						ID:        "post123",
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
				mockRepo.On("GetByID", "post123").Return(existingPost, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := new(postMocks.MockPostRepository)
			mockCommentRepo := new(postMocks.MockCommentRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockPostRepo)
			}

			postService := post.NewPostService(mockPostRepo, mockCommentRepo)

			result, err := postService.UpdatePost(tt.postID, tt.title, tt.content, tt.coverImageURL)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.postID, result.ID)
				assert.Equal(t, tt.title, result.Title)
				assert.Equal(t, tt.content, result.Content)
				assert.Equal(t, tt.coverImageURL, result.CoverImageURL)
				assert.Equal(t, tt.publish, result.IsPublished)
			}

			mockPostRepo.AssertExpectations(t)
		})
	}
}

func TestPostService_DeletePost(t *testing.T) {
	tests := []struct {
		name        string
		postID      string
		expectedErr error
		mockSetup   func(*postMocks.MockPostRepository)
	}{
		{
			name:   "successful post deletion",
			postID: "post123",
			mockSetup: func(mockRepo *postMocks.MockPostRepository) {
				mockRepo.On("Delete", "post123").Return(nil)
			},
		},
		{
			name:        "post not found",
			postID:      "nonexistent",
			expectedErr: errors.New("post not found"),
			mockSetup: func(mockRepo *postMocks.MockPostRepository) {
				mockRepo.On("Delete", "nonexistent").Return(errors.New("post not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := new(postMocks.MockPostRepository)
			mockCommentRepo := new(postMocks.MockCommentRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockPostRepo)
			}

			postService := post.NewPostService(mockPostRepo, mockCommentRepo)

			err := postService.DeletePost(tt.postID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockPostRepo.AssertExpectations(t)
		})
	}
}

func TestPostService_ListPublishedPosts(t *testing.T) {
	mockPostRepo := new(postMocks.MockPostRepository)
	mockCommentRepo := new(postMocks.MockCommentRepository)

	expectedPosts := []*postModel.Post{
		{
			Base: model.Base{
				ID:        "post1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
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
				ID:        "post2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Title:         "Published Post 2",
			Slug:          "published-post-2",
			Content:       "Published content 2",
			AuthorID:      "author456",
			CoverImageURL: "image2.jpg",
			IsPublished:   true,
			PublishedAt:   &time.Time{},
		},
	}

	mockPostRepo.On("ListPublished", 10, 0).Return(expectedPosts, nil)

	postService := post.NewPostService(mockPostRepo, mockCommentRepo)

	posts, err := postService.ListPublished(10, 0)

	assert.NoError(t, err)
	assert.Len(t, posts, 2)
	assert.Equal(t, "post1", posts[0].ID)
	assert.Equal(t, "post2", posts[1].ID)
	assert.True(t, posts[0].IsPublished)
	assert.True(t, posts[1].IsPublished)

	mockPostRepo.AssertExpectations(t)
}

func TestPostService_ListPostsByAuthor(t *testing.T) {
	mockPostRepo := new(postMocks.MockPostRepository)
	mockCommentRepo := new(postMocks.MockCommentRepository)

	expectedPosts := []*postModel.Post{
		{
			Base: model.Base{
				ID:        "post1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Title:         "Author's Post 1",
			Slug:          "authors-post-1",
			Content:       "Content 1",
			AuthorID:      "author123",
			CoverImageURL: "image1.jpg",
			IsPublished:   true,
			PublishedAt:   &time.Time{},
		},
		{
			Base: model.Base{
				ID:        "post2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Title:         "Author's Post 2",
			Slug:          "authors-post-2",
			Content:       "Content 2",
			AuthorID:      "author123",
			CoverImageURL: "image2.jpg",
			IsPublished:   false,
		},
	}

	mockPostRepo.On("ListByAuthor", "author123", 10, 0).Return(expectedPosts, nil)

	postService := post.NewPostService(mockPostRepo, mockCommentRepo)

	posts, err := postService.ListByAuthor("author123", 10, 0)

	assert.NoError(t, err)
	assert.Len(t, posts, 2)
	assert.Equal(t, "author123", posts[0].AuthorID)
	assert.Equal(t, "author123", posts[1].AuthorID)

	mockPostRepo.AssertExpectations(t)
}

func TestPostService_SearchPublishedPosts(t *testing.T) {
	mockPostRepo := new(postMocks.MockPostRepository)
	mockCommentRepo := new(postMocks.MockCommentRepository)

	expectedPosts := []*postModel.Post{
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
	}

	mockPostRepo.On("SearchPublished", "golang", 10, 0).Return(expectedPosts, nil)

	postService := post.NewPostService(mockPostRepo, mockCommentRepo)

	posts, err := postService.SearchPublished("golang", 10, 0)

	assert.NoError(t, err)
	assert.Len(t, posts, 1)
	assert.Equal(t, "post1", posts[0].ID)
	assert.True(t, posts[0].IsPublished)

	mockPostRepo.AssertExpectations(t)
}

func TestPostService_GetPostBySlug(t *testing.T) {
	mockPostRepo := new(postMocks.MockPostRepository)
	mockCommentRepo := new(postMocks.MockCommentRepository)

	expectedPost := &postModel.Post{
		Base: model.Base{
			ID:        "post123",
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

	mockPostRepo.On("GetBySlug", "test-post").Return(expectedPost, nil)

	postService := post.NewPostService(mockPostRepo, mockCommentRepo)

	post, err := postService.GetPostBySlug("test-post")

	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, "test-post", post.Slug)
	assert.Equal(t, "Test Post", post.Title)

	mockPostRepo.AssertExpectations(t)
}

func TestPostService_CreateComment(t *testing.T) {
	tests := []struct {
		name        string
		authorID    string
		content     string
		targetType  string
		targetID    string
		parentID    *string
		expectedErr error
		mockSetup   func(*postMocks.MockCommentRepository)
	}{
		{
			name:       "successful comment creation",
			authorID:   "author123",
			content:    "This is a test comment",
			targetType: "post",
			targetID:   "post123",
			parentID:   nil,
			mockSetup: func(mockRepo *postMocks.MockCommentRepository) {
				mockRepo.On("Create", mock.MatchedBy(func(c *postModel.Comment) bool {
					return c.AuthorID == "author123" &&
						c.Content == "This is a test comment" &&
						c.TargetType == "post" &&
						c.TargetID == "post123" &&
						c.ParentID == nil &&
						c.IsDeleted == false
				})).Return(nil)
			},
		},
		{
			name:        "empty content",
			authorID:    "author123",
			content:     "",
			targetType:  "post",
			targetID:    "post123",
			parentID:    nil,
			expectedErr: errors.New("content must be 1-2000 characters"),
			mockSetup:   func(mockRepo *postMocks.MockCommentRepository) {},
		},
		{
			name:        "repository error",
			authorID:    "author123",
			content:     "Valid content",
			targetType:  "post",
			targetID:    "post123",
			parentID:    nil,
			expectedErr: errors.New("repository error"),
			mockSetup: func(mockRepo *postMocks.MockCommentRepository) {
				mockRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := new(postMocks.MockPostRepository)
			mockCommentRepo := new(postMocks.MockCommentRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockCommentRepo)
			}

			postService := post.NewPostService(mockPostRepo, mockCommentRepo)

			result, err := postService.CreateComment(tt.authorID, tt.targetType, tt.targetID, tt.content, tt.parentID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.authorID, result.AuthorID)
				assert.Equal(t, tt.content, result.Content)
				assert.Equal(t, tt.targetType, result.TargetType)
				assert.Equal(t, tt.targetID, result.TargetID)
				assert.Equal(t, tt.parentID, result.ParentID)
				assert.False(t, result.IsDeleted)
			}

			mockCommentRepo.AssertExpectations(t)
		})
	}
}

func TestPostService_ListCommentsByTarget(t *testing.T) {
	mockPostRepo := new(postMocks.MockPostRepository)
	mockCommentRepo := new(postMocks.MockCommentRepository)

	expectedComments := []*postModel.Comment{
		{
			Base: model.Base{
				ID:        "comment1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
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
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			AuthorID:   "user2",
			Content:    "Second comment",
			TargetType: "post",
			TargetID:   "post123",
			ParentID:   nil,
			IsDeleted:  false,
		},
	}

	mockCommentRepo.On("ListByTarget", "post", "post123", 10, 0).Return(expectedComments, nil)

	postService := post.NewPostService(mockPostRepo, mockCommentRepo)

	comments, err := postService.ListCommentsByTarget("post", "post123", 10, 0)

	assert.NoError(t, err)
	assert.Len(t, comments, 2)
	assert.Equal(t, "comment1", comments[0].ID)
	assert.Equal(t, "comment2", comments[1].ID)
	assert.Equal(t, "post", comments[0].TargetType)
	assert.Equal(t, "post123", comments[0].TargetID)

	mockCommentRepo.AssertExpectations(t)
}

func TestPostService_UpdateComment(t *testing.T) {
	mockPostRepo := new(postMocks.MockPostRepository)
	mockCommentRepo := new(postMocks.MockCommentRepository)

	existingComment := &postModel.Comment{
		Base: model.Base{
			ID:        "comment123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		AuthorID:   "author123",
		Content:    "Original content",
		TargetType: "post",
		TargetID:   "post123",
		ParentID:   nil,
		IsDeleted:  false,
	}

	mockCommentRepo.On("GetByID", "comment123").Return(existingComment, nil)
	mockCommentRepo.On("Update", mock.MatchedBy(func(c *postModel.Comment) bool {
		return c.ID == "comment123" && c.Content == "Updated content"
	})).Return(nil)

	postService := post.NewPostService(mockPostRepo, mockCommentRepo)

	result, err := postService.UpdateComment("comment123", "author123", "Updated content")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "comment123", result.ID)
	assert.Equal(t, "Updated content", result.Content)

	mockCommentRepo.AssertExpectations(t)
}

func TestPostService_DeleteComment(t *testing.T) {
	mockPostRepo := new(postMocks.MockPostRepository)
	mockCommentRepo := new(postMocks.MockCommentRepository)

	existingComment := &postModel.Comment{
		Base: model.Base{
			ID:        "comment123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		AuthorID:   "author123",
		Content:    "Test comment",
		TargetType: "post",
		TargetID:   "post123",
		ParentID:   nil,
		IsDeleted:  false,
	}

	mockCommentRepo.On("GetByID", "comment123").Return(existingComment, nil)
	mockCommentRepo.On("Delete", "comment123").Return(nil)

	postService := post.NewPostService(mockPostRepo, mockCommentRepo)

	err := postService.DeleteComment("comment123", "author123")

	assert.NoError(t, err)

	mockCommentRepo.AssertExpectations(t)
}
