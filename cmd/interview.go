package cmd

import (
	"github.com/andrewhowdencom/vox/internal/ports/cli"
	"github.com/spf13/cobra"
)

// NewInterviewCmd creates the interview command and adds its subcommands.
func NewInterviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interview",
		Short: "Commands for managing interviews",
	}

	cmd.AddCommand(cli.NewStartCmd(nil))
	cmd.AddCommand(cli.NewRepositoryCmd())

	return cmd
}
