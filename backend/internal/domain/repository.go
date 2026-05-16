package domain

import (
	"context"
)

// FileRepository defines the "Contract" for any database implementation.
type FileRepository interface {
	Save(ctx context.Context, file *File) error
	GetByID(ctx context.Context, id string) (*File, error)
}
