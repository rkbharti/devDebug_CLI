package patterns

import (
	"strings"
)

type ErrorMatch struct {
	LineNumber int
	Type       string
	Message    string
	Context    string
	File       string
}

// DetectErrors scans logs and finds errors
func DetectError(line string, lineNum int, context string) *ErrorMatch {
	lower := strings.ToLower(line)

	if strings.Contains(lower, "panic") {
		return &ErrorMatch{
			LineNumber: lineNum,
			Type:       "Panic Error",
			Message:    line,
			Context:    context,
		}
	}

	if strings.Contains(lower, "error") {
		return &ErrorMatch{
			LineNumber: lineNum,
			Type:       "General Error",
			Message:    line,
			Context:    context,
		}
	}

	if strings.Contains(lower, "timeout") {
		return &ErrorMatch{
			LineNumber: lineNum,
			Type:       "Timeout Error",
			Message:    line,
			Context:    context,
		}
	}

	return nil
}
