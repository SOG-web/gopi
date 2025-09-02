package integration

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"gopi.com/api/http/router"
	"gopi.com/config"
	campaignGorm "gopi.com/internal/data/campaign/model/gorm"
	campaignDataRepo "gopi.com/internal/data/campaign/repo"
	challengeGorm "gopi.com/internal/data/challenge/model/gorm"
	challengeDataRepo "gopi.com/internal/data/challenge/repo"
	chatGorm "gopi.com/internal/data/chat/model/gorm"
	chatDataRepo "gopi.com/internal/data/chat/repo"
	postGorm "gopi.com/internal/data/post/model/gorm"
	postDataRepo "gopi.com/internal/data/post/repo"
	userGorm "gopi.com/internal/data/user/model/gorm"
	dataRepo "gopi.com/internal/data/user/repo"
	"gopi.com/internal/db"
	"gopi.com/internal/lib/email"
	"gopi.com/internal/lib/jwt"
	"gopi.com/internal/lib/pwreset"
	"gopi.com/internal/lib/storage"
	serverPkg "gopi.com/internal/server"

	// Import services
	"gopi.com/internal/app/campaign"
	"gopi.com/internal/app/challenge"
	"gopi.com/internal/app/chat"
	postApp "gopi.com/internal/app/post"
	"gopi.com/internal/app/user"
	"gorm.io/gorm"
)

// TestServer holds the complete server setup for integration tests
type TestServer struct {
	suite.Suite
	db       *gorm.DB
	server   *serverPkg.Server
	router   *gin.Engine
	services *TestServices
	repos    *TestRepositories
}

// DB exposes the database for test cleanup
func (ts *TestServer) DB() *gorm.DB {
	return ts.db
}

// TestServices holds all service instances
type TestServices struct {
	UserService      *user.UserService
	CampaignService  *campaign.CampaignService
	ChallengeService *challenge.ChallengeService
	ChatService      *chat.ChatService
	PostService      *postApp.Service
	JWTService       jwt.JWTServiceInterface
	EmailService     *email.EmailService
	PwdResetService  pwreset.PasswordResetServiceInterface
	StorageService   storage.Storage
}

// TestRepositories holds all repository instances
type TestRepositories struct {
	UserRepo            interface{} // dataRepo.UserRepository
	CampaignRepo        interface{} // campaignDataRepo.CampaignRepository
	CampaignRunnerRepo  interface{} // campaignDataRepo.CampaignRunnerRepository
	SponsorCampaignRepo interface{} // campaignDataRepo.SponsorCampaignRepository
	ChallengeRepo       interface{} // challengeDataRepo.ChallengeRepository
	CauseRepo           interface{} // challengeDataRepo.CauseRepository
	CauseRunnerRepo     interface{} // challengeDataRepo.CauseRunnerRepository
	SponsorRepo         interface{} // challengeDataRepo.SponsorChallengeRepository
	SponsorCauseRepo    interface{} // challengeDataRepo.SponsorCauseRepository
	CauseBuyerRepo      interface{} // challengeDataRepo.CauseBuyerRepository
	GroupRepo           interface{} // chatDataRepo.GroupRepository
	MessageRepo         interface{} // chatDataRepo.MessageRepository
	PostRepo            interface{} // postDataRepo.PostRepository
	CommentRepo         interface{} // postDataRepo.CommentRepository
}

// SetupTestServer initializes the complete server setup for integration tests
func SetupTestServer(t *testing.T) *TestServer {
	ts := &TestServer{}

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup logger
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(lg)

	// Create test configuration
	cfg := config.Config{
		RunMode:        "test",
		Port:           "8081",
		DBAddress:      ":memory:", // Use in-memory SQLite for tests
		JWTSecret:      "test-jwt-secret-key-for-integration-tests",
		EmailHost:      "smtp.gmail.com",
		EmailPort:      587,
		EmailUsername:  "test@example.com",
		EmailPassword:  "test-password",
		RedisAddr:      "localhost:6379",
		RedisPassword:  "",
		RedisDB:        0,
		EmailFrom:      "test@example.com",
		UploadBaseDir:  "./test_uploads",
		StorageBackend: "local",
		PublicHost:     "http://localhost:8080",
	}

	// Initialize test database
	var err error
	ts.db, err = db.NewSqliteDb(cfg)
	if err != nil {
		panic(err) // For integration tests, we want to fail fast
	}

	// Auto migrate all models (following main.go structure)
	// User models
	err = ts.db.AutoMigrate(&userGorm.UserGORM{})
	if err != nil {
		panic(err)
	}

	// Challenge models
	challengeGormModels := []interface{}{
		&challengeGorm.Challenge{},
		&challengeGorm.Cause{},
		&challengeGorm.CauseRunner{},
		&challengeGorm.SponsorChallenge{},
		&challengeGorm.SponsorCause{},
		&challengeGorm.CauseBuyer{},
		&challengeGorm.ChallengeMember{},
		&challengeGorm.ChallengeSponsor{},
		&challengeGorm.CauseMember{},
		&challengeGorm.CauseSponsorMember{},
	}
	err = ts.db.AutoMigrate(challengeGormModels...)
	if err != nil {
		panic(err)
	}

	// Campaign models
	campaignGormModels := []interface{}{
		&campaignGorm.Campaign{},
		&campaignGorm.CampaignMember{},
		&campaignGorm.CampaignSponsor{},
		&campaignGorm.CampaignRunner{},
		&campaignGorm.SponsorCampaign{},
	}
	err = ts.db.AutoMigrate(campaignGormModels...)
	if err != nil {
		panic(err)
	}

	// Chat models
	chatGormModels := []interface{}{
		&chatGorm.Group{},
		&chatGorm.Message{},
	}
	err = ts.db.AutoMigrate(chatGormModels...)
	if err != nil {
		panic(err)
	}

	// Post and comment models
	postGormModels := []interface{}{
		&postGorm.Post{},
		&postGorm.Comment{},
	}
	err = ts.db.AutoMigrate(postGormModels...)
	if err != nil {
		panic(err)
	}

	// Initialize services with test configurations
	// Email service (will use test config that doesn't send actual emails)
	emailService := email.NewEmailService(email.EmailConfig{
		Host:     cfg.EmailHost,
		Port:     cfg.EmailPort,
		Username: cfg.EmailUsername,
		Password: cfg.EmailPassword,
		From:     cfg.EmailFrom,
	})

	// JWT service (will use database for testing)
	jwtService := jwt.NewJWTServiceFactory(
		cfg.JWTSecret,
		24*time.Hour,  // Access token expiry
		720*time.Hour, // Refresh token expiry (30 days)
		nil,           // No Redis for integration tests
		ts.db,         // Database connection
	)

	// Password reset service (will use database for testing)
	pwdResetService := pwreset.NewPasswordResetServiceFactory(
		nil,       // No Redis for integration tests
		ts.db,     // Database connection
		time.Hour, // TTL
	)

	// Initialize repositories (following main.go pattern)
	userRepo := dataRepo.NewGormUserRepository(ts.db)
	campaignRepo := campaignDataRepo.NewGormCampaignRepository(ts.db)
	campaignRunnerRepo := campaignDataRepo.NewGormCampaignRunnerRepository(ts.db)
	sponsorCampaignRepo := campaignDataRepo.NewGormSponsorCampaignRepository(ts.db)
	challengeRepo := challengeDataRepo.NewGormChallengeRepository(ts.db)
	causeRepo := challengeDataRepo.NewGormCauseRepository(ts.db)
	causeRunnerRepo := challengeDataRepo.NewGormCauseRunnerRepository(ts.db)
	sponsorRepo := challengeDataRepo.NewGormSponsorChallengeRepository(ts.db)
	sponsorCauseRepo := challengeDataRepo.NewGormSponsorCauseRepository(ts.db)
	causeBuyerRepo := challengeDataRepo.NewGormCauseBuyerRepository(ts.db)

	// Chat repositories
	groupRepo := chatDataRepo.NewGormGroupRepository(ts.db)
	messageRepo := chatDataRepo.NewGormMessageRepository(ts.db)

	// Post repositories
	postRepo := postDataRepo.NewGormPostRepository(ts.db)
	commentRepo := postDataRepo.NewGormCommentRepository(ts.db)

	// Store repositories
	ts.repos = &TestRepositories{
		UserRepo:            userRepo,
		CampaignRepo:        campaignRepo,
		CampaignRunnerRepo:  campaignRunnerRepo,
		SponsorCampaignRepo: sponsorCampaignRepo,
		ChallengeRepo:       challengeRepo,
		CauseRepo:           causeRepo,
		CauseRunnerRepo:     causeRunnerRepo,
		SponsorRepo:         sponsorRepo,
		SponsorCauseRepo:    sponsorCauseRepo,
		CauseBuyerRepo:      causeBuyerRepo,
		GroupRepo:           groupRepo,
		MessageRepo:         messageRepo,
		PostRepo:            postRepo,
		CommentRepo:         commentRepo,
	}

	// Initialize services
	userSvc := user.NewService(userRepo, emailService)
	campaignSvc := campaign.NewCampaignService(campaignRepo, campaignRunnerRepo, sponsorCampaignRepo)
	challengeSvc := challenge.NewChallengeService(challengeRepo, causeRepo, causeRunnerRepo, sponsorRepo, sponsorCauseRepo, causeBuyerRepo)
	chatSvc := chat.NewChatService(groupRepo, messageRepo)
	postSvc := postApp.NewPostService(postRepo, commentRepo)

	// Storage service

	store := storage.NewLocalStorage(cfg.UploadBaseDir, cfg.PublicHost)

	// Store services
	ts.services = &TestServices{
		UserService:      userSvc,
		CampaignService:  campaignSvc,
		ChallengeService: challengeSvc,
		ChatService:      chatSvc,
		PostService:      postSvc,
		JWTService:       jwtService,
		EmailService:     emailService,
		PwdResetService:  pwdResetService,
		StorageService:   store,
	}

	// Create dependencies
	deps := router.Dependencies{
		JWTService:       jwtService,
		UserService:      userSvc,
		CampaignService:  campaignSvc,
		ChallengeService: challengeSvc,
		ChatService:      chatSvc,
		PostService:      postSvc,
		// RedisClient:          nil, // Not needed for integration tests
		SessionMW:            nil, // We'll use JWT instead of sessions
		Storage:              store,
		PasswordResetService: pwdResetService,
		EmailService:         emailService,
		PublicHost:           cfg.PublicHost,
	}

	// Initialize server
	ts.server = serverPkg.New(cfg, deps)
	// Create the actual application router with real routes
	ts.router = router.New(deps)

	return ts
}

// TearDownTestServer cleans up the test server
func (ts *TestServer) TearDownTestServer() {
	if ts.db != nil {
		sqlDB, _ := ts.db.DB()
		sqlDB.Close()
	}

	// Clean up test files
	os.RemoveAll("./test_uploads")
}

// CleanDB cleans up the database between tests
func (ts *TestServer) CleanDB() {
	// Clean up tables (ignore errors for tables that may not exist)
	tables := []string{
		"users", "campaigns", "campaign_runners", "sponsor_campaigns", "campaign_members",
		"challenges", "challenge_members", "challenge_sponsors",
		"cause_runners", "sponsor_causes", "cause_buyers", "cause_members", "cause_sponsor_members",
		"chats", "posts", "comments",
	}

	for _, table := range tables {
		ts.db.Exec("DELETE FROM " + table + " WHERE 1=1")
	}
}

// Helper methods for making HTTP requests
func (ts *TestServer) MakeRequest(method, url string, body interface{}, authToken string) *httptest.ResponseRecorder {
	var jsonBody []byte
	if body != nil {
		jsonBody, _ = json.Marshal(body)
	}

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)
	return w
}

func (ts *TestServer) RegisterUser(email, password, firstName, lastName string) (*httptest.ResponseRecorder, map[string]interface{}) {
	userData := map[string]interface{}{
		"username":   email[:len(email)-len("@example.com")], // Extract username from email
		"email":      email,
		"password":   password,
		"first_name": firstName,
		"last_name":  lastName,
		"height":     175.0,
		"weight":     70.0,
	}

	resp := ts.MakeRequest("POST", "/api/auth/register/", userData, "")
	var responseBody map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &responseBody)

	// If registration was successful, auto-verify the user for testing
	if resp.Code == http.StatusCreated {
		// Update user to be verified in the database
		ts.DB().Exec("UPDATE users SET is_verified = true WHERE email = ?", email)
	}

	return resp, responseBody
}

func (ts *TestServer) RequestPasswordReset(email string) *httptest.ResponseRecorder {
	resetData := map[string]interface{}{
		"email": email,
	}

	return ts.MakeRequest("POST", "/api/auth/password-reset/request/", resetData, "")
}

func (ts *TestServer) ConfirmPasswordReset(email, token, newPassword string) *httptest.ResponseRecorder {
	confirmData := map[string]interface{}{
		"email":        email,
		"token":        token,
		"new_password": newPassword,
	}

	return ts.MakeRequest("POST", "/api/auth/password-reset/confirm/", confirmData, "")
}

func (ts *TestServer) GetUserProfile(token string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/api/user/profile/", nil, token)
}

func (ts *TestServer) UpdateUserProfile(token string, updates map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("PUT", "/api/user/profile/", updates, token)
}

func (ts *TestServer) UploadProfileImage(token string, imageData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/api/user/profile/image/", imageData, token)
}

func (ts *TestServer) GetAllUsers(token string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/api/user/admin/users/", nil, token)
}

func (ts *TestServer) GetStaffUsers(token string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/api/user/admin/staff/", nil, token)
}

func (ts *TestServer) GetVerifiedUsers(token string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/api/user/admin/verified/", nil, token)
}

func (ts *TestServer) GetUnverifiedUsers(token string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/api/user/admin/unverified/", nil, token)
}

func (ts *TestServer) GetUserByID(token, userID string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/api/user/admin/"+userID+"/", nil, token)
}

// Campaign helper methods
func (ts *TestServer) GetCampaigns() *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/campaigns", nil, "")
}

func (ts *TestServer) GetCampaignBySlug(slug string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/campaigns/"+slug, nil, "")
}

func (ts *TestServer) CreateCampaign(token string, campaignData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/campaigns", campaignData, token)
}

func (ts *TestServer) UpdateCampaign(token, slug string, campaignData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("PUT", "/campaigns/"+slug, campaignData, token)
}

func (ts *TestServer) DeleteCampaign(token, slug string) *httptest.ResponseRecorder {
	return ts.MakeRequest("DELETE", "/campaigns/"+slug, nil, token)
}

func (ts *TestServer) GetCampaignsByUser(token string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/campaigns/by_user", nil, token)
}

func (ts *TestServer) GetCampaignsByOthers(token string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/campaigns/by_others", nil, token)
}

func (ts *TestServer) JoinCampaign(token, slug string) *httptest.ResponseRecorder {
	return ts.MakeRequest("PUT", "/campaigns/"+slug+"/join", nil, token)
}

func (ts *TestServer) ParticipateCampaign(token, slug string, participationData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/campaigns/"+slug+"/participate", participationData, token)
}

func (ts *TestServer) SponsorCampaign(token, slug string, sponsorData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/campaigns/"+slug+"/sponsor", sponsorData, token)
}

func (ts *TestServer) GetCampaignLeaderboard(token, slug string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/campaigns/"+slug+"/leaderboard", nil, token)
}

func (ts *TestServer) GetFinishCampaignDetails(token, slug, runnerID string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/campaigns/"+slug+"/finish_campaign/"+runnerID, nil, token)
}

func (ts *TestServer) FinishCampaignRun(token, slug, runnerID string, finishData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("PUT", "/campaigns/"+slug+"/finish_campaign/"+runnerID, finishData, token)
}

// Admin campaign helper methods
func (ts *TestServer) CreateCampaignRunner(token string, runnerData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/admin/campaign-runners", runnerData, token)
}

func (ts *TestServer) GetCampaignRunners(token string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/admin/campaign-runners", nil, token)
}

func (ts *TestServer) GetCampaignRunnerByID(token, runnerID string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/admin/campaign-runners/"+runnerID, nil, token)
}

func (ts *TestServer) UpdateCampaignRunner(token, runnerID string, runnerData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("PUT", "/admin/campaign-runners/"+runnerID, runnerData, token)
}

func (ts *TestServer) DeleteCampaignRunner(token, runnerID string) *httptest.ResponseRecorder {
	return ts.MakeRequest("DELETE", "/admin/campaign-runners/"+runnerID, nil, token)
}

func (ts *TestServer) CreateSponsorCampaign(token string, sponsorData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/admin/sponsor-campaigns", sponsorData, token)
}

func (ts *TestServer) GetSponsorCampaigns(token string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/admin/sponsor-campaigns", nil, token)
}

func (ts *TestServer) GetSponsorCampaignByID(token, sponsorID string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/admin/sponsor-campaigns/"+sponsorID, nil, token)
}

func (ts *TestServer) UpdateSponsorCampaign(token, sponsorID string, sponsorData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("PUT", "/admin/sponsor-campaigns/"+sponsorID, sponsorData, token)
}

func (ts *TestServer) DeleteSponsorCampaign(token, sponsorID string) *httptest.ResponseRecorder {
	return ts.MakeRequest("DELETE", "/admin/sponsor-campaigns/"+sponsorID, nil, token)
}

// Challenge helper methods
func (ts *TestServer) GetChallenges() *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/api/v1/challenges", nil, "")
}

func (ts *TestServer) GetChallengeBySlug(slug string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/api/v1/challenges/slug/"+slug, nil, "")
}

func (ts *TestServer) GetChallengeByID(id string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/api/v1/challenges/id/"+id, nil, "")
}

func (ts *TestServer) GetLeaderboard() *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/api/v1/challenges/leaderboard", nil, "")
}

func (ts *TestServer) GetCausesByChallenge(challengeID string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/api/v1/challenges/"+challengeID+"/causes", nil, "")
}

func (ts *TestServer) CreateChallenge(token string, challengeData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/api/v1/challenges", challengeData, token)
}

func (ts *TestServer) JoinChallenge(token, challengeID string) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/api/v1/challenges/"+challengeID+"/join", nil, token)
}

func (ts *TestServer) SponsorChallenge(token string, sponsorData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/api/v1/challenges/sponsor", sponsorData, token)
}

func (ts *TestServer) GetCauseByID(id string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/api/v1/causes/"+id, nil, "")
}

func (ts *TestServer) CreateCause(token string, causeData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/api/v1/causes", causeData, token)
}

func (ts *TestServer) JoinCause(token, causeID string) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/api/v1/causes/"+causeID+"/join", nil, token)
}

func (ts *TestServer) RecordCauseActivity(token string, activityData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/api/v1/causes/activity", activityData, token)
}

func (ts *TestServer) SponsorCause(token string, sponsorData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/api/v1/causes/sponsor", sponsorData, token)
}

func (ts *TestServer) BuyCause(token string, buyData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/api/v1/causes/buy", buyData, token)
}

// Chat helper methods
func (ts *TestServer) CreateGroup(token string, groupData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/chat/groups", groupData, token)
}

func (ts *TestServer) GetGroups(token string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/chat/groups", nil, token)
}

func (ts *TestServer) GetGroupBySlug(token, slug string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/chat/groups/"+slug, nil, token)
}

func (ts *TestServer) UpdateGroup(token, slug string, groupData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("PUT", "/chat/groups/"+slug, groupData, token)
}

func (ts *TestServer) DeleteGroup(token, slug string) *httptest.ResponseRecorder {
	return ts.MakeRequest("DELETE", "/chat/groups/"+slug, nil, token)
}

func (ts *TestServer) JoinGroup(token, slug string) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/chat/groups/"+slug+"/join", nil, token)
}

func (ts *TestServer) LeaveGroup(token, slug string) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/chat/groups/"+slug+"/leave", nil, token)
}

func (ts *TestServer) SendMessage(token, slug string, messageData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/chat/groups/"+slug+"/messages", messageData, token)
}

func (ts *TestServer) GetMessages(token, slug string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/chat/groups/"+slug+"/messages", nil, token)
}

func (ts *TestServer) AdminSearchGroups(token string, query map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/chat/admin/groups/search", query, token)
}

func (ts *TestServer) WebSocketConnect(token, groupSlug string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/ws/chat/groups/"+groupSlug, nil, token)
}

// Post helper methods
func (ts *TestServer) ListPublishedPosts() *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/posts", nil, "")
}

func (ts *TestServer) GetPostBySlug(slug string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/posts/"+slug, nil, "")
}

func (ts *TestServer) CreatePost(token string, postData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/posts/admin", postData, token)
}

func (ts *TestServer) UpdatePost(token, postID string, postData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("PUT", "/posts/admin/"+postID, postData, token)
}

func (ts *TestServer) PublishPost(token, postID string) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/posts/admin/"+postID+"/publish", nil, token)
}

func (ts *TestServer) UnpublishPost(token, postID string) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/posts/admin/"+postID+"/unpublish", nil, token)
}

func (ts *TestServer) DeletePost(token, postID string) *httptest.ResponseRecorder {
	return ts.MakeRequest("DELETE", "/posts/"+postID, nil, token)
}

func (ts *TestServer) UploadCoverImage(token, postID string, imageData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/posts/"+postID+"/cover", imageData, token)
}

func (ts *TestServer) ListCommentsByTarget(targetType, targetID string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", "/comments/"+targetType+"/"+targetID, nil, "")
}

func (ts *TestServer) CreateComment(token string, commentData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", "/comments", commentData, token)
}

func (ts *TestServer) UpdateComment(token, commentID string, commentData map[string]interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("PUT", "/comments/"+commentID, commentData, token)
}

func (ts *TestServer) DeleteComment(token, commentID string) *httptest.ResponseRecorder {
	return ts.MakeRequest("DELETE", "/comments/"+commentID, nil, token)
}

func (ts *TestServer) LoginUser(email, password string) (*httptest.ResponseRecorder, map[string]interface{}) {
	loginData := map[string]interface{}{
		"email":    email,
		"password": password,
	}

	resp := ts.MakeRequest("POST", "/api/auth/login/", loginData, "")
	var responseBody map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &responseBody)

	return resp, responseBody
}

func (ts *TestServer) VerifyOTP(email, otp string) *httptest.ResponseRecorder {
	verifyData := map[string]interface{}{
		"email": email,
		"otp":   otp,
	}

	return ts.MakeRequest("POST", "/api/auth/verify/", verifyData, "")
}

func (ts *TestServer) ExtractToken(responseBody map[string]interface{}) string {
	if data, exists := responseBody["data"].(map[string]interface{}); exists {
		if token, exists := data["token"].(string); exists {
			return token
		}
	}
	return ""
}
