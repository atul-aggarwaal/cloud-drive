package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/atul-aggarwaal/cloud-drive/internal/domain"
	"github.com/atul-aggarwaal/cloud-drive/internal/pkg/crypto"
)

type FileShareService struct {
	fileRepo      domain.FileRepository
	fileShareRepo domain.FileShareRepository
	userRepo      domain.UserRepository
}

func NewFileShareService(fileRepo domain.FileRepository, fileShareRepository domain.FileShareRepository, userRepository domain.UserRepository) *FileShareService {
	return &FileShareService{
		fileRepo:      fileRepo,
		fileShareRepo: fileShareRepository,
		userRepo:      userRepository,
	}
}

// Shares a file with some user.
func (s *FileShareService) ShareFile(ctx context.Context, ownerID string, request domain.FileShareRequest) error {
	file, err := s.fileRepo.GetFileByID(ctx, request.FileID)

	//find file being shared
	if err != nil {
		return fmt.Errorf("internal error occurred %w", err)
	}
	if file == nil {
		return fmt.Errorf("file resource not found.")
	}

	//Verify ownership of file before sharing
	if file.OwnerID != ownerID {
		return fmt.Errorf("unauthorized : you don't have permission to share this file.")
	}

	//Find target User and related details
	targetUser, err := s.userRepo.GetUserByEmail(ctx, request.TargetEmail)
	if err != nil {
		return fmt.Errorf("internal error occurred: %w", err)
	}
	if targetUser == nil {
		return fmt.Errorf("target user account not found")
	}

	uuid, err := crypto.GenerateUUID7()
	if err != nil {
		return fmt.Errorf("internal error occurred %w", err)
	}

	shareRecord := &domain.FileShare{
		ID:               uuid,
		FileID:           file.ID,
		SharedWithUserID: targetUser.ID,
		PermissionLevel:  request.PermissionLevel,
		CreatedAt:        time.Now(),
	}

	return s.fileShareRepo.CreateShare(ctx, shareRecord)
}
