package analyzer

import "strings"

type Explanation struct {
	Reason     string
	Suggestion string
}

// ExplainError returns human-friendly explanation
func ExplainError(message string) Explanation {

	msg := strings.ToLower(message)

	// Panic / nil pointer
	if strings.Contains(msg, "nil pointer") || strings.Contains(msg, "invalid memory address") {
		return Explanation{
			Reason:     "You are trying to use a variable that is not initialized (nil pointer).",
			Suggestion: "Check if the variable is nil before using it.",
		}
	}

	// Timeout
	if strings.Contains(msg, "timeout") {
		return Explanation{
			Reason:     "The operation took too long and timed out.",
			Suggestion: "Check network, server load, or increase timeout limit.",
		}
	}

	// TypeError (JS)
	if strings.Contains(msg, "cannot read property") {
		return Explanation{
			Reason:     "You are accessing a property on an undefined object.",
			Suggestion: "Ensure the object exists before accessing its properties.",
		}
	}
	//database error(db)
	if strings.Contains(msg, "database") || strings.Contains(msg, "connection failed") {
		return Explanation{
			Reason:     "The application failed to connect to the database.",
			Suggestion: "Check database server, credentials, and network connection.",
		}
	}

	// Default fallback
	return Explanation{
		Reason:     "Unknown error occurred.",
		Suggestion: "Check logs and debug manually.",
	}
}
