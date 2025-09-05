package routes

import (
	"github.com/gin-gonic/gin"
	"gopi.com/api/http/handler"
	"gopi.com/api/http/middleware"
	"gopi.com/internal/app/challenge"
	"gopi.com/internal/app/user"
	"gopi.com/internal/lib/jwt"
)

func RegisterChallengeRoutes(router *gin.Engine, challengeService *challenge.ChallengeService, userService *user.UserService, jwtService jwt.JWTServiceInterface) {
	challengeHandler := handler.NewChallengeHandler(challengeService, userService)

	api := router.Group("/api")

	// Challenge routes
	challenges := api.Group("/challenges")
	{
		challenges.GET("", challengeHandler.GetChallenges)
		challenges.GET("/slug/:slug", challengeHandler.GetChallengeBySlug)
		challenges.GET("/leaderboard", challengeHandler.GetLeaderboard)

		// Challenge-specific cause routes
		challenges.GET("/:challenge_id/causes", challengeHandler.GetCausesByChallenge)
		challenges.GET("/id/:id", challengeHandler.GetChallengeByID)

	}

	// Protected challenge routes
	protectedChallenges := api.Group("/challenges")
	protectedChallenges.Use(middleware.RequireAuth(jwtService))
	{
		protectedChallenges.POST("", challengeHandler.CreateChallenge)
		protectedChallenges.POST("/:id/join", challengeHandler.JoinChallenge)
		protectedChallenges.POST("/sponsor", challengeHandler.SponsorChallenge)
	}

	// Cause routes
	causes := api.Group("/causes")
	{
		causes.GET("/:id", challengeHandler.GetCauseByID)
	}

	// Protected cause routes
	protectedCauses := api.Group("/causes")
	protectedCauses.Use(middleware.RequireAuth(jwtService))
	{
		protectedCauses.POST("", challengeHandler.CreateCause)
		protectedCauses.POST("/:id/join", challengeHandler.JoinCause)
		protectedCauses.POST("/activity", challengeHandler.RecordCauseActivity)
		protectedCauses.POST("/sponsor", challengeHandler.SponsorCause)
		protectedCauses.POST("/buy", challengeHandler.BuyCause)
	}
}
