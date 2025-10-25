package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/andrewhowdencom/vox/internal/config"
	"github.com/andrewhowdencom/vox/internal/domain/interview"
	"github.com/andrewhowdencom/vox/internal/adapters/providers/gemini"
	"github.com/andrewhowdencom/vox/internal/adapters/providers/static"
	"github.com/andrewhowdencom/vox/internal/adapters/storage/bbolt"
	"github.com/andrewhowdencom/vox/internal/adapters/ui/terminal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewStartCmd creates a new cobra command for the "start" command.
func NewStartCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts a new interview",
		Long:  `Starts a new interview with a candidate.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg config.Config
			if err := viper.Unmarshal(&cfg); err != nil {
				return fmt.Errorf("error unmarshalling config: %w", err)
			}

			topicID := viper.GetString("topic")
			apiKey := viper.GetString("api-key")
			model := viper.GetString("model")
			user := viper.GetString("user")

			return runStart(cmd, out, &cfg, topicID, apiKey, model, user)
		},
	}

	cmd.Flags().String("topic", "", "The topic of the interview to start")
	cmd.Flags().String("api-key", "", "The API key for the gemini provider")
	cmd.Flags().String("model", "gemini-1.5-flash", "The model to use for the gemini provider")
	cmd.Flags().String("user", "", "The user conducting the interview")
	cmd.MarkFlagRequired("user")
	viper.BindPFlags(cmd.Flags())

	return cmd
}

// runStart is the main logic for the "start" command.
func runStart(cmd *cobra.Command, out io.Writer, cfg *config.Config, topicID, apiKey, model, user string) error {
	if topicID == "" {
		fmt.Fprintln(out, "Please specify a topic using --topic. Available topics:")
		for _, t := range cfg.Interviews {
			fmt.Fprintf(out, " - %s: %s\n", t.ID, t.Name)
		}
		return nil
	}

	var selectedTopic *config.Topic
	for i, t := range cfg.Interviews {
		if strings.EqualFold(t.ID, topicID) {
			selectedTopic = &cfg.Interviews[i]
			break
		}
	}

	if selectedTopic == nil {
		return fmt.Errorf("topic '%s' not found", topicID)
	}

	questionProvider, err := newQuestionProvider(cmd, cfg, selectedTopic, apiKey, model)
	if err != nil {
		return err
	}

	ui := terminal.New()

	repo, err := bbolt.NewRepository()
	if err != nil {
		return fmt.Errorf("could not create repository: %w", err)
	}
	defer repo.Close()

	interviewToRun := interview.NewInterview(questionProvider, ui, repo)
	err = interviewToRun.Run(user, topicID)
	if err != nil {
		return err
	}

	return nil
}

// newQuestionProvider creates a QuestionProvider based on the selected topic.
func newQuestionProvider(cmd *cobra.Command, cfg *config.Config, topic *config.Topic, apiKey, model string) (interview.QuestionProvider, error) {
	switch strings.ToLower(topic.Provider) {
	case "static":
		return static.New(topic.Questions), nil
	case "gemini":
		if apiKey == "" {
			return nil, fmt.Errorf("api-key is required for gemini provider")
		}
		if !cmd.Flags().Changed("model") && cfg.Providers.Gemini.Model != "" {
			model = cfg.Providers.Gemini.Model
		}

		finalPrompt := buildGeminiPrompt(cfg, topic.Prompt)
		// We need to cast the concrete type to the interface type.
		// Since gemini.New returns (interview.QuestionProvider, error), we can do this.
		p, err := gemini.New(gemini.Model(model), gemini.APIKey(apiKey), gemini.Prompt(finalPrompt))
		if err != nil {
			return nil, err
		}
		return p, nil
	default:
		return nil, fmt.Errorf("unknown provider '%s'", topic.Provider)
	}
}

// buildGeminiPrompt constructs the final prompt for the Gemini provider.
func buildGeminiPrompt(cfg *config.Config, topicPrompt string) string {
	const DefaultSystemPrompt = "You are an interviewer." // This should be defined in a better place.
	systemPrompt := DefaultSystemPrompt
	if cfg.Providers.Gemini.Interviewer.Prompt != "" {
		systemPrompt = cfg.Providers.Gemini.Interviewer.Prompt
	}
	return fmt.Sprintf("%s\n\n%s", systemPrompt, topicPrompt)
}
