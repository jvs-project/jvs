// Package config provides configuration file support for JVS.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the JVS configuration.
type Config struct {
	Engine          string                `yaml:"engine"`
	RetentionPolicy RetentionPolicyConfig `yaml:"retention_policy"`
	Logging         LoggingConfig         `yaml:"logging"`
}

// RetentionPolicyConfig configures GC retention.
type RetentionPolicyConfig struct {
	KeepMinSnapshots int    `yaml:"keep_min_snapshots"`
	KeepMinAge       string `yaml:"keep_min_age"`
}

// LoggingConfig configures logging behavior.
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"` // json, text
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Engine: "auto",
		RetentionPolicy: RetentionPolicyConfig{
			KeepMinSnapshots: 10,
			KeepMinAge:       "24h",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// Load loads configuration from .jvs/config.yaml.
// Returns default config if file doesn't exist.
func Load(repoRoot string) (*Config, error) {
	cfg := Default()
	cfgPath := filepath.Join(repoRoot, ".jvs", "config.yaml")

	data, err := os.ReadFile(cfgPath)
	if os.IsNotExist(err) {
		return cfg, nil // No config file is OK, use defaults
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

// Save writes configuration to .jvs/config.yaml.
func Save(repoRoot string, cfg *Config) error {
	cfgPath := filepath.Join(repoRoot, ".jvs", "config.yaml")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
