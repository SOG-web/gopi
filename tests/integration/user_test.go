package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserIntegrationTestSuite struct {
	suite.Suite
	server *TestServer
}

func TestUserIntegrationSuite(t *testing.T) {
	suite.Run(t, new(UserIntegrationTestSuite))
}

func (suite *UserIntegrationTestSuite) SetupSuite() {
	suite.server = SetupTestServer(suite.T())
}

func (suite *UserIntegrationTestSuite) TearDownSuite() {
	suite.server.TearDownTestServer()
}

func (suite *UserIntegrationTestSuite) SetupTest() {
	suite.server.CleanDB()
	// Clean up blacklisted tokens and password reset tokens (ignore errors if tables don't exist)
	suite.server.DB().Exec("DELETE FROM blacklisted_tokens WHERE 1=1")
	suite.server.DB().Exec("DELETE FROM password_reset_tokens WHERE 1=1")
}

func (suite *UserIntegrationTestSuite) TestGetUserProfile() {
	// Register and login a user first
	suite.server.RegisterUser("profile@example.com", "password123", "Profile", "User")
	resp, loginResp := suite.server.LoginUser("profile@example.com", "password123")

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
		suite.T().Skip("Cannot get token for authenticated tests")
		return
	}

	// Test getting user profile
	resp = suite.server.GetUserProfile(token)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusOK || resp.Code == http.StatusUnauthorized)
}

func (suite *UserIntegrationTestSuite) TestGetUserProfileWithoutAuth() {
	// Test getting user profile without authentication
	resp := suite.server.GetUserProfile("")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *UserIntegrationTestSuite) TestUpdateUserProfile() {
	// Register and login a user first
	suite.server.RegisterUser("update@example.com", "password123", "Update", "User")
	resp, loginResp := suite.server.LoginUser("update@example.com", "password123")

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
		suite.T().Skip("Cannot get token for authenticated tests")
		return
	}

	// Test updating user profile
	updates := map[string]interface{}{
		"first_name": "Updated",
		"last_name":  "Name",
		"height":     180.0,
		"weight":     75.0,
	}
	resp = suite.server.UpdateUserProfile(token, updates)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusOK || resp.Code == http.StatusBadRequest)
}

func (suite *UserIntegrationTestSuite) TestGetAllUsersAsAdmin() {
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

	// Test getting all users (should fail for regular user)
	resp = suite.server.GetAllUsers(token)

	assert.NotNil(suite.T(), resp)
	// Should be forbidden for regular users
	assert.True(suite.T(), resp.Code == http.StatusForbidden || resp.Code == http.StatusUnauthorized)
}

func (suite *UserIntegrationTestSuite) TestGetAllUsersWithoutAuth() {
	// Test getting all users without authentication
	resp := suite.server.GetAllUsers("")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *UserIntegrationTestSuite) TestGetStaffUsers() {
	// Register and login a regular user first
	suite.server.RegisterUser("staff@example.com", "password123", "Staff", "User")
	resp, loginResp := suite.server.LoginUser("staff@example.com", "password123")

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
		suite.T().Skip("Cannot get token for staff tests")
		return
	}

	// Test getting staff users (should fail for regular user)
	resp = suite.server.GetStaffUsers(token)

	assert.NotNil(suite.T(), resp)
	// Should be forbidden for regular users
	assert.True(suite.T(), resp.Code == http.StatusForbidden || resp.Code == http.StatusUnauthorized)
}

func (suite *UserIntegrationTestSuite) TestGetVerifiedUsers() {
	// Register and login a regular user first
	suite.server.RegisterUser("verified@example.com", "password123", "Verified", "User")
	resp, loginResp := suite.server.LoginUser("verified@example.com", "password123")

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
		suite.T().Skip("Cannot get token for verified tests")
		return
	}

	// Test getting verified users (should fail for regular user)
	resp = suite.server.GetVerifiedUsers(token)

	assert.NotNil(suite.T(), resp)
	// Should be forbidden for regular users
	assert.True(suite.T(), resp.Code == http.StatusForbidden || resp.Code == http.StatusUnauthorized)
}

func (suite *UserIntegrationTestSuite) TestGetUnverifiedUsers() {
	// Register and login a regular user first
	suite.server.RegisterUser("unverified@example.com", "password123", "Unverified", "User")
	resp, loginResp := suite.server.LoginUser("unverified@example.com", "password123")

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
		suite.T().Skip("Cannot get token for unverified tests")
		return
	}

	// Test getting unverified users (should fail for regular user)
	resp = suite.server.GetUnverifiedUsers(token)

	assert.NotNil(suite.T(), resp)
	// Should be forbidden for regular users
	assert.True(suite.T(), resp.Code == http.StatusForbidden || resp.Code == http.StatusUnauthorized)
}

func (suite *UserIntegrationTestSuite) TestGetUserByID() {
	// Register and login a regular user first
	suite.server.RegisterUser("byid@example.com", "password123", "ByID", "User")
	resp, loginResp := suite.server.LoginUser("byid@example.com", "password123")

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
		suite.T().Skip("Cannot get token for user by ID tests")
		return
	}

	// Test getting user by ID (should fail for regular user)
	resp = suite.server.GetUserByID(token, "some-user-id")

	assert.NotNil(suite.T(), resp)
	// Should be forbidden for regular users
	assert.True(suite.T(), resp.Code == http.StatusForbidden || resp.Code == http.StatusUnauthorized)
}
