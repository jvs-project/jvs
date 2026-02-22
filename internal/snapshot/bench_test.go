package snapshot_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/jvs-project/jvs/internal/integrity"
	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/pkg/model"
)

// setupBenchRepo creates a repository with test content for benchmarking.
func setupBenchRepo(b *testing.B, contentSize int) string {
	dir := b.TempDir()
	_, err := repo.Init(dir, "bench")
	if err != nil {
		b.Fatal(err)
	}

	mainPath := filepath.Join(dir, "main")

	// Create content of specified size
	if contentSize > 0 {
		data := make([]byte, contentSize)
		for i := range data {
			data[i] = byte(i % 256)
		}
		if err := os.WriteFile(filepath.Join(mainPath, "data.bin"), data, 0644); err != nil {
			b.Fatal(err)
		}
	}

	return dir
}

// setupBenchRepoWithFiles creates a repository with multiple files for benchmarking.
func setupBenchRepoWithFiles(b *testing.B, fileCount int) string {
	dir := b.TempDir()
	_, err := repo.Init(dir, "bench")
	if err != nil {
		b.Fatal(err)
	}

	mainPath := filepath.Join(dir, "main")

	// Create multiple files and directories
	for i := 0; i < fileCount; i++ {
		subDir := filepath.Join(mainPath, "dir", strconv.Itoa(i%10))
		if err := os.MkdirAll(subDir, 0755); err != nil {
			b.Fatal(err)
		}

		data := []byte("test content for benchmarking")
		fileName := "file" + strconv.Itoa(i) + ".txt"
		filePath := filepath.Join(subDir, fileName)
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			b.Fatal(err)
		}
	}

	return dir
}

// BenchmarkSnapshotCreation_CopyEngine_Small benchmarks snapshot creation with small payload using copy engine.
func BenchmarkSnapshotCreation_CopyEngine_Small(b *testing.B) {
	repoPath := setupBenchRepo(b, 1024) // 1KB
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := creator.Create("main", "bench snapshot", nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSnapshotCreation_CopyEngine_Medium benchmarks snapshot creation with medium payload using copy engine.
func BenchmarkSnapshotCreation_CopyEngine_Medium(b *testing.B) {
	repoPath := setupBenchRepo(b, 1024*1024) // 1MB
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := creator.Create("main", "bench snapshot", nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSnapshotCreation_CopyEngine_Large benchmarks snapshot creation with large payload using copy engine.
func BenchmarkSnapshotCreation_CopyEngine_Large(b *testing.B) {
	b.Skip("Skipping large file benchmark in normal test runs")
	repoPath := setupBenchRepo(b, 10*1024*1024) // 10MB
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := creator.Create("main", "bench snapshot", nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSnapshotCreation_ReflinkEngine_Small benchmarks snapshot creation with small payload using reflink engine.
func BenchmarkSnapshotCreation_ReflinkEngine_Small(b *testing.B) {
	repoPath := setupBenchRepo(b, 1024) // 1KB
	creator := snapshot.NewCreator(repoPath, model.EngineReflinkCopy)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := creator.Create("main", "bench snapshot", nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSnapshotCreation_ReflinkEngine_Medium benchmarks snapshot creation with medium payload using reflink engine.
func BenchmarkSnapshotCreation_ReflinkEngine_Medium(b *testing.B) {
	repoPath := setupBenchRepo(b, 1024*1024) // 1MB
	creator := snapshot.NewCreator(repoPath, model.EngineReflinkCopy)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := creator.Create("main", "bench snapshot", nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSnapshotCreation_MultiFile benchmarks snapshot creation with multiple files.
func BenchmarkSnapshotCreation_MultiFile(b *testing.B) {
	repoPath := setupBenchRepoWithFiles(b, 100) // 100 files
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := creator.Create("main", "bench snapshot", nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSnapshotCreation_MultiFile_Large benchmarks snapshot creation with many files.
func BenchmarkSnapshotCreation_MultiFile_Large(b *testing.B) {
	repoPath := setupBenchRepoWithFiles(b, 1000) // 1000 files
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := creator.Create("main", "bench snapshot", nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDescriptorSerialization benchmarks descriptor serialization.
func BenchmarkDescriptorSerialization(b *testing.B) {
	desc := &model.Descriptor{
		SnapshotID:         "0000000000000-abc12345",
		ParentID:           nil,
		WorktreeName:       "main",
		CreatedAt:          time.Now(),
		Note:               "benchmark snapshot descriptor with some content",
		Tags:               []string{"v1.0", "release", "stable"},
		Engine:             model.EngineCopy,
		PayloadRootHash:    "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		DescriptorChecksum: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		IntegrityState:     model.IntegrityVerified,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(desc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDescriptorDeserialization benchmarks descriptor deserialization.
func BenchmarkDescriptorDeserialization(b *testing.B) {
	desc := &model.Descriptor{
		SnapshotID:         "0000000000000-abc12345",
		ParentID:           nil,
		WorktreeName:       "main",
		CreatedAt:          time.Now(),
		Note:               "benchmark snapshot descriptor with some content",
		Tags:               []string{"v1.0", "release", "stable"},
		Engine:             model.EngineCopy,
		PayloadRootHash:    "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		DescriptorChecksum: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		IntegrityState:     model.IntegrityVerified,
	}
	data, err := json.Marshal(desc)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result model.Descriptor
		if err := json.Unmarshal(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLoadDescriptor benchmarks loading a descriptor from disk.
func BenchmarkLoadDescriptor(b *testing.B) {
	repoPath := setupBenchRepo(b, 1024)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "bench snapshot", nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := snapshot.LoadDescriptor(repoPath, desc.SnapshotID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkVerifySnapshot_ChecksumOnly benchmarks snapshot verification without payload hash.
func BenchmarkVerifySnapshot_ChecksumOnly(b *testing.B) {
	repoPath := setupBenchRepo(b, 1024)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "bench snapshot", nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := snapshot.VerifySnapshot(repoPath, desc.SnapshotID, false)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkVerifySnapshot_WithPayloadHash benchmarks snapshot verification with payload hash.
func BenchmarkVerifySnapshot_WithPayloadHash(b *testing.B) {
	repoPath := setupBenchRepo(b, 1024*100) // 100KB
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "bench snapshot", nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := snapshot.VerifySnapshot(repoPath, desc.SnapshotID, true)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkComputeDescriptorChecksum benchmarks descriptor checksum computation.
func BenchmarkComputeDescriptorChecksum(b *testing.B) {
	desc := &model.Descriptor{
		SnapshotID:      "0000000000000-abc12345",
		ParentID:        nil,
		WorktreeName:    "main",
		CreatedAt:       time.Now(),
		Note:            "benchmark snapshot descriptor with some content",
		Tags:            []string{"v1.0", "release", "stable"},
		Engine:          model.EngineCopy,
		PayloadRootHash: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		IntegrityState:  model.IntegrityVerified,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := integrity.ComputeDescriptorChecksum(desc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkListAll_Empty benchmarks listing snapshots in an empty repository.
func BenchmarkListAll_Empty(b *testing.B) {
	repoPath := setupBenchRepo(b, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := snapshot.ListAll(repoPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkListAll_Single benchmarks listing snapshots with one snapshot.
func BenchmarkListAll_Single(b *testing.B) {
	repoPath := setupBenchRepo(b, 1024)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	_, err := creator.Create("main", "bench snapshot", nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := snapshot.ListAll(repoPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkListAll_Many benchmarks listing snapshots with many snapshots.
func BenchmarkListAll_Many(b *testing.B) {
	repoPath := setupBenchRepo(b, 1024)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	// Create 50 snapshots
	for i := 0; i < 50; i++ {
		if _, err := creator.Create("main", "bench snapshot", nil); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := snapshot.ListAll(repoPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFind_ByTag benchmarks finding snapshots by tag.
func BenchmarkFind_ByTag(b *testing.B) {
	repoPath := setupBenchRepo(b, 1024)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	// Create snapshots with various tags
	tags := []string{"v1.0", "v1.1", "v2.0", "release", "wip"}
	for _, tag := range tags {
		_, err := creator.Create("main", "bench snapshot", []string{tag})
		if err != nil {
			b.Fatal(err)
		}
	}

	opts := snapshot.FilterOptions{HasTag: "release"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := snapshot.Find(repoPath, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFind_ByWorktree benchmarks finding snapshots by worktree name.
func BenchmarkFind_ByWorktree(b *testing.B) {
	repoPath := setupBenchRepo(b, 1024)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	// Create multiple snapshots
	for i := 0; i < 10; i++ {
		_, err := creator.Create("main", "bench snapshot", nil)
		if err != nil {
			b.Fatal(err)
		}
	}

	opts := snapshot.FilterOptions{WorktreeName: "main"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := snapshot.Find(repoPath, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFindByTag benchmarks finding a snapshot by tag.
func BenchmarkFindByTag(b *testing.B) {
	repoPath := setupBenchRepo(b, 1024)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	_, err := creator.Create("main", "release snapshot", []string{"v1.0"})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := snapshot.FindByTag(repoPath, "v1.0")
		if err != nil {
			b.Fatal(err)
		}
	}
}
