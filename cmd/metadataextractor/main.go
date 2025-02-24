package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joeldelpilar/vme/internal/exporter"
	"github.com/joeldelpilar/vme/internal/extractor"
)

func main() {
	// Definiera flaggor
	basicFlag := flag.Bool("b", false, "Basic metadata")
	extendedFlag := flag.Bool("e", false, "Extended metadata")
	fullFlag := flag.Bool("f", false, "Full metadata")
	outputFormat := flag.String("o", "", "Output format (json/xml)")

	flag.Parse()

	// Controll if an input file is given
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: metadataextractor [flags] <mp4-fil>")
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

	if *outputFormat != "" {
		err = exporter.ExportMetadata(metadata, *outputFormat)
		if err != nil {
			log.Fatalf("Failed to output metadata: %v", err)
		}
		fmt.Printf("\033[32mSuccessfully\033[0m exported metadata in %s format\n", strings.ToUpper(*outputFormat))
	} else {
		extractor.DisplayMetadata(metadata, level)
	}
}
