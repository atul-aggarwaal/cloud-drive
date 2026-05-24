package domain

import (
	"context"
)

// FileRepository defines the "Contract" for any database implementation.
type FileRepository interface {
	// CreateFile inserts a new file metadata record into the database.
	CreateFile(ctx context.Context, file *File) error

	//CreateVersion inserts a new file version record into the database.
	CreateVersion(ctx context.Context, fileVersion *FileVersion) error

	// Updates the upload status of a file to its current version
	UpdateVersionStatus(ctx context.Context, fileId string, versionNum int, status string) error

	// GetFileByID retrieves a file metadata record by its ID.
	GetFileByID(ctx context.Context, id string) (*File, error)
	GetLatestVersion(ctx context.Context, fileId string) (*FileVersion, error)
}
