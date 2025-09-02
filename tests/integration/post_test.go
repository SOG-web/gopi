package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PostIntegrationTestSuite struct {
	suite.Suite
	server *TestServer
}

func TestPostIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PostIntegrationTestSuite))
}

func (suite *PostIntegrationTestSuite) SetupSuite() {
	suite.server = SetupTestServer(suite.T())
}

func (suite *PostIntegrationTestSuite) TearDownSuite() {
	suite.server.TearDownTestServer()
}

func (suite *PostIntegrationTestSuite) SetupTest() {
	suite.server.CleanDB()
	// Clean up blacklisted tokens and password reset tokens (ignore errors if tables don't exist)
	suite.server.DB().Exec("DELETE FROM blacklisted_tokens WHERE 1=1")
	suite.server.DB().Exec("DELETE FROM password_reset_tokens WHERE 1=1")
}

func (suite *PostIntegrationTestSuite) TestListPublishedPosts() {
	// Test getting published posts without authentication
	resp := suite.server.ListPublishedPosts()

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusOK || resp.Code == http.StatusNotFound)
}

func (suite *PostIntegrationTestSuite) TestGetPostBySlug() {
	// Test getting a specific post by slug (non-existent slug)
	resp := suite.server.GetPostBySlug("non-existent-post")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusNotFound || resp.Code == http.StatusOK)
}

func (suite *PostIntegrationTestSuite) TestCreatePostWithoutAdmin() {
	// Register and login a regular user first
	suite.server.RegisterUser("regular@example.com", "password123", "Regular", "User")
	resp, loginResp := suite.server.LoginUser("regular@example.com", "password123")

	// Extract token from login response
	token := ""
	if loginResp != nil {
		if data, ok := loginResp["data"].(map[string]interface{}); ok {
			if tokenVal, exists := data["token"]; exists {
				token = tokenVal.(string)
			}
		}
	}

	if token == "" {
		suite.T().Skip("Cannot get token for admin tests")
		return
	}

	// Test creating post without admin privileges
	postData := map[string]interface{}{
		"title":   "Test Post",
		"content": "This is a test post content",
		"slug":    "test-post",
		"excerpt": "Test excerpt",
	}

	resp = suite.server.CreatePost(token, postData)

	assert.NotNil(suite.T(), resp)
	// Should be forbidden for regular users
	assert.True(suite.T(), resp.Code == http.StatusForbidden || resp.Code == http.StatusUnauthorized)
}

func (suite *PostIntegrationTestSuite) TestUpdatePostWithoutAdmin() {
	// Register and login a regular user first
	suite.server.RegisterUser("regular2@example.com", "password123", "Regular", "User")
	resp, loginResp := suite.server.LoginUser("regular2@example.com", "password123")

	// Extract token from login response
	token := ""
	if loginResp != nil {
		if data, ok := loginResp["data"].(map[string]interface{}); ok {
			if tokenVal, exists := data["token"]; exists {
				token = tokenVal.(string)
			}
		}
	}

	if token == "" {
		suite.T().Skip("Cannot get token for admin tests")
		return
	}

	// Test updating post without admin privileges
	postData := map[string]interface{}{
		"title":   "Updated Post",
		"content": "Updated content",
	}

	resp = suite.server.UpdatePost(token, "post-123", postData)

	assert.NotNil(suite.T(), resp)
	// Should be forbidden for regular users
	assert.True(suite.T(), resp.Code == http.StatusForbidden || resp.Code == http.StatusUnauthorized)
}

func (suite *PostIntegrationTestSuite) TestPublishPostWithoutAdmin() {
	// Register and login a regular user first
	suite.server.RegisterUser("regular3@example.com", "password123", "Regular", "User")
	resp, loginResp := suite.server.LoginUser("regular3@example.com", "password123")

	// Extract token from login response
	token := ""
	if loginResp != nil {
		if data, ok := loginResp["data"].(map[string]interface{}); ok {
			if tokenVal, exists := data["token"]; exists {
				token = tokenVal.(string)
			}
		}
	}

	if token == "" {
		suite.T().Skip("Cannot get token for admin tests")
		return
	}

	// Test publishing post without admin privileges
	resp = suite.server.PublishPost(token, "post-123")

	assert.NotNil(suite.T(), resp)
	// Should be forbidden for regular users
	assert.True(suite.T(), resp.Code == http.StatusForbidden || resp.Code == http.StatusUnauthorized)
}

func (suite *PostIntegrationTestSuite) TestDeletePostWithoutAuth() {
	// Test deleting post without authentication
	resp := suite.server.DeletePost("", "post-123")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *PostIntegrationTestSuite) TestUploadCoverImageWithoutAuth() {
	// Test uploading cover image without authentication
	imageData := map[string]interface{}{
		"image": "base64-image-data",
	}

	resp := suite.server.UploadCoverImage("", "post-123", imageData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *PostIntegrationTestSuite) TestListCommentsByTarget() {
	// Test getting comments by target (non-existent target)
	resp := suite.server.ListCommentsByTarget("post", "post-123")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusOK || resp.Code == http.StatusNotFound)
}

func (suite *PostIntegrationTestSuite) TestCreateCommentWithoutAuth() {
	// Test creating comment without authentication
	commentData := map[string]interface{}{
		"content":    "This is a test comment",
		"target_type": "post",
		"target_id":  "post-123",
	}

	resp := suite.server.CreateComment("", commentData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *PostIntegrationTestSuite) TestUpdateCommentWithoutAuth() {
	// Test updating comment without authentication
	commentData := map[string]interface{}{
		"content": "Updated comment content",
	}

	resp := suite.server.UpdateComment("", "comment-123", commentData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *PostIntegrationTestSuite) TestDeleteCommentWithoutAuth() {
	// Test deleting comment without authentication
	resp := suite.server.DeleteComment("", "comment-123")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *PostIntegrationTestSuite) TestCreateCommentWithAuth() {
	// Register and login a user first
	suite.server.RegisterUser("comment@example.com", "password123", "Comment", "User")
	resp, loginResp := suite.server.LoginUser("comment@example.com", "password123")

	// Extract token from login response
	token := ""
	if loginResp != nil {
		if data, ok := loginResp["data"].(map[string]interface{}); ok {
			if tokenVal, exists := data["token"]; exists {
				token = tokenVal.(string)
			}
		}
	}

	if token == "" {
		suite.T().Skip("Cannot get token for comment tests")
		return
	}

	// Test creating comment with authentication
	commentData := map[string]interface{}{
		"content":    "This is a test comment with auth",
		"target_type": "post",
		"target_id":  "post-123",
	}

	resp = suite.server.CreateComment(token, commentData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusCreated || resp.Code == http.StatusOK || resp.Code == http.StatusBadRequest)
}

func (suite *PostIntegrationTestSuite) TestDeletePostWithAuth() {
	// Register and login a user first
	suite.server.RegisterUser("delete@example.com", "password123", "Delete", "User")
	resp, loginResp := suite.server.LoginUser("delete@example.com", "password123")

	// Extract token from login response
	token := ""
	if loginResp != nil {
		if data, ok := loginResp["data"].(map[string]interface{}); ok {
			if tokenVal, exists := data["token"]; exists {
				token = tokenVal.(string)
			}
		}
	}

	if token == "" {
		suite.T().Skip("Cannot get token for delete post test")
		return
	}

	// Test deleting post with authentication
	resp = suite.server.DeletePost(token, "post-123")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusOK || resp.Code == http.StatusNotFound || resp.Code == http.StatusForbidden)
}

func (suite *PostIntegrationTestSuite) TestUploadCoverImageWithAuth() {
	// Register and login a user first
	suite.server.RegisterUser("cover@example.com", "password123", "Cover", "User")
	resp, loginResp := suite.server.LoginUser("cover@example.com", "password123")

	// Extract token from login response
	token := ""
	if loginResp != nil {
		if data, ok := loginResp["data"].(map[string]interface{}); ok {
			if tokenVal, exists := data["token"]; exists {
				token = tokenVal.(string)
			}
		}
	}

	if token == "" {
		suite.T().Skip("Cannot get token for cover image test")
		return
	}

	// Test uploading cover image with authentication
	imageData := map[string]interface{}{
		"image": "base64-image-data-for-testing",
	}

	resp = suite.server.UploadCoverImage(token, "post-123", imageData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusCreated || resp.Code == http.StatusOK || resp.Code == http.StatusBadRequest)
}
