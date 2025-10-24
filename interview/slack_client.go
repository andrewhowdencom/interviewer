package interview

import "github.com/slack-go/slack"

// SlackClient is an interface that wraps the slack.Client.
type SlackClient interface {
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
}

// SlackClientWrapper is a wrapper for the slack.Client that implements the SlackClient interface.
type SlackClientWrapper struct {
	*slack.Client
}

// NewSlackClientWrapper creates a new SlackClientWrapper.
func NewSlackClientWrapper(client *slack.Client) *SlackClientWrapper {
	return &SlackClientWrapper{client}
}
