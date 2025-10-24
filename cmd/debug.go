package cmd

import (
	"github.com/spf13/cobra"
)

// debugCmd represents the debug command
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug commands for ruf.",
	Long: `Debug commands for ruf.`,
}

func init() {
	rootCmd.AddCommand(debugCmd)
}
