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

func main() {
	// Defining flags
	basicFlag := flag.Bool("b", false, "Basic metadata")
	extendedFlag := flag.Bool("e", false, "Extended metadata")
	fullFlag := flag.Bool("f", false, "Full metadata")
	outputFormat := flag.String("o", "", "Output format (json/xml)")

	s3Upload := flag.Bool("s3-upload", false, "Upload metadata to S3")
	s3Bucket := flag.String("s3-bucket", "", "S3 bucket name")
	s3Region := flag.String("s3-region", "us-east-1", "S3 region (default: us-east-1)")
	s3Endpoint := flag.String("s3-endpoint", "", "S3 endpoint URL (for MinIO or other S3-compatible services)")
	s3UseSSL := flag.Bool("s3-ssl", true, "Use SSL for S3 connection (default: true)")

	flag.Parse()

	// Control if an input file is given
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: vme [flags] <mp4-file>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	inputFile := args[0]
	absPath, err := filepath.Abs(inputFile)
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}

	// Determine which extraction level to use
	level := "basic" // default
	if *fullFlag {
		level = "full"
	} else if *extendedFlag {
		level = "extended"
	} else if *basicFlag {
		level = "basic"
	}

	// Validate output format if specified
	if *outputFormat != "" {
		format := strings.ToLower(*outputFormat)
		if format != "json" && format != "xml" {
			log.Fatalf("Invalid output format. Use 'json' or 'xml'")
		}
	}

	metadata, err := extractor.ExtractMetadata(absPath, level)
	if err != nil {
		log.Fatalf("Extraction failed: %v", err)
	}

	var outputFile string
	if *outputFormat != "" {
		outputFile = fmt.Sprintf("%s-metadata.%s", metadata.FileInfo.Filename, *outputFormat)
		err = exporter.ExportMetadata(metadata, *outputFormat)
		if err != nil {
			log.Fatalf("Failed to output metadata: %v", err)
		}
		fmt.Printf("\033[32mSuccessfully\033[0m exported metadata in %s format\n", strings.ToUpper(*outputFormat))
		
		// Handle S3 upload if it's enabled and we have created a file
		if *s3Upload {
			if *s3Bucket == "" {
				log.Fatalf("S3 bucket name must be specified with -s3-bucket when using -s3-upload")
			}
			
			// Create S3 configuration with values from the flags
			s3Config := storage.S3Config{
				BucketName: *s3Bucket,
				Region:     *s3Region,
				Endpoint:   *s3Endpoint,
				UseSSL:     *s3UseSSL,
				// Don't set AccessKey and SecretKey here, read from environment variables instead
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
			
			fmt.Printf("\033[32mSuccessfully\033[0m uploaded metadata to S3 bucket %s\n", *s3Bucket)
		}
	} else {
		extractor.DisplayMetadata(metadata, level)
	}
}
