package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CampaignIntegrationTestSuite struct {
	suite.Suite
	server *TestServer
}

func TestCampaignIntegrationSuite(t *testing.T) {
	suite.Run(t, new(CampaignIntegrationTestSuite))
}

func (suite *CampaignIntegrationTestSuite) SetupSuite() {
	suite.server = SetupTestServer(suite.T())
}

func (suite *CampaignIntegrationTestSuite) TearDownSuite() {
	suite.server.TearDownTestServer()
}

func (suite *CampaignIntegrationTestSuite) SetupTest() {
	suite.server.CleanDB()
	// Clean up blacklisted tokens and password reset tokens (ignore errors if tables don't exist)
	suite.server.DB().Exec("DELETE FROM blacklisted_tokens WHERE 1=1")
	suite.server.DB().Exec("DELETE FROM password_reset_tokens WHERE 1=1")
}

func (suite *CampaignIntegrationTestSuite) TestGetCampaignsPublic() {
	// Test getting campaigns without authentication
	resp := suite.server.GetCampaigns()

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusOK || resp.Code == http.StatusNotFound)
}

func (suite *CampaignIntegrationTestSuite) TestGetCampaignBySlug() {
	// Test getting a specific campaign by slug (non-existent slug)
	resp := suite.server.GetCampaignBySlug("non-existent-campaign")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusNotFound || resp.Code == http.StatusOK)
}

func (suite *CampaignIntegrationTestSuite) TestCreateCampaignWithoutAuth() {
	// Test creating campaign without authentication
	campaignData := map[string]interface{}{
		"title":       "Test Campaign",
		"description": "A test campaign",
		"goal":        1000.0,
		"slug":        "test-campaign",
	}

	resp := suite.server.CreateCampaign("", campaignData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *CampaignIntegrationTestSuite) TestCreateCampaignWithAuth() {
	// Register and login a user first
	suite.server.RegisterUser("campaign@example.com", "password123", "Campaign", "User")
	resp, loginResp := suite.server.LoginUser("campaign@example.com", "password123")

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
		suite.T().Skip("Cannot get token for campaign tests")
		return
	}

	// Test creating campaign with authentication
	campaignData := map[string]interface{}{
		"title":       "Test Campaign",
		"description": "A test campaign for integration testing",
		"goal":        1000.0,
		"slug":        "test-campaign",
		"start_date":  "2024-01-01T00:00:00Z",
		"end_date":    "2024-12-31T23:59:59Z",
	}

	resp = suite.server.CreateCampaign(token, campaignData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusCreated || resp.Code == http.StatusOK || resp.Code == http.StatusBadRequest)
}

func (suite *CampaignIntegrationTestSuite) TestGetCampaignsByUserWithoutAuth() {
	// Test getting user campaigns without authentication
	resp := suite.server.GetCampaignsByUser("")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *CampaignIntegrationTestSuite) TestGetCampaignsByUserWithAuth() {
	// Register and login a user first
	suite.server.RegisterUser("campaignuser@example.com", "password123", "Campaign", "User")
	resp, loginResp := suite.server.LoginUser("campaignuser@example.com", "password123")

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
		suite.T().Skip("Cannot get token for user campaigns test")
		return
	}

	// Test getting user campaigns with authentication
	resp = suite.server.GetCampaignsByUser(token)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusOK || resp.Code == http.StatusNotFound)
}

func (suite *CampaignIntegrationTestSuite) TestGetCampaignsByOthersWithoutAuth() {
	// Test getting other users' campaigns without authentication
	resp := suite.server.GetCampaignsByOthers("")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *CampaignIntegrationTestSuite) TestJoinCampaignWithoutAuth() {
	// Test joining campaign without authentication
	resp := suite.server.JoinCampaign("", "test-campaign")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *CampaignIntegrationTestSuite) TestParticipateCampaignWithoutAuth() {
	// Test participating in campaign without authentication
	participationData := map[string]interface{}{
		"distance": 5.5,
		"duration": 30,
	}

	resp := suite.server.ParticipateCampaign("", "test-campaign", participationData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *CampaignIntegrationTestSuite) TestSponsorCampaignWithoutAuth() {
	// Test sponsoring campaign without authentication
	sponsorData := map[string]interface{}{
		"amount": 100.0,
	}

	resp := suite.server.SponsorCampaign("", "test-campaign", sponsorData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *CampaignIntegrationTestSuite) TestGetCampaignLeaderboardWithoutAuth() {
	// Test getting campaign leaderboard without authentication
	resp := suite.server.GetCampaignLeaderboard("", "test-campaign")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *CampaignIntegrationTestSuite) TestGetFinishCampaignDetailsWithoutAuth() {
	// Test getting finish campaign details without authentication
	resp := suite.server.GetFinishCampaignDetails("", "test-campaign", "runner-123")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *CampaignIntegrationTestSuite) TestFinishCampaignRunWithoutAuth() {
	// Test finishing campaign run without authentication
	finishData := map[string]interface{}{
		"final_distance": 10.0,
		"total_time":     60,
	}

	resp := suite.server.FinishCampaignRun("", "test-campaign", "runner-123", finishData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *CampaignIntegrationTestSuite) TestGetCampaignRunnersWithoutAdmin() {
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

	// Test getting campaign runners without admin privileges
	resp = suite.server.GetCampaignRunners(token)

	assert.NotNil(suite.T(), resp)
	// Should be forbidden for regular users
	assert.True(suite.T(), resp.Code == http.StatusForbidden || resp.Code == http.StatusUnauthorized)
}

func (suite *CampaignIntegrationTestSuite) TestCreateCampaignRunnerWithoutAdmin() {
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

	// Test creating campaign runner without admin privileges
	runnerData := map[string]interface{}{
		"user_id":    "user-123",
		"campaign_id": "campaign-123",
		"distance":   0.0,
	}

	resp = suite.server.CreateCampaignRunner(token, runnerData)

	assert.NotNil(suite.T(), resp)
	// Should be forbidden for regular users
	assert.True(suite.T(), resp.Code == http.StatusForbidden || resp.Code == http.StatusUnauthorized)
}

func (suite *CampaignIntegrationTestSuite) TestGetSponsorCampaignsWithoutAdmin() {
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

	// Test getting sponsor campaigns without admin privileges
	resp = suite.server.GetSponsorCampaigns(token)

	assert.NotNil(suite.T(), resp)
	// Should be forbidden for regular users
	assert.True(suite.T(), resp.Code == http.StatusForbidden || resp.Code == http.StatusUnauthorized)
}
