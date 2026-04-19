package patterns

import (
	"strings"

	"github.com/rkbharti/devdebug/internal/config"
	"github.com/rkbharti/devdebug/internal/input"
)

type ErrorMatch struct {
	LineNumber int
	Type       string
	Message    string
	Context    string
	File       string
}

// ─────────────────────────────────────────────────────────────────────────────
// DetectError inspects a ParsedLine and returns an ErrorMatch if it is an
// error. Returns nil for info/debug lines and lines with no error keywords.
// ─────────────────────────────────────────────────────────────────────────────
func DetectError(parsed input.ParsedLine, lineNum int, context string, cfg *config.Config) *ErrorMatch {

	// ── skip empty lines ──────────────────────────────────────────────────────
	if strings.TrimSpace(parsed.Raw) == "" {
		return nil
	}

	// ── JSON path: use level field if available ───────────────────────────────
	if parsed.IsJSON {
		return detectFromJSON(parsed, lineNum, context, cfg)
	}

	// ── plain text path: use keyword matching on raw line ─────────────────────
	return detectFromPlainText(parsed.Raw, lineNum, context, cfg)
}

// ─────────────────────────────────────────────────────────────────────────────
// detectFromJSON — uses the extracted level + message from JSON logs.
// Level field is the source of truth when present.
// Falls back to keyword matching on message when level is absent/ambiguous.
// ─────────────────────────────────────────────────────────────────────────────
func detectFromJSON(parsed input.ParsedLine, lineNum int, context string, cfg *config.Config) *ErrorMatch {

	level := parsed.Level // already lowercased by ParseLine

	// ── explicitly non-error levels → skip ───────────────────────────────────
	if level == "info" || level == "debug" || level == "trace" || level == "warn" || level == "warning" {
		return nil
	}

	// ── explicitly error levels → detect ─────────────────────────────────────
	isErrorLevel := level == "error" || level == "err" ||
		level == "fatal" || level == "critical" || level == "panic"

	if isErrorLevel {
		// classify using the extracted message content
		errType := classifyMessage(parsed.Message, cfg)
		if errType == "" {
			errType = "General Error" // level says error — trust it
		}
		return &ErrorMatch{
			LineNumber: lineNum,
			Type:       errType,
			Message:    parsed.Message,
			Context:    context,
		}
	}

	// ── no level field or unknown level → fall back to keyword matching ───────
	return detectFromPlainText(parsed.Message, lineNum, context, cfg)
}

// ─────────────────────────────────────────────────────────────────────────────
// detectFromPlainText — original keyword-based matching on a plain text line.
// ─────────────────────────────────────────────────────────────────────────────
func detectFromPlainText(line string, lineNum int, context string, cfg *config.Config) *ErrorMatch {

	// ── noise filter ──────────────────────────────────────────────────────────
	lower := strings.ToLower(line)

	if strings.Contains(lower, "info") || strings.Contains(lower, "debug") {
		return nil
	}

	// ── custom config patterns ────────────────────────────────────────────────
	if cfg != nil {
		for _, p := range cfg.Patterns {
			keyword := strings.TrimSpace(p.Keyword)
			if keyword == "" {
				continue
			}
			if strings.Contains(lower, strings.ToLower(keyword)) {
				return &ErrorMatch{
					LineNumber: lineNum,
					Type:       p.Name,
					Message:    line,
					Context:    context,
				}
			}
		}
	}

	// ── default keyword patterns ──────────────────────────────────────────────
	errType := classifyMessage(line, nil)
	if errType == "" {
		return nil
	}

	return &ErrorMatch{
		LineNumber: lineNum,
		Type:       errType,
		Message:    line,
		Context:    context,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// classifyMessage returns the error type for a message string.
// Used by both JSON and plain text paths.
// Returns "" if no pattern matches.
// ─────────────────────────────────────────────────────────────────────────────
func classifyMessage(message string, cfg *config.Config) string {
	lower := strings.ToLower(message)

	// custom config first
	if cfg != nil {
		for _, p := range cfg.Patterns {
			keyword := strings.TrimSpace(p.Keyword)
			if keyword == "" {
				continue
			}
			if strings.Contains(lower, strings.ToLower(keyword)) {
				return p.Name
			}
		}
	}

	if strings.Contains(lower, "panic") {
		return "Panic Error"
	}
	if strings.Contains(lower, "error:") ||
		strings.HasPrefix(lower, "error ") ||
		strings.Contains(lower, " exception") {
		return "General Error"
	}
	if strings.Contains(lower, "timeout ") ||
		strings.Contains(lower, "request timeout") ||
		strings.Contains(lower, "timed out") ||
		strings.Contains(lower, "connection timeout") {
		return "Timeout Error"
	}

	return ""
}
