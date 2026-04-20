package analyzer

import "github.com/rkbharti/LogSensei_CLI/internal/patterns"

// made struct for error count and total erros

type Summary struct {
	TotalErrors int
	ErrorCount  map[string]int
}

// AggregateErrors groups errors by type
func AggregateErrors(errors []patterns.ErrorMatch) Summary {
	countMap := make(map[string]int)

	for _, e := range errors {
		countMap[e.Type]++
	}

	return Summary{
		TotalErrors: len(errors),
		ErrorCount:  countMap,
	}
}
