package gemini

import (
	"context"
	"testing"

	"github.com/andrewhowdencom/vox/internal/domain"
	"github.com/google/generative-ai-go/genai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGeminiClient is a mock implementation of the GeminiClient interface.
type MockGeminiClient struct {
	mock.Mock
}

func (m *MockGeminiClient) GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	args := m.Called(ctx, parts)
	return args.Get(0).(*genai.GenerateContentResponse), args.Error(1)
}

func (m *MockGeminiClient) StartChat() ChatSession {
	args := m.Called()
	return args.Get(0).(ChatSession)
}

// MockChatSession is a mock implementation of the ChatSession interface.
type MockChatSession struct {
	mock.Mock
}

func (m *MockChatSession) SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	args := m.Called(ctx, parts)
	return args.Get(0).(*genai.GenerateContentResponse), args.Error(1)
}


func TestGeminiQuestionProvider_Summarize(t *testing.T) {
	// Create a mock GeminiClient
	mockClient := new(MockGeminiClient)

	// Create a QuestionProvider with the mock client
	provider := &QuestionProvider{
		client: mockClient,
	}

	// Create a sample transcript
	transcript := &domain.Transcript{
		Entries: []struct {
			Question string `json:"question"`
			Answer   string `json:"answer"`
		}{
			{Question: "What is your name?", Answer: "My name is Jules."},
			{Question: "What is your quest?", Answer: "To seek the Holy Grail."},
		},
	}

	// Set up the expected response from the mock client
	expectedSummary := "This is a summary of the interview."
	mockResponse := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []genai.Part{
						genai.Text(expectedSummary),
					},
				},
			},
		},
	}
	mockClient.On("GenerateContent", mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Call the Summarize method
	summary, err := provider.Summarize(transcript)

	// Assert that the summary is correct and there are no errors
	assert.NoError(t, err)
	assert.Equal(t, expectedSummary, summary)

	// Assert that the mock client's GenerateContent method was called
	mockClient.AssertExpectations(t)
}
