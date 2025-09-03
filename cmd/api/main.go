// Package main bootstraps the House of Rou API server.
//
// @title           Gopi Backend API
// @version         1.0
// @description     Complete user management API with authentication, admin features, and Django equivalent functionality.
// @schemes         http https
// @host            localhost:8080
// @BasePath        /api
//
// @securityDefinitions.apikey Bearer
// @in              header
// @name            Authorization
// @description     Enter JWT Bearer token in the format: Bearer {token}
//
// @securityDefinitions.apikey Session
// @in              cookie
// @name            hor_session
package main

import (
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"gopi.com/api/http/router"
	"gopi.com/config"
	docs "gopi.com/docs"
	"gopi.com/internal/app/campaign"
	"gopi.com/internal/app/challenge"
	"gopi.com/internal/app/chat"
	postApp "gopi.com/internal/app/post"
	"gopi.com/internal/app/user"
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
	jwtLib "gopi.com/internal/lib/jwt"
	"gopi.com/internal/lib/pwreset"
	pwresetGorm "gopi.com/internal/lib/pwreset"
	"gopi.com/internal/lib/storage"
	"gopi.com/internal/logger"
	"gopi.com/internal/server"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Envs

	// Logger
	lg := logger.New(cfg)
	slog.SetDefault(lg)

	// Configure swagger metadata at runtime
	docs.SwaggerInfo.BasePath = "/v1"

	slog.Info("creating db", "cfg", cfg)
	var gdb *gorm.DB
	var err error
	if cfg.DBDriver == "sqlite" {
		gdb, err = db.NewSqliteDb(cfg)
	} else {
		gdb, err = db.NewMysqlDb(cfg)
	}
	if err != nil {
		slog.Error("db error", "err", err)
		return
	}
	slog.Info("db created")

	slog.Info("migrating db")
	// User models
	if err := gdb.AutoMigrate(&userGorm.UserGORM{}); err != nil {
		slog.Error("user migrate error", "err", err)
		return
	}

	// Challenge models (to be added once imports are fixed)
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
	if err := gdb.AutoMigrate(challengeGormModels...); err != nil {
		slog.Error("challenge migrate error", "err", err)
		return
	}

	// campaign models
	campaignGormModels := []interface{}{
		&campaignGorm.Campaign{},
		&campaignGorm.CampaignMember{},
		&campaignGorm.CampaignSponsor{},
		&campaignGorm.CampaignRunner{},
		&campaignGorm.SponsorCampaign{},
	}
	if err := gdb.AutoMigrate(campaignGormModels...); err != nil {
		slog.Error("campaign migrate error", "err", err)
		return
	}

	// chat models
	chatGormModels := []interface{}{
		&chatGorm.Group{},
		&chatGorm.Message{},
	}
	if err := gdb.AutoMigrate(chatGormModels...); err != nil {
		slog.Error("chat migrate error", "err", err)
		return
	}

	// post and comment models
	postGormModels := []interface{}{
		&postGorm.Post{},
		&postGorm.Comment{},
	}
	if err := gdb.AutoMigrate(postGormModels...); err != nil {
		slog.Error("post migrate error", "err", err)
		return
	}

	// JWT and Password Reset models (only if using database implementations)
	if cfg.UseDatabaseJWT || cfg.UseDatabasePWReset {
		serviceModels := []interface{}{}
		if cfg.UseDatabaseJWT {
			serviceModels = append(serviceModels, &jwtLib.BlacklistedToken{})
		}
		if cfg.UseDatabasePWReset {
			serviceModels = append(serviceModels, &pwresetGorm.PasswordResetToken{})
		}
		if len(serviceModels) > 0 {
			if err := gdb.AutoMigrate(serviceModels...); err != nil {
				slog.Error("service models migrate error", "err", err)
				return
			}
			slog.Info("service models migrated")
		}
	}

	slog.Info("db migrated")

	slog.Info("creating services")
	// Email service configuration
	emailConfig := email.EmailConfig{
		Host:     cfg.EmailHost,
		Port:     cfg.EmailPort,
		Username: cfg.EmailUsername,
		Password: cfg.EmailPassword,
		From:     cfg.EmailFrom,
	}

	// Email service (using factory pattern)
	var emailService email.EmailServiceInterface
	emailService, err = email.NewEmailServiceFactory(
		emailConfig,
		cfg.UseLocalEmail,
		cfg.EmailLogPath, // Configurable log file path
	)
	if err != nil {
		slog.Error("failed to create email service", "err", err)
		return
	}

	if cfg.UseLocalEmail {
		slog.Info("using local email service - emails will be logged to ./logs/emails.log")
	} else {
		slog.Info("using production email service")
	}

	// Redis configuration (only if needed)
	var redisClient *redis.Client
	if !cfg.UseDatabaseJWT || !cfg.UseDatabasePWReset {
		slog.Info("connecting to redis")
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
		slog.Info("redis connected")
	}

	// JWT service configuration (using factory)
	jwtService := jwtLib.NewJWTServiceFactory(
		cfg.JWTSecret,
		24*time.Hour,  // Access token expiry
		720*time.Hour, // Refresh token expiry (30 days)
		redisClient,   // Redis client for token blacklisting (nil if using database)
		gdb,           // Database connection
	)
	slog.Info("jwt service created")

	// Password reset service (using factory)
	pwResetService := pwreset.NewPasswordResetServiceFactory(
		redisClient, // Redis client (nil if using database)
		gdb,         // Database connection
		time.Hour,   // TTL
	)

	slog.Info("creating repos")
	userRepo := dataRepo.NewGormUserRepository(gdb)
	campaignRepo := campaignDataRepo.NewGormCampaignRepository(gdb)
	campaignRunnerRepo := campaignDataRepo.NewGormCampaignRunnerRepository(gdb)
	campaignSponRepo := campaignDataRepo.NewGormSponsorCampaignRepository(gdb)
	challengeRepo := challengeDataRepo.NewGormChallengeRepository(gdb)
	causeRepo := challengeDataRepo.NewGormCauseRepository(gdb)
	causeRunnerRepo := challengeDataRepo.NewGormCauseRunnerRepository(gdb)
	sponsorRepo := challengeDataRepo.NewGormSponsorChallengeRepository(gdb)
	sponsorCauseRepo := challengeDataRepo.NewGormSponsorCauseRepository(gdb)
	causeBuyerRepo := challengeDataRepo.NewGormCauseBuyerRepository(gdb)

	// Chat repositories
	groupRepo := chatDataRepo.NewGormGroupRepository(gdb)
	messageRepo := chatDataRepo.NewGormMessageRepository(gdb)
	// Post repositories
	postRepo := postDataRepo.NewGormPostRepository(gdb)
	commentRepo := postDataRepo.NewGormCommentRepository(gdb)
	slog.Info("repos created")

	userSvc := user.NewUserService(userRepo, emailService)
	campaignSvc := campaign.NewCampaignService(campaignRepo, campaignRunnerRepo, campaignSponRepo)
	challengeSvc := challenge.NewChallengeService(challengeRepo, causeRepo, causeRunnerRepo, sponsorRepo, sponsorCauseRepo, causeBuyerRepo)
	chatSvc := chat.NewChatService(groupRepo, messageRepo)
	postSvc := postApp.NewPostService(postRepo, commentRepo)
	slog.Info("services created")

	// Storage initialization
	var store storage.Storage
	switch cfg.StorageBackend {
	case "s3":
		slog.Info("initializing s3 storage")
		s3Store, err := storage.NewS3Storage(storage.Config{
			Backend:           "s3",
			S3Endpoint:        cfg.S3Endpoint,
			S3Region:          cfg.S3Region,
			S3Bucket:          cfg.S3Bucket,
			S3AccessKeyID:     cfg.S3AccessKeyID,
			S3SecretAccessKey: cfg.S3SecretAccessKey,
			S3UseSSL:          cfg.S3UseSSL,
			S3ForcePathStyle:  cfg.S3ForcePathStyle,
			S3PublicBaseURL:   cfg.S3PublicBaseURL,
		})
		if err != nil {
			slog.Error("failed to init s3 storage, aborting", "err", err)
			return
		}
		store = s3Store
	default:
		slog.Info("initializing local storage")
		store = storage.NewLocalStorage(cfg.UploadBaseDir, cfg.UploadPublicBaseURL)
	}

	slog.Info("creating handlers")
	slog.Info("handlers created")

	slog.Info("creating server")
	deps := router.Dependencies{
		JWTService:           jwtService,
		UserService:          userSvc,
		CampaignService:      campaignSvc,
		ChallengeService:     challengeSvc,
		ChatService:          chatSvc,
		PostService:          postSvc,
		RedisClient:          redisClient,
		SessionMW:            nil, // We'll use JWT instead of sessions
		Storage:              store,
		PasswordResetService: pwResetService,
		EmailService:         emailService,
		PublicHost:           cfg.PublicHost,
	}
	srv := server.New(cfg, deps)
	slog.Info("server created")

	slog.Info("running server")
	if err := srv.Run(); err != nil {
		slog.Error("server error", "err", err)
		return
	}
}
