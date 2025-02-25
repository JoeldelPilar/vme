package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Config struct {
	BucketName  string
	Region      string
	Endpoint    string
	UseSSL      bool
	AccessKey   string
	SecretKey   string
}

type S3Client struct {
	client     *s3.Client
	bucketName string
}

// NewS3Client creates a new S3 client with the provided configuration
func NewS3Client(cfg S3Config) (*S3Client, error) {
	options := s3.Options{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		),
	}

	if cfg.Endpoint != "" {
		options.BaseEndpoint = &cfg.Endpoint
		options.UsePathStyle = true // Important for MinIO and other S3-compatible services
	}

	client := s3.New(s3.Options(options))
	return &S3Client{
		client:     client,
		bucketName: cfg.BucketName,
	}, nil
}

// UploadFile uploads a file to S3
func (s *S3Client) UploadFile(ctx context.Context, filePath string, key string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return nil
}
