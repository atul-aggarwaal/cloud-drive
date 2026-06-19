package domain

import (
	"context"
	"time"
)

// Struct represents a single file share record
type FileShare struct {
	ID               string    `json:"id"`               //Unique Share ID
	FileID           string    `json:"file_id"`          // ID of the file being shared
	SharedWithUserID string    `json:"user_id"`          // ID of the user with whome file is being shared
	PermissionLevel  string    `json:"permission_level"` // Level of permission to the user "read" / "read_write"
	CreatedAt        time.Time `json:"created_at"`       // Time at which permission was given
}

type FileShareRequest struct {
	FileID          string `json:"file_id"`
	TargetEmail     string `json:"target_email"`
	PermissionLevel string `json:"permission_level"`
}
type FileShareRepository interface {
	CreateShare(ctx context.Context, share *FileShare) error
	HasAccess(ctx context.Context, fileId string, userId string) (bool, error)
}
