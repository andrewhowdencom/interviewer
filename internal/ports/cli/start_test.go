package cli

import (
	"fmt"
	"testing"

	"github.com/andrewhowdencom/vox/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestBuildGeminiPrompt(t *testing.T) {
	t.Run("should use the default system prompt when no custom prompt is configured", func(t *testing.T) {
		cfg := &config.Config{}
		topicPrompt := "Topic-specific prompt"

		// Assuming DefaultSystemPrompt is accessible or redefined for the test
		const DefaultSystemPrompt = "You are an interviewer."
		expectedPrompt := fmt.Sprintf("%s\n\n%s", DefaultSystemPrompt, topicPrompt)
		actualPrompt := buildGeminiPrompt(cfg, topicPrompt)

		assert.Equal(t, expectedPrompt, actualPrompt)
	})

	t.Run("should use a custom system prompt when one is configured", func(t *testing.T) {
		cfg := &config.Config{
			Providers: struct {
				Gemini struct {
					APIKey      string `yaml:"api_key"`
					Model       string
					Interviewer struct {
						Prompt string
					}
				}
			}{
				Gemini: struct {
					APIKey      string `yaml:"api_key"`
					Model       string
					Interviewer struct {
						Prompt string
					}
				}{
					Interviewer: struct {
						Prompt string
					}{
						Prompt: "Custom system prompt",
					},
				},
			},
		}
		topicPrompt := "Topic-specific prompt"

		expectedPrompt := "Custom system prompt\n\nTopic-specific prompt"
		actualPrompt := buildGeminiPrompt(cfg, topicPrompt)

		assert.Equal(t, expectedPrompt, actualPrompt)
	})
}
