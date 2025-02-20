package extractor

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path"
	"strings"
)

type FFProbeOutput struct {
	Format  Format   `json:"format"`
	Streams []Stream `json:"streams"`
}

type Format struct {
	Filename  string            `json:"filename"`
	Duration  string            `json:"duration"`
	Size      string            `json:"size"`
	BitRate   string            `json:"bit_rate"`
	Format    string            `json:"format_name"`
	Tags      map[string]string `json:"tags"`
}

type Stream struct {
	CodecType string `json:"codec_type"`
	CodecName string `json:"codec_name"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

type StreamInfo struct {
	Index      int
	Type       string
	Codec      string
	Resolution string
}

type MediaMetadata struct {
	// File level info (basic)
	FileInfo struct {
		Filename    string
		Size        string
		Format      string
	}

	// Movie metadata (extended)
	MovieInfo struct {
		Title       string
		Duration    string
		Tags        map[string]string
	}

	// Track metadata (full)
	TrackInfo struct {
		Streams     []StreamInfo
		BitRate     string
	}
}

func displayMetadata(metadata MediaMetadata, level string) {
	// File level info (always shown)
	fmt.Println("\033[32m----- File Information -----\033[0m")
	fmt.Printf("\nFilename: %s\n", metadata.FileInfo.Filename)
	fmt.Printf("Size: %s bytes\n", metadata.FileInfo.Size)
	fmt.Printf("Format: %s\n", metadata.FileInfo.Format)

	// Movie metadata (extended and full)
	if level == "extended" || level == "full" {
		fmt.Println("\n\033[33m----- Movie Information -----\033[0m")
		if metadata.MovieInfo.Title != "" {
			fmt.Printf("\nTitle: %s\n", metadata.MovieInfo.Title)
		}
		fmt.Printf("Duration: %s seconds\n", metadata.MovieInfo.Duration)
		
		// Print metadata tags
		for key, value := range metadata.MovieInfo.Tags {
			if key != "title" { // Skip title as we already printed it
				if key == "compatible_brands" {
					fmt.Printf("%s: ", key)
					for i := 0; i < len(value); i += 4 {
						end := i + 4
						if end > len(value) {
							end = len(value)
						}
						if i > 0 {
							fmt.Print(", ")
						}
						fmt.Print(value[i:end])
					}
					fmt.Println()
				} else {
					fmt.Printf("%s: %s\n", key, value)
				}
			}
		}
	}

	// Track metadata (full only)
	if level == "full" {
		fmt.Println("\n\033[96m----- Track Information -----\033[0m")
		fmt.Printf("\nBitrate: %s bits/s", metadata.TrackInfo.BitRate)
		
		for _, stream := range metadata.TrackInfo.Streams {
			fmt.Printf("\n%s Track - Codec: %s", 
				strings.ToUpper(stream.Type), stream.Codec)
			if stream.Resolution != "" {
				fmt.Printf(", \nResolution: %s\n", stream.Resolution)
			}
			fmt.Println()
		}
	}
}

func extractMediaMetadata(data FFProbeOutput) MediaMetadata {
	var metadata MediaMetadata

	// File level info
	metadata.FileInfo.Filename = path.Base(data.Format.Filename)
	metadata.FileInfo.Size = data.Format.Size
	metadata.FileInfo.Format = data.Format.Format

	// Movie metadata
	metadata.MovieInfo.Duration = data.Format.Duration
	metadata.MovieInfo.Tags = data.Format.Tags
	if title, ok := data.Format.Tags["title"]; ok {
		metadata.MovieInfo.Title = title
	}

	// Track metadata
	metadata.TrackInfo.BitRate = data.Format.BitRate
	for i, stream := range data.Streams {
		streamInfo := StreamInfo{
			Index: i,
			Type:  stream.CodecType,
			Codec: stream.CodecName,
		}
		if stream.CodecType == "video" {
			streamInfo.Resolution = fmt.Sprintf("%dx%d", stream.Width, stream.Height)
		}
		metadata.TrackInfo.Streams = append(metadata.TrackInfo.Streams, streamInfo)
	}

	return metadata
}

// ExtractMetadata extracts metadata from an MP4 file based on the given level.
// level can be "basic", "extended" or "full".
func ExtractMetadata(filepath string, level string) error {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filepath)

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("ffprobe failed: %w", err)
	}

	var data FFProbeOutput
	if err := json.Unmarshal(output, &data); err != nil {
		return fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	metadata := extractMediaMetadata(data)
	displayMetadata(metadata, level)

	return nil
}
