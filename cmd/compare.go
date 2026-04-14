package cmd

import (
	"fmt"
	"os"

	"github.com/rkbharti/devdebug/internal/analyzer"
	"github.com/rkbharti/devdebug/internal/input"
	"github.com/rkbharti/devdebug/internal/patterns"
	"github.com/spf13/cobra"
)

var compareCmd = &cobra.Command{
	Use:   "compare [old] [new]",
	Short: "Compare two log files",
	Args:  cobra.ExactArgs(2),

	Run: func(cmd *cobra.Command, args []string) {

		oldFile := args[0]
		newFile := args[1]

		// check files
		if _, err := os.Stat(oldFile); os.IsNotExist(err) {
			fmt.Println("❌ Old file not found:", oldFile)
			return
		}
		if _, err := os.Stat(newFile); os.IsNotExist(err) {
			fmt.Println("❌ New file not found:", newFile)
			return
		}

		fmt.Println("🔍 Comparing logs...")

		oldErrors := analyzeFile(oldFile)
		newErrors := analyzeFile(newFile)

		oldSummary := analyzer.AggregateErrors(oldErrors)
		newSummary := analyzer.AggregateErrors(newErrors)

		fmt.Println("📊 COMPARISON RESULT")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		fmt.Printf("Old Errors: %d\n", oldSummary.TotalErrors)
		fmt.Printf("New Errors: %d\n\n", newSummary.TotalErrors)

		if newSummary.TotalErrors > oldSummary.TotalErrors {
			fmt.Println("🚨 Regression detected!")
		} else if newSummary.TotalErrors < oldSummary.TotalErrors {
			fmt.Println("✅ Improvement detected!")
		} else {
			fmt.Println("⚖️ No change in error count")
		}
		oldMap := buildErrorMap(oldErrors, oldFile)
		newMap := buildErrorMap(newErrors, newFile)

		fmt.Println("\n🔍 DETAILED DIFF")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// 🚨 New Errors
		fmt.Println("\n🚨 New Errors:")

		foundNew := false

		for msg, files := range newMap {
			if _, exists := oldMap[msg]; !exists {

				count := len(files)

				if count == 1 {
					fmt.Printf("+ %s (%d time)\n", msg, count)
				} else {
					fmt.Printf("+ %s (%d times)\n", msg, count)
				}

				fmt.Println("  → File:", files[0])

				foundNew = true
			}
		}

		if !foundNew {
			fmt.Println("None")
		}

		// ✅ Fixed Errors
		fmt.Println("\n✅ Fixed Errors:")

		foundFixed := false

		for msg, files := range oldMap {
			if _, exists := newMap[msg]; !exists {

				count := len(files)

				if count == 1 {
					fmt.Printf("- %s (%d time)\n", msg, count)
				} else {
					fmt.Printf("- %s (%d times)\n", msg, count)
				}

				fmt.Println("  → File:", files[0])

				foundFixed = true // ✅ IMPORTANT
			}
		}

		if !foundFixed {
			fmt.Println("None")
		}

		// ⚖️ Unchanged Errors
		fmt.Println("\n⚖️ Unchanged Errors:")
		foundSame := false
		for msg, files := range newMap {
			if _, exists := oldMap[msg]; exists {

				count := len(files)

				if count == 1 {
					fmt.Printf("= %s (%d time)\n", msg, count)
				} else {
					fmt.Printf("= %s (%d times)\n", msg, count)
				}

				fmt.Println("  → File:", newFile)
				foundSame = true
			}
		}
		if !foundSame {
			fmt.Println("None")
		}
	},
}

// reuse your existing logic
func analyzeFile(file string) []patterns.ErrorMatch {

	var errors []patterns.ErrorMatch
	var lastError *patterns.ErrorMatch

	input.ProcessFile(file, func(line string, lineNum int) {

		if lastError != nil && line != "" {
			lastError.Context = line
			lastError = nil
			return
		}

		e := patterns.DetectError(line, lineNum, "")
		if e != nil {
			errors = append(errors, *e)
			lastError = &errors[len(errors)-1]
		}
	})

	return errors
}

func buildErrorMap(errors []patterns.ErrorMatch, fileName string) map[string][]string {
	m := make(map[string][]string)

	for _, e := range errors {
		m[e.Message] = append(m[e.Message], fileName)
	}

	return m
}

func init() {
	rootCmd.AddCommand(compareCmd)
}
