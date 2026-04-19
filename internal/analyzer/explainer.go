package analyzer

import "strings"

// Explanation holds a human-readable reason and fix suggestion.
type Explanation struct {
	Reason     string
	Suggestion string
}

// ─────────────────────────────────────────────────────────────────────────────
// explanationRule defines one matchable error pattern.
// keyword   — substring to match against the lowercased message
// reason    — human-readable cause
// suggestion — actionable fix
// ─────────────────────────────────────────────────────────────────────────────

type explanationRule struct {
	keyword    string
	reason     string
	suggestion string
}

// ─────────────────────────────────────────────────────────────────────────────
// explanationRegistry is the single source of truth for all error explanations.
//
// HOW TO ADD A NEW ERROR TYPE:
//   Just append a new explanationRule{} to this slice.
//   No logic changes needed anywhere else.
//
// ORDER MATTERS — first match wins.
// Put more specific patterns before generic ones.
// ─────────────────────────────────────────────────────────────────────────────

var explanationRegistry = []explanationRule{

	// ── Go runtime errors ─────────────────────────────────────────────────────
	{
		keyword:    "nil pointer",
		reason:     "A variable is being used before it was initialized (nil pointer dereference).",
		suggestion: "Check if the variable is nil before using it. Use guard clauses: if x == nil { return }",
	},
	{
		keyword:    "invalid memory address",
		reason:     "A variable is being used before it was initialized (nil pointer dereference).",
		suggestion: "Check if the variable is nil before using it. Use guard clauses: if x == nil { return }",
	},
	{
		keyword:    "index out of range",
		reason:     "You are accessing a slice or array at an index that does not exist.",
		suggestion: "Check the slice length before indexing. Use len(slice) > index as a guard.",
	},
	{
		keyword:    "stack overflow",
		reason:     "A function is calling itself infinitely (infinite recursion).",
		suggestion: "Check your recursive function for a proper base/termination case.",
	},
	{
		keyword:    "deadlock",
		reason:     "All goroutines are waiting on each other — the program is stuck.",
		suggestion: "Check for circular mutex locks or channels that are never read/written.",
	},
	{
		keyword:    "goroutine leak",
		reason:     "A goroutine was started but never terminated, causing memory to grow.",
		suggestion: "Ensure every goroutine has a clear exit path. Use context cancellation.",
	},

	// ── network / timeout ─────────────────────────────────────────────────────
	{
		keyword:    "connection refused",
		reason:     "The target server is not running or is not accepting connections.",
		suggestion: "Verify the server is running and the port is correct. Check firewall rules.",
	},
	{
		keyword:    "connection reset",
		reason:     "The connection was forcibly closed by the remote server.",
		suggestion: "Check server-side logs for why the connection was dropped. May be a crash or restart.",
	},
	{
		keyword:    "timeout",
		reason:     "The operation took too long and timed out.",
		suggestion: "Check network latency, server load, or increase the timeout limit.",
	},
	{
		keyword:    "no such host",
		reason:     "The hostname could not be resolved via DNS.",
		suggestion: "Check the hostname for typos. Verify DNS settings and network connectivity.",
	},
	{
		keyword:    "tls",
		reason:     "A TLS/SSL handshake or certificate error occurred.",
		suggestion: "Check certificate validity, expiry, and whether the CA is trusted.",
	},

	// ── database errors ───────────────────────────────────────────────────────
	{
		keyword:    "connection failed",
		reason:     "The application failed to connect to the database.",
		suggestion: "Check database host, port, credentials, and whether the DB server is running.",
	},
	{
		keyword:    "database",
		reason:     "A database operation failed.",
		suggestion: "Check database server status, credentials, and network connection.",
	},
	{
		keyword:    "deadlock found",
		reason:     "Two database transactions are blocking each other.",
		suggestion: "Review transaction order and add retry logic for deadlock errors.",
	},
	{
		keyword:    "duplicate entry",
		reason:     "A unique constraint was violated in the database.",
		suggestion: "Check for duplicate data before inserting. Use INSERT IGNORE or ON CONFLICT.",
	},

	// ── authentication / authorization ────────────────────────────────────────
	{
		keyword:    "unauthorized",
		reason:     "The request lacks valid authentication credentials.",
		suggestion: "Check API keys, JWT tokens, or session credentials. Ensure they are not expired.",
	},
	{
		keyword:    "forbidden",
		reason:     "The authenticated user does not have permission for this action.",
		suggestion: "Check role/permission configuration. Verify the user has required access level.",
	},
	{
		keyword:    "token expired",
		reason:     "The authentication token has expired.",
		suggestion: "Implement token refresh logic. Check token TTL configuration.",
	},

	// ── file system errors ────────────────────────────────────────────────────
	{
		keyword:    "no such file or directory",
		reason:     "A required file or directory does not exist at the expected path.",
		suggestion: "Verify the file path. Check working directory and relative vs absolute paths.",
	},
	{
		keyword:    "permission denied",
		reason:     "The process does not have permission to access the file or resource.",
		suggestion: "Check file permissions with ls -la. Run with appropriate privileges or fix ownership.",
	},
	{
		keyword:    "disk full",
		reason:     "The disk has run out of space.",
		suggestion: "Free up disk space. Check with df -h. Consider log rotation or storage expansion.",
	},

	// ── JavaScript / frontend ─────────────────────────────────────────────────
	{
		keyword:    "cannot read property",
		reason:     "You are accessing a property on an undefined or null object.",
		suggestion: "Ensure the object exists before accessing its properties. Use optional chaining: obj?.property",
	},
	{
		keyword:    "is not a function",
		reason:     "You are calling something that is not a function.",
		suggestion: "Check the variable type before calling it. It may be undefined or overwritten.",
	},
	{
		keyword:    "syntaxerror",
		reason:     "There is a JavaScript syntax error in the code.",
		suggestion: "Check for missing brackets, quotes, or semicolons near the reported line.",
	},

	// ── Python errors ─────────────────────────────────────────────────────────
	{
		keyword:    "traceback",
		reason:     "A Python exception occurred — see the traceback for the call stack.",
		suggestion: "Read the last line of the traceback for the actual error. Fix the root cause shown there.",
	},
	{
		keyword:    "attributeerror",
		reason:     "You are accessing an attribute that does not exist on the object.",
		suggestion: "Check the object type with type(obj). Verify the attribute name is correct.",
	},
	{
		keyword:    "importerror",
		reason:     "A Python module could not be imported.",
		suggestion: "Run pip install <module> or check your virtual environment is activated.",
	},
	{
		keyword:    "keyerror",
		reason:     "A dictionary key does not exist.",
		suggestion: "Use dict.get(key, default) instead of dict[key] to safely access keys.",
	},
}

// ─────────────────────────────────────────────────────────────────────────────
// ExplainError returns a human-readable explanation for a given error message.
// It matches the first rule in explanationRegistry whose keyword appears
// in the lowercased message. Falls back to a generic explanation if no match.
// ─────────────────────────────────────────────────────────────────────────────

func ExplainError(message string) Explanation {
	lower := strings.ToLower(message)

	for _, rule := range explanationRegistry {
		if strings.Contains(lower, rule.keyword) {
			return Explanation{
				Reason:     rule.reason,
				Suggestion: rule.suggestion,
			}
		}
	}

	// fallback — unknown error
	return Explanation{
		Reason:     "Unknown error occurred.",
		Suggestion: "Check logs and debug manually.",
	}
}
