package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Storage implements domain.BlobStorage using AWS S3.
type S3Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucketName    string
}

// NewS3Storage creates a new instance of S3Storage.
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

// GenerateUploadUrl creates a presigned URL that can be used to upload an object to S3.
func (this *S3Storage) GenerateUploadUrl(ctx context.Context, key string, fileHash string, expireIn time.Duration) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(this.bucketName),
		Key:         aws.String(key),
		ChecksumMD5: aws.String(fileHash),
		ChecksumAlgorithm: types.ChecksumAlgorithmMd5, // expect two additional headers in the request: x-amz-checksum-algorithm and x-amz-checksum-md5
	}

	request, err := this.presignClient.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {opts.Expires = expireIn})
	if err != nil {
		return "", fmt.Errorf("Failed to generate presigned URL for upload %v", err)
	}
	return request.URL, nil
}

func (this *S3Storage) GenerateDownloadUrl(ctx context.Context, key string, expiredIn time.Duration) (string, error) {
	request, err := this.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(this.bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiredIn))

	if err != nil {
		return "", fmt.Errorf("Failed to generate presigned URL for download %v", err)
	}
	return request.URL, nil
}
