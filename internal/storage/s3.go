package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Config struct {
	BucketName string
	Region     string
	Endpoint   string
	UseSSL     bool
	AccessKey  string
	SecretKey  string
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

// LoadS3ConfigFromEnv loads S3 credentials from environment variables if they are not already set
func LoadS3ConfigFromEnv(config S3Config) S3Config {
	// Endast läs känsliga uppgifter från miljövariabler
	if config.AccessKey == "" {
		config.AccessKey = os.Getenv("VME_S3_ACCESS_KEY")
	}
	if config.SecretKey == "" {
		config.SecretKey = os.Getenv("VME_S3_SECRET_KEY")
	}

	return config
}

// DownloadFile downloads a file from S3 to a local temporary file
func (s *S3Client) DownloadFile(ctx context.Context, key string) (string, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "vme-download-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tmpFile.Close()

	// Get the object from S3
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	// Copy the contents to the temporary file
	_, err = io.Copy(tmpFile, result.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy S3 object to file: %w", err)
	}

	return tmpFile.Name(), nil
}

// ParseS3URI parses an S3 URI (s3://bucket/key) into bucket and key
func ParseS3URI(uri string) (bucket string, key string, err error) {
	if !strings.HasPrefix(uri, "s3://") {
		return "", "", fmt.Errorf("invalid S3 URI format: %s", uri)
	}

	parts := strings.SplitN(strings.TrimPrefix(uri, "s3://"), "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid S3 URI format: %s", uri)
	}

	return parts[0], parts[1], nil
}
