package cli

import (
	"fmt"

	"github.com/andrewhowdencom/vox/internal/adapters/storage/bbolt"
	"github.com/andrewhowdencom/vox/internal/domain/storage"
	"github.com/spf13/cobra"
)

// NewRepositoryViewCmd creates a new cobra command for the "repository view" command.
func NewRepositoryViewCmd() *cobra.Command {
	return newRepositoryViewCmd(func() (storage.Repository, error) {
		return bbolt.NewRepository()
	})
}

func newRepositoryViewCmd(repoFn func() (storage.Repository, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view [id]",
		Short: "View a single interview",
		Long:  `View a single interview.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			full, _ := cmd.Flags().GetBool("full")

			repo, err := repoFn()
			if err != nil {
				return fmt.Errorf("could not create repository: %w", err)
			}

			interview, err := repo.GetInterview(id)
			if err != nil {
				return fmt.Errorf("could not get interview: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Interview ID: %s\n", interview.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "User: %s\n", interview.UserID)
			fmt.Fprintf(cmd.OutOrStdout(), "Project: %s\n", interview.ProjectID)
			fmt.Fprintf(cmd.OutOrStdout(), "Created At: %s\n", interview.CreatedAt.String())

			if full {
				transcript, err := repo.GetTranscript(id)
				if err != nil {
					return fmt.Errorf("could not get transcript: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "\n--- Transcript ---")
				for _, entry := range transcript.Entries {
					fmt.Fprintf(cmd.OutOrStdout(), "Q: %s\nA: %s\n\n", entry.Question, entry.Answer)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "------------------")
			} else {
				summary, err := repo.GetSummary(id)
				if err != nil {
					return fmt.Errorf("could not get summary: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "\n--- Summary ---")
				fmt.Fprintln(cmd.OutOrStdout(), summary.Text)
				fmt.Fprintln(cmd.OutOrStdout(), "---------------")
			}

			return nil
		},
	}
	cmd.Flags().Bool("full", false, "Show the full transcript instead of the summary")
	return cmd
}
