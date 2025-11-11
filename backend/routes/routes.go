package routes

import (
	"email-app-backend/handlers"
	"email-app-backend/middleware"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// CORS middleware
	config := cors.DefaultConfig()

	// Allow multiple origins for development and production
	allowedOrigins := []string{
		"http://localhost:3000",  // Local development
		"https://localhost:3000", // Local HTTPS
	}

	// Add production frontend URL if set
	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		allowedOrigins = append(allowedOrigins, frontendURL)
	}

	config.AllowOrigins = allowedOrigins
	config.AllowCredentials = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	// Public routes (no authentication required)
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
		auth.POST("/google", handlers.GoogleAuth)
		auth.POST("/google/callback", handlers.HandleGoogleCallback)
	}

	// Protected routes (authentication required)
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// User profile
		api.GET("/profile", handlers.GetProfile)

		// Gmail integration
		gmail := api.Group("/gmail")
		{
			gmail.GET("/auth-url", handlers.GetGmailAuthURL)
			gmail.GET("/status", handlers.GetGmailStatus)
			gmail.POST("/send", handlers.SendEmail)
		}
	}

	return r
}
