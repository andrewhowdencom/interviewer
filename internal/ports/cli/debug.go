package cli

import (
	"github.com/spf13/cobra"
)

// NewDebugCmd creates the debug command and adds its subcommands.
func NewDebugCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Debugging commands",
	}

	cmd.AddCommand(newDebugConfigCmd())

	return cmd
}
