package input

import (
	"encoding/json"
	"strings"
)

// ParsedLine is a normalised log line — regardless of whether the
// original was plain text or JSON, everything comes out the same shape.
type ParsedLine struct {
	Raw     string            // original line as-is
	Message string            // extracted or original message
	Level   string            // "error", "warn", "info", "debug" — lowercased
	IsJSON  bool              // true if line was valid JSON
	Fields  map[string]string // all other string fields from JSON
}

// ─────────────────────────────────────────────────────────────────────────────
// knownMessageKeys — JSON field names commonly used for the main message.
// Checked in order — first match wins.
// ─────────────────────────────────────────────────────────────────────────────
var knownMessageKeys = []string{
	"message", "msg", "error", "err",
	"log", "text", "body", "detail",
}

// knownLevelKeys — JSON field names commonly used for log level.
var knownLevelKeys = []string{
	"level", "severity", "lvl", "log_level", "loglevel",
}

// ─────────────────────────────────────────────────────────────────────────────
// ParseLine takes a raw log line and returns a ParsedLine.
// If the line is valid JSON it extracts message + level from known fields.
// If it is plain text it wraps it as-is.
// ─────────────────────────────────────────────────────────────────────────────
func ParseLine(raw string) ParsedLine {
	trimmed := strings.TrimSpace(raw)

	// fast path — must start with { to be JSON
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return plainLine(raw)
	}

	// try to parse as JSON
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &obj); err != nil {
		// not valid JSON — treat as plain text
		return plainLine(raw)
	}

	// ── extract message ───────────────────────────────────────────────────────
	message := ""
	for _, key := range knownMessageKeys {
		if val, ok := obj[key]; ok {
			if str, ok := val.(string); ok && str != "" {
				message = str
				break
			}
		}
	}

	// fallback — no known message key found, use full raw line
	if message == "" {
		message = raw
	}

	// ── extract level ─────────────────────────────────────────────────────────
	level := ""
	for _, key := range knownLevelKeys {
		if val, ok := obj[key]; ok {
			if str, ok := val.(string); ok {
				level = strings.ToLower(str)
				break
			}
		}
	}

	// ── collect remaining string fields (for context) ─────────────────────────
	fields := make(map[string]string)
	for k, v := range obj {
		if str, ok := v.(string); ok {
			fields[k] = str
		}
	}

	return ParsedLine{
		Raw:     raw,
		Message: message,
		Level:   level,
		IsJSON:  true,
		Fields:  fields,
	}
}

// plainLine wraps a plain text line into a ParsedLine with no JSON fields.
func plainLine(raw string) ParsedLine {
	return ParsedLine{
		Raw:     raw,
		Message: raw,
		Level:   "",
		IsJSON:  false,
		Fields:  nil,
	}
}
