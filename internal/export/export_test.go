package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rkbharti/LogSensei_CLI/internal/patterns"
)

// ─────────────────────────────────────────────────────────────────────────────
// helper — run export function in a temp dir and return file contents
// ─────────────────────────────────────────────────────────────────────────────

func runInTempDir(t *testing.T, fn func()) string {
	t.Helper()

	// save and restore working dir
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}
	defer os.Chdir(original)

	// switch to a temp dir so report files don't pollute the project
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir to temp: %v", err)
	}

	fn()
	return tmp
}

func sampleErrors() []patterns.ErrorMatch {
	return []patterns.ErrorMatch{
		{
			LineNumber: 10,
			Type:       "Panic Error",
			Message:    "panic: nil pointer dereference",
			Context:    "goroutine 1 [running]",
			File:       "main.go",
			Timestamp:  time.Time{},
		},
		{
			LineNumber: 42,
			Type:       "Timeout Error",
			Message:    "request timed out after 30s",
			Context:    "",
			File:       "server.go",
			Timestamp:  time.Time{},
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ExportJSON tests
// ─────────────────────────────────────────────────────────────────────────────

func TestExportJSON(t *testing.T) {

	t.Run("creates report.json file", func(t *testing.T) {
		dir := runInTempDir(t, func() {
			if err := ExportJSON(sampleErrors()); err != nil {
				t.Errorf("ExportJSON returned error: %v", err)
			}
		})

		path := filepath.Join(dir, "report.json")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("report.json was not created")
		}
	})

	t.Run("report.json contains valid JSON", func(t *testing.T) {
		dir := runInTempDir(t, func() {
			ExportJSON(sampleErrors())
		})

		data, err := os.ReadFile(filepath.Join(dir, "report.json"))
		if err != nil {
			t.Fatalf("failed to read report.json: %v", err)
		}

		var result []patterns.ErrorMatch
		if err := json.Unmarshal(data, &result); err != nil {
			t.Errorf("report.json is not valid JSON: %v", err)
		}
	})

	t.Run("report.json contains all errors", func(t *testing.T) {
		dir := runInTempDir(t, func() {
			ExportJSON(sampleErrors())
		})

		data, _ := os.ReadFile(filepath.Join(dir, "report.json"))
		var result []patterns.ErrorMatch
		json.Unmarshal(data, &result)

		if len(result) != 2 {
			t.Errorf("expected 2 errors in JSON, got %d", len(result))
		}
		if result[0].Type != "Panic Error" {
			t.Errorf("first error type: got %q, want %q", result[0].Type, "Panic Error")
		}
		if result[1].LineNumber != 42 {
			t.Errorf("second error line: got %d, want 42", result[1].LineNumber)
		}
	})

	t.Run("empty errors slice produces valid empty JSON array", func(t *testing.T) {
		dir := runInTempDir(t, func() {
			ExportJSON([]patterns.ErrorMatch{})
		})

		data, _ := os.ReadFile(filepath.Join(dir, "report.json"))
		var result []patterns.ErrorMatch
		if err := json.Unmarshal(data, &result); err != nil {
			t.Errorf("empty export is not valid JSON: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty array, got %d items", len(result))
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// ExportMarkdown tests
// ─────────────────────────────────────────────────────────────────────────────

func TestExportMarkdown(t *testing.T) {

	t.Run("creates report.md file", func(t *testing.T) {
		dir := runInTempDir(t, func() {
			if err := ExportMarkdown(sampleErrors()); err != nil {
				t.Errorf("ExportMarkdown returned error: %v", err)
			}
		})

		path := filepath.Join(dir, "report.md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("report.md was not created")
		}
	})

	t.Run("report.md contains title header", func(t *testing.T) {
		dir := runInTempDir(t, func() {
			ExportMarkdown(sampleErrors())
		})

		data, _ := os.ReadFile(filepath.Join(dir, "report.md"))
		content := string(data)

		if !strings.Contains(content, "# DevDebug Report") {
			t.Error("report.md missing title header")
		}
	})

	t.Run("report.md contains error type and message", func(t *testing.T) {
		dir := runInTempDir(t, func() {
			ExportMarkdown(sampleErrors())
		})

		data, _ := os.ReadFile(filepath.Join(dir, "report.md"))
		content := string(data)

		if !strings.Contains(content, "Panic Error") {
			t.Error("report.md missing error type 'Panic Error'")
		}
		if !strings.Contains(content, "panic: nil pointer dereference") {
			t.Error("report.md missing error message")
		}
		if !strings.Contains(content, "Timeout Error") {
			t.Error("report.md missing second error type")
		}
	})

	t.Run("report.md contains line numbers", func(t *testing.T) {
		dir := runInTempDir(t, func() {
			ExportMarkdown(sampleErrors())
		})

		data, _ := os.ReadFile(filepath.Join(dir, "report.md"))
		content := string(data)

		if !strings.Contains(content, "Line 10") {
			t.Error("report.md missing 'Line 10'")
		}
		if !strings.Contains(content, "Line 42") {
			t.Error("report.md missing 'Line 42'")
		}
	})

	t.Run("empty errors slice produces only title", func(t *testing.T) {
		dir := runInTempDir(t, func() {
			ExportMarkdown([]patterns.ErrorMatch{})
		})

		data, _ := os.ReadFile(filepath.Join(dir, "report.md"))
		content := string(data)

		if !strings.Contains(content, "# DevDebug Report") {
			t.Error("report.md missing title even for empty export")
		}
		if strings.Contains(content, "## Error at Line") {
			t.Error("report.md should have no error sections for empty input")
		}
	})
}
