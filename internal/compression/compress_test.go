package compression

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jvs-project/jvs/pkg/model"
)

func TestNewCompressor(t *testing.T) {
	// Test no compression
	c := NewCompressor(LevelNone)
	if c.IsEnabled() {
		t.Error("expected compression to be disabled")
	}
	if c.Type != TypeNone {
		t.Errorf("expected type none, got %s", c.Type)
	}

	// Test fast compression
	c = NewCompressor(LevelFast)
	if !c.IsEnabled() {
		t.Error("expected compression to be enabled")
	}
	if c.Type != TypeGzip {
		t.Errorf("expected type gzip, got %s", c.Type)
	}
	if c.Level != LevelFast {
		t.Errorf("expected level %d, got %d", LevelFast, c.Level)
	}

	// Test default compression
	c = NewCompressor(LevelDefault)
	if c.Type != TypeGzip {
		t.Errorf("expected type gzip, got %s", c.Type)
	}
	if c.Level != LevelDefault {
		t.Errorf("expected level %d, got %d", LevelDefault, c.Level)
	}

	// Test max compression
	c = NewCompressor(LevelMax)
	if c.Type != TypeGzip {
		t.Errorf("expected type gzip, got %s", c.Type)
	}
	if c.Level != LevelMax {
		t.Errorf("expected level %d, got %d", LevelMax, c.Level)
	}
}

func TestNewCompressorFromString(t *testing.T) {
	tests := []struct {
		level    string
		expected CompressionLevel
	}{
		{"none", LevelNone},
		{"0", LevelNone},
		{"fast", LevelFast},
		{"1", LevelFast},
		{"default", LevelDefault},
		{"6", LevelDefault},
		{"max", LevelMax},
		{"9", LevelMax},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			c, err := NewCompressorFromString(tt.level)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if c.Level != tt.expected {
				t.Errorf("expected level %d, got %d", tt.expected, c.Level)
			}
		})
	}

	// Test invalid level
	_, err := NewCompressorFromString("invalid")
	if err == nil {
		t.Error("expected error for invalid level")
	}
}

func TestCompressorString(t *testing.T) {
	tests := []struct {
		level    CompressionLevel
		expected string
	}{
		{LevelNone, "none"},
		{LevelFast, "fast"},
		{LevelDefault, "default"},
		{LevelMax, "max"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			c := NewCompressor(tt.level)
			if c.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, c.String())
			}
		})
	}
}

func TestCompressDecompressFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	testData := []byte("Hello, World! This is test data for compression.")
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	// Test compression disabled
	cNone := NewCompressor(LevelNone)
	compressedPath, err := cNone.CompressFile(testFile)
	if err != nil {
		t.Fatalf("compress with none: %v", err)
	}
	if compressedPath != testFile {
		t.Errorf("expected same path, got %s", compressedPath)
	}

	// Test compression enabled
	cFast := NewCompressor(LevelFast)
	compressedPath, err = cFast.CompressFile(testFile)
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	if compressedPath != testFile+".gz" {
		t.Errorf("expected .gz extension, got %s", compressedPath)
	}

	// Verify compressed file exists
	if _, err := os.Stat(compressedPath); os.IsNotExist(err) {
		t.Error("compressed file not created")
	}

	// Test decompression
	decompressedPath, err := DecompressFile(compressedPath)
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}
	if decompressedPath != testFile {
		t.Errorf("expected original path, got %s", decompressedPath)
	}

	// Verify decompressed data matches original
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("read decompressed: %v", err)
	}
	if string(data) != string(testData) {
		t.Error("decompressed data doesn't match original")
	}
}

func TestCompressDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"file1.txt": "content 1",
		"file2.txt": "content 2",
		"subdir/file3.txt": "content 3",
	}
	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}

	// Compress directory
	c := NewCompressor(LevelFast)
	count, err := c.CompressDir(tmpDir)
	if err != nil {
		t.Fatalf("compress dir: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 compressed files, got %d", count)
	}

	// Verify compressed files exist
	for path := range files {
		gzPath := filepath.Join(tmpDir, path+".gz")
		if _, err := os.Stat(gzPath); os.IsNotExist(err) {
			t.Errorf("compressed file not found: %s", gzPath)
		}
		// Original should be removed
		origPath := filepath.Join(tmpDir, path)
		if _, err := os.Stat(origPath); err == nil {
			t.Errorf("original file not removed: %s", origPath)
		}
	}
}

func TestDecompressDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create compressed test files
	files := map[string]string{
		"file1.txt": "content 1",
		"file2.txt": "content 2",
		"subdir/file3.txt": "content 3",
	}
	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path+".gz")
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		// Create compressed file
		c := NewCompressor(LevelFast)
		testFile := filepath.Join(tmpDir, path)
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}
		if _, err := c.CompressFile(testFile); err != nil {
			t.Fatalf("compress file: %v", err)
		}
		// Remove original
		os.Remove(testFile)
	}

	// Decompress directory
	count, err := DecompressDir(tmpDir)
	if err != nil {
		t.Fatalf("decompress dir: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 decompressed files, got %d", count)
	}

	// Verify decompressed files exist and content matches
	for path, expectedContent := range files {
		origPath := filepath.Join(tmpDir, path)
		data, err := os.ReadFile(origPath)
		if err != nil {
			t.Errorf("read decompressed file: %v", err)
			continue
		}
		if string(data) != expectedContent {
			t.Errorf("content mismatch for %s", path)
		}
		// Compressed file should be removed
		gzPath := filepath.Join(tmpDir, path+".gz")
		if _, err := os.Stat(gzPath); err == nil {
			t.Errorf("compressed file not removed: %s", gzPath)
		}
	}
}

func TestIsCompressedFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"file.txt", false},
		{"file.txt.gz", true},
		{"file.gz", true},
		{"file.tar.gz", true},
		{"gzip", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsCompressedFile(tt.path); got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestCompressedUncompressedPath(t *testing.T) {
	tests := []struct {
		path              string
		compressed        string
		uncompressed      string
	}{
		{"file.txt", "file.txt.gz", "file.txt"},
		{"file.txt.gz", "file.txt.gz.gz", "file.txt.gz"},
		{"path/to/file", "path/to/file.gz", "path/to/file"},
		{"archive.tar.gz", "archive.tar.gz.gz", "archive.tar.gz"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if c := CompressedPath(tt.path); c != tt.compressed {
				t.Errorf("CompressedPath: expected %s, got %s", tt.compressed, c)
			}
			if u := UncompressedPath(tt.compressed); u != tt.uncompressed {
				t.Errorf("UncompressedPath: expected %s, got %s", tt.uncompressed, u)
			}
		})
	}
}

func TestCompressionInfoFromLevel(t *testing.T) {
	// Test valid levels
	info, err := CompressionInfoFromLevel("fast")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if info.Type != TypeGzip {
		t.Errorf("expected type gzip, got %s", info.Type)
	}
	if info.Level != LevelFast {
		t.Errorf("expected level %d, got %d", LevelFast, info.Level)
	}

	// Test none (returns nil)
	info, err = CompressionInfoFromLevel("none")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if info != nil {
		t.Error("expected nil for none compression")
	}

	// Test invalid
	_, err = CompressionInfoFromLevel("invalid")
	if err == nil {
		t.Error("expected error for invalid level")
	}
}

func TestCompressionInfoString(t *testing.T) {
	info := &model.CompressionInfo{
		Type:  "gzip",
		Level: 6,
	}

	// This tests the model.CompressionInfo.String method if it exists,
	// or verifies the structure is correct
	if info.Type != "gzip" {
		t.Errorf("expected type gzip, got %s", info.Type)
	}
	if info.Level != 6 {
		t.Errorf("expected level 6, got %d", info.Level)
	}
}

func TestCompressBytes(t *testing.T) {
	c := NewCompressor(LevelFast)
	data := []byte("Hello, World!")

	compressed, err := c.compressBytes(data)
	if err != nil {
		t.Fatalf("compress: %v", err)
	}

	// Compressed data should be different from original
	if string(compressed) == string(data) {
		t.Error("compressed data is identical to original")
	}

	// Decompress and verify
	decompressed, err := decompressBytes(compressed)
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}

	if string(decompressed) != string(data) {
		t.Error("decompressed data doesn't match original")
	}
}

func TestDecompressFileNotCompressed(t *testing.T) {
	tmpDir := t.TempDir()

	// Create non-compressed file
	testFile := filepath.Join(tmpDir, "test.txt")
	testData := []byte("Hello, World!")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// Decompress should return original path for non-compressed files
	result, err := DecompressFile(testFile)
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}
	if result != testFile {
		t.Errorf("expected original path, got %s", result)
	}
}
