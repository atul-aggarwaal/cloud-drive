package domain

import (
	"context"
)

// FileRepository defines the "Contract" for any database implementation.
type FileRepository interface {
	// Save inserts a new file record into the database.
	Save(ctx context.Context, file *File) error

	// GetByID retrieves a file record by its ID.
	GetByID(ctx context.Context, id string) (*File, error)
}
