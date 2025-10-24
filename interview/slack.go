package interview

import (
	"fmt"
	"strings"

	"github.com/slack-go/slack"
)

// SlackUI handles the user interface for the interview in Slack.
type SlackUI struct {
	Client     *slack.Client
	ChannelID  string
	UserID     string
	AnswerChan chan string
}

// NewSlackUI creates a new SlackUI.
func NewSlackUI(client *slack.Client, channelID, userID string) *SlackUI {
	return &SlackUI{
		Client:     client,
		ChannelID:  channelID,
		UserID:     userID,
		AnswerChan: make(chan string),
	}
}

// Ask sends a question to the user on Slack and waits for their answer.
func (s *SlackUI) Ask(question string) (string, error) {
	_, _, err := s.Client.PostMessage(s.ChannelID, slack.MsgOptionText(question, false))
	if err != nil {
		return "", fmt.Errorf("failed to post message to slack: %w", err)
	}
	// Wait for the answer from the event handler via the channel
	answer := <-s.AnswerChan
	return answer, nil
}

// DisplaySummary sends the interview summary to the user on Slack.
func (s *SlackUI) DisplaySummary(qas []QuestionAndAnswer) {
	var summary strings.Builder
	summary.WriteString("*--- Interview Summary ---*\n")
	for _, qa := range qas {
		summary.WriteString(fmt.Sprintf("*Q:* %s\n", qa.Question))
		summary.WriteString(fmt.Sprintf("*A:* %s\n", qa.Answer))
	}
	summary.WriteString("*-----------------------*\n")

	_, _, err := s.Client.PostMessage(s.ChannelID, slack.MsgOptionText(summary.String(), false))
	if err != nil {
		// We can't return an error here, so we'll just log it.
		// In a real application, you'd want to use a proper logger.
		fmt.Printf("Error displaying summary: %v\n", err)
	}
}
