package routes

import (
	"github.com/gin-gonic/gin"
	"gopi.com/api/http/handler"
	"gopi.com/api/http/middleware"
	postApp "gopi.com/internal/app/post"
	"gopi.com/internal/lib/jwt"
	"gopi.com/internal/lib/storage"
)

func RegisterPostRoutes(router *gin.Engine, postService *postApp.Service, jwtService *jwt.JWTService, st storage.Storage) {
	postHandler := handler.NewPostHandler(postService, st)

	// Public post routes
	public := router.Group("/posts")
	{
		public.GET("", postHandler.ListPublishedPosts)
		public.GET("/:slug", postHandler.GetPostBySlug)
	}

	// Admin-only post routes
	admin := router.Group("/posts/admin")
	admin.Use(middleware.RequireAuth(jwtService))
	admin.Use(middleware.RequireStaff())
	{
		admin.POST("", postHandler.CreatePost)
		admin.PUT("/:id", postHandler.UpdatePost)
		admin.POST("/:id/publish", postHandler.PublishPost)
		admin.POST("/:id/unpublish", postHandler.UnpublishPost)
	}

	// Authenticated post routes (non-staff)
	auth := router.Group("/posts")
	auth.Use(middleware.RequireAuth(jwtService))
	{
		auth.DELETE("/:id", postHandler.DeletePost)
		auth.POST("/:id/cover", postHandler.UploadCoverImage)
	}

	// Comments
	comments := router.Group("/comments")
	{
		// Public read
		comments.GET("/:targetType/:targetID", postHandler.ListCommentsByTarget)
	}

	// Auth required for write operations
	commentsAuth := router.Group("/comments")
	commentsAuth.Use(middleware.RequireAuth(jwtService))
	{
		commentsAuth.POST("", postHandler.CreateComment)
		commentsAuth.PUT("/:id", postHandler.UpdateComment)
		commentsAuth.DELETE("/:id", postHandler.DeleteComment)
	}
}
