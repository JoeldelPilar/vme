# 🎬 vme - Video Metadata Extractor

[![Go Report Card](https://goreportcard.com/badge/github.com/joeldelpilar/vme)](https://goreportcard.com/report/github.com/joeldelpilar/vme)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A command-line tool to extract metadata from MP4 video files.

## 🌟 Features

*   Extract basic, extended, or full metadata from MP4 files.
*   Supports file information, movie information (tags, title, duration), and track information (bitrate, codecs, resolution).
*   Clear and formatted output to the console.

## 🛠️ Installation

1.  **Prerequisites:**
    *   [Go](https://go.dev/dl/) (version 1.22 or later)
    *   [FFmpeg](https://ffmpeg.org/download.html) (with `ffprobe` available in your system's PATH)

2.  **Get the package:**

    ```bash
    go install github.com/joeldelpilar/vme@latest
    ```

## 🚀 Usage

```bash
vme [flags] <mp4-file>
```

### 🔍 Flags

*   `-b`: Basic metadata (file information only).
*   `-e`: Extended metadata (file and movie information).
*   `-f`: Full metadata (file, movie, and track information).

### Examples

*   Extract basic metadata:

    ```bash
    vme -b video.mp4
    ```

*   Extract extended metadata:

    ```bash
    vme -e video.mp4
    ```

*   Extract full metadata:

    ```bash
    vme -f video.mp4
    ```

## ⚙️ How it Works

The tool uses `ffprobe` (from the FFmpeg suite) to analyze the MP4 file and extract metadata. The extracted data is then formatted and displayed in the console.

*   `cmd/metadataextractor/main.go`: startLine: 13 endLine: 49 -  The main entry point of the application, handling command-line arguments and calling the extraction logic.
*   `internal/extractor/extractor.go`: startLine: 147 endLine: 171 - Contains the core logic for extracting metadata using `ffprobe` and formatting the output.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## 👨‍💻 Author

*   Joel del Pilar - [GitHub](https://github.com/joeldelpilar)

## 🙏 Acknowledgments

*   Uses the [FFmpeg](https://ffmpeg.org/) project for media analysis.
*   Inspired by the need for a simple and effective MP4 metadata extraction tool.
