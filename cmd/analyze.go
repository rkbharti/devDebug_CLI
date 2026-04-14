package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rkbharti/devdebug/internal/analyzer"
	"github.com/rkbharti/devdebug/internal/config"
	"github.com/rkbharti/devdebug/internal/export"
	"github.com/rkbharti/devdebug/internal/input"
	"github.com/rkbharti/devdebug/internal/patterns"
	"github.com/rkbharti/devdebug/internal/stacktrace"
	"github.com/rkbharti/devdebug/internal/ui"

	"github.com/spf13/cobra"
)

var filterType string
var outputFormat string
var follow bool

var analyzeCmd = &cobra.Command{
	Use:   "analyze [file]",
	Short: "Analyze log file for errors",
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {

		cfg, err := config.LoadConfig("devdebug.yaml")
		if err != nil {
			fmt.Println("⚠️ Config not loaded (using default rules)")
			cfg = nil
		}

		file := args[0]

		info, err := os.Stat(file)
		if err != nil {
			fmt.Println("❌ Error:", err)
			return
		}

		fmt.Println(ui.SuccessStyle.Render("✅ File found: " + file))
		fmt.Println(ui.InfoStyle.Render("🔍 Starting analysis..."))

		// 🔥 WATCH MODE
		if follow {
			fmt.Println("👀 Watching log file in real-time...")

			err := input.FollowFile(file, func(line string) {

				e := patterns.DetectError(line, 0, "", cfg)

				if e != nil {
					fmt.Println(ui.ErrorStyle.Render("\n🔴 ERROR DETECTED"))
					fmt.Println("Type:", e.Type)
					fmt.Println("Message:", e.Message)

					exp := analyzer.ExplainError(e.Message)

					fmt.Println("\nExplanation:")
					fmt.Println(exp.Reason)

					fmt.Println("\nSuggestion:")
					fmt.Println(exp.Suggestion)
				}
			})

			if err != nil {
				fmt.Println("❌ Watch failed:", err)
			}

			return
		}

		var errors []patterns.ErrorMatch

		// 🔥 FOLDER MODE
		if info.IsDir() {

			files, err := os.ReadDir(file)
			if err != nil {
				fmt.Println("❌ Failed to read directory:", err)
				return
			}

			fmt.Println("📂 Scanning folder:", file)

			for _, f := range files {

				if !strings.HasSuffix(f.Name(), ".log") {
					continue
				}

				fullPath := file + "/" + f.Name()

				fmt.Println("📄 Processing:", f.Name())

				var lastError *patterns.ErrorMatch

				input.ProcessFile(fullPath, func(line string, lineNum int) {

					if lastError != nil {

						if strings.TrimSpace(line) == "" {
							lastError = nil
							return
						}

						lastError.Context += "\n" + line
						return
					}

					e := patterns.DetectError(line, lineNum, "", cfg)
					if e != nil {
						e.File = f.Name()
						errors = append(errors, *e)
						lastError = &errors[len(errors)-1]
					}
				})
			}

		} else {

			// 🔥 SINGLE FILE MODE
			var lastError *patterns.ErrorMatch

			input.ProcessFile(file, func(line string, lineNum int) {

				if lastError != nil {

					if strings.TrimSpace(line) == "" {
						lastError = nil
						return
					}

					lastError.Context += "\n" + line
					return
				}

				e := patterns.DetectError(line, lineNum, "", cfg)
				if e != nil {
					e.File = file
					errors = append(errors, *e)
					lastError = &errors[len(errors)-1]
				}
			})
		}

		fmt.Println(ui.TitleStyle.Render("\n🚨 ERROR REPORT"))

		if filterType != "" {
			fmt.Println(
				ui.InfoStyle.Render("🔍 Showing only:") + " " +
					ui.WarningStyle.Render(filterType) + " errors",
			)
		}

		// 🔥 FILTER
		var summaryData []patterns.ErrorMatch
		for _, e := range errors {
			if filterType != "" {
				if !strings.Contains(strings.ToLower(e.Type), strings.ToLower(filterType)) {
					continue
				}
			}
			summaryData = append(summaryData, e)
		}

		// 🔥 GROUPING (Phase 14 Step 3)
		grouped := make(map[string][]patterns.ErrorMatch)
		for _, e := range summaryData {
			grouped[e.Message] = append(grouped[e.Message], e)
		}

		// 🔥 PRINT GROUPED OUTPUT
		for msg, group := range grouped {

			count := len(group)
			e := group[0]

			fmt.Println(ui.ErrorStyle.Render("🔴 ERROR DETECTED"))

			if count == 1 {
				fmt.Println(ui.InfoStyle.Render(
					fmt.Sprintf("Log Location: %s (Line %d)", e.File, e.LineNumber),
				))
			} else {
				fmt.Println(ui.WarningStyle.Render(
					fmt.Sprintf("Occurred %d times", count),
				))
			}

			fmt.Println("Type:", e.Type)
			fmt.Println(ui.InfoStyle.Render("Message:"), msg)

			exp := analyzer.ExplainError(msg)

			fmt.Println(ui.WarningStyle.Render("\nExplanation:"))
			fmt.Println(exp.Reason)

			fmt.Println(ui.SuccessStyle.Render("\nSuggestion:"))
			fmt.Println(exp.Suggestion)

			combined := e.Message + " " + e.Context
			info := stacktrace.ExtractFileLine(combined)

			fmt.Println(ui.InfoStyle.Render("\n Code Location:"))
			fmt.Println("→ File:", info.File)
			fmt.Println("→ Line:", info.Line)

			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		}

		// 🔥 SUMMARY
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

		// 🔥 FILE SUMMARY
		fileCount := make(map[string]int)
		for _, e := range summaryData {
			fileCount[e.File]++
		}

		fmt.Println("\n📂 FILE SUMMARY")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		for file, count := range fileCount {
			if count == 1 {
				fmt.Printf("%s → %d error\n", file, count)
			} else {
				fmt.Printf("%s → %d errors\n", file, count)
			}
		}

		// 🔥 EXPORT
		if outputFormat != "" {

			var exportErr error

			switch outputFormat {
			case "json":
				exportErr = export.ExportJSON(summaryData)
				fmt.Println("📁 Report exported as report.json")

			case "md":
				exportErr = export.ExportMarkdown(summaryData)
				fmt.Println("📁 Report exported as report.md")

			default:
				fmt.Println("❌ Unsupported format. Use json or md")
				return
			}

			if exportErr != nil {
				fmt.Println("❌ Export failed:", exportErr)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVarP(&filterType, "type", "t", "", "Filter errors by type (panic, error, timeout)")
	analyzeCmd.Flags().StringVarP(&outputFormat, "format", "f", "", "Export Format : json or md")
	analyzeCmd.Flags().BoolVarP(&follow, "follow", "", false, "Follow log file in real time")
}
