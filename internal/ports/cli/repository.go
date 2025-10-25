package cli

import "github.com/spf13/cobra"

// NewRepositoryCmd creates a new cobra command for the "repository" command.
func NewRepositoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repository",
		Short: "Interact with the interview repository",
		Long:  `Interact with the interview repository.`,
	}

	cmd.AddCommand(NewRepositoryListCmd())
	cmd.AddCommand(NewRepositoryViewCmd())
	cmd.AddCommand(NewRepositoryExportCmd())

	return cmd
}
