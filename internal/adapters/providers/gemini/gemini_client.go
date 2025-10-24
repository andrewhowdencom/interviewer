package gemini

import (
	"context"

	"github.com/google/generative-ai-go/genai"
)

// GeminiClient is an interface that wraps the genai.GenerativeModel.
// This is useful for testing and abstracting away the concrete implementation.
type GeminiClient interface {
	StartChat() ChatSession
	GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

// generativeModelWrapper is a wrapper around the genai.GenerativeModel to implement the GeminiClient interface.
type generativeModelWrapper struct {
	*genai.GenerativeModel
}

// StartChat starts a chat session.
func (w *generativeModelWrapper) StartChat() ChatSession {
	return NewGenaiChatSessionWrapper(w.GenerativeModel.StartChat())
}

// Ensure the real client implements the interface
var _ GeminiClient = (*generativeModelWrapper)(nil)
