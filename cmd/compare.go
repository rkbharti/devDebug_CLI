package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rkbharti/LogSensei_CLI/internal/analyzer"
	"github.com/rkbharti/LogSensei_CLI/internal/config"
	"github.com/rkbharti/LogSensei_CLI/internal/input"
	"github.com/rkbharti/LogSensei_CLI/internal/patterns"

	"github.com/spf13/cobra"
)

var compareCmd = &cobra.Command{
	Use:   "compare [old] [new]",
	Short: "Compare two log files",
	Args:  cobra.ExactArgs(2),

	Run: func(cmd *cobra.Command, args []string) {

		cfg, err := config.LoadConfig("devdebug.yaml")
		if err != nil {
			fmt.Println("⚠️ Config not loaded (using default rules)")
			cfg = nil
		}
		compiled := cfg.Compile()

		oldFile := args[0]
		newFile := args[1]

		if _, err := os.Stat(oldFile); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "❌ Old file not found:", oldFile)
			os.Exit(2)
		}
		if _, err := os.Stat(newFile); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "❌ New file not found:", newFile)
			os.Exit(2)
		}

		fmt.Println("🔍 Comparing logs...")

		oldErrors := analyzeFile(oldFile, compiled)
		newErrors := analyzeFile(newFile, compiled)

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

		// ── new errors ────────────────────────────────────────────────────────
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

		// ── fixed errors ──────────────────────────────────────────────────────
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
				foundFixed = true
			}
		}
		if !foundFixed {
			fmt.Println("None")
		}

		// ── unchanged errors ──────────────────────────────────────────────────
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
				_ = files
				foundSame = true
			}
		}
		if !foundSame {
			fmt.Println("None")
		}
	},
}

// ─────────────────────────────────────────────────────────────────────────────
// analyzeFile reads a log file and returns all detected errors.
// Used by compare command for both old and new files.
// ─────────────────────────────────────────────────────────────────────────────

func analyzeFile(file string, compiled []config.CompiledPattern) []patterns.ErrorMatch {

	var errors []patterns.ErrorMatch
	var lastError *patterns.ErrorMatch

	input.ProcessFile(file, func(parsed input.ParsedLine, lineNum int) {

		// ── context accumulation ──────────────────────────────────────────────
		if lastError != nil {
			if strings.TrimSpace(parsed.Raw) == "" { // ✅ fixed
				lastError = nil
				return
			}
			lastError.Context += "\n" + parsed.Raw // ✅ fixed
			return
		}

		// ── error detection ───────────────────────────────────────────────────
		e := patterns.DetectError(parsed, lineNum, "", compiled)
		if e != nil {
			errors = append(errors, *e)
			lastError = &errors[len(errors)-1]
		}
	})

	return errors
}

// buildErrorMap groups errors by message, mapping message → list of filenames.
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
