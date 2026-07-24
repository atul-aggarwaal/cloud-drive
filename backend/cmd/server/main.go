package main

import (
	// This is the driver you just fetched.
	// The underscore registers it so database/sql can use it.

	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atul-aggarwaal/cloud-drive/internal/api/handler"
	"github.com/atul-aggarwaal/cloud-drive/internal/repository"
	"github.com/atul-aggarwaal/cloud-drive/internal/storage"
	"github.com/atul-aggarwaal/cloud-drive/internal/usecase"
	"github.com/atul-aggarwaal/cloud-drive/internal/worker"
	"github.com/aws/aws-sdk-go-v2/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

// main is the entry point of the server application.
// It sets up the database connection, AWS configuration,
// initializes the layers, and starts the HTTP server.
func main() {
	wd, _ := os.Getwd()
	log.Printf("WORKING DIR = %s", wd)

	//1. Establish top level context lifecycle listeners
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture termination events from OS container environments
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 1. Setup Postgres connection
	connStr := "host=localhost port=5432 user=drive_admin password=drive_secret dbname=cloud_drive sslmode=disable"
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal("Postgres connection error :", err)
	}
	defer db.Close()

	//Run DB Migration
	if err := repository.RunDatabaseMigration(db); err != nil {
		log.Fatal("CRITICAL: Database synchronization failed", err)
	}

	//2. setup aws config for localstack pointing to http://localhost:4566 instead of real AWS
	err = godotenv.Load("../.env")
	if err != nil {
		log.Println("No .env file found, falling back to system env variables")
	}

	// No Need to provide credentials hare. SDK will automatically read the credentials
	// from .env file loaded by "os" above and since names AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
	// are standard, it will work.
	defaultConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		log.Fatal("AWS Config error", err)
	}

	//3. initialize the layers
	//Intilize User Layer
	userRepo := repository.NewPostgresUserRepository(db)
	userService := usecase.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	//Initilize file Layer
	fileRepo := repository.NewPostresFileRepository(db)
	fileShareRepo := repository.NewPostgresFileShareRepository(db)
	blobStorage := storage.NewS3Storage(defaultConfig, os.Getenv("AWS_S3_BUCKET_NAME"))
	fileService := usecase.NewFileService(fileRepo, userRepo, blobStorage, fileShareRepo)
	fileHandler := handler.NewFileHandler(fileService)

	fileShareService := usecase.NewFileShareService(fileRepo, fileShareRepo, userRepo)
	fileShareHandler := handler.NewFileShareHandler(fileShareService)
	// 1. Initialize and boot the asynchronous event consumer loop
	queueURL := os.Getenv("AWS_SQS_QUEUE_URL")
	uploadWorker := worker.NewUploadWorker(defaultConfig, queueURL, fileService)
	// Spin processing off into a separate background thread (Goroutine)
	go uploadWorker.Start(ctx)
	
	deletionWorker := worker.NewFileDeletionWorker(fileRepo, blobStorage);
	scheduler := worker.NewScheduler(deletionWorker, time.Minute)
	go scheduler.Run(ctx)

	//4. Define routes
	// Public interface like login and new user registration don't require JWT tokens
	http.HandleFunc("/user/register", userHandler.HandleRegister)
	http.HandleFunc("/user/login", userHandler.HandleLogin)
	//Wrap secure paths with AuthInterceptor which validates a valid JWT token before allowing upload/download
	http.HandleFunc("/upload/initiate", handler.AuthInterceptor(fileHandler.HandleInitiateUpload))
	http.HandleFunc("/file/download", handler.AuthInterceptor(fileHandler.DownloadFile))
	http.HandleFunc("/file/share", handler.AuthInterceptor(fileShareHandler.NewFileShareRequest))
	http.HandleFunc("/files", handler.AuthInterceptor(fileHandler.ListFiles))
	http.HandleFunc("/file", handler.AuthInterceptor(fileHandler.DeleteFile))
	/* http.HandleFunc("/janitor", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Starting janitor")
		if err := worker.NewFileDeletionWorker(fileRepo, blobStorage).RunOnce(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("janitor completed"))
	}) */
	//5. Create an HTTP server to server incoming requests
	server := &http.Server{
		Addr:    ":8081",
		Handler: nil,
	}

	//6. Run http server asynchronously in an independent thread to avoid blocking main thread.
	go func() {
		//5. start server
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}() //calling this function with ()

	log.Println("Server is up and listening on: 8081 Press CTRL+C to stop")
	// Wait for an OS interrupt signal to stop everything gracefully
	<-sigChan

	log.Println("Shutdown signal received. Cleansing worker locks...")
	cancel() // This stops the worker loop context safely

	//7. Force  HTTP server to shudown and release port 8080
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error Shutting down server: %v", err)
	}
}
