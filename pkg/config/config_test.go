package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jvs-project/jvs/pkg/model"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.DefaultEngine != "" {
		t.Errorf("expected empty engine, got %s", cfg.DefaultEngine)
	}
	if cfg.DefaultTags != nil {
		t.Errorf("expected nil tags, got %v", cfg.DefaultTags)
	}
	if cfg.OutputFormat != "" {
		t.Errorf("expected empty output_format, got %s", cfg.OutputFormat)
	}
}

func TestLoad_NotExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jvs-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Error("expected config, got nil")
	}
	// Should return default config
	if cfg.DefaultEngine != "" {
		t.Errorf("expected empty default_engine, got %s", cfg.DefaultEngine)
	}
}

func TestLoad_Exists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jvs-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .jvs directory and config file
	jvsDir := filepath.Join(tmpDir, ".jvs")
	if err := os.MkdirAll(jvsDir, 0755); err != nil {
		t.Fatal(err)
	}

	configContent := `
default_engine: juicefs-clone
default_tags:
  - auto
  - dev
output_format: json
progress_enabled: false
`
	configPath := filepath.Join(jvsDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cfg.DefaultEngine != "juicefs-clone" {
		t.Errorf("expected default_engine 'juicefs-clone', got %s", cfg.DefaultEngine)
	}
	if len(cfg.DefaultTags) != 2 {
		t.Errorf("expected 2 default tags, got %d", len(cfg.DefaultTags))
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("expected output_format 'json', got %s", cfg.OutputFormat)
	}
	if cfg.ProgressEnabled == nil || *cfg.ProgressEnabled != false {
		t.Error("expected progress_enabled to be false")
	}
}

func TestSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jvs-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	enabled := false
	cfg := &Config{
		DefaultEngine:   "copy",
		DefaultTags:     []string{"test"},
		OutputFormat:    "text",
		ProgressEnabled: &enabled,
	}

	if err := Save(tmpDir, cfg); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tmpDir, ".jvs", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file not created")
	}

	// Load and verify
	loaded, err := Load(tmpDir)
	if err != nil {
		t.Errorf("unexpected error loading: %v", err)
	}
	if loaded.DefaultEngine != "copy" {
		t.Errorf("expected engine 'copy', got %s", loaded.DefaultEngine)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jvs-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	jvsDir := filepath.Join(tmpDir, ".jvs")
	if err := os.MkdirAll(jvsDir, 0755); err != nil {
		t.Fatal(err)
	}

	invalidYAML := `
default_engine: [this is invalid yaml
`
	configPath := filepath.Join(jvsDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = Load(tmpDir)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoad_WithNewConfigFields(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jvs-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .jvs directory and config file with new fields
	jvsDir := filepath.Join(tmpDir, ".jvs")
	if err := os.MkdirAll(jvsDir, 0755); err != nil {
		t.Fatal(err)
	}

	configContent := `
default_engine: juicefs-clone
default_tags:
  - auto
  - dev
output_format: json
progress_enabled: false
`
	configPath := filepath.Join(jvsDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(cfg.DefaultEngine) != "juicefs-clone" {
		t.Errorf("expected default_engine 'juicefs-clone', got %s", cfg.DefaultEngine)
	}
	if len(cfg.DefaultTags) != 2 {
		t.Errorf("expected 2 default tags, got %d", len(cfg.DefaultTags))
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("expected output_format 'json', got %s", cfg.OutputFormat)
	}
	if cfg.ProgressEnabled == nil || *cfg.ProgressEnabled != false {
		t.Error("expected progress_enabled to be false")
	}
}

func TestConfig_GetDefaultEngine(t *testing.T) {
	cfg := &Config{DefaultEngine: "juicefs-clone"}
	if cfg.GetDefaultEngine() != "juicefs-clone" {
		t.Error("expected juicefs-clone")
	}

	cfg2 := &Config{DefaultEngine: ""}
	if cfg2.GetDefaultEngine() != "" {
		t.Error("expected empty string for unset engine")
	}

	cfg3 := &Config{DefaultEngine: "auto"}
	if cfg3.GetDefaultEngine() != "" {
		t.Error("expected empty string for auto engine")
	}
}

func TestConfig_GetDefaultTags(t *testing.T) {
	cfg := &Config{DefaultTags: []string{"auto", "dev"}}
	tags := cfg.GetDefaultTags()
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}

	cfg2 := &Config{DefaultTags: nil}
	if cfg2.GetDefaultTags() != nil {
		t.Error("expected nil for unset tags")
	}
}

func TestConfig_GetOutputFormat(t *testing.T) {
	cfg := &Config{OutputFormat: "json"}
	if cfg.GetOutputFormat() != "json" {
		t.Error("expected json")
	}

	cfg2 := &Config{OutputFormat: ""}
	if cfg2.GetOutputFormat() != "" {
		t.Error("expected empty string for unset format")
	}
}

func TestConfig_Set(t *testing.T) {
	cfg := &Config{}

	// Set default_engine
	if err := cfg.Set("default_engine", "juicefs-clone"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cfg.DefaultEngine != "juicefs-clone" {
		t.Errorf("expected juicefs-clone, got %s", cfg.DefaultEngine)
	}

	// Set default_tags
	if err := cfg.Set("default_tags", "[\"auto\", \"dev\"]"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(cfg.DefaultTags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(cfg.DefaultTags))
	}

	// Set output_format
	if err := cfg.Set("output_format", "json"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("expected json, got %s", cfg.OutputFormat)
	}

	// Set progress_enabled
	if err := cfg.Set("progress_enabled", "true"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cfg.ProgressEnabled == nil || !*cfg.ProgressEnabled {
		t.Error("expected progress_enabled to be true")
	}

	// Test invalid key
	if err := cfg.Set("invalid_key", "value"); err == nil {
		t.Error("expected error for invalid key")
	}

	// Test invalid progress_enabled value
	if err := cfg.Set("progress_enabled", "invalid"); err == nil {
		t.Error("expected error for invalid progress_enabled value")
	}
}

func TestConfig_Get(t *testing.T) {
	enabled := true
	cfg := &Config{
		DefaultEngine:   "juicefs-clone",
		DefaultTags:     []string{"auto", "dev"},
		OutputFormat:    "json",
		ProgressEnabled: &enabled,
	}

	// Get each value
	val, err := cfg.Get("default_engine")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != "juicefs-clone" {
		t.Errorf("expected juicefs-clone, got %s", val)
	}

	val, err = cfg.Get("default_tags")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val == "" || val == "[]" {
		t.Error("expected tags list")
	}

	val, err = cfg.Get("output_format")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != "json" {
		t.Errorf("expected json, got %s", val)
	}

	val, err = cfg.Get("progress_enabled")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != "true" {
		t.Errorf("expected true, got %s", val)
	}

	// Test invalid key
	_, err = cfg.Get("invalid_key")
	if err == nil {
		t.Error("expected error for invalid key")
	}
}

func TestConfig_Validate(t *testing.T) {
	cfg := &Config{}

	// Valid engines
	validEngines := []model.EngineType{"juicefs-clone", "reflink-copy", "copy", "auto"}
	for _, engine := range validEngines {
		cfg.DefaultEngine = engine
		if err := cfg.validate(); err != nil {
			t.Errorf("unexpected error for engine %s: %v", engine, err)
		}
	}

	// Invalid engine
	cfg.DefaultEngine = "invalid"
	if err := cfg.validate(); err == nil {
		t.Error("expected error for invalid engine")
	}

	// Valid output formats
	validFormats := []string{"text", "json", ""}
	for _, format := range validFormats {
		cfg.DefaultEngine = "" // reset to valid
		cfg.OutputFormat = format
		if err := cfg.validate(); err != nil {
			t.Errorf("unexpected error for format %s: %v", format, err)
		}
	}

	// Invalid output format
	cfg.OutputFormat = "invalid"
	if err := cfg.validate(); err == nil {
		t.Error("expected error for invalid output format")
	}
}

func TestKeys(t *testing.T) {
	keys := Keys()
	if len(keys) != 4 {
		t.Errorf("expected 4 keys, got %d", len(keys))
	}

	expectedKeys := map[string]bool{
		"default_engine":   false,
		"default_tags":     false,
		"output_format":    false,
		"progress_enabled": false,
	}

	for _, key := range keys {
		if _, ok := expectedKeys[key]; !ok {
			t.Errorf("unexpected key: %s", key)
		}
		expectedKeys[key] = true
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("missing key: %s", key)
		}
	}
}
