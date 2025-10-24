package interview

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Model is a type for the Gemini model name.
type Model string

// APIKey is a type for the Gemini API key.
type APIKey string

// Prompt is a type for the interview prompt.
type Prompt string

const InterviewStructure = `You are to ask maximally one question at a time, and then wait for the users response. Then, use the
users prompt and the information supplied in the context so far to ask the next question.`

// GeminiQuestionProvider provides questions from the Gemini API.
type GeminiQuestionProvider struct {
	client         GeminiClient
	conversational ChatSession
	questionCount  int
	maxQuestions   int
	summary        string
}

// generativeModelWrapper wraps a genai.GenerativeModel to implement the GeminiClient interface.
type generativeModelWrapper struct {
	*genai.GenerativeModel
}

func (w *generativeModelWrapper) SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	return w.GenerativeModel.GenerateContent(ctx, parts...)
}

// NewGeminiQuestionProvider creates a new GeminiQuestionProvider.
func NewGeminiQuestionProvider(model Model, apiKey APIKey, prompt Prompt) (QuestionProvider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(string(apiKey)))
	if err != nil {
		return nil, err
	}

	generativeModel := client.GenerativeModel(string(model))
	wrappedModel := &generativeModelWrapper{generativeModel}

	cs := wrappedModel.StartChat()
	cs.History = []*genai.Content{
		{
			Parts: []genai.Part{
				genai.Text(prompt),
				genai.Text(InterviewStructure),
			},
			Role: "model",
		},
	}

	return &GeminiQuestionProvider{
		client:         wrappedModel,
		conversational: NewGenaiChatSessionWrapper(cs),
		maxQuestions:   20,
	}, nil
}

// NextQuestion returns the next question from the Gemini API.
func (p *GeminiQuestionProvider) NextQuestion(previousAnswer string) (string, bool) {
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
func (p *GeminiQuestionProvider) generateSummary() {
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
func (p *GeminiQuestionProvider) Summary() string {
	return p.summary
}
