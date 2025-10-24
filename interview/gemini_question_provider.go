package interview

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

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
func NewGeminiQuestionProvider(model, apiKey, prompt string) (*GeminiQuestionProvider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	generativeModel := client.GenerativeModel(model)
	wrappedModel := &generativeModelWrapper{generativeModel}

	cs := wrappedModel.StartChat()
	cs.History = []*genai.Content{
		{
			Parts: []genai.Part{
				genai.Text(prompt),
			},
			Role: "user",
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
		// In a real application, you'd want to handle this error more gracefully.
		// For now, we'll just end the interview.
		if strings.Contains(err.Error(), "404") {
			fmt.Println("Error: The configured model was not found. Please check the model name and ensure your API key has the correct permissions.")
		}
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
