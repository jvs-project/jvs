package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Engine != "auto" {
		t.Errorf("expected engine 'auto', got %s", cfg.Engine)
	}
	if cfg.RetentionPolicy.KeepMinSnapshots != 10 {
		t.Errorf("expected KeepMinSnapshots 10, got %d", cfg.RetentionPolicy.KeepMinSnapshots)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("expected logging level 'info', got %s", cfg.Logging.Level)
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
	if cfg.Engine != "auto" {
		t.Errorf("expected default engine, got %s", cfg.Engine)
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
engine: juicefs
retention_policy:
  keep_min_snapshots: 20
  keep_min_age: 48h
logging:
  level: debug
  format: json
`
	configPath := filepath.Join(jvsDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cfg.Engine != "juicefs" {
		t.Errorf("expected engine 'juicefs', got %s", cfg.Engine)
	}
	if cfg.RetentionPolicy.KeepMinSnapshots != 20 {
		t.Errorf("expected KeepMinSnapshots 20, got %d", cfg.RetentionPolicy.KeepMinSnapshots)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected logging level 'debug', got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("expected logging format 'json', got %s", cfg.Logging.Format)
	}
}

func TestSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jvs-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &Config{
		Engine: "reflink",
		RetentionPolicy: RetentionPolicyConfig{
			KeepMinSnapshots: 15,
			KeepMinAge:       "12h",
		},
		Logging: LoggingConfig{
			Level:  "warn",
			Format: "text",
		},
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
	if loaded.Engine != "reflink" {
		t.Errorf("expected engine 'reflink', got %s", loaded.Engine)
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
engine: [this is invalid yaml
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
