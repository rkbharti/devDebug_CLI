package patterns

import "strings"

type ErrorMatch struct {
	LineNumber int
	Type       string
	Message    string
}

// DetectErrors scans logs and finds errors
func DetectErrors(lines []string) []ErrorMatch {
	var errors []ErrorMatch

	for i, line := range lines {

		lower := strings.ToLower(line)

		if strings.Contains(lower, "panic") {
			errors = append(errors, ErrorMatch{
				LineNumber: i + 1,
				Type:       "Panic Error",
				Message:    line,
			})
		} else if strings.Contains(lower, "error") {
			errors = append(errors, ErrorMatch{
				LineNumber: i + 1,
				Type:       "General Error",
				Message:    line,
			})
		} else if strings.Contains(lower, "timeout") {
			errors = append(errors, ErrorMatch{
				LineNumber: i + 1,
				Type:       "Timeout Error",
				Message:    line,
			})
		}
	}

	return errors
}
