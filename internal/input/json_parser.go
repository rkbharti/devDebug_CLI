package input

import (
	"encoding/json"
	"strings"
	"time"
)

// ParsedLine is a normalised log line — regardless of whether the
// original was plain text or JSON, everything comes out the same shape.
type ParsedLine struct {
	Raw       string            // original line as-is
	Message   string            // extracted or original message
	Level     string            // "error", "warn", "info", "debug" — lowercased
	IsJSON    bool              // true if line was valid JSON
	Fields    map[string]string // all other string fields from JSON
	Timestamp time.Time         // 🆕 extracted timestamp (zero value if not found)
}

// ─────────────────────────────────────────────────────────────────────────────
// knownMessageKeys — JSON field names commonly used for the main message.
// ─────────────────────────────────────────────────────────────────────────────
var knownMessageKeys = []string{
	"message", "msg", "error", "err",
	"log", "text", "body", "detail",
}

// knownLevelKeys — JSON field names commonly used for log level.
var knownLevelKeys = []string{
	"level", "severity", "lvl", "log_level", "loglevel",
}

// knownTimestampKeys — JSON field names commonly used for timestamps.
var knownTimestampKeys = []string{
	"timestamp", "time", "ts", "@timestamp",
	"datetime", "date", "logged_at", "created_at",
}

// timestampFormats — common timestamp formats found in real log files.
// Tried in order — first successful parse wins.
var timestampFormats = []string{
	time.RFC3339Nano,               // 2026-04-19T07:41:00.000Z
	time.RFC3339,                   // 2026-04-19T07:41:00Z
	"2006-01-02T15:04:05.000Z0700", // ISO8601 with milliseconds
	"2006-01-02T15:04:05",          // ISO8601 without timezone
	"2006-01-02 15:04:05.000",      // Space-separated with ms
	"2006-01-02 15:04:05",          // Space-separated plain
	"2006/01/02 15:04:05",          // Slash-separated (nginx style)
	"02/Jan/2006:15:04:05 -0700",   // Apache access log format
	"Jan 02 15:04:05",              // Syslog format
}

// ─────────────────────────────────────────────────────────────────────────────
// ParseLine takes a raw log line and returns a ParsedLine.
// If the line is valid JSON it extracts message + level + timestamp.
// If it is plain text it wraps it as-is and tries to parse timestamp from text.
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

	// ── extract timestamp 🆕 ──────────────────────────────────────────────────
	ts := extractTimestampFromJSON(obj)

	// ── collect remaining string fields ──────────────────────────────────────
	fields := make(map[string]string)
	for k, v := range obj {
		if str, ok := v.(string); ok {
			fields[k] = str
		}
	}

	return ParsedLine{
		Raw:       raw,
		Message:   message,
		Level:     level,
		IsJSON:    true,
		Fields:    fields,
		Timestamp: ts, // 🆕
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// extractTimestampFromJSON tries known timestamp keys and formats.
// Returns zero time.Time if nothing matches.
// ─────────────────────────────────────────────────────────────────────────────
func extractTimestampFromJSON(obj map[string]interface{}) time.Time {

	for _, key := range knownTimestampKeys {
		val, ok := obj[key]
		if !ok {
			continue
		}

		switch v := val.(type) {

		case string:
			// try each known format
			if t := parseTimestampString(v); !t.IsZero() {
				return t
			}

		case float64:
			// Unix timestamp as number
			// detect seconds vs milliseconds by magnitude
			if v > 1e12 {
				// milliseconds
				return time.UnixMilli(int64(v)).UTC()
			}
			// seconds
			return time.Unix(int64(v), 0).UTC()
		}
	}

	return time.Time{} // zero — no timestamp found
}

// parseTimestampString tries all known formats against a string value.
func parseTimestampString(s string) time.Time {
	s = strings.TrimSpace(s)
	for _, format := range timestampFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

// plainLine wraps a plain text line into a ParsedLine with no JSON fields.
func plainLine(raw string) ParsedLine {
	return ParsedLine{
		Raw:       raw,
		Message:   raw,
		Level:     "",
		IsJSON:    false,
		Fields:    nil,
		Timestamp: time.Time{}, // zero — plain text has no parsed timestamp
	}
}
