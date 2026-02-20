//go:build conformance

package conformance

import (
	"strings"
	"testing"
)

// Test 1: Lock acquire succeeds
func TestLock_Acquire(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	stdout, stderr, code := runJVSInRepo(t, repoPath, "lock", "acquire")
	if code != 0 {
		t.Fatalf("lock acquire failed: %s", stderr)
	}
	if !strings.Contains(stdout, "Lock acquired") {
		t.Errorf("expected 'Lock acquired' in output, got: %s", stdout)
	}
}

// Test 2: Lock conflict on double acquire
func TestLock_ConflictOnDoubleAcquire(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// First acquire
	_, _, code := runJVSInRepo(t, repoPath, "lock", "acquire")
	if code != 0 {
		t.Fatalf("first lock acquire failed")
	}

	// Second acquire should fail
	_, _, code = runJVSInRepo(t, repoPath, "lock", "acquire")
	if code == 0 {
		t.Error("second lock acquire should have failed")
	}
}

// Test 3: Lock release succeeds
func TestLock_Release(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Acquire
	runJVSInRepo(t, repoPath, "lock", "acquire")

	// Release
	stdout, stderr, code := runJVSInRepo(t, repoPath, "lock", "release")
	if code != 0 {
		t.Fatalf("lock release failed: %s", stderr)
	}
	if !strings.Contains(stdout, "Lock released") {
		t.Errorf("expected 'Lock released' in output, got: %s", stdout)
	}

	// Should be able to acquire again
	_, _, code = runJVSInRepo(t, repoPath, "lock", "acquire")
	if code != 0 {
		t.Error("should be able to acquire after release")
	}
}
