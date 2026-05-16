package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// s3Storage implements domain.BlobStorage
type S3Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucketName    string
}

func NewS3Storage(cfg aws.Config, bucketName string, endpoint string) *S3Storage {
	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		// This is the non-deprecated, modern way to point to LocalStack
		if endpoint != "" {
			options.BaseEndpoint = aws.String(endpoint)
		}
		options.UsePathStyle = true
	})

	return &S3Storage{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bucketName:    bucketName,
	}
}

func (this *S3Storage) GenerateUploadUrl(ctx context.Context, key string, expireIn time.Duration) (string, error) {
	request, err := this.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(this.bucketName),
		Key:    aws.String(key)}, func(opts *s3.PresignOptions) {
		opts.Expires = expireIn
	})
	if err != nil {
		return "", fmt.Errorf("Failed to generate presigned URL for upload %v", err)
	}
	return request.URL, nil
}
