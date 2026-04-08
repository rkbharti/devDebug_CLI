package patterns

import (
	"strings"
)

type ErrorMatch struct {
	LineNumber int
	Type       string
	Message    string
	Context    string
}

// DetectErrors scans logs and finds errors
func DetectErrors(lines []string) []ErrorMatch {
	var errors []ErrorMatch

	for i, line := range lines {

		lower := strings.ToLower(line)

		// to get next line of log/ code / or any file
		var context string

		// check next lines safely
		for j := i + 1; j < len(lines); j++ {
			next := strings.TrimSpace(lines[j])

			if next == "" {
				continue
			}

			context = next
			break
		}

		if strings.Contains(lower, "panic") {
			errors = append(errors, ErrorMatch{
				LineNumber: i + 1,
				Type:       "Panic Error",
				Message:    line,
				Context:    context,
			})
		} else if strings.Contains(lower, "error") {
			errors = append(errors, ErrorMatch{
				LineNumber: i + 1,
				Type:       "General Error",
				Message:    line,
				Context:    context,
			})
		} else if strings.Contains(lower, "timeout") {
			errors = append(errors, ErrorMatch{
				LineNumber: i + 1,
				Type:       "Timeout Error",
				Message:    line,
				Context:    context,
			})
		}
	}

	return errors
}
