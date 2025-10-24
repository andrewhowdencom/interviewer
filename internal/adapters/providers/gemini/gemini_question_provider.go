package gemini

import (
	"context"
	"fmt"
	"strings"

	"github.com/andrewhowdencom/vox/internal/domain/interview"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const InterviewStructure = `You are to ask maximally one question at a time, and then wait for the users response. Then, use the
users prompt and the information supplied in the context so far to ask the next question.`

// QuestionProvider provides questions from the Gemini API.
type QuestionProvider struct {
	client         GeminiClient
	conversational ChatSession
	questionCount  int
	maxQuestions   int
	summary        string
}

// New creates a new GeminiQuestionProvider.
func New(model Model, apiKey APIKey, prompt Prompt) (interview.QuestionProvider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(string(apiKey)))
	if err != nil {
		return nil, err
	}

	generativeModel := client.GenerativeModel(string(model))
	wrappedModel := &generativeModelWrapper{generativeModel}

	// The chat session needs to be initialized with history.
	// We'll do this by accessing the underlying genai.ChatSession.
	chat := wrappedModel.GenerativeModel.StartChat()
	chat.History = []*genai.Content{
		{
			Parts: []genai.Part{
				genai.Text(prompt),
				genai.Text(InterviewStructure),
			},
			Role: "model",
		},
	}

	// Now, we wrap the initialized chat session.
	cs := NewGenaiChatSessionWrapper(chat)

	return &QuestionProvider{
		client:         wrappedModel,
		conversational: cs,
		maxQuestions:   20,
	}, nil
}

// NextQuestion returns the next question from the Gemini API.
func (p *QuestionProvider) NextQuestion(previousAnswer string) (string, bool) {
	if p.questionCount >= p.maxQuestions {
		// Generate summary before finishing
		p.generateSummary()
		return "", false
	}

	ctx := context.Background()
	var parts []genai.Part
	if previousAnswer != "" {
		parts = append(parts, genai.Text(previousAnswer))
	}

	resp, err := p.conversational.SendMessage(ctx, parts...)
	if err != nil {
		fmt.Println("Error getting next question:", err)
		return "", false
	}

	if len(resp.Candidates) > 0 {
		content := resp.Candidates[0].Content
		if len(content.Parts) > 0 {
			if text, ok := content.Parts[0].(genai.Text); ok {
				question := string(text)
				if strings.Contains(question, "INTERVIEW_COMPLETE") {
					// Extract summary if Gemini provides it
					p.summary = strings.TrimSpace(strings.Replace(question, "INTERVIEW_COMPLETE", "", 1))
					return "", false
				}
				p.questionCount++
				return question, true
			}
		}
	}

	return "", false
}

// generateSummary generates a summary of the interview using the Gemini API.
func (p *QuestionProvider) generateSummary() {
	ctx := context.Background()
	resp, err := p.conversational.SendMessage(ctx, genai.Text("Please summarize our conversation."))
	if err != nil {
		p.summary = "Error generating summary: " + err.Error()
		return
	}

	if len(resp.Candidates) > 0 {
		content := resp.Candidates[0].Content
		if len(content.Parts) > 0 {
			if text, ok := content.Parts[0].(genai.Text); ok {
				p.summary = string(text)
			}
		}
	}
}

// Summary returns the summary of the interview.
func (p *QuestionProvider) Summary() string {
	return p.summary
}

var _ interview.QuestionProvider = (*QuestionProvider)(nil)
