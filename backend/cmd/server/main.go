package main

import (
	// This is the driver you just fetched.
	// The underscore registers it so database/sql can use it.

	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/atul-aggarwaal/cloud-drive/internal/api/handler"
	"github.com/atul-aggarwaal/cloud-drive/internal/repository"
	"github.com/atul-aggarwaal/cloud-drive/internal/storage"
	"github.com/atul-aggarwaal/cloud-drive/internal/usecase"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// main is the entry point of the server application.
// It sets up the database connection, AWS configuration,
// initializes the layers, and starts the HTTP server.
func main() {
	ctx := context.Background()

	// 1. Setup Postgres connection
	connStr := "host=localhost port=5432 user=drive_admin password=drive_secret dbname=cloud_drive sslmode=disable"
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal("Postgres connection error :", err)
	}
	defer db.Close()

	//2. setup aws config for localstack pointing to http://localhost:4566 instead of real AWS
	defaultConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("mock-key", "mock-secret", "")))
	if err != nil {
		log.Fatal("AWS Config error", err)
	}

	//3. initialize the layers
	fileRepo := repository.NewPostresFileRepository(db)
	blobStorage := storage.NewS3Storage(defaultConfig, "my-cloud-bucket", "http://localhost:4566")

	fileService := usecase.NewFileService(fileRepo, blobStorage)
	fileHandler := handler.NewFileHandler(fileService)

	//4. Define routes
	http.HandleFunc("/upload/initiate", fileHandler.HandleInitiateUpload)

	//5. start server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
