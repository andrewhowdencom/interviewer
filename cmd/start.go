/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/andrewhowdencom/interviewer/interview"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a new interview",
	Long: `Starts a new interview with a candidate.`,
	Run: func(cmd *cobra.Command, args []string) {
		// For now, we'll use the static question provider and terminal UI.
		// In the future, we can use flags to specify different providers and UIs.
		questionProvider := interview.NewStaticQuestionProvider()
		ui := interview.NewTerminalUI()
		runInterview(cmd, questionProvider, ui)
	},
}

func runInterview(cmd *cobra.Command, questionProvider interview.QuestionProvider, ui interview.InterviewUI) {
	var qas []interview.QuestionAndAnswer

	for {
		question, hasMore := questionProvider.NextQuestion()
		if !hasMore {
			break
		}

		answer, err := ui.Ask(question)
		if err != nil {
			cmd.ErrOrStderr().Write([]byte("Error asking question: " + err.Error()))
			return
		}

		qas = append(qas, interview.QuestionAndAnswer{
			Question: question,
			Answer:   answer,
		})
	}

	ui.DisplaySummary(qas)
}

func init() {
	interviewCmd.AddCommand(startCmd)
}
