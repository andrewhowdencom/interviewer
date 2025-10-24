package gemini

import (
	"context"

	"github.com/google/generative-ai-go/genai"
)

// ChatSession is an interface that wraps the genai.ChatSession.
type ChatSession interface {
	SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

// genaiChatSessionWrapper is a concrete implementation of the ChatSession interface
// that wraps the genai.ChatSession.
type genaiChatSessionWrapper struct {
	session *genai.ChatSession
}

// NewGenaiChatSessionWrapper creates a new genaiChatSessionWrapper.
func NewGenaiChatSessionWrapper(session *genai.ChatSession) ChatSession {
	return &genaiChatSessionWrapper{
		session: session,
	}
}

// SendMessage sends a message to the chat session.
func (w *genaiChatSessionWrapper) SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	return w.session.SendMessage(ctx, parts...)
}

// Ensure the wrapper implements the interface
var _ ChatSession = (*genaiChatSessionWrapper)(nil)
