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
	repo          domain.FileRepository
	userRepo      domain.UserRepository
	storage       domain.BlobStorage
	fileShareRepo domain.FileShareRepository
}

// NewFileService creates a new instance of FileService.
func NewFileService(repo domain.FileRepository, userRepo domain.UserRepository, storage domain.BlobStorage, fileShareRepo domain.FileShareRepository) *FileService {
	return &FileService{
		repo:          repo,
		userRepo:      userRepo,
		storage:       storage,
		fileShareRepo: fileShareRepo,
	}
}

// InitiateUpload initiates an upload for a file by storing its metadata in the DB
// and providing a presigned URL to the S3 bucket to upload the actual file.
func (s *FileService) InitiateUpload(ctx context.Context, ownerId string, fileName string, fileHash string, fileSize int64) (*domain.File, string, error) {
	fileId := uuid.New().String()

	user, err := s.userRepo.GetUserByID(ctx, ownerId)
	if user == nil {
		return nil, "", fmt.Errorf("file owner validation failed: %w", err)
	}

	existingFile, err := s.repo.GetFileFileByOwnerIdNameAndStatus(ctx, ownerId, fileName, domain.FileStatusActive)
	if err != nil {
		return nil, "", fmt.Errorf("retreiving existing file: %w ", err)
	}

	new_version := 1
	if existingFile != nil {
		lastVersion, err := s.repo.GetLatestVersion(ctx, existingFile.ID)
		if err != nil {
			return nil, "", fmt.Errorf("reteriving existing version: %w", err)
		}
		new_version = lastVersion.VersionNum + 1
		fileId = existingFile.ID
	}

	//Create Metadata for File
	file := &domain.File{
		ID:       fileId,
		OwnerID:  ownerId,
		FileName: fileName,
		IsFolder: false,
	}

	//Create Metadata for FileVersion
	fileVersion := &domain.FileVersion{
		FileId:     fileId,
		VersionNum: new_version, // Static for new file
		FileHash:   fileHash,
		Size:       fileSize,
		Status:     "PENDING",
	}

	// Construct S3 path: user/<user_id>/<file_id>/<version_num>
	objectKey := fmt.Sprintf("user/%s/%s/%d", ownerId, fileId, new_version)
	presignedURL, err := s.storage.GenerateUploadUrl(ctx, objectKey, fileHash, 15*time.Minute)
	if err != nil {
		return nil, "", fmt.Errorf("generating upload url: %w", err)
	}

	if existingFile != nil {
		if err := s.repo.CreateNewFileVersion(ctx, fileVersion); err != nil {
			return nil, "", fmt.Errorf("error while creating new file version: %w", err)
		}
	} else {
		if err := s.repo.CreateFileWithInitialVersion(ctx, file, fileVersion); err != nil {
			return nil, "", fmt.Errorf("error while creating first file version: %w", err)
		}
	}
	return file, presignedURL, nil
}

// CompleteUpload marks a file upload as complete.
func (s *FileService) CompleteUpload(ctx context.Context, fileId string, versionNum int) error {
	log.Printf("File Service] Completing upload for file ID :%s", fileId)
	return s.repo.UpdateFileVersionStatus(ctx, fileId, versionNum, domain.FileStatusPending, domain.FileStatusAvailable)
}

// InitiateDownload validates if a file is available for download and accordingly generates a presigned URL with 15 minute expiry.
func (s *FileService) InitiateDownload(ctx context.Context, fileId string, userId string) (string, error) {
	log.Println("Initiating download for file")

	file, err := s.repo.GetFileByID(ctx, fileId)
	fileVersion, err2 := s.repo.GetLatestVersion(ctx, fileId)

	if err != nil {
		return "", fmt.Errorf("fail to find file: %v", err)
	}
	if err2 != nil {
		return "", fmt.Errorf("fail to find file version: %v", err2)
	}

	if file == nil {
		return "", fmt.Errorf("file is null, fail to find file")
	}
	if fileVersion == nil {
		return "", fmt.Errorf("failed to locate latest file version")
	}

	// Only allow valid user to download file
	// Validate file access, if user id doesn't match with owner id of file.
	if userId != file.OwnerID {

		isAccesible, err := s.fileShareRepo.HasAccess(ctx, fileId, userId)
		if err != nil {
			return "", fmt.Errorf("error validating file authorization %w", err)
		}
		if !isAccesible {
			return "", fmt.Errorf("unauthorized access")
		}
	}

	if fileVersion.Status != "AVAILABLE" {
		return "", fmt.Errorf("file is not available for download")
	}

	objectKey := fmt.Sprintf("user/%s/%s/v%d", file.OwnerID, file.ID, fileVersion.VersionNum)
	downloadUrl, err := s.storage.GenerateDownloadUrl(ctx, objectKey, 15*time.Minute)

	if err != nil {
		return "", fmt.Errorf("storage provider error: %w", err)
	}

	return downloadUrl, nil
}

// Returns list of files Owned or Shared with user
func (s *FileService) ListFilesForUser(ctx context.Context, userId string) ([]*domain.File, error) {
	log.Printf("executing service method to list files.")

	files, error := s.repo.GetFiles(ctx, userId)

	if error != nil {
		return nil, fmt.Errorf("listingFilesForUser: %w", error)
	}

	if len(files) == 0 {
		user, err := s.userRepo.GetUserByID(ctx, userId)
		if err != nil {
			log.Printf("No files found for user {%s}", user.UserName)
		}
		return nil, nil
	}
	return files, nil
}

// marks file for deletion, later to be picked up by a janitor
func (s *FileService) RequestFileDeletion(ctx context.Context, userId string, fileId string) error {
	file, err := s.repo.GetFileByID(ctx, fileId)
	if err != nil {
		return fmt.Errorf("get file: %w", err)
	}

	if file.OwnerID != userId {
		return domain.ErrorPermissionDenied
	}

	fileVersions, err := s.repo.GetVersions(ctx, fileId)
	if err != nil {
		return fmt.Errorf("load versions for file %S: %w", fileId, err)
	}
	if len(fileVersions) == 0 {
		return fmt.Errorf("file %s has no versions: %w", fileId, domain.ErorFileCorrupted)
	}
	// soft delete file metadata from db
	err = s.repo.RequestDelete(ctx, userId, fileId)
	if err != nil {
		return fmt.Errorf("mark requested delete: %w", err)
	}
	return nil
}
