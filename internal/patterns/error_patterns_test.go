package patterns

import (
	"testing"

	"github.com/rkbharti/LogSensei_CLI/internal/config"
	"github.com/rkbharti/LogSensei_CLI/internal/input"
)

func TestDetectError(t *testing.T) {

	tests := []struct {
		name         string
		line         string
		cfg          *config.Config
		wantNil      bool
		wantType     string
		wantContains string
	}{
		// ── nil / noise ───────────────────────────────────────────────────────
		{name: "empty line is ignored", line: "", wantNil: true},
		{name: "whitespace-only line is ignored", line: "     ", wantNil: true},
		{name: "INFO log is ignored", line: "[INFO] Server started on port 8080", wantNil: true},
		{name: "DEBUG log is ignored", line: "[DEBUG] connecting to database", wantNil: true},
		{name: "INFO log with mixed case is ignored", line: "[Info] Request received", wantNil: true},

		// ── panic ─────────────────────────────────────────────────────────────
		{
			name:         "lowercase panic is detected",
			line:         "panic: runtime error: invalid memory address",
			wantNil:      false,
			wantType:     "Panic Error",
			wantContains: "panic",
		},
		{
			name:         "uppercase PANIC is detected",
			line:         "PANIC: goroutine died",
			wantNil:      false,
			wantType:     "Panic Error",
			wantContains: "PANIC",
		},

		// ── general error ─────────────────────────────────────────────────────
		{
			name:         "line with 'error:' is detected",
			line:         "error: failed to open config file",
			wantNil:      false,
			wantType:     "General Error",
			wantContains: "error",
		},
		{
			name:         "line starting with 'error ' is detected",
			line:         "error reading socket",
			wantNil:      false,
			wantType:     "General Error",
			wantContains: "error reading",
		},
		{
			name:         "line with exception is detected",
			line:         "caught NullPointer exception at main.go:42",
			wantNil:      false,
			wantType:     "General Error",
			wantContains: "exception",
		},

		// ── timeout ───────────────────────────────────────────────────────────
		{
			name:         "request timeout is detected",
			line:         "request timeout after 30s",
			wantNil:      false,
			wantType:     "Timeout Error",
			wantContains: "timeout",
		},
		{
			name:         "connection timeout is detected",
			line:         "connection timeout to redis:6379",
			wantNil:      false,
			wantType:     "Timeout Error",
			wantContains: "connection timeout",
		},

		// ── ignored plain lines ───────────────────────────────────────────────
		{name: "normal log line with no keywords is ignored", line: "Server running on port 3000", wantNil: true},
		{name: "warning log without error keyword is ignored", line: "WARN: memory usage high", wantNil: true},

		// ── JSON log detection ────────────────────────────────────────────────
		{
			name:         "JSON error level is detected",
			line:         `{"level":"error","message":"database connection failed","service":"api"}`,
			wantNil:      false,
			wantType:     "General Error",
			wantContains: "database connection failed",
		},
		{
			name:         "JSON fatal level is detected",
			line:         `{"level":"fatal","msg":"server crashed","ts":1713541200}`,
			wantNil:      false,
			wantType:     "General Error",
			wantContains: "server crashed",
		},
		{
			name:         "JSON panic level is detected as Panic Error",
			line:         `{"level":"error","message":"panic: nil pointer dereference"}`,
			wantNil:      false,
			wantType:     "Panic Error",
			wantContains: "panic",
		},
		{name: "JSON info level is ignored", line: `{"level":"info","message":"Server started","port":8080}`, wantNil: true},
		{name: "JSON debug level is ignored", line: `{"level":"debug","msg":"connecting to redis"}`, wantNil: true},
		{name: "JSON warn level is ignored", line: `{"level":"warn","msg":"high memory usage"}`, wantNil: true},
		{
			name:         "JSON with no level falls back to keyword matching",
			line:         `{"message":"error: config file not found","service":"loader"}`,
			wantNil:      false,
			wantType:     "General Error",
			wantContains: "error",
		},
		{
			name:    "JSON with no level and no error keyword is ignored",
			line:    `{"message":"successfully connected","service":"db"}`,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := input.ParseLine(tt.line)
			result := DetectError(parsed, 1, "", tt.cfg.Compile())

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil but got: Type=%q Message=%q", result.Type, result.Message)
				}
				return
			}

			if result == nil {
				t.Fatalf("expected ErrorMatch but got nil for line: %q", tt.line)
			}
			if result.Type != tt.wantType {
				t.Errorf("Type: got %q, want %q", result.Type, tt.wantType)
			}
			if tt.wantContains != "" {
				found := false
				for i := 0; i <= len(result.Message)-len(tt.wantContains); i++ {
					if result.Message[i:i+len(tt.wantContains)] == tt.wantContains {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Message %q does not contain %q", result.Message, tt.wantContains)
				}
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Custom config tests
// ─────────────────────────────────────────────────────────────────────────────

func TestDetectError_CustomConfig(t *testing.T) {

	cfg := &config.Config{
		Patterns: []config.Pattern{
			{Name: "DB Down", Keyword: "database unreachable"},
			{Name: "Auth Fail", Keyword: "unauthorized"},
			{Name: "Empty Pattern", Keyword: ""},
			{Name: "Whitespace Pattern", Keyword: "   "},
		},
	}

	tests := []struct {
		name     string
		line     string
		wantNil  bool
		wantType string
	}{
		{name: "custom keyword matched", line: "database unreachable: host=db01", wantNil: false, wantType: "DB Down"},
		{name: "custom keyword case insensitive", line: "Unauthorized access attempt from 192.168.1.1", wantNil: false, wantType: "Auth Fail"},
		{name: "empty keyword pattern does not panic", line: "some random log line", wantNil: true},
		{name: "whitespace keyword pattern does not panic", line: "another log line", wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := input.ParseLine(tt.line) // ✅ fixed
			result := DetectError(parsed, 1, "", cfg.Compile())

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil but got: Type=%q", result.Type)
				}
				return
			}
			if result == nil {
				t.Fatalf("expected match but got nil for line: %q", tt.line)
			}
			if result.Type != tt.wantType {
				t.Errorf("Type: got %q, want %q", result.Type, tt.wantType)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Line number test
// ─────────────────────────────────────────────────────────────────────────────

func TestDetectError_LineNumber(t *testing.T) {
	parsed := input.ParseLine("panic: nil pointer dereference") // ✅ fixed
	result := DetectError(parsed, 42, "", nil)

	if result == nil {
		t.Fatal("expected ErrorMatch, got nil")
	}
	if result.LineNumber != 42 {
		t.Errorf("LineNumber: got %d, want 42", result.LineNumber)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Regex pattern tests
// ─────────────────────────────────────────────────────────────────────────────

func TestDetectError_RegexPatterns(t *testing.T) {

	tests := []struct {
		name     string
		line     string
		cfg      *config.Config
		wantNil  bool
		wantType string
	}{
		{
			name: "regex matches HTTP 5xx errors",
			line: "HTTP 500 Internal Server Error",
			cfg: &config.Config{
				Patterns: []config.Pattern{
					{Name: "5xx Error", Regex: `HTTP [5][0-9]{2}`},
				},
			},
			wantNil:  false,
			wantType: "5xx Error",
		},
		{
			name: "regex does not match HTTP 4xx",
			line: "HTTP 404 Not Found",
			cfg: &config.Config{
				Patterns: []config.Pattern{
					{Name: "5xx Error", Regex: `HTTP [5][0-9]{2}`},
				},
			},
			wantNil: true,
		},
		{
			name: "regex matches retry exhaustion pattern",
			line: "failed after 5 retries — giving up",
			cfg: &config.Config{
				Patterns: []config.Pattern{
					{Name: "Retry Exhausted", Regex: `failed after [0-9]+ retr`},
				},
			},
			wantNil:  false,
			wantType: "Retry Exhausted",
		},
		{
			name: "keyword takes priority over regex — keyword matched first",
			line: "database unreachable: connection refused",
			cfg: &config.Config{
				Patterns: []config.Pattern{
					{Name: "DB Down", Keyword: "database unreachable"},
					{Name: "Conn Refused", Regex: `connection refused`},
				},
			},
			wantNil:  false,
			wantType: "DB Down", // keyword matched first — order matters
		},
		{
			name: "regex is case-sensitive by default",
			line: "UNAUTHORIZED access attempt",
			cfg: &config.Config{
				Patterns: []config.Pattern{
					{Name: "Auth Fail", Regex: `unauthorized`}, // lowercase regex
				},
			},
			wantNil: true, // regex is case-sensitive — UPPERCASE won't match
		},
		{
			name: "regex with case-insensitive flag (?i) matches uppercase",
			line: "UNAUTHORIZED access attempt",
			cfg: &config.Config{
				Patterns: []config.Pattern{
					{Name: "Auth Fail", Regex: `(?i)unauthorized`}, // (?i) = case-insensitive
				},
			},
			wantNil:  false,
			wantType: "Auth Fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := input.ParseLine(tt.line)
			result := DetectError(parsed, 1, "", tt.cfg.Compile())

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil but got: Type=%q Message=%q", result.Type, result.Message)
				}
				return
			}

			if result == nil {
				t.Fatalf("expected ErrorMatch but got nil for line: %q", tt.line)
			}

			if result.Type != tt.wantType {
				t.Errorf("Type: got %q, want %q", result.Type, tt.wantType)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Config validation tests
// ─────────────────────────────────────────────────────────────────────────────

func TestConfig_ValidatePatterns(t *testing.T) {

	tests := []struct {
		name    string
		cfg     config.Config
		wantErr bool
	}{
		{
			name: "valid keyword pattern",
			cfg: config.Config{
				Patterns: []config.Pattern{{Name: "DB Down", Keyword: "database unreachable"}},
			},
			wantErr: false,
		},
		{
			name: "valid regex pattern",
			cfg: config.Config{
				Patterns: []config.Pattern{{Name: "5xx", Regex: `HTTP [5][0-9]{2}`}},
			},
			wantErr: false,
		},
		{
			name: "invalid regex fails validation",
			cfg: config.Config{
				Patterns: []config.Pattern{{Name: "Bad", Regex: `[invalid(`}},
			},
			wantErr: true,
		},
		{
			name: "pattern with no keyword and no regex fails",
			cfg: config.Config{
				Patterns: []config.Pattern{{Name: "Empty", Keyword: "", Regex: ""}},
			},
			wantErr: true,
		},
		{
			name: "pattern with no name fails",
			cfg: config.Config{
				Patterns: []config.Pattern{{Name: "", Keyword: "something"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.ValidatePatterns()
			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}
