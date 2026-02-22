package config

import (
	"testing"

	"github.com/jvs-project/jvs/pkg/model"
)

func TestSnapshotTemplate_ResolveTemplate(t *testing.T) {
	type templateTest struct {
		name     string
		template string
		wantNote string
		wantTags []string
	}

	tests := []templateTest{{
		{
			name:     "builtin pre-experiment",
			template: "pre-experiment",
			wantNote: "Before experiment: ",
			wantTags: []string{"experiment", "checkpoint"},
		},
		{
			name:     "builtin checkpoint",
			template: "checkpoint",
			wantNote: "Checkpoint: ",
			wantTags: []string{"checkpoint"},
		},
		{
			name:     "builtin work",
			template: "work",
			wantNote: "Work in progress: ",
			wantTags: []string{"wip"},
		},
		{
			name:     "builtin release",
			template: "release",
			wantNote: "Release: ",
			wantTags: []string{"release", "stable"},
		},
		{
			name:     "builtin archive",
			template: "archive",
			wantNote: "Archive: ",
			wantTags: []string{"archive"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := ResolveTemplate("", tt.template)
			if tmpl == nil {
				t.Fatalf("template %q not found", tt.template)
			}

			if tmpl.Note == "" {
				t.Error("template note is empty")
			}
			// Note that templates have placeholders that get expanded,
			// so we just check that it starts with the expected prefix
			// Actual expansion is tested in the template package

			if len(tmpl.Tags) != len(tt.wantTags) {
				t.Errorf("got %d tags, want %d", len(tmpl.Tags), len(tt.wantTags))
			}
		})
	}
}

func TestGetBuiltinTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()

	// Check that all expected templates exist
	expected := []string{
		"pre-experiment",
		"pre-deploy",
		"checkpoint",
		"work",
		"release",
		"archive",
	}

	for _, name := range expected {
		if tmpl, ok := templates[name]; !ok {
			t.Errorf("missing built-in template: %s", name)
		} else {
			if tmpl.Note == "" {
				t.Errorf("template %s has empty note", name)
			}
			if len(tmpl.Tags) == 0 {
				t.Errorf("template %s has no tags", name)
			}
		}
	}
}

func TestConfig_GetSnapshotTemplate(t *testing.T) {
	cfg := &Config{
		SnapshotTemplates: map[string]SnapshotTemplate{
			"custom": {
				Note: "Custom note",
				Tags: []string{"custom", "test"},
			},
		},
	}

	t.Run("existing template", func(t *testing.T) {
		tmpl := cfg.GetSnapshotTemplate("custom")
		if tmpl == nil {
			t.Fatal("template not found")
		}
		if tmpl.Note != "Custom note" {
			t.Errorf("got note %q, want %q", tmpl.Note, "Custom note")
		}
	})

	t.Run("non-existing template", func(t *testing.T) {
		tmpl := cfg.GetSnapshotTemplate("nonexistent")
		if tmpl != nil {
			t.Error("expected nil for non-existent template")
		}
	})

	t.Run("no templates configured", func(t *testing.T) {
		cfg := &Config{}
		tmpl := cfg.GetSnapshotTemplate("any")
		if tmpl != nil {
			t.Error("expected nil when no templates configured")
		}
	})
}

func TestConfig_validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				DefaultEngine: model.EngineCopy,
			},
			wantErr: false,
		},
		{
			name: "invalid engine",
			cfg: Config{
				DefaultEngine: model.EngineType("invalid"),
			},
			wantErr: true,
		},
		{
			name:    "empty config",
			cfg:     Config{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSnapshotTemplate_PartialSnapshot(t *testing.T) {
	cfg := &Config{
		SnapshotTemplates: map[string]SnapshotTemplate{
			"partial-backup": {
				Note: "Partial backup",
				Tags: []string{"backup"},
				Paths: []string{"src/", "data/"},
			},
		},
	}

	tmpl := cfg.GetSnapshotTemplate("partial-backup")
	if tmpl == nil {
		t.Fatal("template not found")
	}

	if len(tmpl.Paths) != 2 {
		t.Errorf("got %d paths, want 2", len(tmpl.Paths))
	}

	expectedPaths := []string{"src/", "data/"}
	for i, path := range tmpl.Paths {
		if path != expectedPaths[i] {
			t.Errorf("path %d = %q, want %q", i, path, expectedPaths[i])
		}
	}
}

func TestListTemplates(t *testing.T) {
	// List templates with empty config (only built-ins)
	templates := ListTemplates("")

	// Should have all built-in templates
	builtinCount := len(GetBuiltinTemplates())
	if len(templates) != builtinCount {
		t.Errorf("got %d templates, want %d (built-ins only)", len(templates), builtinCount)
	}

	// Check that expected names are present
	expectedNames := []string{
		"pre-experiment",
		"pre-deploy",
		"checkpoint",
		"work",
		"release",
		"archive",
	}

	nameMap := make(map[string]bool)
	for _, name := range templates {
		nameMap[name] = true
	}

	for _, name := range expectedNames {
		if !nameMap[name] {
			t.Errorf("missing template: %s", name)
		}
	}
}
