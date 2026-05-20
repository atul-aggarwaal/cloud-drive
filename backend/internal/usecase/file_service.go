package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/atul-aggarwaal/cloud-drive/internal/domain"
	"github.com/google/uuid"
)

// FileService handles the business logic for file operations.
// It orchestrates the interaction between the database (metadata)
// and the blob storage (actual bytes).
type FileService struct {
	repo    domain.FileRepository
	storage domain.BlobStorage
}

// NewFileService creates a new instance of FileService.
func NewFileService(repo domain.FileRepository, storage domain.BlobStorage) *FileService {
	return &FileService{
		repo:    repo,
		storage: storage,
	}
}

// InitiateUpload initiates an upload for a file by storing its metadata in the DB
// and providing a presigned URL to the S3 bucket to upload the actual file.
func (s *FileService) InitiateUpload(ctx context.Context, userId, fileName string, size int64) (*domain.File, string, error) {
	fileId := uuid.New().String()

	file := &domain.File{
		ID:        fileId,
		UserID:    userId,
		FileName:  fileName,
		Size:      size,
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}

	if err := s.repo.Save(ctx, file); err != nil {
		return nil, "", fmt.Errorf("metadata persistance failure: %w", err)
	}

	// Construct S3 path: user/<user_id>/<file_id>
	objectKey := fmt.Sprintf("user/%s/%s", userId, fileId)
	presignedURL, err := s.storage.GenerateUploadUrl(ctx, objectKey, 15*time.Minute)

	if err != nil {
		return nil, "", fmt.Errorf("storage provider error: %w", err)
	}

	return file, presignedURL, nil
}

// CompleteUpload marks a file upload as complete.
func (s *FileService) CompleteUpload(ctx context.Context, fileId string) error {
	log.Printf("File Service] Completing upload for file ID :%s", fileId)
	return s.repo.UpdateStatus(ctx, fileId, "AVAILABLE")
}

/*
Validates if a file is available for download and accordingly generates a presigned URL with 15 minute expiry
*/
func (s *FileService) InitiateDownload(ctx context.Context, fileId string, UserId string) (string, error) {
	file, err := s.repo.GetByID(ctx, fileId)

	if err != nil {
		return "", fmt.Errorf("fail to find file. Error: %v", err)
	}
	if file == nil {
		return "", fmt.Errorf("File is null. fail to find file. Error")
	}
	// Only allow valid user to download file
	if file.UserID != UserId {
		return "", fmt.Errorf("Unauthorized user. Error: %v", err)
	}
	if file.Status != "AVAILABLE" {
		return "", fmt.Errorf("file is not available for download")
	}

	objectKey := fmt.Sprintf("user/%s/%s", UserId, file.ID)

	downloadUrl, err := s.storage.GenerateDownloadUrl(ctx, objectKey, 15*time.Minute)

	if err != nil {
		return "", fmt.Errorf("Storage provider error %w", err)
	}

	return downloadUrl, nil
}
