package engine

import (
	"os"
	"os/exec"

	"github.com/jvs-project/jvs/pkg/model"
)

// JuiceFSEngine performs clone using `juicefs clone` command.
type JuiceFSEngine struct {
	CopyEngine *CopyEngine // Fallback
}

// NewJuiceFSEngine creates a new JuiceFSEngine.
func NewJuiceFSEngine() *JuiceFSEngine {
	return &JuiceFSEngine{
		CopyEngine: NewCopyEngine(),
	}
}

// Name returns the engine type.
func (e *JuiceFSEngine) Name() model.EngineType {
	return model.EngineJuiceFSClone
}

// Clone performs a juicefs clone if available, falls back to copy otherwise.
func (e *JuiceFSEngine) Clone(src, dst string) (*CloneResult, error) {
	// Check if juicefs command is available
	if !e.isJuiceFSAvailable() {
		// Fall back to copy engine
		result, err := e.CopyEngine.Clone(src, dst)
		if err != nil {
			return nil, err
		}
		result.Degraded = true
		result.Degradations = append(result.Degradations, "juicefs-not-available")
		return result, nil
	}

	// Check if source is on JuiceFS
	if !e.isOnJuiceFS(src) {
		// Fall back to copy engine
		result, err := e.CopyEngine.Clone(src, dst)
		if err != nil {
			return nil, err
		}
		result.Degraded = true
		result.Degradations = append(result.Degradations, "not-on-juicefs")
		return result, nil
	}

	// Execute juicefs clone
	cmd := exec.Command("juicefs", "clone", src, dst, "-p")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Fall back to copy on failure
		result, err := e.CopyEngine.Clone(src, dst)
		if err != nil {
			return nil, err
		}
		result.Degraded = true
		result.Degradations = append(result.Degradations, "juicefs-clone-failed")
		return result, nil
	}

	return &CloneResult{Degraded: false}, nil
}

func (e *JuiceFSEngine) isJuiceFSAvailable() bool {
	_, err := exec.LookPath("juicefs")
	return err == nil
}

func (e *JuiceFSEngine) isOnJuiceFS(path string) bool {
	// Check if path is on JuiceFS by looking at mount info
	// This is a simplified check - in production would check /proc/mounts
	var s syscallStat_t
	if err := doStat(path, &s); err != nil {
		return false
	}
	// JuiceFS typically has specific filesystem type
	// For now, return true if juicefs command exists
	return e.isJuiceFSAvailable()
}

// Minimal syscall stat for filesystem detection
type syscallStat_t struct {
	Dev  uint64
	Ino  uint64
	Mode uint32
	// ... other fields omitted
}

func doStat(path string, s *syscallStat_t) error {
	// Simplified - would use syscall.Stat in production
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	s.Dev = uint64(info.ModTime().UnixNano()) // placeholder
	return nil
}

// Engine detection function
func DetectEngine(repoRoot string) (Engine, error) {
	// Check environment variable first
	if engineType := os.Getenv("JVS_ENGINE"); engineType != "" {
		switch engineType {
		case "juicefs":
			return NewJuiceFSEngine(), nil
		case "reflink":
			return NewReflinkEngine(), nil
		case "copy":
			return NewCopyEngine(), nil
		}
	}

	// Auto-detect based on filesystem
	// 1. Check if on JuiceFS
	juicefsEngine := NewJuiceFSEngine()
	if juicefsEngine.isOnJuiceFS(repoRoot) && juicefsEngine.isJuiceFSAvailable() {
		return juicefsEngine, nil
	}

	// 2. Check if reflink is supported (btrfs, xfs, apfs)
	reflinkEngine := NewReflinkEngine()
	// Try a test reflink
	testDir, err := os.MkdirTemp("", "jvs-reflink-test-")
	if err == nil {
		testFile := testDir + "/test"
		os.WriteFile(testFile, []byte("test"), 0644)
		testClone := testDir + "/clone"
		info, _ := os.Stat(testFile)
		if reflinkEngine.reflinkFile(testFile, testClone, info) == nil {
			os.RemoveAll(testDir)
			return reflinkEngine, nil
		}
		os.RemoveAll(testDir)
	}

	// 3. Fall back to copy
	return NewCopyEngine(), nil
}
