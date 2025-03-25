package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joeldelpilar/vme/internal/exporter"
	"github.com/joeldelpilar/vme/internal/extractor"
	"github.com/joeldelpilar/vme/internal/storage"
)

// Flags holds all command line flags
type Flags struct {
	Basic        bool
	Extended     bool
	Full         bool
	OutputFormat string
	S3Upload     bool
	S3Bucket     string
	S3Region     string
	S3Endpoint   string
	S3UseSSL     bool
}

func main() {
	flags := parseFlags()
	input := validateInput()

	var inputFile string
	var cleanup func()

	// Check if input is an S3 URI
	if strings.HasPrefix(input, "s3://") {
		var err error
		inputFile, cleanup, err = handleS3Input(input, flags)
		if err != nil {
			log.Fatalf("Error handling S3 input: %v", err)
		}
		if cleanup != nil {
			defer cleanup()
		}
	} else {
		inputFile = input
	}

	level := determineExtractionLevel(flags)

	absPath, err := filepath.Abs(inputFile)
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}

	metadata, err := extractor.ExtractMetadata(absPath, level)
	if err != nil {
		log.Fatalf("Extraction failed: %v", err)
	}

	if flags.OutputFormat != "" {
		outputFile := fmt.Sprintf("%s-metadata.%s", metadata.FileInfo.Filename, flags.OutputFormat)
		err = exporter.ExportMetadata(metadata, flags.OutputFormat)
		if err != nil {
			log.Fatalf("Failed to output metadata: %v", err)
		}
		fmt.Printf("\033[32mSuccessfully\033[0m exported metadata in %s format\n", strings.ToUpper(flags.OutputFormat))

		if flags.S3Upload {
			uploadToS3(flags, outputFile)
		}
	} else {
		extractor.DisplayMetadata(metadata, level)
	}
}

// parseFlags parses command line flags and returns them in a struct
func parseFlags() Flags {
	flags := Flags{}

	flag.BoolVar(&flags.Basic, "b", false, "Basic metadata")
	flag.BoolVar(&flags.Extended, "e", false, "Extended metadata")
	flag.BoolVar(&flags.Full, "f", false, "Full metadata")
	flag.StringVar(&flags.OutputFormat, "o", "", "Output format (json/xml)")

	flag.BoolVar(&flags.S3Upload, "s3-upload", false, "Upload metadata to S3")
	flag.StringVar(&flags.S3Bucket, "s3-bucket", "", "S3 bucket name")
	flag.StringVar(&flags.S3Region, "s3-region", "us-east-1", "S3 region (default: us-east-1)")
	flag.StringVar(&flags.S3Endpoint, "s3-endpoint", "", "S3 endpoint URL (for MinIO or other S3-compatible services)")
	flag.BoolVar(&flags.S3UseSSL, "s3-ssl", true, "Use SSL for S3 connection (default: true)")

	flag.Parse()

	// Validate output format if specified
	if flags.OutputFormat != "" {
		format := strings.ToLower(flags.OutputFormat)
		if format != "json" && format != "xml" {
			log.Fatalf("Invalid output format. Use 'json' or 'xml'")
		}
		flags.OutputFormat = format
	}

	return flags
}

// validateInput checks if input file is provided and returns it
func validateInput() string {
	if len(flag.Args()) != 1 {
		log.Fatal("Please provide exactly one input file or S3 URI")
	}
	return flag.Args()[0]
}

// handleS3Input handles input if it's an S3 URI
func handleS3Input(input string, flags Flags) (string, func(), error) {
	bucket, key, err := storage.ParseS3URI(input)
	if err != nil {
		return "", nil, err
	}

	// Create S3 configuration
	s3Config := storage.S3Config{
		BucketName: bucket,
		Region:     flags.S3Region,
		Endpoint:   flags.S3Endpoint,
		UseSSL:     flags.S3UseSSL,
	}

	// Load access key and secret key from environment variables
	s3Config = storage.LoadS3ConfigFromEnv(s3Config)

	// Check that necessary credentials exist
	if s3Config.AccessKey == "" || s3Config.SecretKey == "" {
		return "", nil, fmt.Errorf("S3 access key and secret key must be set via VME_S3_ACCESS_KEY and VME_S3_SECRET_KEY environment variables")
	}

	// Create S3 client
	client, err := storage.NewS3Client(s3Config)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create S3 client: %v", err)
	}

	// Download the file
	tmpFile, err := client.DownloadFile(context.Background(), key)
	if err != nil {
		return "", nil, fmt.Errorf("failed to download file from S3: %v", err)
	}

	cleanup := func() {
		os.Remove(tmpFile)
	}

	return tmpFile, cleanup, nil
}

// determineExtractionLevel determines which extraction level to use based on flags
func determineExtractionLevel(flags Flags) string {
	if flags.Full {
		return "full"
	} else if flags.Extended {
		return "extended"
	} else if flags.Basic {
		return "basic"
	}
	return "basic" // default
}

// uploadToS3 handles uploading a file to S3
func uploadToS3(flags Flags, outputFile string) {
	if flags.S3Bucket == "" {
		log.Fatalf("S3 bucket name must be specified with -s3-bucket when using -s3-upload")
	}

	// Create S3 configuration
	s3Config := storage.S3Config{
		BucketName: flags.S3Bucket,
		Region:     flags.S3Region,
		Endpoint:   flags.S3Endpoint,
		UseSSL:     flags.S3UseSSL,
	}

	// Load access key and secret key from environment variables
	s3Config = storage.LoadS3ConfigFromEnv(s3Config)

	// Check that necessary credentials exist
	if s3Config.AccessKey == "" || s3Config.SecretKey == "" {
		log.Fatalf("S3 access key and secret key must be set via VME_S3_ACCESS_KEY and VME_S3_SECRET_KEY environment variables")
	}

	// Create S3 client
	client, err := storage.NewS3Client(s3Config)
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}

	// Upload the file to S3
	ctx := context.Background()
	key := filepath.Base(outputFile)
	err = client.UploadFile(ctx, outputFile, key)
	if err != nil {
		log.Fatalf("Failed to upload file to S3: %v", err)
	}

	fmt.Printf("\033[32mSuccessfully\033[0m uploaded metadata to S3 bucket %s\n", flags.S3Bucket)
}
