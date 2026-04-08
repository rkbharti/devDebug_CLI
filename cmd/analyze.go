package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rkbharti/devdebug/internal/analyzer"
	"github.com/rkbharti/devdebug/internal/input"
	"github.com/rkbharti/devdebug/internal/patterns"
	"github.com/rkbharti/devdebug/internal/stacktrace"
	"github.com/rkbharti/devdebug/internal/ui"

	"github.com/spf13/cobra"
)

var filterType string

var analyzeCmd = &cobra.Command{
	Use:   "analyze [file]",
	Short: "Analyze log file for errors",
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]

		// check file exists
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Println(ui.ErrorStyle.Render("❌ File does not exist: " + file))
			return
		}

		fmt.Println(ui.SuccessStyle.Render("✅ File found: " + file))
		fmt.Println(ui.InfoStyle.Render("🔍 Starting analysis..."))

		// read file
		lines, err := input.ReadFile(file)
		if err != nil {
			fmt.Println(ui.ErrorStyle.Render("❌ Error reading file: " + err.Error()))
			return
		}

		input.PrintLines(lines)

		// detect errors
		errors := patterns.DetectErrors(lines)

		fmt.Println(ui.TitleStyle.Render("\n🚨 ERROR REPORT"))

		// 🔥 Show filter ONCE (correct placement)
		if filterType != "" {
			fmt.Println(
				ui.InfoStyle.Render("🔍 Showing only:") + " " +
					ui.WarningStyle.Render(filterType) + " errors",
			)
		}
		var filteredErrors []patterns.ErrorMatch

		for _, e := range errors {

			// 🔥 Apply filter
			if filterType != "" {
				if !strings.Contains(strings.ToLower(e.Type), strings.ToLower(filterType)) {
					continue
				}
			}
			filteredErrors = append(filteredErrors, e)

			fmt.Println(ui.ErrorStyle.Render(
				fmt.Sprintf("🔴 ERROR DETECTED (Line %d)", e.LineNumber),
			))

			fmt.Println("Type:", e.Type)

			fmt.Println(ui.InfoStyle.Render("Message:"), e.Message)

			// 🔥 Explanation (Phase 6)
			exp := analyzer.ExplainError(e.Message)

			fmt.Println(ui.WarningStyle.Render("\nExplanation:"))
			fmt.Println(exp.Reason)

			fmt.Println(ui.SuccessStyle.Render("\nSuggestion:"))
			fmt.Println(exp.Suggestion)

			// 🔥 Location extraction (robust)
			combined := e.Message + " " + e.Context
			info := stacktrace.ExtractFileLine(combined)

			fmt.Println(ui.InfoStyle.Render("\nLocation:"))
			fmt.Println("→ File:", info.File)
			fmt.Println("→ Line:", info.Line)

			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		}

		// 🔥 Summary
		var summaryData []patterns.ErrorMatch

		if filterType != "" {
			summaryData = filteredErrors
		} else {
			summaryData = errors
		}

		summary := analyzer.AggregateErrors(summaryData)
		fmt.Println(ui.TitleStyle.Render("\n📊 SUMMARY REPORT"))
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		fmt.Printf("Total Errors: %d\n\n", summary.TotalErrors)

		fmt.Println("Top Issues:")

		for errType, count := range summary.ErrorCount {
			if count == 1 {
				fmt.Printf("• %s → %d time\n", errType, count)
			} else {
				fmt.Printf("• %s → %d times\n", errType, count)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVarP(
		&filterType,
		"type",
		"t",
		"",
		"Filter errors by type (panic, error, timeout)",
	)
}
