package config

import (
	"fmt"
	"log"
	"os"

	"email-app-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	var dsn string

	// Check if DATABASE_URL is provided (common in production)
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		dsn = databaseURL
	} else {
		// Fallback to individual environment variables
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")

		sslmode := "disable"
		if os.Getenv("GIN_MODE") == "release" {
			sslmode = "require" // Use SSL in production
		}

		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate the schema
	err = database.AutoMigrate(&models.User{}, &models.GmailToken{}, &models.EmailHistory{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	DB = database
	log.Println("Database connected successfully")
}
