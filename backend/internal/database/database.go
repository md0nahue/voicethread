package database

import (
	"fmt"
	"os"

	"voicethread/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.Recording{}, &models.AudioChunk{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	DB = db
	return nil
}
