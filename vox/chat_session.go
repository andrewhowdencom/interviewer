package interview

import (
	"context"
	"github.com/google/generative-ai-go/genai"
)

// ChatSession is an interface for a conversational chat session.
type ChatSession interface {
	SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

// genaiChatSessionWrapper wraps a *genai.ChatSession to implement the ChatSession interface.
type genaiChatSessionWrapper struct {
	session *genai.ChatSession
}

// NewGenaiChatSessionWrapper creates a new wrapper for a genai.ChatSession.
func NewGenaiChatSessionWrapper(session *genai.ChatSession) ChatSession {
	return &genaiChatSessionWrapper{session: session}
}

// SendMessage sends a message in the chat session.
func (w *genaiChatSessionWrapper) SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	return w.session.SendMessage(ctx, parts...)
}
