package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestNewS3Client(t *testing.T) {
	tests := []struct {
		name    string
		cfg     S3Config
		wantErr bool
	}{
		{
			name: "Valid MinIO config",
			cfg: S3Config{
				BucketName: "test-bucket",
				Region:     "us-east-1",
				Endpoint:   "http://localhost:9000",
				UseSSL:     false,
				AccessKey:  "minioadmin",
				SecretKey:  "minioadmin",
			},
			wantErr: false,
		},
		{
			name: "Valid AWS config",
			cfg: S3Config{
				BucketName: "test-bucket",
				Region:     "us-east-1",
				AccessKey:  "test",
				SecretKey:  "test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewS3Client(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewS3Client() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Integration test - requires running MinIO
func TestS3Client_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	endpoint := "http://localhost:9000"

	config := S3Config{
		BucketName: "test-bucket",
		Region:     "us-east-1",
		Endpoint:   endpoint,
		UseSSL:     false,
		AccessKey:  "minioadmin",
		SecretKey:  "minioadmin",
	}

	ctx := context.Background()

	// Create S3 client for MinIO
	client, err := NewS3Client(config)
	if err != nil {
		t.Fatalf("Failed to create S3 client: %v", err)
	}

	// Try to create bucket if it doesn't exist
	_, err = client.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &config.BucketName,
	})
	if err != nil {
		t.Logf("Bucket does not exist, creating it...")
		// Bucket doesn't exist, create it
		_, err = client.client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: &config.BucketName,
		})
		if err != nil {
			t.Fatalf("Failed to create bucket: %v", err)
		}
		t.Logf("Created bucket: %s", config.BucketName)
	} else {
		t.Logf("Bucket %s already exists", config.BucketName)
	}

	// Create a temporary test file
	tmpfile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write some test data
	if _, err := tmpfile.Write([]byte("test content")); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	testKey := "test-key.txt"

	// Test file upload
	err = client.UploadFile(ctx, tmpfile.Name(), testKey)
	if err != nil {
		t.Errorf("UploadFile() error = %v", err)
	}

	t.Log("Successfully uploaded test file to MinIO")

	// Cleanup: Remove file from S3
	_, err = client.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &config.BucketName,
		Key:    &testKey,
	})
	if err != nil {
		t.Errorf("Failed to delete test file from S3: %v", err)
	} else {
		t.Log("Successfully deleted test file from S3")
	}

	// Give MinIO a moment to process the deletion
	time.Sleep(1 * time.Second)

	// Cleanup: Remove bucket
	_, err = client.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: &config.BucketName,
	})
	if err != nil {
		t.Errorf("Failed to delete test bucket: %v", err)
	} else {
		t.Log("Successfully deleted test bucket")
	}
}
