// Package config provides configuration file support for JVS.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jvs-project/jvs/pkg/model"
	"github.com/jvs-project/jvs/pkg/webhook"
	"gopkg.in/yaml.v3"
)

var (
	// cache is a per-repo config cache
	cache   = make(map[string]*Config)
	cacheMu sync.RWMutex
)

// Config represents the JVS configuration.
type Config struct {
	// DefaultEngine is the default snapshot engine to use.
	DefaultEngine model.EngineType `yaml:"default_engine,omitempty"`

	// DefaultTags are tags automatically added to each snapshot.
	DefaultTags []string `yaml:"default_tags,omitempty"`

	// OutputFormat is the default output format (text or json).
	OutputFormat string `yaml:"output_format,omitempty"`

	// ProgressEnabled enables progress bars by default.
	ProgressEnabled *bool `yaml:"progress_enabled,omitempty"`

	// SnapshotTemplates defines pre-configured snapshot patterns.
	SnapshotTemplates map[string]SnapshotTemplate `yaml:"snapshot_templates,omitempty"`

	// Webhooks defines webhook notification configuration.
	Webhooks *WebhookConfig `yaml:"webhooks,omitempty"`

	// Legacy fields for backward compatibility
	Engine          string                `yaml:"engine,omitempty"`
	RetentionPolicy RetentionPolicyConfig `yaml:"retention_policy,omitempty"`
	Logging         LoggingConfig         `yaml:"logging,omitempty"`
}

// SnapshotTemplate defines a pre-configured snapshot pattern.
type SnapshotTemplate struct {
	// Note is the note template for the snapshot.
	// Supports placeholders: {date}, {time}, {datetime}, {user}, {hostname}
	Note string `yaml:"note,omitempty"`

	// Tags are tags to automatically apply.
	Tags []string `yaml:"tags,omitempty"`

	// Compression is the compression level to use.
	Compression string `yaml:"compression,omitempty"` // none, fast, default, max

	// Paths are paths to include for partial snapshots (empty = full snapshot).
	Paths []string `yaml:"paths,omitempty"`
}

// RetentionPolicyConfig configures GC retention.
type RetentionPolicyConfig struct {
	KeepMinSnapshots int    `yaml:"keep_min_snapshots,omitempty"`
	KeepMinAge       string `yaml:"keep_min_age,omitempty"`
}

// LoggingConfig configures logging behavior.
type LoggingConfig struct {
	Level  string `yaml:"level,omitempty"`
	Format string `yaml:"format,omitempty"` // json, text
}

// WebhookConfig represents webhook configuration for event notifications.
type WebhookConfig struct {
	// Enabled enables or disables webhook notifications.
	Enabled bool `yaml:"enabled,omitempty"`

	// MaxRetries is the number of retries for failed webhook deliveries.
	MaxRetries int `yaml:"max_retries,omitempty"`

	// RetryDelay is the delay between retries.
	RetryDelay string `yaml:"retry_delay,omitempty"` // e.g., "5s", "1m"

	// AsyncQueueSize is the size of the async webhook queue.
	AsyncQueueSize int `yaml:"async_queue_size,omitempty"`

	// Hooks is the list of webhook endpoints.
	Hooks []WebhookHook `yaml:"hooks,omitempty"`
}

// WebhookHook represents a single webhook endpoint.
type WebhookHook struct {
	// URL is the webhook endpoint URL.
	URL string `yaml:"url"`

	// Secret is the HMAC secret for signature verification (optional).
	Secret string `yaml:"secret,omitempty"`

	// Events is the list of events to send to this webhook.
	// Use "*" to receive all events.
	Events []string `yaml:"events,omitempty"`

	// Timeout is the HTTP request timeout (optional).
	Timeout string `yaml:"timeout,omitempty"` // e.g., "10s"

	// Enabled enables or disables this specific hook.
	Enabled bool `yaml:"enabled,omitempty"`
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		DefaultEngine:   "",
		DefaultTags:     nil,
		OutputFormat:    "",
		ProgressEnabled: nil,
		// Legacy defaults
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
	// Check cache first
	cacheMu.RLock()
	if cfg, ok := cache[repoRoot]; ok {
		cacheMu.RUnlock()
		return cfg, nil
	}
	cacheMu.RUnlock()

	cfg := Default()
	cfgPath := filepath.Join(repoRoot, ".jvs", "config.yaml")

	data, err := os.ReadFile(cfgPath)
	if os.IsNotExist(err) {
		// No config file is OK, cache and return defaults
		cacheAndReturn(repoRoot, cfg)
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Validate config
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	cacheAndReturn(repoRoot, cfg)
	return cfg, nil
}

// Save writes configuration to .jvs/config.yaml.
func Save(repoRoot string, cfg *Config) error {
	if err := cfg.validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

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

	// Update cache
	cacheAndReturn(repoRoot, cfg)
	return nil
}

// validate checks if the configuration is valid.
func (c *Config) validate() error {
	// Validate default_engine if set
	if c.DefaultEngine != "" {
		switch c.DefaultEngine {
		case model.EngineJuiceFSClone, model.EngineReflinkCopy, model.EngineCopy, "auto":
			// Valid
		default:
			return fmt.Errorf("invalid default_engine: %s (must be juicefs-clone, reflink-copy, copy, or auto)", c.DefaultEngine)
		}
	}

	// Validate output_format if set
	if c.OutputFormat != "" && c.OutputFormat != "text" && c.OutputFormat != "json" {
		return fmt.Errorf("invalid output_format: %s (must be text or json)", c.OutputFormat)
	}

	// Validate webhooks if configured
	if c.Webhooks != nil {
		if err := c.Webhooks.Validate(); err != nil {
			return fmt.Errorf("webhooks: %w", err)
		}
	}

	return nil
}

// GetDefaultEngine returns the default engine, or empty string if not set.
func (c *Config) GetDefaultEngine() model.EngineType {
	if c.DefaultEngine != "" && c.DefaultEngine != "auto" {
		return c.DefaultEngine
	}
	return ""
}

// GetDefaultTags returns the default tags.
func (c *Config) GetDefaultTags() []string {
	if c.DefaultTags != nil {
		return c.DefaultTags
	}
	return nil
}

// GetOutputFormat returns the output format, or empty string if not set.
func (c *Config) GetOutputFormat() string {
	return c.OutputFormat
}

// GetProgressEnabled returns whether progress is enabled.
// Returns nil if not configured (auto-detect based on terminal).
func (c *Config) GetProgressEnabled() *bool {
	return c.ProgressEnabled
}

// Set sets a configuration value by key.
func (c *Config) Set(key, value string) error {
	switch key {
	case "default_engine":
		c.DefaultEngine = model.EngineType(value)
	case "default_tags":
		// Parse as YAML list
		if err := yaml.Unmarshal([]byte(value), &c.DefaultTags); err != nil {
			return fmt.Errorf("parse tags: %w", err)
		}
	case "output_format":
		c.OutputFormat = value
	case "progress_enabled":
		var enabled bool
		if value == "true" {
			enabled = true
		} else if value == "false" {
			enabled = false
		} else {
			return fmt.Errorf("invalid progress_enabled value: %s (must be true or false)", value)
		}
		c.ProgressEnabled = &enabled
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

// Get gets a configuration value by key as a string.
func (c *Config) Get(key string) (string, error) {
	switch key {
	case "default_engine":
		if c.DefaultEngine == "" {
			return "", nil
		}
		return string(c.DefaultEngine), nil
	case "default_tags":
		if c.DefaultTags == nil {
			return "[]", nil
		}
		data, err := yaml.Marshal(c.DefaultTags)
		if err != nil {
			return "", err
		}
		return string(data), nil
	case "output_format":
		return c.OutputFormat, nil
	case "progress_enabled":
		if c.ProgressEnabled == nil {
			return "", nil
		}
		if *c.ProgressEnabled {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// Keys returns all valid configuration keys.
func Keys() []string {
	return []string{
		"default_engine",
		"default_tags",
		"output_format",
		"progress_enabled",
	}
}

// InvalidateCache clears the cached config for a repository.
func InvalidateCache(repoRoot string) {
	cacheMu.Lock()
	delete(cache, repoRoot)
	cacheMu.Unlock()
}

// cacheAndReturn stores the config in cache.
func cacheAndReturn(repoRoot string, cfg *Config) {
	cacheMu.Lock()
	cache[repoRoot] = cfg
	cacheMu.Unlock()
}

// GetSnapshotTemplate returns a snapshot template by name.
// Returns nil if the template doesn't exist.
func (c *Config) GetSnapshotTemplate(name string) *SnapshotTemplate {
	if c.SnapshotTemplates == nil {
		return nil
	}
	if tmpl, ok := c.SnapshotTemplates[name]; ok {
		return &tmpl
	}
	return nil
}

// GetBuiltinTemplates returns the built-in snapshot templates.
func GetBuiltinTemplates() map[string]SnapshotTemplate {
	return map[string]SnapshotTemplate{
		"pre-experiment": {
			Note: "Before experiment: {datetime}",
			Tags: []string{"experiment", "checkpoint"},
		},
		"pre-deploy": {
			Note: "Pre-deployment checkpoint: {datetime}",
			Tags: []string{"pre-deploy", "release"},
		},
		"checkpoint": {
			Note: "Checkpoint: {datetime}",
			Tags: []string{"checkpoint"},
		},
		"work": {
			Note: "Work in progress: {datetime}",
			Tags: []string{"wip"},
		},
		"release": {
			Note: "Release: {datetime}",
			Tags: []string{"release", "stable"},
		},
		"archive": {
			Note: "Archive: {datetime}",
			Tags: []string{"archive"},
			Compression: "max",
		},
	}
}

// ResolveTemplate resolves a template name to an actual SnapshotTemplate.
// Checks user-defined templates first, then built-in templates.
func ResolveTemplate(repoRoot string, name string) *SnapshotTemplate {
	cfg, _ := Load(repoRoot)

	// Check user-defined templates first
	if tmpl := cfg.GetSnapshotTemplate(name); tmpl != nil {
		return tmpl
	}

	// Check built-in templates
	builtin := GetBuiltinTemplates()
	if tmpl, ok := builtin[name]; ok {
		return &tmpl
	}

	return nil
}

// ListTemplates returns all available template names (user + built-in).
func ListTemplates(repoRoot string) []string {
	cfg, _ := Load(repoRoot)
	names := make(map[string]bool)

	// Add user-defined templates
	if cfg.SnapshotTemplates != nil {
		for name := range cfg.SnapshotTemplates {
			names[name] = true
		}
	}

	// Add built-in templates
	for name := range GetBuiltinTemplates() {
		names[name] = true
	}

	// Convert to sorted slice
	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, name)
	}
	return result
}

// Validate checks if the webhook configuration is valid.
func (w *WebhookConfig) Validate() error {
	for i, hook := range w.Hooks {
		if hook.URL == "" {
			return fmt.Errorf("hook[%d]: url is required", i)
		}
		if len(hook.Events) == 0 {
			return fmt.Errorf("hook[%d]: at least one event must be specified", i)
		}
	}
	return nil
}

// ToWebhookConfig converts WebhookConfig to webhook.Config for use by the webhook package.
func (w *WebhookConfig) ToWebhookConfig() *webhook.Config {
	if w == nil {
		return nil
	}

	cfg := &webhook.Config{
		Enabled: w.Enabled,
		Hooks:   make([]webhook.HookConfig, 0, len(w.Hooks)),
	}

	// Apply defaults
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if w.MaxRetries > 0 {
		cfg.MaxRetries = w.MaxRetries
	}
	if w.RetryDelay != "" {
		if d, err := parseDuration(w.RetryDelay); err == nil {
			cfg.RetryDelay = d
		}
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 5 * 1000000000 // 5 seconds in nanoseconds
	}
	if w.AsyncQueueSize > 0 {
		cfg.AsyncQueueSize = w.AsyncQueueSize
	}
	if cfg.AsyncQueueSize == 0 {
		cfg.AsyncQueueSize = 100
	}

	// Convert hooks
	for _, h := range w.Hooks {
		hookCfg := webhook.HookConfig{
			URL:     h.URL,
			Secret:  h.Secret,
			Events:  make([]webhook.EventType, 0, len(h.Events)),
			Enabled: h.Enabled,
		}

		// Convert event strings to EventType
		for _, e := range h.Events {
			hookCfg.Events = append(hookCfg.Events, webhook.EventType(e))
		}

		// Parse timeout if specified
		if h.Timeout != "" {
			if d, err := parseDuration(h.Timeout); err == nil {
				hookCfg.Timeout = d
			}
		}

		cfg.Hooks = append(cfg.Hooks, hookCfg)
	}

	return cfg
}

// parseDuration parses a duration string like "5s", "1m", "1h".
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

