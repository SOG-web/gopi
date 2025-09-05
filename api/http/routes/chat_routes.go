package routes

import (
	"github.com/gin-gonic/gin"
	"gopi.com/api/http/handler"
	"gopi.com/api/http/middleware"
	"gopi.com/api/ws"
	"gopi.com/internal/app/chat"
	"gopi.com/internal/app/user"
	"gopi.com/internal/lib/jwt"
)

func RegisterChatRoutes(router *gin.Engine, chatService *chat.ChatService, userService *user.UserService, jwtService jwt.JWTServiceInterface) {
	chatHandler := handler.NewChatHandler(chatService, userService)

	// WebSocket setup
	wsManager := ws.NewWebSocketManager(chatService, userService)
	wsHandler := ws.NewChatWebSocketHandler(wsManager)

	// Start WebSocket manager in a goroutine
	go wsManager.Run()

	// Protected chat routes (require authentication)
	protectedChat := router.Group("/api/chat")
	protectedChat.Use(middleware.RequireAuth(jwtService))
	{
		// Group management routes
		protectedChat.POST("/groups", chatHandler.CreateGroup)
		protectedChat.GET("/groups", chatHandler.GetGroups)
		protectedChat.GET("/groups/:slug", chatHandler.GetGroupBySlug)
		protectedChat.PUT("/groups/:slug", chatHandler.UpdateGroup)
		protectedChat.DELETE("/groups/:slug", chatHandler.DeleteGroup)

		// Group membership routes
		protectedChat.POST("/groups/:slug/join", chatHandler.JoinGroup)
		protectedChat.POST("/groups/:slug/leave", chatHandler.LeaveGroup)

		// Message routes
		protectedChat.POST("/groups/:slug/messages", chatHandler.SendMessage)
		protectedChat.GET("/groups/:slug/messages", chatHandler.GetMessages)
	}

	// Admin-only chat routes
	adminChat := router.Group("/api/chat/admin")
	adminChat.Use(middleware.RequireAuth(jwtService))
	adminChat.Use(middleware.RequireStaff())
	{
		adminChat.GET("/groups/search", chatHandler.AdminSearchGroups)
	}

	// WebSocket routes (also protected)
	protectedWS := router.Group("/ws/chat")
	protectedWS.Use(middleware.RequireAuth(jwtService))
	{
		protectedWS.GET("/groups/:groupSlug", wsHandler.HandleWebSocket)
		// Alias route for compatibility
		protectedWS.GET("/group/:groupSlug", wsHandler.HandleWebSocket)
	}
}
