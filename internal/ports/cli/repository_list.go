package cli

import (
	"fmt"

	"github.com/andrewhowdencom/vox/internal/adapters/storage/bbolt"
	"github.com/andrewhowdencom/vox/internal/domain/storage"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

// NewRepositoryListCmd creates a new cobra command for the "repository list" command.
func NewRepositoryListCmd() *cobra.Command {
	return newRepositoryListCmd(func() (storage.Repository, error) {
		return bbolt.NewRepository()
	})
}

func newRepositoryListCmd(repoFn func() (storage.Repository, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all interviews in the repository",
		Long:  `List all interviews in the repository.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := repoFn()
			if err != nil {
				return fmt.Errorf("could not create repository: %w", err)
			}
			// Note: We can't close the repo here, as it might be a mock.
			// The caller of the factory function is responsible for closing the repo.

			interviews, err := repo.ListInterviews()
			if err != nil {
				return fmt.Errorf("could not list interviews: %w", err)
			}

			if len(interviews) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No interviews found.")
				return nil
			}

			tbl := table.New("ID", "User", "Project", "Created At")
			tbl.WithWriter(cmd.OutOrStdout())

			for _, i := range interviews {
				tbl.AddRow(i.ID, i.UserID, i.ProjectID, i.CreatedAt.String())
			}

			tbl.Print()

			return nil
		},
	}
	return cmd
}
