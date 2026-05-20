package domain

import (
	"context"
	"time"
)

// BlobStorage defines the contract for any blob storage implementation.

type BlobStorage interface {

	// GenerateUploadUrl creates a temporary, secure URL for the client to upload a file directly
	GenerateUploadUrl(ctx context.Context, key string, expireIn time.Duration) (string, error)

	/*
		Gemerates a pre-signed download url for a given file
	*/
	GenerateDownloadUrl(ctx context.Context, key string, expireIn time.Duration) (string, error)
}
