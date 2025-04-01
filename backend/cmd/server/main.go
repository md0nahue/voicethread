package main

import (
	"log"
	"os"

	"voicethread/internal/database"
	"voicethread/internal/server"
	"voicethread/internal/storage"
)

func main() {
	// Initialize database connection
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize storage
	storageType := os.Getenv("STORAGE_TYPE")
	var store storage.Storage

	switch storageType {
	case "s3":
		accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		region := os.Getenv("AWS_REGION")
		bucket := os.Getenv("AWS_BUCKET_NAME")

		if accessKey == "" || secretKey == "" || region == "" || bucket == "" {
			log.Println("Warning: S3 credentials not provided, falling back to local storage")
			store = storage.NewLocalStorage("storage")
		} else {
			store = storage.NewS3Storage(accessKey, secretKey, region, bucket)
		}
	default:
		store = storage.NewLocalStorage("storage")
	}

	// Start server
	srv := server.New(store)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
