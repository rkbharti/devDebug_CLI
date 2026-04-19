package analyzer

import (
	"strings"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// Table-driven tests for ExplainError()
// ─────────────────────────────────────────────────────────────────────────────

func TestExplainError(t *testing.T) {

	tests := []struct {
		name               string
		message            string
		wantReasonContains string // substring expected in Reason
		wantSuggContains   string // substring expected in Suggestion
	}{
		// ── nil pointer ───────────────────────────────────────────────────────
		{
			name:               "nil pointer message",
			message:            "panic: runtime error: nil pointer dereference",
			wantReasonContains: "nil pointer",
			wantSuggContains:   "nil",
		},
		{
			name:               "invalid memory address message",
			message:            "invalid memory address or nil pointer",
			wantReasonContains: "nil pointer",
			wantSuggContains:   "nil",
		},

		// ── timeout ───────────────────────────────────────────────────────────
		{
			name:               "timeout message",
			message:            "request timeout after 30s",
			wantReasonContains: "too long",
			wantSuggContains:   "timeout",
		},

		// ── JS / frontend error ───────────────────────────────────────────────
		{
			name:               "cannot read property (JS TypeError)",
			message:            "TypeError: cannot read property 'name' of undefined",
			wantReasonContains: "undefined",
			wantSuggContains:   "exists",
		},

		// ── database error ────────────────────────────────────────────────────
		{
			name:               "database connection failed",
			message:            "database connection failed: timeout",
			wantReasonContains: "database",
			wantSuggContains:   "database",
		},
		{
			name:               "connection failed keyword",
			message:            "connection failed to postgres://localhost:5432",
			wantReasonContains: "database",
			wantSuggContains:   "database",
		},

		// ── unknown / fallback ────────────────────────────────────────────────
		{
			name:               "unrecognised message falls back to default",
			message:            "something completely unknown happened",
			wantReasonContains: "Unknown",
			wantSuggContains:   "manually",
		},
		{
			name:               "empty message uses fallback",
			message:            "",
			wantReasonContains: "Unknown",
			wantSuggContains:   "manually",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			result := ExplainError(tt.message)

			// ── reason check ──────────────────────────────────────────────────
			if !strings.Contains(result.Reason, tt.wantReasonContains) {
				t.Errorf(
					"Reason: got %q, expected it to contain %q",
					result.Reason,
					tt.wantReasonContains,
				)
			}

			// ── suggestion check ──────────────────────────────────────────────
			if !strings.Contains(result.Suggestion, tt.wantSuggContains) {
				t.Errorf(
					"Suggestion: got %q, expected it to contain %q",
					result.Suggestion,
					tt.wantSuggContains,
				)
			}

			// ── both fields must be non-empty ─────────────────────────────────
			if result.Reason == "" {
				t.Error("Reason must not be empty")
			}
			if result.Suggestion == "" {
				t.Error("Suggestion must not be empty")
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Test that Explanation struct always returns both fields populated
// ─────────────────────────────────────────────────────────────────────────────

func TestExplainError_AlwaysReturnsBothFields(t *testing.T) {
	inputs := []string{
		"panic: nil pointer",
		"timeout occurred",
		"database connection failed",
		"cannot read property of undefined",
		"",
		"gibberish xyzabc 12345",
	}

	for _, msg := range inputs {
		t.Run(msg, func(t *testing.T) {
			result := ExplainError(msg)
			if result.Reason == "" {
				t.Errorf("Reason is empty for message: %q", msg)
			}
			if result.Suggestion == "" {
				t.Errorf("Suggestion is empty for message: %q", msg)
			}
		})
	}
}
