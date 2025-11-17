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

	config := cors.DefaultConfig()

	allowedOrigins := []string{
		"http://localhost:3000",  // Local development
		"https://localhost:3000", // Local HTTPS
		"https://email-mkt.netlify.app",
	}

	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		allowedOrigins = append(allowedOrigins, frontendURL)
	}

	config.AllowOrigins = allowedOrigins
	config.AllowCredentials = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	auth := r.Group("/api/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
		auth.POST("/google", handlers.GoogleAuth)
		auth.POST("/google/callback", handlers.HandleGoogleCallback)
	}

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/profile", handlers.GetProfile)

		gmail := api.Group("/gmail")
		{
			gmail.GET("/auth-url", handlers.GetGmailAuthURL)
			gmail.GET("/status", handlers.GetGmailStatus)
			gmail.DELETE("/disconnect", handlers.DisconnectGmail)
			gmail.POST("/send", handlers.SendEmail)
			gmail.POST("/process-csv", handlers.ProcessCSV)
			gmail.POST("/send-bulk", handlers.SendBulkEmails)
			gmail.GET("/history", handlers.GetEmailHistory)
			gmail.GET("/history/stats", handlers.GetEmailHistoryStats)
		}
	}

	return r
}
