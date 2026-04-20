package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// These are set at build time via -ldflags.
// They fall back to safe defaults when built without flags (local dev).
var (
	Version    = "dev"
	CommitHash = "none"
	BuildDate  = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print LogSensei version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("LogSensei %s\n", Version)
		fmt.Printf("Commit : %s\n", CommitHash)
		fmt.Printf("Built  : %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
