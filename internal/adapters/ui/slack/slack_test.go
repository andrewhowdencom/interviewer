package slack_test

import (
	"errors"
	"testing"

	"github.com/andrewhowdencom/vox/internal/adapters/ui/slack"
	goslack "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSlackClient is a mock implementation of the SlackClient interface.
type MockSlackClient struct {
	mock.Mock
}

func (m *MockSlackClient) PostMessage(channelID string, options ...goslack.MsgOption) (string, string, error) {
	args := m.Called(channelID, options)
	return args.String(0), args.String(1), args.Error(2)
}

func TestSlackUI_Ask(t *testing.T) {
	t.Run("should send a question to Slack and return the answer", func(t *testing.T) {
		mockClient := new(MockSlackClient)
		ui := slack.New(mockClient, "C12345", "U12345")

		mockClient.On("PostMessage", "C12345", mock.Anything).Return("", "", nil)

		go func() {
			ui.AnswerChan <- "This is the answer."
		}()

		answer, err := ui.Ask("What is your name?")
		assert.NoError(t, err)
		assert.Equal(t, "This is the answer.", answer)
		mockClient.AssertExpectations(t)
	})

	t.Run("should return an error if sending the message fails", func(t *testing.T) {
		mockClient := new(MockSlackClient)
		ui := slack.New(mockClient, "C12345", "U12345")

		mockClient.On("PostMessage", "C12345", mock.Anything).Return("", "", errors.New("API error"))

		_, err := ui.Ask("What is your name?")
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestSlackUI_DisplaySummary(t *testing.T) {
	t.Run("should send the summary to Slack", func(t *testing.T) {
		mockClient := new(MockSlackClient)
		ui := slack.New(mockClient, "C12345", "U12345")

		summary := "This is a summary."

		mockClient.On("PostMessage", "C12345", mock.Anything).Return("", "", nil)

		ui.DisplaySummary(summary)
		mockClient.AssertExpectations(t)
	})
}
