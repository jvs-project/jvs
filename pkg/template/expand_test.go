package template

import (
	"testing"
)

func TestExpand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		vars     map[string]string
		contains []string // strings that should be in the output
	}{
		{
			name:     "date placeholder",
			input:    "Snapshot {date}",
			contains: []string{"Snapshot ", "20"}, // year starts with 20
		},
		{
			name:     "datetime placeholder",
			input:    "Snapshot {datetime}",
			contains: []string{"Snapshot ", "-", ":"},
		},
		{
			name:     "user placeholder",
			input:    "User {user}",
			contains: []string{"User "},
		},
		{
			name:     "hostname placeholder",
			input:    "Host {hostname}",
			contains: []string{"Host "},
		},
		{
			name:     "arch placeholder",
			input:    "Arch {arch}",
			contains: []string{"Arch "},
		},
		{
			name:     "multiple placeholders",
			input:    "{date} {time} {user}",
			contains: []string{"-", ":"},
		},
		{
			name:     "custom var",
			input:    "Branch {branch}",
			vars:     map[string]string{"branch": "main"},
			contains: []string{"Branch main"},
		},
		{
			name:     "custom var overrides built-in",
			input:    "Date {date}",
			vars:     map[string]string{"date": "2024-01-01"},
			contains: []string{"Date 2024-01-01"},
		},
		{
			name:     "no placeholders",
			input:    "Simple text",
			contains: []string{"Simple text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Expand(tt.input, tt.vars)

			// Check that expected substrings are present
			for _, contain := range tt.contains {
				if !containsIn(result, contain) {
					t.Errorf("Expand(%q) = %q, does not contain %q", tt.input, result, contain)
				}
			}
		})
	}
}

func TestExpandNote(t *testing.T) {
	note := ExpandNote("Checkpoint: {datetime}")
	if note == "Checkpoint: {datetime}" {
		t.Error("placeholder not expanded")
	}
	if note == "" {
		t.Error("empty result")
	}
}

func containsIn(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
