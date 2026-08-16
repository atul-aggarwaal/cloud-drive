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

	//Get all files owned by or shared with the user
	GetFiles(ctx context.Context, userId string) ([]*File, error)

	//Get lateest version of a file by file id
	GetLatestVersion(ctx context.Context, fileId string) (*FileVersion, error)

	//List all the versions related to a file
	GetVersions(ctx context.Context, fileId string) ([]*FileVersion, error)

	//Delete metadata related to file present in s3 bucket
	RequestDelete(ctx context.Context, userId string, fileId string) error

	//Returns list of files which are marked for deletion
	GetFilesMarkedForDeletion(ctx context.Context) ([]*File, error)

	//Claim set of limited files for deletion
	ClaimFilesForDeletion(ctx context.Context, limit int) ([]*File, error)

	//Deletes a specific version of a file
	DeleteFileVersion(ctx context.Context, versionNum int, fileID string) error

	//Soft delete file
	MarkFileDeleted(ctx context.Context, fileID string) error
}
