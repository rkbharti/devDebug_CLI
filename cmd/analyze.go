package cmd

import (
	"fmt"
	"os"

	"github.com/rkbharti/devdebug/internal/input"
	"github.com/rkbharti/devdebug/internal/patterns"
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
			fmt.Println("-----------------------------")
		}

		fmt.Printf("\n Total Erros Found : %d\n", len(errors))

	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
