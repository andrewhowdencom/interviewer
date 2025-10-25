package gemini

import (
	"context"
	"fmt"
	"strings"

	"github.com/andrewhowdencom/vox/internal/domain"
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
					return "", false
				}
				p.questionCount++
				return question, true
			}
		}
	}

	return "", false
}

// Summarize generates a summary of the interview transcript.
func (p *QuestionProvider) Summarize(transcript *domain.Transcript) (string, error) {
	// Format the transcript into a single string for the prompt.
	var transcriptText string
	for _, entry := range transcript.Entries {
		transcriptText += fmt.Sprintf("Q: %s\nA: %s\n\n", entry.Question, entry.Answer)
	}

	// Create the prompt for summarization.
	prompt := fmt.Sprintf("Please summarize the following interview transcript:\n\n%s", transcriptText)

	// Call the Gemini API to generate the summary.
	resp, err := p.client.GenerateContent(context.Background(), genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("could not generate summary: %w", err)
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no summary response from Gemini")
	}

	return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
}

var _ interview.QuestionProvider = (*QuestionProvider)(nil)
var _ interview.Summarizer = (*QuestionProvider)(nil)
