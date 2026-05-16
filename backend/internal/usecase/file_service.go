package usecase

import (
	"context"
	"fmt"
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

func NewFileService(repo domain.FileRepository, storage domain.BlobStorage) *FileService {
	return &FileService{
		repo:    repo,
		storage: storage,
	}
}

// InitiateUpload Initiate an upload for a file by storing its metadata in DB and by providing a presigned URL
// to S3 bucket to updload actual file.
func (this *FileService) InitiateUpload(ctx context.Context, userId, fileName string, size int64) (*domain.File, string, error) {
	fileId := uuid.New().String()

	file := &domain.File{
		ID:        fileId,
		UserID:    userId,
		FileName:  fileName,
		Size:      size,
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}

	if err := this.repo.Save(ctx, file); err != nil {
		return nil, "", fmt.Errorf("Metadata persistance failure: %v", err)
	}

	// Construct S3 path: user/<user_id>/<file_id>
	objectKey := fmt.Sprintf("user/%s/%s", userId, fileId)

	presignedURL, err := this.storage.GenerateUploadUrl(ctx, objectKey, 15*time.Minute)

	if err != nil {
		return nil, "", fmt.Errorf("Storage Provider error: %v", err)
	}

	return file, presignedURL, nil
}
