package domain

import (
	"context"
	"time"
)

// BlobStorage defines the contract for our storage layer
type BlobStorage interface {
	// creates a temporary, secure URL for the client to upload a file directly
	GenerateUploadUrl(ctx context.Context, key string, expireIn time.Duration) (string, error)

	//TODO : support for Download URL
}
