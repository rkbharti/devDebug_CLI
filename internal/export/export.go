package export

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rkbharti/LogSensei_CLI/internal/patterns"
)

// exportJSON writes errors to JSON file

func ExportJSON(errors []patterns.ErrorMatch) error {
	file, err := os.Create("report.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")
	return encoder.Encode(errors)
}

// exportMarkdown write errors to markdown file
func ExportMarkdown(errors []patterns.ErrorMatch) error {
	file, err := os.Create("report.md")
	if err != nil {
		return err
	}
	defer file.Close()

	content := "# DevDebug Report\n\n"

	for _, e := range errors {
		content += fmt.Sprintf("## Error at Line %d\n", e.LineNumber)
		content += fmt.Sprintf("- Type: %s\n", e.Type)
		content += fmt.Sprintf("- Message: %s\n", e.Message)
		content += fmt.Sprintf("- Context: %s\n\n", e.Context)

	}
	_, err = file.WriteString(content)
	return err
}
