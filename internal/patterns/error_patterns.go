package patterns

import (
	"strings"
	"time"

	"github.com/rkbharti/devdebug/internal/config"
	"github.com/rkbharti/devdebug/internal/input"
)

type ErrorMatch struct {
	LineNumber int
	Type       string
	Message    string
	Context    string
	File       string
	Timestamp  time.Time
}

// ─────────────────────────────────────────────────────────────────────────────
// DetectError inspects a ParsedLine and returns an ErrorMatch if it is an
// error. Returns nil for info/debug lines and lines with no error keywords.
// ─────────────────────────────────────────────────────────────────────────────
func DetectError(parsed input.ParsedLine, lineNum int, context string, cfg *config.Config) *ErrorMatch {

	if strings.TrimSpace(parsed.Raw) == "" {
		return nil
	}

	var match *ErrorMatch

	if parsed.IsJSON {
		match = detectFromJSON(parsed, lineNum, context, cfg)
	} else {
		match = detectFromPlainText(parsed.Raw, lineNum, context, cfg)
	}

	// 🆕 attach timestamp from parsed line to the match
	if match != nil {
		match.Timestamp = parsed.Timestamp
	}

	return match
}

// ─────────────────────────────────────────────────────────────────────────────
// detectFromJSON — uses the extracted level + message from JSON logs.
// Level field is the source of truth when present.
// Falls back to keyword matching on message when level is absent/ambiguous.
// ─────────────────────────────────────────────────────────────────────────────
func detectFromJSON(parsed input.ParsedLine, lineNum int, context string, cfg *config.Config) *ErrorMatch {

	level := parsed.Level

	if level == "info" || level == "debug" || level == "trace" || level == "warn" || level == "warning" {
		return nil
	}

	isErrorLevel := level == "error" || level == "err" ||
		level == "fatal" || level == "critical" || level == "panic"

	if isErrorLevel {
		errType := classifyMessage(parsed.Message, cfg)
		if errType == "" {
			errType = "General Error"
		}
		return &ErrorMatch{
			LineNumber: lineNum,
			Type:       errType,
			Message:    parsed.Message,
			Context:    context,
		}
	}

	return detectFromPlainText(parsed.Message, lineNum, context, cfg)
}

// ─────────────────────────────────────────────────────────────────────────────
// detectFromPlainText — original keyword-based matching on a plain text line.
// ─────────────────────────────────────────────────────────────────────────────
func detectFromPlainText(line string, lineNum int, context string, cfg *config.Config) *ErrorMatch {

	lower := strings.ToLower(line)

	if strings.Contains(lower, "info") || strings.Contains(lower, "debug") {
		return nil
	}

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
