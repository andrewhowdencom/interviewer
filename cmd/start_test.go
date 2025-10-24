package cmd

import (
	"bytes"
	"fmt"
	"testing"

	interview "github.com/andrewhowdencom/vox/interview"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Mock GeminiQuestionProvider for testing
type MockGeminiQuestionProvider struct {
	interview.QuestionProvider
	model  string
	apiKey string
	prompt string
}

func (m *MockGeminiQuestionProvider) NextQuestion(previousAnswer string) (string, bool) {
	return "", false
}

func (m *MockGeminiQuestionProvider) Summary() string {
	return ""
}

func newMockGeminiQuestionProvider(model, apiKey, prompt string) (interview.QuestionProvider, error) {
	return &MockGeminiQuestionProvider{model: model, apiKey: apiKey, prompt: prompt}, nil
}

func TestNewQuestionProvider(t *testing.T) {
	// a dummy command
	cmd := &cobra.Command{}
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	t.Run("should prepend the default system prompt when no custom prompt is configured", func(t *testing.T) {
		config := &interview.Config{
			Interviews: []interview.Topic{
				{ID: "test", Provider: "gemini", Prompt: "Topic-specific prompt"},
			},
		}

		// replace the real function with the mock
		originalNewGeminiQuestionProvider := interview.NewGeminiQuestionProvider
		interview.NewGeminiQuestionProvider = newMockGeminiQuestionProvider
		defer func() { interview.NewGeminiQuestionProvider = originalNewGeminiQuestionProvider }()


		provider, err := newQuestionProvider(cmd, config, &config.Interviews[0], "test-api-key", "test-model")
		assert.NoError(t, err)

		mockProvider := provider.(*MockGeminiQuestionProvider)
		expectedPrompt := fmt.Sprintf("%s\n\n%s", interview.DefaultSystemPrompt, "Topic-specific prompt")
		assert.Equal(t, expectedPrompt, mockProvider.prompt)
	})

	t.Run("should prepend a custom system prompt when one is configured", func(t *testing.T) {
		config := &interview.Config{
			Providers: interview.Providers{
				Gemini: interview.Gemini{
					Interviewer: interview.Interviewer{
						Prompt: "Custom system prompt",
					},
				},
			},
			Interviews: []interview.Topic{
				{ID: "test", Provider: "gemini", Prompt: "Topic-specific prompt"},
			},
		}

		// replace the real function with the mock
		originalNewGeminiQuestionProvider := interview.NewGeminiQuestionProvider
		interview.NewGeminiQuestionProvider = newMockGeminiQuestionProvider
		defer func() { interview.NewGeminiQuestionProvider = originalNewGeminiQuestionProvider }()

		provider, err := newQuestionProvider(cmd, config, &config.Interviews[0], "test-api-key", "test-model")
		assert.NoError(t, err)

		mockProvider := provider.(*MockGeminiQuestionProvider)
		expectedPrompt := "Custom system prompt\n\nTopic-specific prompt"
		assert.Equal(t, expectedPrompt, mockProvider.prompt)
	})
}
