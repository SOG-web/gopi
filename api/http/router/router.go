package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gopi.com/api/http/handler"
	"gopi.com/api/http/routes"
	"gopi.com/internal/app/campaign"
	"gopi.com/internal/app/challenge"
	"gopi.com/internal/app/chat"
	"gopi.com/internal/app/post"
	"gopi.com/internal/app/user"
	"gopi.com/internal/lib/jwt"
	"gopi.com/internal/lib/storage"
	"gopi.com/internal/lib/pwreset"
	"gopi.com/internal/lib/email"
)

type Dependencies struct {
	SessionMW        gin.HandlerFunc
	JWTService       *jwt.JWTService
	UserService      *user.UserService
	CampaignService  *campaign.CampaignService
	ChallengeService *challenge.ChallengeService
	ChatService      *chat.ChatService
	PostService      *post.Service
	RedisClient      *redis.Client
	Storage          storage.Storage
	PasswordResetService *pwreset.Service
	EmailService         *email.EmailService
	PublicHost           string
}

func New(deps Dependencies) *gin.Engine {
	r := gin.Default()

	// Limit multipart memory to 16 MiB (tunable)
	r.MaxMultipartMemory = 16 << 20

	// Session middleware (optional, for backward compatibility)
	if deps.SessionMW != nil {
		r.Use(deps.SessionMW)
	}

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Health check endpoint
	if deps.RedisClient != nil {
		r.GET("/health", handler.HealthWithRedis(deps.RedisClient))
	} else {
		r.GET("/health", handler.Health)
	}

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/doc.json")))
	r.GET("/doc.json", func(c *gin.Context) {
		c.File("docs/swagger.json")
	})

	// Serve static uploads (profile images, etc.)
	r.Static("/uploads", "./uploads")

	// Enhanced user system routes
	if deps.JWTService != nil && deps.UserService != nil {
		routes.SetupAuthRoutes(r, deps.UserService, deps.JWTService)
		routes.SetupUserRoutes(r, deps.UserService, deps.JWTService, deps.Storage)
		routes.SetupAdminRoutes(r, deps.UserService, deps.JWTService)
	}

	// Password reset routes (no auth required)
	if deps.UserService != nil && deps.PasswordResetService != nil {
		routes.SetupPasswordResetRoutes(r, deps.UserService, deps.PasswordResetService, deps.EmailService, deps.PublicHost)
	}

	// Campaign, challenge, and chat routes integration
	if deps.CampaignService != nil && deps.JWTService != nil && deps.UserService != nil {
		routes.RegisterCampaignRoutes(r, deps.CampaignService, deps.UserService, deps.JWTService)
	}
	if deps.ChallengeService != nil && deps.JWTService != nil && deps.UserService != nil {
		routes.RegisterChallengeRoutes(r, deps.ChallengeService, deps.UserService, deps.JWTService)
	}
	if deps.ChatService != nil && deps.JWTService != nil && deps.UserService != nil {
		routes.RegisterChatRoutes(r, deps.ChatService, deps.UserService, deps.JWTService)
	}

	// Posts and comments routes
	if deps.PostService != nil && deps.JWTService != nil {
		routes.RegisterPostRoutes(r, deps.PostService, deps.JWTService, deps.Storage)
	}

	return r
}
