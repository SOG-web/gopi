package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string                 `json:"status"`
	Time   string                 `json:"time"`
	Checks map[string]interface{} `json:"checks"`
}

// Health endpoint for health checks
// @Summary Health Check
// @Description Health check endpoint to verify service status including Redis connection
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "Service is healthy"
// @Failure 503 {object} HealthResponse "Service is unhealthy"
// @Router /health [get]
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}

// HealthWithRedis creates a health check handler that includes Redis status
func HealthWithRedis(redisClient *redis.Client) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		checks := make(map[string]interface{})
		overall := "healthy"
		statusCode := http.StatusOK

		// Check Redis connection
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, err := redisClient.Ping(ctx).Result()
		if err != nil {
			checks["redis"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			overall = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		} else {
			checks["redis"] = map[string]interface{}{
				"status": "healthy",
			}
		}

		response := HealthResponse{
			Status: overall,
			Time:   time.Now().Format(time.RFC3339),
			Checks: checks,
		}

		c.JSON(statusCode, response)
	})
}
