//go:build go1.18
// +build go1.18

// Fuzzing tests for JVS critical functions
//
// This package contains fuzz targets for testing critical parsing and validation
// functions with randomized inputs. Fuzzing helps find edge cases, panics, and
// security vulnerabilities that might be missed with traditional unit tests.
//
// Running fuzz tests:
//   go test -fuzz=FuzzParseSnapshotID -fuzztime=30s ./test/fuzz/...
//   go test -fuzz=. -fuzztime=1m ./test/fuzz/...
//
// For more information on Go fuzzing, see:
// https://go.dev/doc/tutorial/fuzz

package fuzz

import (
	"encoding/json"
	"testing"

	"github.com/jvs-project/jvs/pkg/jsonutil"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/jvs-project/jvs/pkg/pathutil"
)

// FuzzValidateName tests worktree name validation with random inputs.
//
// This fuzz target ensures ValidateName handles arbitrary input without panicking
// and correctly validates/rejects inputs according to the specification.
func FuzzValidateName(f *testing.F) {
	// Seed corpus with edge cases
	f.Add("")                           // empty string
	f.Add("valid-name-123")             // valid name
	f.Add("..")                          // path traversal
	f.Add("../escape")                   // path traversal attempt
	f.Add("name/with/slash")             // invalid separator
	f.Add(`name\with\backslash`)        // invalid separator
	f.Add("name\twith\tcontrol")         // control character
	f.Add("name\x00null")               // null byte
	f.Add("a")                           // single char
	f.Add("aa")                          // two chars
	f.Add("a.b")                         // dot
	f.Add("a-b")                         // hyphen
	f.Add("a_b")                         // underscore
	f.Add("very-long-name-with-many-chars-that-still-valid-123456789")

	f.Fuzz(func(t *testing.T, name string) {
		// Should not panic on any input
		err := pathutil.ValidateName(name)

		// If we got a result, it should be consistent
		// (same input should give same result)
		err2 := pathutil.ValidateName(name)
		if (err == nil) != (err2 == nil) {
			t.Errorf("inconsistent validation for %q: %v vs %v", name, err, err2)
		}
	})
}

// FuzzValidateTag tests tag validation with random inputs.
//
// Tags have the same validation rules as worktree names.
func FuzzValidateTag(f *testing.F) {
	// Seed corpus
	f.Add("")
	f.Add("v1.0")
	f.Add("stable")
	f.Add("release_2024")
	f.Add("tag/with/slash")
	f.Add("../escape")
	f.Add("tag with spaces")
	f.Add("tag\twith\tcontrol")

	f.Fuzz(func(t *testing.T, tag string) {
		// Should not panic
		err := pathutil.ValidateTag(tag)

		// Consistency check
		err2 := pathutil.ValidateTag(tag)
		if (err == nil) != (err2 == nil) {
			t.Errorf("inconsistent validation for %q: %v vs %v", tag, err, err2)
		}
	})
}

// FuzzParseSnapshotID tests snapshot ID parsing with random inputs.
//
// Snapshot IDs have the format: <unix_ms>-<rand8hex>
// This fuzz target ensures parsing doesn't crash on malformed input.
func FuzzParseSnapshotID(f *testing.F) {
	// Seed corpus
	f.Add("1708300800000-a3f7c1b2")        // valid
	f.Add("")                               // empty
	f.Add("-")                              // just separator
	f.Add("123")                            // too short
	f.Add("1708300800000-")                 // missing random part
	f.Add("-a3f7c1b2")                      // missing timestamp
	f.Add("1708300800000-abc")             // random too short
	f.Add("not-a-number-abc")              // invalid timestamp
	f.Add("1708300800000-gghhii")           // invalid hex
	f.Add("1708300800000-xxxxxxxx")         // valid hex format

	f.Fuzz(func(t *testing.T, data string) {
		// Convert to SnapshotID type - should not panic
		id := model.SnapshotID(data)

		// Calling methods should not panic
		_ = id.ShortID()
		_ = id.String()

		// String() should return the original input
		if id.String() != data {
			t.Errorf("String() returned %q, expected %q", id.String(), data)
		}
	})
}

// FuzzCanonicalMarshal tests canonical JSON marshaling with random inputs.
//
// This ensures the JSON canonicalization doesn't panic and produces
// consistent output for the same input.
func FuzzCanonicalMarshal(f *testing.F) {
	// Seed corpus with various Go types
	f.Add(int(42))
	f.Add(int64(12345678901234))
	f.Add("test string")
	f.Add(true)
	f.Add(false)
	f.Add(nil)
	f.Add([]int{1, 2, 3})
	f.Add(map[string]any{"a": 1, "b": 2})
	f.Add([]any{1, "two", map[string]int{"c": 3}})

	// Add JSON-like structures
	f.Add([]byte(`{"name":"test","value":123}`))
	f.Add([]byte(`{"nested":{"key":"value"}}`))
	f.Add([]byte(`[1,2,3]`))
	f.Add([]byte(`null`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// First, try to interpret data as JSON
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			// Not valid JSON, skip this iteration
			return
		}

		// CanonicalMarshal should not panic on valid JSON-derived data
		result1, err := jsonutil.CanonicalMarshal(v)
		if err != nil {
			// Some inputs may legitimately fail (e.g., channels, functions)
			// This is okay as long as we don't panic
			return
		}

		// Result should be valid JSON
		if !json.Valid(result1) {
			t.Errorf("CanonicalMarshal produced invalid JSON: %q", result1)
		}

		// Result should be deterministic (same input → same output)
		result2, err := jsonutil.CanonicalMarshal(v)
		if err != nil {
			t.Errorf("CanonicalMarshal inconsistent error: %v", err)
			return
		}
		if string(result1) != string(result2) {
			t.Errorf("CanonicalMarshal not deterministic: %q vs %q", result1, result2)
		}
	})
}

// FuzzDescriptorJSON tests descriptor JSON marshaling/unmarshaling with random data.
//
// This ensures the Descriptor struct handles various JSON inputs without panicking.
func FuzzDescriptorJSON(f *testing.F) {
	// Seed corpus with valid descriptors
	validDesc := model.Descriptor{
		SnapshotID:      "1708300800000-a3f7c1b2",
		WorktreeName:    "main",
		Engine:          model.EngineCopy,
		PayloadRootHash: "abc123",
	}
	validJSON, _ := json.Marshal(validDesc)

	f.Add(validJSON)
	f.Add([]byte(`{}`))                           // empty object
	f.Add([]byte(`{"snapshot_id":"test"}`))      // minimal
	f.Add([]byte(`invalid json`))                // invalid JSON
	f.Add([]byte(`{"snapshot_id":123}`))         // wrong type
	f.Add([]byte(`{"extra":"field"}`))           // extra field

	f.Fuzz(func(t *testing.T, data []byte) {
		// Unmarshal should not panic
		var desc model.Descriptor
		err := json.Unmarshal(data, &desc)

		// If unmarshal succeeded, marshal should also succeed and not panic
		if err == nil {
			_, err := json.Marshal(desc)
			if err != nil {
				t.Errorf("Marshal failed after successful Unmarshal: %v", err)
			}
		}
	})
}

// FuzzSnapshotIDString tests SnapshotID.String() consistency.
func FuzzSnapshotIDString(f *testing.F) {
	// Seed corpus
	f.Add("1708300800000-a3f7c1b2")
	f.Add("")
	f.Add("-")
	f.Add("abc")

	f.Fuzz(func(t *testing.T, data string) {
		id := model.SnapshotID(data)

		// String() should always return the original input
		if got := id.String(); got != data {
			t.Errorf("String() = %q, want %q", got, data)
		}

		// ShortID should return first 8 chars or less
		short := id.ShortID()
		if len(short) > 8 {
			t.Errorf("ShortID() returned %d chars, want at most 8", len(short))
		}
		if len(short) > len(data) {
			t.Errorf("ShortID() longer than input: %q vs %q", short, data)
		}
	})
}

// FuzzNewSnapshotID tests snapshot ID generation doesn't panic.
// Note: This is a regular test, not fuzz, since it uses randomness internally.
func TestNewSnapshotID(t *testing.T) {
	// Generate many IDs and check format
	for i := 0; i < 1000; i++ {
		id := model.NewSnapshotID()

		str := id.String()

		// Should contain exactly one hyphen
		hyphenCount := 0
		for _, c := range str {
			if c == '-' {
				hyphenCount++
			}
		}
		if hyphenCount != 1 {
			t.Errorf("expected 1 hyphen in SnapshotID, got %d: %s", hyphenCount, str)
		}

		// ShortID should be 8 chars
		if len(id.ShortID()) != 8 {
			t.Errorf("ShortID() returned %d chars, want 8", len(id.ShortID()))
		}
	}
}
