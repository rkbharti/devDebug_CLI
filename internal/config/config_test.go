package config

import (
	"os"
	"path/filepath"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────────

// writeTempYAML writes content to a temp file and returns its path.
// The file is automatically removed when the test ends.
func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "devdebug.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp YAML: %v", err)
	}
	return path
}

// ─────────────────────────────────────────────────────────────────────────────
// LoadConfig tests
// ─────────────────────────────────────────────────────────────────────────────

func TestLoadConfig(t *testing.T) {

	tests := []struct {
		name      string
		yaml      string
		wantErr   bool
		wantCount int // expected number of patterns loaded
	}{
		{
			name: "valid keyword pattern loads correctly",
			yaml: `
patterns:
  - name: "DB Down"
    keyword: "database unreachable"
`,
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "valid regex pattern loads correctly",
			yaml: `
patterns:
  - name: "5xx Error"
    regex: "HTTP [5][0-9]{2}"
`,
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "multiple patterns all load",
			yaml: `
patterns:
  - name: "DB Down"
    keyword: "database unreachable"
  - name: "Auth Fail"
    regex: "(?i)unauthorized"
  - name: "5xx Error"
    regex: "HTTP [5][0-9]{2}"
`,
			wantErr:   false,
			wantCount: 3,
		},
		{
			name: "empty patterns list is valid",
			yaml: `
patterns: []
`,
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "invalid regex is rejected at load time",
			yaml: `
patterns:
  - name: "Bad Pattern"
    regex: "[invalid("
`,
			wantErr: true,
		},
		{
			name: "pattern with no keyword and no regex is rejected",
			yaml: `
patterns:
  - name: "Empty"
    keyword: ""
    regex: ""
`,
			wantErr: true,
		},
		{
			name: "pattern with no name is rejected",
			yaml: `
patterns:
  - name: ""
    keyword: "something"
`,
			wantErr: true,
		},
		{
			name:    "invalid YAML syntax returns error",
			yaml:    `patterns: [[[`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempYAML(t, tt.yaml)
			cfg, err := LoadConfig(path)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(cfg.Patterns) != tt.wantCount {
				t.Errorf("pattern count: got %d, want %d", len(cfg.Patterns), tt.wantCount)
			}
		})
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent_path/devdebug.yaml")
	if err == nil {
		t.Error("expected error for missing file but got nil")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Compile tests
// ─────────────────────────────────────────────────────────────────────────────

func TestCompile(t *testing.T) {

	t.Run("nil config returns nil slice", func(t *testing.T) {
		var cfg *Config
		result := cfg.Compile()
		if result != nil {
			t.Errorf("expected nil but got slice of length %d", len(result))
		}
	})

	t.Run("keyword is lowercased in compiled pattern", func(t *testing.T) {
		cfg := &Config{
			Patterns: []Pattern{
				{Name: "DB Down", Keyword: "DATABASE Unreachable"},
			},
		}
		compiled := cfg.Compile()

		if len(compiled) != 1 {
			t.Fatalf("expected 1 compiled pattern, got %d", len(compiled))
		}
		if compiled[0].Keyword != "database unreachable" {
			t.Errorf("Keyword: got %q, want %q", compiled[0].Keyword, "database unreachable")
		}
		if compiled[0].Regex != nil {
			t.Error("Regex should be nil for keyword-only pattern")
		}
	})

	t.Run("regex pattern has compiled Regexp", func(t *testing.T) {
		cfg := &Config{
			Patterns: []Pattern{
				{Name: "5xx Error", Regex: `HTTP [5][0-9]{2}`},
			},
		}
		compiled := cfg.Compile()

		if len(compiled) != 1 {
			t.Fatalf("expected 1 compiled pattern, got %d", len(compiled))
		}
		if compiled[0].Regex == nil {
			t.Fatal("Regex should not be nil for regex pattern")
		}
		if compiled[0].Keyword != "" {
			t.Errorf("Keyword should be empty for regex-only pattern, got %q", compiled[0].Keyword)
		}
	})

	t.Run("pattern name is preserved correctly", func(t *testing.T) {
		cfg := &Config{
			Patterns: []Pattern{
				{Name: "Auth Attack", Regex: `(?i)unauthorized`},
			},
		}
		compiled := cfg.Compile()

		if compiled[0].Name != "Auth Attack" {
			t.Errorf("Name: got %q, want %q", compiled[0].Name, "Auth Attack")
		}
	})

	t.Run("compiled slice length matches pattern count", func(t *testing.T) {
		cfg := &Config{
			Patterns: []Pattern{
				{Name: "A", Keyword: "alpha"},
				{Name: "B", Keyword: "beta"},
				{Name: "C", Regex: `gamma.*delta`},
			},
		}
		compiled := cfg.Compile()

		if len(compiled) != 3 {
			t.Errorf("expected 3 compiled patterns, got %d", len(compiled))
		}
	})

	t.Run("regex actually matches expected input", func(t *testing.T) {
		cfg := &Config{
			Patterns: []Pattern{
				{Name: "5xx Error", Regex: `HTTP [5][0-9]{2}`},
			},
		}
		compiled := cfg.Compile()
		re := compiled[0].Regex

		if !re.MatchString("HTTP 500 Internal Server Error") {
			t.Error("regex should match HTTP 500")
		}
		if re.MatchString("HTTP 404 Not Found") {
			t.Error("regex should not match HTTP 404")
		}
	})

	t.Run("empty config returns empty slice not nil", func(t *testing.T) {
		cfg := &Config{Patterns: []Pattern{}}
		compiled := cfg.Compile()

		// empty slice is not nil — capacity was allocated
		if compiled == nil {
			t.Error("expected empty slice but got nil")
		}
		if len(compiled) != 0 {
			t.Errorf("expected length 0 but got %d", len(compiled))
		}
	})
}
