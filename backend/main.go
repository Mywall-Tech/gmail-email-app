package main

import (
	"log"
	"os"

	"email-app-backend/config"
	"email-app-backend/routes"

	"github.com/joho/godotenv"
)

func init() {
	// Load environment variables before any other package init functions
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	} else {
		log.Println("Successfully loaded .env file")
	}
}

func main() {

	// Connect to database
	config.ConnectDatabase()

	// Setup routes
	r := routes.SetupRoutes()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
