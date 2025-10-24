package cmd

import (
	"fmt"
	"testing"

	interview "github.com/andrewhowdencom/vox/interview"
	"github.com/stretchr/testify/assert"
)

func TestBuildGeminiPrompt(t *testing.T) {
	t.Run("should use the default system prompt when no custom prompt is configured", func(t *testing.T) {
		config := &interview.Config{}
		topicPrompt := "Topic-specific prompt"

		expectedPrompt := fmt.Sprintf("%s\n\n%s", interview.DefaultSystemPrompt, topicPrompt)
		actualPrompt := buildGeminiPrompt(config, topicPrompt)

		assert.Equal(t, expectedPrompt, actualPrompt)
	})

	t.Run("should use a custom system prompt when one is configured", func(t *testing.T) {
		config := &interview.Config{
			Providers: interview.Providers{
				Gemini: interview.Gemini{
					Interviewer: interview.Interviewer{
						Prompt: "Custom system prompt",
					},
				},
			},
		}
		topicPrompt := "Topic-specific prompt"

		expectedPrompt := "Custom system prompt\n\nTopic-specific prompt"
		actualPrompt := buildGeminiPrompt(config, topicPrompt)

		assert.Equal(t, expectedPrompt, actualPrompt)
	})
}
