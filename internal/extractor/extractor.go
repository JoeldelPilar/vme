package extractor

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
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

type MetadataTag struct {
	Name  string `json:"name" xml:"name"`
	Value string `json:"value" xml:"value"`
}

type MediaMetadata struct {
	XMLName   xml.Name `json:"-" xml:"MediaMetadata"`
	FileInfo  struct {
		Filename string `json:"filename" xml:"filename"`
		Size     string `json:"size" xml:"size"`
		Format   string `json:"format" xml:"format"`
	} `json:"fileInfo" xml:"FileInfo"`
	MovieInfo struct {
		Title    string        `json:"title" xml:"title"`
		Duration string        `json:"duration" xml:"duration"`
		Tags     []MetadataTag `json:"tags" xml:"tag"`
	} `json:"movieInfo" xml:"MovieInfo"`
	TrackInfo struct {
		Streams []StreamInfo `json:"streams" xml:"streams"`
		BitRate string       `json:"bitRate" xml:"bitRate"`
	} `json:"trackInfo" xml:"TrackInfo"`
}

type TagGroup struct {
    Name string
    Tags []string
}

var metadataGroups = []TagGroup{
    {
        Name: "Temporal Information",
        Tags: []string{"creation_time", "date", "year"},
    },
    {
        Name: "Content Information",
        Tags: []string{"title", "description", "synopsis", "comment", "copyright"},
    },
    {
        Name: "Creator Information",
        Tags: []string{"artist", "album_artist", "composer", "author", "director", "producer"},
    },
    {
        Name: "Categorization",
        Tags: []string{"genre", "album", "show", "episode_id", "network", "season_number", "episode_sort"},
    },
    {
        Name: "Technical Information",
        Tags: []string{"encoder", "encoder_version", "compatible_brands", "major_brand", "minor_version"},
    },
    {
        Name: "Location and Language",
        Tags: []string{"location", "language", "country"},
    },
    {
        Name: "Media Information",
        Tags: []string{"media_type", "rating", "purchase_date", "sort_name", "artwork_url"},
    },
    {
        Name: "Distribution",
        Tags: []string{"publisher", "publisher_id", "content_id", "isrc"},
    },
}

func DisplayMetadata(metadata MediaMetadata, level string) {
	fmt.Println("\033[32m----- File Information -----\033[0m")
	fmt.Printf("\nFilename: %s\n", metadata.FileInfo.Filename)
	fmt.Printf("Size: %s bytes\n", metadata.FileInfo.Size)
	fmt.Printf("Format: %s\n", metadata.FileInfo.Format)

	if level == "extended" || level == "full" {
		fmt.Println("\n\033[33m----- Movie Information -----\033[0m")
		if metadata.MovieInfo.Title != "" {
			fmt.Printf("\nTitle: %s\n", metadata.MovieInfo.Title)
		}
		fmt.Printf("Duration: %s seconds\n", metadata.MovieInfo.Duration)

		// Gruppera och visa tags
		tagsByGroup := make(map[string][]MetadataTag)
		for _, tag := range metadata.MovieInfo.Tags {
			for _, group := range metadataGroups {
				for _, groupTag := range group.Tags {
					if tag.Name == groupTag {
						tagsByGroup[group.Name] = append(tagsByGroup[group.Name], tag)
					}
				}
			}
		}

		// Visa grupperade tags
		for groupName, tags := range tagsByGroup {
			if len(tags) > 0 {
				fmt.Printf("\n\033[36m%s:\033[0m\n", groupName)
				for _, tag := range tags {
					fmt.Printf("  %s: %s\n", tag.Name, tag.Value)
				}
			}
		}
	}

	if level == "full" {
		fmt.Println("\n\033[96m----- Track Information -----\033[0m")
		fmt.Printf("\nBitrate: %s bits/s\n", metadata.TrackInfo.BitRate)
		
		for _, stream := range metadata.TrackInfo.Streams {
			fmt.Printf("\n%s Track - Codec: %s", 
				strings.ToUpper(stream.Type), stream.Codec)
			if stream.Resolution != "" {
				fmt.Printf(", Resolution: %s", stream.Resolution)
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
	if title, ok := data.Format.Tags["title"]; ok {
		metadata.MovieInfo.Title = title
	}

	// Extract desired tags
	for _, group := range metadataGroups {
		for _, tagName := range group.Tags {
			if value, exists := data.Format.Tags[tagName]; exists {
				metadata.MovieInfo.Tags = append(metadata.MovieInfo.Tags, MetadataTag{
					Name:  tagName,
					Value: value,
				})
			}
		}
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
func ExtractMetadata(filepath string, level string) (MediaMetadata, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filepath)

	output, err := cmd.Output()
	if err != nil {
		return MediaMetadata{}, fmt.Errorf("ffprobe failed: %w", err)
	}

	var data FFProbeOutput
	if err := json.Unmarshal(output, &data); err != nil {
		return MediaMetadata{}, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	metadata := extractMediaMetadata(data)
	return metadata, nil
}
// Lägg till ny funktion för att hantera output i olika format
func OutputMetadata(metadata MediaMetadata, format string) error {
	// Skapa output-filnamn baserat på formatet
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

	// Skriv till fil
	err = os.WriteFile(outputFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
