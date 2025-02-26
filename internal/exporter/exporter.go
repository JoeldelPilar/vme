package exporter

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/joeldelpilar/vme/internal/extractor"
)

// ExportMetadata exports metadata to a file in the specified format
func ExportMetadata(metadata extractor.MediaMetadata, format string) error {
	// Create output filename based on format
	outputFile := fmt.Sprintf("%s-metadata.%s", metadata.FileInfo.Filename, format)

	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(metadata, "", "  ")
	case "xml":
		data, err = xml.MarshalIndent(metadata, "", "  ")
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Write to file
	err = os.WriteFile(outputFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
