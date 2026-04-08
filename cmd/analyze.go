package cmd

import (
	"fmt"
	"os"

	"github.com/rkbharti/devdebug/internal/analyzer"
	"github.com/rkbharti/devdebug/internal/input"
	"github.com/rkbharti/devdebug/internal/patterns"
	"github.com/rkbharti/devdebug/internal/stacktrace"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [file]",
	Short: "Analyze log file for errors",
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]

		// check file exists
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Println("❌ File does not exist:", file)
			return
		}
		fmt.Println("✅ File found:", file)
		fmt.Println("🔍 Starting analysis...")

		// New Logic
		lines, err := input.ReadFile(file)
		if err != nil {
			fmt.Println("❌ Error reading file:", err)
			return
		}

		input.PrintLines(lines)

		// Detect erros
		errors := patterns.DetectErrors(lines)

		fmt.Println("\n🚨 Error Report:")

		for _, e := range errors {
			fmt.Printf("🔴 ERROR DETECTED (Line %d)\n", e.LineNumber)
			fmt.Println("Type:", e.Type)

			fmt.Println("Message:", e.Message)

			// 🔥 PHASE 6 ADD HERE
			exp := analyzer.ExplainError(e.Message)

			fmt.Println("\nExplanation:")
			fmt.Println(exp.Reason)

			fmt.Println("\nSuggestion:")
			fmt.Println(exp.Suggestion)

			// existing stack trace logic
			info := stacktrace.ExtractFileLine(e.Context)

			fmt.Println("\nLocation:")
			fmt.Println("→ File:", info.File)
			fmt.Println("→ Line:", info.Line)

			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		}

		// 🔥 Aggregate Errors``
		summary := analyzer.AggregateErrors(errors)

		fmt.Println("\n📊 SUMMARY REPORT")
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
}
