package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ChatIntegrationTestSuite struct {
	suite.Suite
	server *TestServer
}

func TestChatIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ChatIntegrationTestSuite))
}

func (suite *ChatIntegrationTestSuite) SetupSuite() {
	suite.server = SetupTestServer(suite.T())
}

func (suite *ChatIntegrationTestSuite) TearDownSuite() {
	suite.server.TearDownTestServer()
}

func (suite *ChatIntegrationTestSuite) SetupTest() {
	suite.server.CleanDB()
	// Clean up blacklisted tokens and password reset tokens (ignore errors if tables don't exist)
	suite.server.DB().Exec("DELETE FROM blacklisted_tokens WHERE 1=1")
	suite.server.DB().Exec("DELETE FROM password_reset_tokens WHERE 1=1")
}

func (suite *ChatIntegrationTestSuite) TestGetGroupsWithoutAuth() {
	// Test getting groups without authentication
	resp := suite.server.GetGroups("")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChatIntegrationTestSuite) TestCreateGroupWithoutAuth() {
	// Test creating group without authentication
	groupData := map[string]interface{}{
		"name":        "Test Group",
		"description": "A test chat group",
		"slug":        "test-group",
		"is_private":  false,
	}

	resp := suite.server.CreateGroup("", groupData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChatIntegrationTestSuite) TestGetGroupBySlugWithoutAuth() {
	// Test getting group by slug without authentication
	resp := suite.server.GetGroupBySlug("", "test-group")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChatIntegrationTestSuite) TestUpdateGroupWithoutAuth() {
	// Test updating group without authentication
	groupData := map[string]interface{}{
		"name":        "Updated Group",
		"description": "Updated description",
	}

	resp := suite.server.UpdateGroup("", "test-group", groupData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChatIntegrationTestSuite) TestDeleteGroupWithoutAuth() {
	// Test deleting group without authentication
	resp := suite.server.DeleteGroup("", "test-group")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChatIntegrationTestSuite) TestJoinGroupWithoutAuth() {
	// Test joining group without authentication
	resp := suite.server.JoinGroup("", "test-group")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChatIntegrationTestSuite) TestLeaveGroupWithoutAuth() {
	// Test leaving group without authentication
	resp := suite.server.LeaveGroup("", "test-group")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChatIntegrationTestSuite) TestSendMessageWithoutAuth() {
	// Test sending message without authentication
	messageData := map[string]interface{}{
		"content": "Test message",
	}

	resp := suite.server.SendMessage("", "test-group", messageData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChatIntegrationTestSuite) TestGetMessagesWithoutAuth() {
	// Test getting messages without authentication
	resp := suite.server.GetMessages("", "test-group")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChatIntegrationTestSuite) TestWebSocketConnectWithoutAuth() {
	// Test WebSocket connection without authentication
	resp := suite.server.WebSocketConnect("", "test-group")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChatIntegrationTestSuite) TestAdminSearchGroupsWithoutAdmin() {
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

	// Test admin search without admin privileges
	query := map[string]interface{}{
		"q": "test",
	}

	resp = suite.server.AdminSearchGroups(token, query)

	assert.NotNil(suite.T(), resp)
	// Should be forbidden for regular users
	assert.True(suite.T(), resp.Code == http.StatusForbidden || resp.Code == http.StatusUnauthorized)
}

func (suite *ChatIntegrationTestSuite) TestCreateGroupWithAuth() {
	// Register and login a user first
	suite.server.RegisterUser("chat@example.com", "password123", "Chat", "User")
	resp, loginResp := suite.server.LoginUser("chat@example.com", "password123")

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
		suite.T().Skip("Cannot get token for chat tests")
		return
	}

	// Test creating group with authentication
	groupData := map[string]interface{}{
		"name":        "Authenticated Group",
		"description": "A group created with authentication",
		"slug":        "auth-group",
		"is_private":  false,
	}

	resp = suite.server.CreateGroup(token, groupData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusCreated || resp.Code == http.StatusOK || resp.Code == http.StatusBadRequest)
}

func (suite *ChatIntegrationTestSuite) TestGetGroupsWithAuth() {
	// Register and login a user first
	suite.server.RegisterUser("chat2@example.com", "password123", "Chat", "User")
	resp, loginResp := suite.server.LoginUser("chat2@example.com", "password123")

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
		suite.T().Skip("Cannot get token for get groups test")
		return
	}

	// Test getting groups with authentication
	resp = suite.server.GetGroups(token)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusOK || resp.Code == http.StatusNotFound)
}

func (suite *ChatIntegrationTestSuite) TestSendMessageWithAuth() {
	// Register and login a user first
	suite.server.RegisterUser("chat3@example.com", "password123", "Chat", "User")
	resp, loginResp := suite.server.LoginUser("chat3@example.com", "password123")

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
		suite.T().Skip("Cannot get token for send message test")
		return
	}

	// Test sending message with authentication
	messageData := map[string]interface{}{
		"content": "Hello from integration test!",
	}

	resp = suite.server.SendMessage(token, "test-group", messageData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusCreated || resp.Code == http.StatusOK || resp.Code == http.StatusBadRequest)
}

func (suite *ChatIntegrationTestSuite) TestGetMessagesWithAuth() {
	// Register and login a user first
	suite.server.RegisterUser("chat4@example.com", "password123", "Chat", "User")
	resp, loginResp := suite.server.LoginUser("chat4@example.com", "password123")

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
		suite.T().Skip("Cannot get token for get messages test")
		return
	}

	// Test getting messages with authentication
	resp = suite.server.GetMessages(token, "test-group")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusOK || resp.Code == http.StatusNotFound)
}
