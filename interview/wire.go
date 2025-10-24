//go:build wireinject
// +build wireinject

package interview

import (
	"github.com/google/wire"
	"github.com/slack-go/slack"
)

func InitializeStaticQuestionProvider(questions []string) *StaticQuestionProvider {
	wire.Build(NewStaticQuestionProvider)
	return nil
}

func InitializeGeminiQuestionProvider(model Model, apiKey APIKey, prompt Prompt) (QuestionProvider, error) {
	wire.Build(NewGeminiQuestionProvider)
	return nil, nil
}

func InitializeTerminalUI() *TerminalUI {
	wire.Build(NewTerminalUI)
	return nil
}

var slackSet = wire.NewSet(
	NewSlackUI,
	NewSlackClientWrapper,
	wire.Bind(new(SlackClient), new(*SlackClientWrapper)),
)

func InitializeSlackUI(client *slack.Client, channelID ChannelID, userID UserID) *SlackUI {
	wire.Build(slackSet)
	return nil
}
