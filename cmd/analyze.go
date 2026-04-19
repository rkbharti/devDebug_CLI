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

// ── package-level flag vars ───────────────────────────────────────────────────
var filterType string
var outputFormat string
var follow bool
var quiet bool // 🆕 --quiet flag

// ── command definition ────────────────────────────────────────────────────────
var analyzeCmd = &cobra.Command{
	Use:   "analyze [file or folder]",
	Short: "Analyze log file or folder for errors",
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {

		// ── load config ───────────────────────────────────────────────────────
		cfg, err := config.LoadConfig("devdebug.yaml")
		if err != nil {
			printInfo("⚠️  Config not loaded (using default rules)", quiet)
			cfg = nil
		}

		file := args[0]

		info, err := os.Stat(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, "❌ Error:", err)
			os.Exit(2) // exit 2 = usage/input error (not a log error)
		}

		printInfo(ui.SuccessStyle.Render("✅ File found: "+file), quiet)
		printInfo(ui.InfoStyle.Render("🔍 Starting analysis..."), quiet)

		// ── watch mode (blocking — exits when interrupted) ────────────────────
		if follow {
			handleWatchMode(file, cfg, quiet)
			return
		}

		// ── collect errors (folder or single file) ────────────────────────────
		var errors []patterns.ErrorMatch

		if info.IsDir() {
			errors = handleFolderMode(file, cfg, quiet)
		} else {
			errors = handleSingleFileMode(file, cfg)
		}

		// ── filter ────────────────────────────────────────────────────────────
		summaryData := applyFilter(errors, filterType)

		// ── print report ──────────────────────────────────────────────────────
		printReport(summaryData, filterType, quiet)

		// ── export ────────────────────────────────────────────────────────────
		handleExport(summaryData, outputFormat, quiet)

		// ── CI gate: exit 1 if any errors were found ──────────────────────────
		if len(summaryData) > 0 {
			os.Exit(1)
		}
	},
}

// ── init ──────────────────────────────────────────────────────────────────────
func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVarP(&filterType, "type", "t", "", "Filter errors by type (panic, error, timeout)")
	analyzeCmd.Flags().StringVarP(&outputFormat, "format", "f", "", "Export format: json or md")
	analyzeCmd.Flags().BoolVar(&follow, "follow", false, "Follow log file in real time (watch mode)")
	analyzeCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress output — only exit code (for CI use)") // 🆕
}

// ─────────────────────────────────────────────────────────────────────────────
// STEP 1 — Watch Mode
// ─────────────────────────────────────────────────────────────────────────────

func handleWatchMode(file string, cfg *config.Config, quiet bool) {
	printInfo("👀 Watching log file in real-time... (Ctrl+C to stop)", quiet)

	err := input.FollowFile(file, func(line string) {
		e := patterns.DetectError(line, 0, "", cfg)
		if e == nil {
			return
		}

		// always print errors even in quiet mode — watch mode is interactive
		fmt.Println(ui.ErrorStyle.Render("\n🔴 ERROR DETECTED"))
		fmt.Println("Type   :", e.Type)
		fmt.Println("Message:", e.Message)

		exp := analyzer.ExplainError(e.Message)
		fmt.Println("\nExplanation:", exp.Reason)
		fmt.Println("Suggestion :", exp.Suggestion)
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Watch failed:", err)
		os.Exit(2)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// STEP 2 — Folder Mode
// ─────────────────────────────────────────────────────────────────────────────

func handleFolderMode(dir string, cfg *config.Config, quiet bool) []patterns.ErrorMatch {
	printInfo("📂 Scanning folder: "+dir, quiet)

	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Failed to read directory:", err)
		os.Exit(2)
	}

	var allErrors []patterns.ErrorMatch

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".log") {
			continue
		}

		printInfo("📄 Processing: "+f.Name(), quiet)

		fullPath := dir + "/" + f.Name()
		fileErrors := collectErrors(fullPath, f.Name(), cfg)
		allErrors = append(allErrors, fileErrors...)
	}

	return allErrors
}

// ─────────────────────────────────────────────────────────────────────────────
// STEP 3 — Single File Mode
// ─────────────────────────────────────────────────────────────────────────────

func handleSingleFileMode(file string, cfg *config.Config) []patterns.ErrorMatch {
	return collectErrors(file, file, cfg)
}

// ─────────────────────────────────────────────────────────────────────────────
// SHARED — collectErrors (used by both folder and single file)
// ─────────────────────────────────────────────────────────────────────────────

func collectErrors(filepath string, label string, cfg *config.Config) []patterns.ErrorMatch {
	var errors []patterns.ErrorMatch
	var lastError *patterns.ErrorMatch

	input.ProcessFile(filepath, func(line string, lineNum int) {

		// ── context accumulation ──────────────────────────────────────────────
		if lastError != nil {
			if strings.TrimSpace(line) == "" {
				lastError = nil
				return
			}
			lastError.Context += "\n" + line
			return
		}

		// ── error detection ───────────────────────────────────────────────────
		e := patterns.DetectError(line, lineNum, "", cfg)
		if e != nil {
			e.File = label
			errors = append(errors, *e)
			lastError = &errors[len(errors)-1]
		}
	})

	return errors
}

// ─────────────────────────────────────────────────────────────────────────────
// STEP 4 — Filter
// ─────────────────────────────────────────────────────────────────────────────

func applyFilter(errors []patterns.ErrorMatch, filterType string) []patterns.ErrorMatch {
	if filterType == "" {
		return errors
	}

	var filtered []patterns.ErrorMatch
	for _, e := range errors {
		if strings.Contains(strings.ToLower(e.Type), strings.ToLower(filterType)) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// ─────────────────────────────────────────────────────────────────────────────
// STEP 5 — Print Report
// ─────────────────────────────────────────────────────────────────────────────

func printReport(summaryData []patterns.ErrorMatch, filterType string, quiet bool) {
	if quiet {
		return // silent mode — no output, only exit code matters
	}

	fmt.Println(ui.TitleStyle.Render("\n🚨 ERROR REPORT"))

	if filterType != "" {
		fmt.Println(
			ui.InfoStyle.Render("🔍 Showing only:") + " " +
				ui.WarningStyle.Render(filterType) + " errors",
		)
	}

	// ── group by message ──────────────────────────────────────────────────────
	grouped := make(map[string][]patterns.ErrorMatch)
	for _, e := range summaryData {
		grouped[e.Message] = append(grouped[e.Message], e)
	}

	// ── print each group ──────────────────────────────────────────────────────
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

		fmt.Println("Type   :", e.Type)
		fmt.Println(ui.InfoStyle.Render("Message:"), msg)

		exp := analyzer.ExplainError(msg)
		fmt.Println(ui.WarningStyle.Render("\nExplanation:"))
		fmt.Println(exp.Reason)
		fmt.Println(ui.SuccessStyle.Render("\nSuggestion:"))
		fmt.Println(exp.Suggestion)

		combined := e.Message + " " + e.Context
		loc := stacktrace.ExtractFileLine(combined)
		fmt.Println(ui.InfoStyle.Render("\nCode Location:"))
		fmt.Println("→ File:", loc.File)
		fmt.Println("→ Line:", loc.Line)

		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	}

	// ── summary ───────────────────────────────────────────────────────────────
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

	// ── file summary ──────────────────────────────────────────────────────────
	fileCount := make(map[string]int)
	for _, e := range summaryData {
		fileCount[e.File]++
	}

	fmt.Println("\n📂 FILE SUMMARY")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for f, count := range fileCount {
		if count == 1 {
			fmt.Printf("%s → %d error\n", f, count)
		} else {
			fmt.Printf("%s → %d errors\n", f, count)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// STEP 6 — Export
// ─────────────────────────────────────────────────────────────────────────────

func handleExport(summaryData []patterns.ErrorMatch, outputFormat string, quiet bool) {
	if outputFormat == "" {
		return
	}

	var exportErr error

	switch outputFormat {
	case "json":
		exportErr = export.ExportJSON(summaryData)
		printInfo("📁 Report exported as report.json", quiet)

	case "md":
		exportErr = export.ExportMarkdown(summaryData)
		printInfo("📁 Report exported as report.md", quiet)

	default:
		fmt.Fprintln(os.Stderr, "❌ Unsupported format. Use: json or md")
		os.Exit(2)
	}

	if exportErr != nil {
		fmt.Fprintln(os.Stderr, "❌ Export failed:", exportErr)
		os.Exit(2)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// HELPER — printInfo respects --quiet flag
// ─────────────────────────────────────────────────────────────────────────────

func printInfo(msg string, quiet bool) {
	if !quiet {
		fmt.Println(msg)
	}
}
