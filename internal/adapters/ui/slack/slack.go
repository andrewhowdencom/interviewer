package slack

import (
	"fmt"
	"log/slog"

	"github.com/andrewhowdencom/vox/internal/domain/interview"
	"github.com/slack-go/slack"
)

// ChannelID is a type for Slack channel IDs.
type ChannelID string

// UserID is a type for Slack user IDs.
type UserID string

// UI handles the user interface for the interview in Slack.
type UI struct {
	Client     SlackClient
	ChannelID  ChannelID
	UserID     UserID
	AnswerChan chan string
}

// New creates a new SlackUI.
func New(client SlackClient, channelID ChannelID, userID UserID) *UI {
	return &UI{
		Client:     client,
		ChannelID:  channelID,
		UserID:     userID,
		AnswerChan: make(chan string),
	}
}

// Ask sends a question to the user on Slack and waits for their answer.
func (s *UI) Ask(question string) (string, error) {
	slog.Debug("Asking question on slack", "channel_id", s.ChannelID, "user_id", s.UserID, "question", question)
	_, _, err := s.Client.PostMessage(string(s.ChannelID), slack.MsgOptionText(question, false))
	if err != nil {
		slog.Error("Failed to post message to slack", "error", err, "channel_id", s.ChannelID, "user_id", s.UserID)
		return "", fmt.Errorf("failed to post message to slack: %w", err)
	}
	// Wait for the answer from the event handler via the channel
	slog.Debug("Waiting for answer from user", "channel_id", s.ChannelID, "user_id", s.UserID)
	answer := <-s.AnswerChan
	slog.Debug("Received answer from user", "channel_id", s.ChannelID, "user_id", s.UserID, "answer", answer)
	return answer, nil
}

// DisplaySummary sends the interview summary to the user on Slack.
func (s *UI) DisplaySummary(summary string) {
	if summary != "" {
		formattedSummary := fmt.Sprintf("*--- Interview Summary ---*\n%s\n*-----------------------*", summary)
		slog.Debug("Displaying summary on slack", "channel_id", s.ChannelID, "user_id", s.UserID, "summary", formattedSummary)

		_, _, err := s.Client.PostMessage(string(s.ChannelID), slack.MsgOptionText(formattedSummary, false))
		if err != nil {
			slog.Error("Error displaying summary", "error", err, "channel_id", s.ChannelID, "user_id", s.UserID)
		}
	}
}

// Ensure UI implements the domain interface.
var _ interview.InterviewUI = (*UI)(nil)
