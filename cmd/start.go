package cmd

import (
	"fmt"
	"io"
	"strings"

	interview "github.com/andrewhowdencom/vox/interview"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newStartCmd creates a new cobra command for the "start" command.
func newStartCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts a new interview",
		Long:  `Starts a new interview with a candidate.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var config interview.Config
			if err := viper.Unmarshal(&config); err != nil {
				return fmt.Errorf("error unmarshalling config: %w", err)
			}

			topicID := viper.GetString("topic")
			apiKey := viper.GetString("api-key")
			model := viper.GetString("model")

			return runStart(cmd, out, &config, topicID, apiKey, model)
		},
	}

	cmd.Flags().String("topic", "", "The topic of the interview to start")
	cmd.Flags().String("api-key", "", "The API key for the gemini provider")
	cmd.Flags().String("model", "gemini-1.5-flash", "The model to use for the gemini provider")
	viper.BindPFlags(cmd.Flags())

	return cmd
}

// runStart is the main logic for the "start" command.
func runStart(cmd *cobra.Command, out io.Writer, config *interview.Config, topicID, apiKey, model string) error {
	if topicID == "" {
		fmt.Fprintln(out, "Please specify a topic using --topic. Available topics:")
		for _, t := range config.Interviews {
			fmt.Fprintf(out, " - %s: %s\n", t.ID, t.Name)
		}
		return nil
	}

	var selectedTopic *interview.Topic
	for i, t := range config.Interviews {
		if strings.EqualFold(t.ID, topicID) {
			selectedTopic = &config.Interviews[i]
			break
		}
	}

	if selectedTopic == nil {
		return fmt.Errorf("topic '%s' not found", topicID)
	}

	questionProvider, err := newQuestionProvider(cmd, config, selectedTopic, apiKey, model)
	if err != nil {
		return err
	}

	ui := interview.NewTerminalUI()
	return runInterview(out, questionProvider, ui)
}

// newQuestionProvider creates a QuestionProvider based on the selected topic.
func newQuestionProvider(cmd *cobra.Command, config *interview.Config, topic *interview.Topic, apiKey, model string) (interview.QuestionProvider, error) {
	switch strings.ToLower(topic.Provider) {
	case "static":
		return interview.NewStaticQuestionProvider(topic.Questions), nil
	case "gemini":
		if apiKey == "" {
			return nil, fmt.Errorf("api-key is required for gemini provider")
		}
		if !cmd.Flags().Changed("model") && config.Providers.Gemini.Model != "" {
			model = config.Providers.Gemini.Model
		}
		systemPrompt := interview.DefaultSystemPrompt
		if config.Providers.Gemini.Interviewer.Prompt != "" {
			systemPrompt = config.Providers.Gemini.Interviewer.Prompt
		}

		finalPrompt := fmt.Sprintf("%s\n\n%s", systemPrompt, topic.Prompt)
		return interview.NewGeminiQuestionProvider(model, apiKey, finalPrompt)
	default:
		return nil, fmt.Errorf("unknown provider '%s'", topic.Provider)
	}
}

// runInterview executes the interview loop.
func runInterview(out io.Writer, questionProvider interview.QuestionProvider, ui interview.InterviewUI) error {
	var qas []interview.QuestionAndAnswer
	var answer string
	var err error

	for {
		question, hasMore := questionProvider.NextQuestion(answer)
		if !hasMore {
			break
		}

		answer, err = ui.Ask(question)
		if err != nil {
			return fmt.Errorf("error asking question: %w", err)
		}

		qas = append(qas, interview.QuestionAndAnswer{
			Question: question,
			Answer:   answer,
		})
	}

	if gp, ok := questionProvider.(*interview.GeminiQuestionProvider); ok {
		summary := gp.Summary()
		if summary != "" {
			fmt.Fprintln(out, "\n--- Gemini Summary ---")
			fmt.Fprintln(out, summary)
			fmt.Fprintln(out, "--------------------")
		}
	}

	ui.DisplaySummary(qas)
	return nil
}

