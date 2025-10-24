/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/andrewhowdencom/vox/interview"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a new interview",
	Long:  `Starts a new interview with a candidate.`,
	Run: func(cmd *cobra.Command, args []string) {
		var config interview.Config
		if err := viper.Unmarshal(&config); err != nil {
			cmd.ErrOrStderr().Write([]byte(fmt.Sprintf("Error unmarshalling config: %v\n", err)))
			return
		}

		topicID := viper.GetString("topic")

		if topicID == "" {
			cmd.Println("Please specify a topic using --topic. Available topics:")
			for _, t := range config.Interviews {
				cmd.Printf(" - %s: %s\n", t.ID, t.Name)
			}
			return
		}

		var selectedTopic *interview.Topic
		for i, t := range config.Interviews {
			if strings.EqualFold(t.ID, topicID) {
				selectedTopic = &config.Interviews[i]
				break
			}
		}

		if selectedTopic == nil {
			cmd.ErrOrStderr().Write([]byte(fmt.Sprintf("Error: topic '%s' not found\n", topicID)))
			return
		}

		var questionProvider interview.QuestionProvider
		var err error

		switch strings.ToLower(selectedTopic.Provider) {
		case "static":
			questionProvider = interview.NewStaticQuestionProvider(selectedTopic.Questions)
		case "gemini":
			apiKey := viper.GetString("api-key")
			if apiKey == "" {
				cmd.ErrOrStderr().Write([]byte("Error: api-key is required for gemini provider\n"))
				return
			}
			questionProvider, err = interview.NewGeminiQuestionProvider("gemini-1.5-flash", apiKey, selectedTopic.Prompt)
			if err != nil {
				cmd.ErrOrStderr().Write([]byte(fmt.Sprintf("Error creating gemini provider: %v\n", err)))
				return
			}
		default:
			cmd.ErrOrStderr().Write([]byte(fmt.Sprintf("Error: unknown provider '%s'\n", selectedTopic.Provider)))
			return
		}

		ui := interview.NewTerminalUI()
		runInterview(cmd, questionProvider, ui)
	},
}

func runInterview(cmd *cobra.Command, questionProvider interview.QuestionProvider, ui interview.InterviewUI) {
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
			cmd.ErrOrStderr().Write([]byte("Error asking question: " + err.Error()))
			return
		}

		qas = append(qas, interview.QuestionAndAnswer{
			Question: question,
			Answer:   answer,
		})
	}

	if gp, ok := questionProvider.(*interview.GeminiQuestionProvider); ok {
		summary := gp.Summary()
		if summary != "" {
			fmt.Println("\n--- Gemini Summary ---")
			fmt.Println(summary)
			fmt.Println("--------------------")
		}
	}

	ui.DisplaySummary(qas)
}

func init() {
	interviewCmd.AddCommand(startCmd)
	startCmd.Flags().String("topic", "", "The topic of the interview to start")
	startCmd.Flags().String("api-key", "", "The API key for the gemini provider")
	viper.BindPFlags(startCmd.Flags())
}
