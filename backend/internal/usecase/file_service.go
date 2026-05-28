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
	repo     domain.FileRepository
	userRepo domain.UserRepository
	storage  domain.BlobStorage
}

// NewFileService creates a new instance of FileService.
func NewFileService(repo domain.FileRepository, userRepo domain.UserRepository, storage domain.BlobStorage) *FileService {
	return &FileService{
		repo:     repo,
		userRepo: userRepo,
		storage:  storage,
	}
}

// InitiateUpload initiates an upload for a file by storing its metadata in the DB
// and providing a presigned URL to the S3 bucket to upload the actual file.
func (s *FileService) InitiateUpload(ctx context.Context, ownerId string, fileName string, fileHash string, fileSize int64) (*domain.File, string, error) {
	fileId := uuid.New().String()

	//Create Metadata for File
	file := &domain.File{
		ID:       fileId,
		OwnerID:  ownerId,
		FileName: fileName,
		IsFolder: false,
	}
	user, err := s.userRepo.GetUserByID(ctx, ownerId)
	if user == nil {
		return nil, "", fmt.Errorf("File Owner validation failed: %w", err)
	}

	if err := s.repo.CreateFile(ctx, file); err != nil {
		return nil, "", fmt.Errorf("Error while creating file: %w", err)
	}

	//Create Metadata for FileVersion
	fileVersion := &domain.FileVersion{
		FileId:     fileId,
		VersionNum: 1, // Static for new file
		FileHash:   fileHash,
		Size:       fileSize,
		Status:     "PENDING",
	}
	if err := s.repo.CreateVersion(ctx, fileVersion); err != nil {
		return nil, "", fmt.Errorf("Error while creating first file version: %w", err)
	}

	// Construct S3 path: user/<user_id>/<file_id>/<version_num>
	objectKey := fmt.Sprintf("user/%s/%s/v1", ownerId, fileId)
	presignedURL, err := s.storage.GenerateUploadUrl(ctx, objectKey, 15*time.Minute)

	if err != nil {
		return nil, "", fmt.Errorf("storage provider error: %w", err)
	}

	return file, presignedURL, nil
}

// CompleteUpload marks a file upload as complete.
func (s *FileService) CompleteUpload(ctx context.Context, fileId string, versionNum int) error {
	log.Printf("File Service] Completing upload for file ID :%s", fileId)
	return s.repo.UpdateVersionStatus(ctx, fileId, versionNum, "AVAILABLE")
}

/*
Validates if a file is available for download and accordingly generates a presigned URL with 15 minute expiry
*/
func (s *FileService) InitiateDownload(ctx context.Context, fileId string, UserId string) (string, error) {
	file, err := s.repo.GetFileByID(ctx, fileId)
	fileVersion, err2 := s.repo.GetLatestVersion(ctx, fileId)

	if err != nil {
		return "", fmt.Errorf("fail to find file. Error: %v", err)
	}
	if err2 != nil {
		return "", fmt.Errorf("fail to find file. Error: %v", err2)
	}

	if file == nil {
		return "", fmt.Errorf("File is null. fail to find file. Error")
	}
	if fileVersion == nil {
		return "", fmt.Errorf("File to locate latest file version. Error")
	}
	// Only allow valid user to download file

	if file.OwnerID != UserId {
		return "", fmt.Errorf("Unauthorized user. Error: %v", err)
	}
	if fileVersion.Status != "AVAILABLE" {
		return "", fmt.Errorf("file is not available for download")
	}

	objectKey := fmt.Sprintf("user/%s/%s/v%d", UserId, file.ID, fileVersion.VersionNum)
	downloadUrl, err := s.storage.GenerateDownloadUrl(ctx, objectKey, 15*time.Minute)

	if err != nil {
		return "", fmt.Errorf("Storage provider error %w", err)
	}

	return downloadUrl, nil
}
