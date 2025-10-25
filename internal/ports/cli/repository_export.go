package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/andrewhowdencom/vox/internal/domain"
	"github.com/andrewhowdencom/vox/internal/adapters/storage/bbolt"
	"github.com/andrewhowdencom/vox/internal/domain/storage"
	"github.com/spf13/cobra"
)

// NewRepositoryExportCmd creates a new cobra command for the "repository export" command.
func NewRepositoryExportCmd() *cobra.Command {
	return newRepositoryExportCmd(func() (storage.Repository, error) {
		return bbolt.NewRepository()
	})
}

func newRepositoryExportCmd(repoFn func() (storage.Repository, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [id]",
		Short: "Export a single interview",
		Long:  `Export a single interview.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			format, _ := cmd.Flags().GetString("format")

			repo, err := repoFn()
			if err != nil {
				return fmt.Errorf("could not create repository: %w", err)
			}

			interview, err := repo.GetInterview(id)
			if err != nil {
				return fmt.Errorf("could not get interview: %w", err)
			}
			transcript, err := repo.GetTranscript(id)
			if err != nil {
				return fmt.Errorf("could not get transcript: %w", err)
			}
			summary, err := repo.GetSummary(id)
			if err != nil {
				return fmt.Errorf("could not get summary: %w", err)
			}

			output := struct {
				*domain.Interview
				*domain.Transcript
				*domain.Summary
			}{
				interview,
				transcript,
				summary,
			}

			switch strings.ToLower(format) {
			case "json":
				b, err := json.MarshalIndent(output, "", "  ")
				if err != nil {
					return fmt.Errorf("could not marshal interview: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(b))
			case "text":
				fmt.Fprintf(cmd.OutOrStdout(), "Interview ID: %s\n", interview.ID)
				fmt.Fprintf(cmd.OutOrStdout(), "User: %s\n", interview.UserID)
				fmt.Fprintf(cmd.OutOrStdout(), "Project: %s\n", interview.ProjectID)
				fmt.Fprintf(cmd.OutOrStdout(), "Created At: %s\n\n", interview.CreatedAt.String())
				fmt.Fprintln(cmd.OutOrStdout(), "--- Transcript ---")
				for _, entry := range transcript.Entries {
					fmt.Fprintf(cmd.OutOrStdout(), "Q: %s\nA: %s\n\n", entry.Question, entry.Answer)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "--- Summary ---")
				fmt.Fprintln(cmd.OutOrStdout(), summary.Text)
			default:
				return fmt.Errorf("unknown format: %s", format)
			}

			return nil
		},
	}
	cmd.Flags().String("format", "json", "The format to export the interview in (json, text)")
	return cmd
}
