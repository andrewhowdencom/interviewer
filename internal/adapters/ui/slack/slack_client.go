package slack

import "github.com/slack-go/slack"

// SlackClient is an interface that wraps the slack.Client.
// This is useful for testing and abstracting away the concrete implementation.
type SlackClient interface {
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
}

// Ensure the real client implements the interface
var _ SlackClient = (*slack.Client)(nil)
