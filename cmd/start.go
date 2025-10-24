/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/andrewhowdencom/interviewer/interview"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a new interview",
	Long:  `Starts a new interview with a candidate.`,
	Run: func(cmd *cobra.Command, args []string) {
		provider := viper.GetString("provider")
		var questionProvider interview.QuestionProvider

		switch strings.ToLower(provider) {
		case "static":
			questionProvider = interview.NewStaticQuestionProvider()
		case "gemini":
			model := viper.GetString("model")
			apiKey := viper.GetString("api-key")
			if apiKey == "" {
				cmd.ErrOrStderr().Write([]byte("Error: api-key is required for gemini provider\n"))
				return
			}
			var err error
			questionProvider, err = interview.NewGeminiQuestionProvider(model, apiKey)
			if err != nil {
				cmd.ErrOrStderr().Write([]byte(fmt.Sprintf("Error creating gemini provider: %v\n", err)))
				return
			}
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
	startCmd.Flags().String("provider", "static", "The question provider to use (static or gemini)")
	startCmd.Flags().String("model", "gemini-1.5-flash", "The gemini model to use")
	startCmd.Flags().String("api-key", "", "The API key for the gemini provider")
	viper.BindPFlags(startCmd.Flags())
}
