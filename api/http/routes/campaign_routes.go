package routes

import (
	"github.com/gin-gonic/gin"
	"gopi.com/api/http/handler"
	"gopi.com/api/http/middleware"
	"gopi.com/internal/app/campaign"
	"gopi.com/internal/app/user"
	"gopi.com/internal/lib/jwt"
)

func RegisterCampaignRoutes(router *gin.Engine, campaignService *campaign.CampaignService, userService *user.UserService, jwtService *jwt.JWTService) {
	campaignHandler := handler.NewCampaignHandler(campaignService, userService)
	campaignAdminHandler := handler.NewCampaignAdminHandler(campaignService, userService)

	// Public campaign routes
	campaigns := router.Group("/campaigns")
	{
		campaigns.GET("", campaignHandler.GetCampaigns)            // tested
		campaigns.GET("/:slug", campaignHandler.GetCampaignBySlug) // tested
	}

	// Protected campaign routes (require authentication)
	protectedCampaigns := router.Group("/campaigns")
	protectedCampaigns.Use(middleware.RequireAuth(jwtService))
	{
		protectedCampaigns.POST("", campaignHandler.CreateCampaign)         // tested
		protectedCampaigns.PUT("/:slug", campaignHandler.UpdateCampaign)    // tested
		protectedCampaigns.DELETE("/:slug", campaignHandler.DeleteCampaign) // tested

		// User-specific routes
		protectedCampaigns.GET("/by_user", campaignHandler.GetCampaignsByUser)     // tested
		protectedCampaigns.GET("/by_others", campaignHandler.GetCampaignsByOthers) // tested

		// Campaign interaction routes
		protectedCampaigns.PUT("/:slug/join", campaignHandler.JoinCampaign)                // tested
		protectedCampaigns.POST("/:slug/participate", campaignHandler.ParticipateCampaign) // tested
		protectedCampaigns.POST("/:slug/sponsor", campaignHandler.SponsorCampaign)         // tested

		// Campaign info routes
		protectedCampaigns.GET("/:slug/leaderboard", campaignHandler.GetCampaignLeaderboard) // tested

		// Campaign finish routes
		protectedCampaigns.GET("/:slug/finish_campaign/:runner_id", campaignHandler.GetFinishCampaignDetails) // tested
		protectedCampaigns.PUT("/:slug/finish_campaign/:runner_id", campaignHandler.FinishCampaignRun)        // tested
	}

	// Admin routes for campaign management
	adminCampaigns := router.Group("/admin")
	adminCampaigns.Use(middleware.RequireAuth(jwtService))
	adminCampaigns.Use(middleware.RequireStaff()) // Require staff privileges for admin routes
	{
		// Campaign Runner admin routes
		adminCampaigns.POST("/campaign-runners", campaignAdminHandler.CreateCampaignRunner)
		adminCampaigns.GET("/campaign-runners", campaignAdminHandler.GetCampaignRunners)
		adminCampaigns.GET("/campaign-runners/:id", campaignAdminHandler.GetCampaignRunnerByID)
		adminCampaigns.PUT("/campaign-runners/:id", campaignAdminHandler.UpdateCampaignRunner)
		adminCampaigns.DELETE("/campaign-runners/:id", campaignAdminHandler.DeleteCampaignRunner)

		// Sponsor Campaign admin routes
		adminCampaigns.POST("/sponsor-campaigns", campaignAdminHandler.CreateSponsorCampaign)
		adminCampaigns.GET("/sponsor-campaigns", campaignAdminHandler.GetSponsorCampaigns)
		adminCampaigns.GET("/sponsor-campaigns/:id", campaignAdminHandler.GetSponsorCampaignByID)
		adminCampaigns.PUT("/sponsor-campaigns/:id", campaignAdminHandler.UpdateSponsorCampaign)
		adminCampaigns.DELETE("/sponsor-campaigns/:id", campaignAdminHandler.DeleteSponsorCampaign)
	}
}
