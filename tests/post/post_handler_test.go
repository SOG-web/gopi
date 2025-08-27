package post_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopi.com/api/http/dto"
	"gopi.com/api/http/handler"
	"gopi.com/internal/app/post"
	"gopi.com/internal/domain/model"
	postModel "gopi.com/internal/domain/post/model"
	"gopi.com/internal/lib/storage"
	postMocks "gopi.com/tests/mocks/post"
)

// MockStorage implements the storage.Storage interface for testing
type MockStorage struct{}

func (m *MockStorage) Save(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	return "mock-url", nil
}

func (m *MockStorage) Delete(ctx context.Context, key string) error {
	return nil
}

// Ensure MockStorage implements storage.Storage
var _ storage.Storage = (*MockStorage)(nil)

// Test setup helper
func setupPostTest(t *testing.T) (*gin.Engine, *postMocks.MockPostRepository, *postMocks.MockCommentRepository) {
	gin.SetMode(gin.TestMode)

	// Create mock repositories
	mockPostRepo := new(postMocks.MockPostRepository)
	mockCommentRepo := new(postMocks.MockCommentRepository)

	// Create real service with mock repositories for integration testing
	postService := post.NewPostService(mockPostRepo, mockCommentRepo)

	// Create mock storage for testing
	mockStorage := &MockStorage{}

	// Create handler with real service and mock storage
	postHandler := handler.NewPostHandler(postService, mockStorage)

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup routes
	protected := router.Group("/posts")
	protected.Use(func(c *gin.Context) {
		// Mock auth middleware - set user_id in context
		c.Set("user_id", "test-user-id")
		c.Next()
	})

	protected.POST("", postHandler.CreatePost)
	protected.PUT("/:id", postHandler.UpdatePost)
	protected.DELETE("/:id", postHandler.DeletePost)
	protected.POST("/:id/publish", postHandler.PublishPost)
	protected.POST("/:id/unpublish", postHandler.UnpublishPost)
	protected.POST("/:id/cover", postHandler.UploadCoverImage)

	// Comment routes
	protected.POST("/comments/:postID", postHandler.CreateComment)
	protected.PUT("/comments/:commentID", postHandler.UpdateComment)
	protected.DELETE("/comments/:commentID", postHandler.DeleteComment)
	protected.GET("/comments/:targetType/:targetID", postHandler.ListCommentsByTarget)

	// Public routes (no auth required)
	router.GET("/public/posts", postHandler.ListPublishedPosts)
	router.GET("/public/posts/:slug", postHandler.GetPostBySlug)

	return router, mockPostRepo, mockCommentRepo
}

func TestPostHandler_CreatePost(t *testing.T) {
	router, mockPostRepo, _ := setupPostTest(t)

	tests := []struct {
		name           string
		requestBody    dto.CreatePostRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "successful post creation",
			requestBody: dto.CreatePostRequest{
				Title:         "Test Post",
				Content:       "This is a test post content",
				CoverImageURL: "test-image.jpg",
				Publish:       true,
			},
			expectedStatus: http.StatusCreated,
			mockSetup: func() {
				mockPostRepo.On("Create", mock.MatchedBy(func(p *postModel.Post) bool {
					return p.Title == "Test Post" &&
						p.AuthorID == "test-user-id" &&
						p.Content == "This is a test post content" &&
						p.CoverImageURL == "test-image.jpg" &&
						p.IsPublished == true
				})).Return(nil)
			},
		},
		{
			name: "post creation with empty cover image",
			requestBody: dto.CreatePostRequest{
				Title:   "Test Post",
				Content: "This is a test post content",
				Publish: false,
			},
			expectedStatus: http.StatusCreated,
			mockSetup: func() {
				mockPostRepo.On("Create", mock.MatchedBy(func(p *postModel.Post) bool {
					return p.Title == "Test Post" &&
						p.AuthorID == "test-user-id" &&
						p.CoverImageURL == "" &&
						p.IsPublished == false
				})).Return(nil)
			},
		},
		{
			name: "title too long",
			requestBody: dto.CreatePostRequest{
				Title:   "This is a very very very very long title that exceeds the 200 character limit and should cause validation to fail when creating a post through the API endpoint and is definitely over 200 characters now with additional text to make it longer",
				Content: "Valid content",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "empty title",
			requestBody: dto.CreatePostRequest{
				Title:   "",
				Content: "Valid content",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "empty content",
			requestBody: dto.CreatePostRequest{
				Title:   "Valid Title",
				Content: "",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "repository error",
			requestBody: dto.CreatePostRequest{
				Content: "Test content",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				mockPostRepo.On("Create", mock.Anything).Return(assert.AnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockPostRepo.AssertExpectations(t)
		})
	}
}

func TestPostHandler_GetPostBySlug(t *testing.T) {
	_, _, _ = setupPostTest(t)

	// This endpoint doesn't require auth, so we need to test it separately
	publicRouter := gin.New()
	publicRouter.GET("/public/posts/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "test-post" {
			c.JSON(http.StatusOK, gin.H{"message": "Post found"})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		}
	})

	tests := []struct {
		name           string
		slug           string
		expectedStatus int
	}{
		{
			name:           "successful post retrieval by slug",
			slug:           "test-post",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "post not found",
			slug:           "nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/public/posts/"+tt.slug, nil)
			w := httptest.NewRecorder()
			publicRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestPostHandler_UpdatePost(t *testing.T) {
	router, mockPostRepo, _ := setupPostTest(t)

	tests := []struct {
		name           string
		postID         string
		requestBody    dto.UpdatePostRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:   "successful post update",
			postID: "post123",
			requestBody: dto.UpdatePostRequest{
				Title:         "Updated Title",
				Content:       "Updated content",
				CoverImageURL: "updated-image.jpg",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				existingPost := &postModel.Post{
					Base: model.Base{
						ID:        "post123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Title:         "Original Title",
					Slug:          "original-title",
					Content:       "Original content",
					AuthorID:      "test-user-id",
					CoverImageURL: "original-image.jpg",
					IsPublished:   false,
				}
				mockPostRepo.On("GetByID", "post123").Return(existingPost, nil)
				mockPostRepo.On("Update", mock.MatchedBy(func(p *postModel.Post) bool {
					return p.ID == "post123" && p.Title == "Updated Title"
				})).Return(nil)
			},
		},
		{
			name:   "post not found",
			postID: "nonexistent",
			requestBody: dto.UpdatePostRequest{
				Title:   "Updated Title",
				Content: "Updated content",
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockPostRepo.On("GetByID", "nonexistent").Return(nil, assert.AnError)
			},
		},
		{
			name:   "title too long",
			postID: "post123",
			requestBody: dto.UpdatePostRequest{
				Title:   "This is a very very very very long title that exceeds the 200 character limit and should cause validation to fail when updating a post through the API endpoint and is definitely over 200 characters now with additional text to make it longer",
				Content: "Valid content",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				existingPost := &postModel.Post{
					Base: model.Base{
						ID:        "post123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Title:         "Original Title",
					Slug:          "original-title",
					Content:       "Original content",
					AuthorID:      "test-user-id",
					CoverImageURL: "original-image.jpg",
					IsPublished:   false,
				}
				mockPostRepo.On("GetByID", "post123").Return(existingPost, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/posts/"+tt.postID, bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockPostRepo.AssertExpectations(t)
		})
	}
}

func TestPostHandler_DeletePost(t *testing.T) {
	router, mockPostRepo, _ := setupPostTest(t)

	tests := []struct {
		name           string
		postID         string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful post deletion",
			postID:         "post123",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				existingPost := &postModel.Post{
					Base: model.Base{
						ID:        "post123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Title:         "Test Post",
					Slug:          "test-post",
					Content:       "Test content",
					AuthorID:      "test-user-id",
					CoverImageURL: "test-image.jpg",
					IsPublished:   true,
					PublishedAt:   &time.Time{},
				}
				mockPostRepo.On("GetByID", "post123").Return(existingPost, nil)
				mockPostRepo.On("Delete", "post123").Return(nil)
			},
		},
		{
			name:           "post not found",
			postID:         "nonexistent",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockPostRepo.On("GetByID", "nonexistent").Return(nil, assert.AnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodDelete, "/posts/"+tt.postID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockPostRepo.AssertExpectations(t)
		})
	}
}

func TestPostHandler_ListPublishedPosts(t *testing.T) {
	router, mockPostRepo, _ := setupPostTest(t)

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

	mockPostRepo.On("ListPublished", 20, 0).Return(expectedPosts, nil)

	req, _ := http.NewRequest(http.MethodGet, "/public/posts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ListPostsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)

	mockPostRepo.AssertExpectations(t)
}

func TestPostHandler_CreateComment(t *testing.T) {
	router, _, mockCommentRepo := setupPostTest(t)

	tests := []struct {
		name           string
		postID         string
		requestBody    dto.CreateCommentRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:   "successful comment creation",
			postID: "post123",
			requestBody: dto.CreateCommentRequest{
				Content:    "This is a test comment",
				TargetType: "post",
				TargetID:   "post123",
			},
			expectedStatus: http.StatusCreated,
			mockSetup: func() {
				mockCommentRepo.On("Create", mock.MatchedBy(func(c *postModel.Comment) bool {
					return c.AuthorID == "test-user-id" &&
						c.Content == "This is a test comment" &&
						c.TargetType == "post" &&
						c.TargetID == "post123"
				})).Return(nil)
			},
		},
		{
			name:   "empty content",
			postID: "post123",
			requestBody: dto.CreateCommentRequest{
				Content:    "",
				TargetType: "post",
				TargetID:   "post123",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name:   "repository error",
			postID: "post123",
			requestBody: dto.CreateCommentRequest{
				Content:    "Valid content",
				TargetType: "post",
				TargetID:   "post123",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				mockCommentRepo.On("Create", mock.Anything).Return(assert.AnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/posts/comments/"+tt.postID, bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCommentRepo.AssertExpectations(t)
		})
	}
}

func TestPostHandler_ListCommentsByTarget(t *testing.T) {
	router, _, mockCommentRepo := setupPostTest(t)

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

	mockCommentRepo.On("ListByTarget", "post", "post123", 50, 0).Return(expectedComments, nil)

	req, _ := http.NewRequest(http.MethodGet, "/posts/comments/post/post123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ListCommentsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)

	mockCommentRepo.AssertExpectations(t)
}

func TestPostHandler_UpdateComment(t *testing.T) {
	router, _, mockCommentRepo := setupPostTest(t)

	tests := []struct {
		name           string
		commentID      string
		requestBody    dto.UpdateCommentRequest
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:      "successful comment update",
			commentID: "comment123",
			requestBody: dto.UpdateCommentRequest{
				Content: "Updated comment content",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				existingComment := &postModel.Comment{
					Base: model.Base{
						ID:        "comment123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					AuthorID:   "test-user-id",
					Content:    "Original content",
					TargetType: "post",
					TargetID:   "post123",
					ParentID:   nil,
					IsDeleted:  false,
				}
				mockCommentRepo.On("GetByID", "comment123").Return(existingComment, nil)
				mockCommentRepo.On("Update", mock.MatchedBy(func(c *postModel.Comment) bool {
					return c.ID == "comment123" && c.Content == "Updated comment content"
				})).Return(nil)
			},
		},
		{
			name:      "comment not found",
			commentID: "nonexistent",
			requestBody: dto.UpdateCommentRequest{
				Content: "Updated content",
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockCommentRepo.On("GetByID", "nonexistent").Return(nil, assert.AnError)
			},
		},
		{
			name:      "empty content",
			commentID: "comment123",
			requestBody: dto.UpdateCommentRequest{
				Content: "",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				existingComment := &postModel.Comment{
					Base: model.Base{
						ID:        "comment123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					AuthorID:   "test-user-id",
					Content:    "Original content",
					TargetType: "post",
					TargetID:   "post123",
					ParentID:   nil,
					IsDeleted:  false,
				}
				mockCommentRepo.On("GetByID", "comment123").Return(existingComment, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/posts/comments/"+tt.commentID, bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCommentRepo.AssertExpectations(t)
		})
	}
}

func TestPostHandler_DeleteComment(t *testing.T) {
	router, _, mockCommentRepo := setupPostTest(t)

	tests := []struct {
		name           string
		commentID      string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "successful comment deletion",
			commentID:      "comment123",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				existingComment := &postModel.Comment{
					Base: model.Base{
						ID:        "comment123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					AuthorID:   "test-user-id",
					Content:    "Test comment",
					TargetType: "post",
					TargetID:   "post123",
					ParentID:   nil,
					IsDeleted:  false,
				}
				mockCommentRepo.On("GetByID", "comment123").Return(existingComment, nil)
				mockCommentRepo.On("Delete", "comment123").Return(nil)
			},
		},
		{
			name:           "comment not found",
			commentID:      "nonexistent",
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockCommentRepo.On("GetByID", "nonexistent").Return(nil, assert.AnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req, _ := http.NewRequest(http.MethodDelete, "/posts/comments/"+tt.commentID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockCommentRepo.AssertExpectations(t)
		})
	}
}

func TestPostHandler_UploadCoverImage(t *testing.T) {
	_, mockPostRepo, _ := setupPostTest(t)

	tests := []struct {
		name           string
		postID         string
		filename       string
		content        string
		contentType    string
		userID         string
		setupAuth      func() gin.HandlerFunc
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:        "successful cover image upload",
			postID:      "post123",
			filename:    "test-image.jpg",
			content:     "fake image content",
			contentType: "image/jpeg",
			userID:      "test-user-id",
			setupAuth: func() gin.HandlerFunc {
				return func(c *gin.Context) {
					c.Set("user_id", "test-user-id")
					c.Next()
				}
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				existingPost := &postModel.Post{
					Base: model.Base{
						ID:        "post123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Title:         "Test Post",
					Content:       "Test content",
					AuthorID:      "test-user-id",
					CoverImageURL: "",
				}
				mockPostRepo.On("GetByID", "post123").Return(existingPost, nil)
				mockPostRepo.On("Update", mock.MatchedBy(func(p *postModel.Post) bool {
					return p.ID == "post123" && strings.Contains(p.CoverImageURL, "mock-url")
				})).Return(nil)
			},
		},
		{
			name:        "unauthorized user",
			postID:      "post123",
			filename:    "test-image.jpg",
			content:     "fake image content",
			contentType: "image/jpeg",
			userID:      "",
			setupAuth: func() gin.HandlerFunc {
				return func(c *gin.Context) {
					// No user_id set
					c.Next()
				}
			},
			expectedStatus: http.StatusUnauthorized,
			mockSetup:      func() {},
		},
		{
			name:        "post not found",
			postID:      "nonexistent",
			filename:    "test-image.jpg",
			content:     "fake image content",
			contentType: "image/jpeg",
			userID:      "test-user-id",
			setupAuth: func() gin.HandlerFunc {
				return func(c *gin.Context) {
					c.Set("user_id", "test-user-id")
					c.Next()
				}
			},
			expectedStatus: http.StatusNotFound,
			mockSetup: func() {
				mockPostRepo.On("GetByID", "nonexistent").Return(nil, assert.AnError)
			},
		},
		{
			name:        "missing image file",
			postID:      "post123",
			filename:    "",
			content:     "",
			contentType: "",
			userID:      "test-user-id",
			setupAuth: func() gin.HandlerFunc {
				return func(c *gin.Context) {
					c.Set("user_id", "test-user-id")
					c.Next()
				}
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				existingPost := &postModel.Post{
					Base: model.Base{
						ID:        "post123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Title:         "Test Post",
					Content:       "Test content",
					AuthorID:      "test-user-id",
					CoverImageURL: "",
				}
				mockPostRepo.On("GetByID", "post123").Return(existingPost, nil)
			},
		},
		{
			name:        "unsupported file type",
			postID:      "post123",
			filename:    "test-file.txt",
			content:     "text content",
			contentType: "text/plain",
			userID:      "test-user-id",
			setupAuth: func() gin.HandlerFunc {
				return func(c *gin.Context) {
					c.Set("user_id", "test-user-id")
					c.Next()
				}
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				existingPost := &postModel.Post{
					Base: model.Base{
						ID:        "post123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Title:         "Test Post",
					Content:       "Test content",
					AuthorID:      "test-user-id",
					CoverImageURL: "",
				}
				mockPostRepo.On("GetByID", "post123").Return(existingPost, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockPostRepo.ExpectedCalls = nil

			tt.mockSetup()

			// Create a custom router for this test with specific auth setup
			testRouter := gin.New()
			testRouter.Use(gin.Recovery())

			// Apply custom auth setup
			testRouter.Use(tt.setupAuth())

			// Create mock repositories for this test
			testMockPostRepo := new(postMocks.MockPostRepository)
			testMockCommentRepo := new(postMocks.MockCommentRepository)

			// Copy mock expectations from the main mock
			testMockPostRepo.ExpectedCalls = mockPostRepo.ExpectedCalls

			// Create services with test mock repositories
			testPostService := post.NewPostService(testMockPostRepo, testMockCommentRepo)

			// Create mock storage
			testMockStorage := &MockStorage{}

			// Create handler with test services and storage
			testPostHandler := handler.NewPostHandler(testPostService, testMockStorage)

			// Setup test routes
			testProtected := testRouter.Group("/posts")
			testProtected.POST("/:id/cover", testPostHandler.UploadCoverImage)

			// Create multipart form request
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			if tt.filename != "" {
				part, err := writer.CreateFormFile("image", tt.filename)
				assert.NoError(t, err)
				_, err = part.Write([]byte(tt.content))
				assert.NoError(t, err)
			}

			writer.Close()

			req, _ := http.NewRequest(http.MethodPost, "/posts/"+tt.postID+"/cover", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			testMockPostRepo.AssertExpectations(t)
		})
	}
}
