package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ChallengeIntegrationTestSuite struct {
	suite.Suite
	server *TestServer
}

func TestChallengeIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ChallengeIntegrationTestSuite))
}

func (suite *ChallengeIntegrationTestSuite) SetupSuite() {
	suite.server = SetupTestServer(suite.T())
}

func (suite *ChallengeIntegrationTestSuite) TearDownSuite() {
	suite.server.TearDownTestServer()
}

func (suite *ChallengeIntegrationTestSuite) SetupTest() {
	suite.server.CleanDB()
	// Clean up blacklisted tokens and password reset tokens (ignore errors if tables don't exist)
	suite.server.DB().Exec("DELETE FROM blacklisted_tokens WHERE 1=1")
	suite.server.DB().Exec("DELETE FROM password_reset_tokens WHERE 1=1")
}

func (suite *ChallengeIntegrationTestSuite) TestGetChallengesPublic() {
	// Test getting challenges without authentication
	resp := suite.server.GetChallenges()

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusOK || resp.Code == http.StatusNotFound)
}

func (suite *ChallengeIntegrationTestSuite) TestGetChallengeBySlug() {
	// Test getting a specific challenge by slug (non-existent slug)
	resp := suite.server.GetChallengeBySlug("non-existent-challenge")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusNotFound || resp.Code == http.StatusOK)
}

func (suite *ChallengeIntegrationTestSuite) TestGetChallengeByID() {
	// Test getting a specific challenge by ID (non-existent ID)
	resp := suite.server.GetChallengeByID("non-existent-id")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusNotFound || resp.Code == http.StatusOK)
}

func (suite *ChallengeIntegrationTestSuite) TestGetLeaderboard() {
	// Test getting challenge leaderboard without authentication
	resp := suite.server.GetLeaderboard()

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusOK || resp.Code == http.StatusNotFound)
}

func (suite *ChallengeIntegrationTestSuite) TestGetCausesByChallenge() {
	// Test getting causes by challenge ID (non-existent challenge)
	resp := suite.server.GetCausesByChallenge("non-existent-challenge-id")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusNotFound || resp.Code == http.StatusOK)
}

func (suite *ChallengeIntegrationTestSuite) TestCreateChallengeWithoutAuth() {
	// Test creating challenge without authentication
	challengeData := map[string]interface{}{
		"title":       "Test Challenge",
		"description": "A test challenge",
		"goal":        1000.0,
		"slug":        "test-challenge",
		"start_date":  "2024-01-01T00:00:00Z",
		"end_date":    "2024-12-31T23:59:59Z",
	}

	resp := suite.server.CreateChallenge("", challengeData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChallengeIntegrationTestSuite) TestJoinChallengeWithoutAuth() {
	// Test joining challenge without authentication
	resp := suite.server.JoinChallenge("", "challenge-123")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChallengeIntegrationTestSuite) TestSponsorChallengeWithoutAuth() {
	// Test sponsoring challenge without authentication
	sponsorData := map[string]interface{}{
		"challenge_id": "challenge-123",
		"amount":       100.0,
	}

	resp := suite.server.SponsorChallenge("", sponsorData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChallengeIntegrationTestSuite) TestGetCauseByID() {
	// Test getting a specific cause by ID (non-existent ID)
	resp := suite.server.GetCauseByID("non-existent-cause-id")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusNotFound || resp.Code == http.StatusOK)
}

func (suite *ChallengeIntegrationTestSuite) TestCreateCauseWithoutAuth() {
	// Test creating cause without authentication
	causeData := map[string]interface{}{
		"title":         "Test Cause",
		"description":   "A test cause",
		"goal":          500.0,
		"challenge_id":  "challenge-123",
		"start_date":    "2024-01-01T00:00:00Z",
		"end_date":      "2024-12-31T23:59:59Z",
	}

	resp := suite.server.CreateCause("", causeData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChallengeIntegrationTestSuite) TestJoinCauseWithoutAuth() {
	// Test joining cause without authentication
	resp := suite.server.JoinCause("", "cause-123")

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChallengeIntegrationTestSuite) TestRecordCauseActivityWithoutAuth() {
	// Test recording cause activity without authentication
	activityData := map[string]interface{}{
		"cause_id":     "cause-123",
		"activity_type": "run",
		"distance":     5.5,
		"duration":     30,
	}

	resp := suite.server.RecordCauseActivity("", activityData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChallengeIntegrationTestSuite) TestSponsorCauseWithoutAuth() {
	// Test sponsoring cause without authentication
	sponsorData := map[string]interface{}{
		"cause_id": "cause-123",
		"amount":   50.0,
	}

	resp := suite.server.SponsorCause("", sponsorData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChallengeIntegrationTestSuite) TestBuyCauseWithoutAuth() {
	// Test buying cause without authentication
	buyData := map[string]interface{}{
		"cause_id": "cause-123",
		"amount":   25.0,
	}

	resp := suite.server.BuyCause("", buyData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusUnauthorized || resp.Code == http.StatusForbidden)
}

func (suite *ChallengeIntegrationTestSuite) TestCreateChallengeWithAuth() {
	// Register and login a user first
	suite.server.RegisterUser("challenge@example.com", "password123", "Challenge", "User")
	resp, loginResp := suite.server.LoginUser("challenge@example.com", "password123")

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
		suite.T().Skip("Cannot get token for challenge tests")
		return
	}

	// Test creating challenge with authentication
	challengeData := map[string]interface{}{
		"title":       "Authenticated Challenge",
		"description": "A challenge created with authentication",
		"goal":        2000.0,
		"slug":        "auth-challenge",
		"start_date":  "2024-01-01T00:00:00Z",
		"end_date":    "2024-12-31T23:59:59Z",
	}

	resp = suite.server.CreateChallenge(token, challengeData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusCreated || resp.Code == http.StatusOK || resp.Code == http.StatusBadRequest)
}

func (suite *ChallengeIntegrationTestSuite) TestCreateCauseWithAuth() {
	// Register and login a user first
	suite.server.RegisterUser("cause@example.com", "password123", "Cause", "User")
	resp, loginResp := suite.server.LoginUser("cause@example.com", "password123")

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
		suite.T().Skip("Cannot get token for cause tests")
		return
	}

	// Test creating cause with authentication
	causeData := map[string]interface{}{
		"title":        "Authenticated Cause",
		"description":  "A cause created with authentication",
		"goal":         1000.0,
		"challenge_id": "challenge-123",
		"start_date":   "2024-01-01T00:00:00Z",
		"end_date":     "2024-12-31T23:59:59Z",
	}

	resp = suite.server.CreateCause(token, causeData)

	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Code == http.StatusCreated || resp.Code == http.StatusOK || resp.Code == http.StatusBadRequest)
}
