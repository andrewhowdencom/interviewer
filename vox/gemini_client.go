package interview

import (
	"context"

	"github.com/google/generative-ai-go/genai"
)

// GeminiClient defines the interface for the Gemini client.
type GeminiClient interface {
	StartChat() *genai.ChatSession
	SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}
