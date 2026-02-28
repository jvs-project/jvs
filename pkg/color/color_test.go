package color

import (
	"os"
	"sync"
	"testing"
)

func TestEnabled(t *testing.T) {
	origEnabled := state.enabled.Load()
	origOverridden := state.overridden.Load()

	Enable()
	if !Enabled() {
		t.Error("expected colors to be enabled")
	}

	Disable()
	if Enabled() {
		t.Error("expected colors to be disabled")
	}

	state.enabled.Store(origEnabled)
	state.overridden.Store(origOverridden)
}

func TestEnableDisable(t *testing.T) {
	origEnabled := state.enabled.Load()
	origOverridden := state.overridden.Load()

	Enable()
	if !Enabled() {
		t.Error("expected colors to be enabled after Enable()")
	}

	Disable()
	if Enabled() {
		t.Error("expected colors to be disabled after Disable()")
	}

	state.enabled.Store(origEnabled)
	state.overridden.Store(origOverridden)
}

func TestColorFuncs(t *testing.T) {
	Enable()

	tests := []struct {
		name     string
		fn       func(string) string
		input    string
		contains string
	}{
		{"Redf", Redf, "test", Red},
		{"Greenf", Greenf, "test", Green},
		{"Yellowf", Yellowf, "test", Yellow},
		{"Bluef", Bluef, "test", Blue},
		{"Cyanf", Cyanf, "test", Cyan},
		{"Boldf", Boldf, "test", Bold},
		{"Dimf", Dimf, "test", DimCode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if !containsString(result, tt.contains) {
				t.Errorf("%s(%q) = %q, expected to contain %q", tt.name, tt.input, result, tt.contains)
			}
			// Should always end with Reset
			if !containsString(result, Reset) {
				t.Errorf("%s(%q) = %q, expected to contain reset code", tt.name, tt.input, result)
			}
		})
	}
}

func TestColorFuncsDisabled(t *testing.T) {
	Disable()

	tests := []struct {
		name  string
		fn    func(string) string
		input string
	}{
		{"Redf", Redf, "test"},
		{"Greenf", Greenf, "test"},
		{"Success", Success, "test"},
		{"Error", Error, "test"},
		{"Warning", Warning, "test"},
		{"Info", Info, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if result != tt.input {
				t.Errorf("%s(%q) = %q, expected %q (no color when disabled)", tt.name, tt.input, result, tt.input)
			}
		})
	}
}

func TestSpecializedFormatters(t *testing.T) {
	Enable()

	tests := []struct {
		name  string
		fn    func(string) string
		input string
		color string
	}{
		{"Success", Success, "ok", Green},
		{"Error", Error, "fail", Red},
		{"Warning", Warning, "warn", Yellow},
		{"Info", Info, "info", Cyan},
		{"SnapshotID", SnapshotID, "abc123", Cyan},
		{"Tag", Tag, "v1.0", Blue},
		{"Header", Header, "Title", Bold},
		{"Dim", Dim, "subtle", DimCode},
		{"Highlight", Highlight, "important", Yellow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if !containsString(result, tt.color) {
				t.Errorf("%s(%q) = %q, expected to contain color code", tt.name, tt.input, result)
			}
		})
	}
}

func TestFormattedFunctions(t *testing.T) {
	Enable()

	if result := Successf("test %d", 123); !containsString(result, Green) {
		t.Errorf("Successf() should contain green color code, got %q", result)
	}

	if result := Errorf("err %s", "x"); !containsString(result, Red) {
		t.Errorf("Errorf() should contain red color code, got %q", result)
	}

	if result := Warningf("warn %d", 42); !containsString(result, Yellow) {
		t.Errorf("Warningf() should contain yellow color code, got %q", result)
	}

	if result := Infof("info %s", "test"); !containsString(result, Cyan) {
		t.Errorf("Infof() should contain cyan color code, got %q", result)
	}
}

func TestCode(t *testing.T) {
	Enable()

	result := Code("jvs init")
	if !containsString(result, Bold) {
		t.Errorf("Code() should contain bold code, got %q", result)
	}
	if !containsString(result, Reset) {
		t.Errorf("Code() should contain reset code, got %q", result)
	}

	Disable()
	result = Code("test")
	if result != "test" {
		t.Errorf("Code() disabled should return plain text, got %q", result)
	}
	Enable()
}

func TestInitRespectsNoColorEnv(t *testing.T) {
	origNoColor, exists := os.LookupEnv("NO_COLOR")

	os.Setenv("NO_COLOR", "1")
	state.overridden.Store(false)
	state.enabled.Store(true)
	state.once = sync.Once{}

	Init(false)
	if Enabled() {
		t.Error("expected colors to be disabled when NO_COLOR is set")
	}

	if exists {
		os.Setenv("NO_COLOR", origNoColor)
	} else {
		os.Unsetenv("NO_COLOR")
	}
	state.once = sync.Once{}
}

func TestInitRespectsNoColorFlag(t *testing.T) {
	state.overridden.Store(false)
	state.enabled.Store(true)
	state.once = sync.Once{}

	Init(true)
	if Enabled() {
		t.Error("expected colors to be disabled when noColorFlag is true")
	}

	state.once = sync.Once{}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
