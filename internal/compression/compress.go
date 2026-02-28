// Package compression provides compression support for JVS snapshots.
// It supports gzip compression at configurable levels for snapshot data.
package compression

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CompressionLevel represents the compression level.
type CompressionLevel int

const (
	// LevelNone disables compression.
	LevelNone CompressionLevel = 0
	// LevelFast uses fastest compression (gzip level 1).
	LevelFast CompressionLevel = 1
	// LevelDefault uses default compression (gzip level 6).
	LevelDefault CompressionLevel = 6
	// LevelMax uses maximum compression (gzip level 9).
	LevelMax CompressionLevel = 9
)

// CompressionType represents the compression algorithm.
type CompressionType string

const (
	// TypeGzip uses gzip compression.
	TypeGzip CompressionType = "gzip"
	// TypeNone indicates no compression.
	TypeNone CompressionType = "none"
)

// Compressor handles compression operations.
type Compressor struct {
	Type  CompressionType
	Level CompressionLevel
}

// NewCompressor creates a new compressor with the specified level.
// Level 0 means no compression.
func NewCompressor(level CompressionLevel) *Compressor {
	if level <= LevelNone {
		return &Compressor{Type: TypeNone, Level: LevelNone}
	}
	return &Compressor{Type: TypeGzip, Level: level}
}

// NewCompressorFromString creates a compressor from a string level.
// Valid values: "none", "fast", "default", "max"
func NewCompressorFromString(level string) (*Compressor, error) {
	switch strings.ToLower(level) {
	case "none", "0":
		return NewCompressor(LevelNone), nil
	case "fast", "1":
		return NewCompressor(LevelFast), nil
	case "default", "6":
		return NewCompressor(LevelDefault), nil
	case "max", "9":
		return NewCompressor(LevelMax), nil
	default:
		return nil, fmt.Errorf("invalid compression level: %s (must be none, fast, default, or max)", level)
	}
}

// IsEnabled returns true if compression is enabled.
func (c *Compressor) IsEnabled() bool {
	return c.Type != TypeNone
}

// String returns the string representation of the compressor.
func (c *Compressor) String() string {
	switch c.Level {
	case LevelNone:
		return "none"
	case LevelFast:
		return "fast"
	case LevelDefault:
		return "default"
	case LevelMax:
		return "max"
	default:
		return fmt.Sprintf("level-%d", c.Level)
	}
}

// CompressFile compresses a file and returns the compressed path.
// The compressed file has a .gz extension added.
// If compression is disabled, returns the original path.
func (c *Compressor) CompressFile(path string) (string, error) {
	if !c.IsEnabled() {
		return path, nil
	}

	// Read original file
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	// Compress
	compressed, err := c.compressBytes(data)
	if err != nil {
		return "", fmt.Errorf("compress: %w", err)
	}

	// Write compressed file
	compressedPath := path + ".gz"
	if err := os.WriteFile(compressedPath, compressed, 0600); err != nil {
		return "", fmt.Errorf("write compressed file: %w", err)
	}

	return compressedPath, nil
}

// DecompressFile decompresses a .gz file and returns the decompressed path.
// If the file is not compressed, returns the original path.
func DecompressFile(path string) (string, error) {
	// Check if file is compressed
	if !strings.HasSuffix(path, ".gz") {
		return path, nil
	}

	// Read compressed file
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read compressed file: %w", err)
	}

	// Decompress
	decompressed, err := decompressBytes(data)
	if err != nil {
		return "", fmt.Errorf("decompress: %w", err)
	}

	// Write decompressed file (remove .gz extension)
	decompressedPath := strings.TrimSuffix(path, ".gz")
	if err := os.WriteFile(decompressedPath, decompressed, 0644); err != nil {
		return "", fmt.Errorf("write decompressed file: %w", err)
	}

	return decompressedPath, nil
}

// CompressDir compresses all files in a directory tree.
// Returns the count of compressed files and any error.
func (c *Compressor) CompressDir(root string) (int, error) {
	if !c.IsEnabled() {
		return 0, nil
	}

	count := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and already compressed files
		if info.IsDir() || strings.HasSuffix(path, ".gz") {
			return nil
		}

		// Compress file
		_, err = c.CompressFile(path)
		if err != nil {
			return fmt.Errorf("compress %s: %w", path, err)
		}

		// Remove original file after successful compression
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove original %s: %w", path, err)
		}

		count++
		return nil
	})

	return count, err
}

// DecompressDir decompresses all .gz files in a directory tree.
// Returns the count of decompressed files and any error.
func DecompressDir(root string) (int, error) {
	count := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-compressed files
		if info.IsDir() || !strings.HasSuffix(path, ".gz") {
			return nil
		}

		// Skip .READY.gz markers (metadata, don't decompress)
		if strings.HasPrefix(filepath.Base(path), ".READY") {
			return nil
		}

		// Decompress file
		_, err = DecompressFile(path)
		if err != nil {
			return fmt.Errorf("decompress %s: %w", path, err)
		}

		// Remove compressed file after successful decompression
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove compressed %s: %w", path, err)
		}

		count++
		return nil
	})

	return count, err
}

// compressBytes compresses a byte slice using gzip.
func (c *Compressor) compressBytes(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, int(c.Level))
	if err != nil {
		return nil, fmt.Errorf("create gzip writer: %w", err)
	}

	if _, err := w.Write(data); err != nil {
		w.Close()
		return nil, fmt.Errorf("write: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("close: %w", err)
	}

	return buf.Bytes(), nil
}

// decompressBytes decompresses a gzipped byte slice.
func decompressBytes(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create gzip reader: %w", err)
	}
	defer r.Close()

	result, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	return result, nil
}

// IsCompressedFile returns true if the file path indicates a compressed file.
func IsCompressedFile(path string) bool {
	return strings.HasSuffix(path, ".gz")
}

// CompressedPath returns the compressed path for a file.
func CompressedPath(path string) string {
	return path + ".gz"
}

// UncompressedPath returns the uncompressed path for a file.
func UncompressedPath(path string) string {
	return strings.TrimSuffix(path, ".gz")
}

// SnapshotCompressionInfo stores compression metadata in the descriptor.
type SnapshotCompressionInfo struct {
	Type  CompressionType     `json:"type,omitempty"`
	Level CompressionLevel     `json:"level,omitempty"`
}

// CompressionInfoFromLevel creates compression info from a level string.
func CompressionInfoFromLevel(level string) (*SnapshotCompressionInfo, error) {
	c, err := NewCompressorFromString(level)
	if err != nil {
		return nil, err
	}
	if !c.IsEnabled() {
		return nil, nil
	}
	return &SnapshotCompressionInfo{
		Type:  c.Type,
		Level: c.Level,
	}, nil
}

// String returns the string representation of the compression info.
func (ci *SnapshotCompressionInfo) String() string {
	if ci == nil || ci.Type == TypeNone {
		return "none"
	}
	return fmt.Sprintf("%s-%d", ci.Type, ci.Level)
}
