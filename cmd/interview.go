package cmd

import (
	"github.com/andrewhowdencom/vox/internal/ports/cli"
	"github.com/spf13/cobra"
)

// newInterviewCmd creates the interview command and adds its subcommands.
// This is a provider for Wire.
func newInterviewCmd(startCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interview",
		Short: "Commands for interviewing customers",
	}

	// The start command is now created by its own package and passed in.
	cmd.AddCommand(startCmd)

	return cmd
}

// provideStartCmd is a Wire provider that creates the start command.
func provideStartCmd() *cobra.Command {
	return cli.NewStartCmd(nil) // Passing nil for out, as cobra handles it.
}
