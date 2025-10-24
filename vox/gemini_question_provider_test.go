package interview

import (
	"context"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChatSession is a mock implementation of the ChatSession for testing.
type MockChatSession struct {
	mock.Mock
}

func (m *MockChatSession) SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	args := m.Called(ctx, parts)
	return args.Get(0).(*genai.GenerateContentResponse), args.Error(1)
}


func TestGeminiQuestionProvider_NextQuestion(t *testing.T) {
	// Create a mock ChatSession
	mockChat := new(MockChatSession)

	// Create the GeminiQuestionProvider with the mock chat
	provider := &GeminiQuestionProvider{
		conversational: mockChat,
		maxQuestions:   5, // Use a smaller number for testing
	}

	// Test the first question
	question, hasMore := provider.NextQuestion("")
	assert.True(t, hasMore)
	assert.Equal(t, "how are you?", question)

	// Test the next 3 questions
	for i := 0; i < 3; i++ {
		mockResp := &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{
							genai.Text("Another question"),
						},
					},
				},
			},
		}
		mockChat.On("SendMessage", mock.Anything, mock.Anything).Return(mockResp, nil).Once()

		question, hasMore = provider.NextQuestion("Some answer")
		assert.True(t, hasMore)
		assert.Equal(t, "Another question", question)
	}

	// Test the 5th question (should be the last one)
	mockResp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []genai.Part{
						genai.Text("Final question"),
					},
				},
			},
		},
	}
	mockChat.On("SendMessage", mock.Anything, mock.Anything).Return(mockResp, nil).Once()

	question, hasMore = provider.NextQuestion("Some answer")
	assert.True(t, hasMore)
	assert.Equal(t, "Final question", question)

	// Test that the 6th question terminates the interview and generates a summary
	summaryResp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []genai.Part{
						genai.Text("This is the summary."),
					},
				},
			},
		},
	}
	mockChat.On("SendMessage", mock.Anything, mock.Anything).Return(summaryResp, nil).Once()

	_, hasMore = provider.NextQuestion("Some answer")
	assert.False(t, hasMore)
	assert.Equal(t, "This is the summary.", provider.Summary())
}

func TestGeminiQuestionProvider_NextQuestion_InterviewComplete(t *testing.T) {
	// Create a mock ChatSession
	mockChat := new(MockChatSession)

	// Create the provider
	provider := &GeminiQuestionProvider{
		conversational: mockChat,
		maxQuestions:   20,
	}

	// Test the first question
	question, hasMore := provider.NextQuestion("")
	assert.True(t, hasMore)
	assert.Equal(t, "how are you?", question)

	// Mock the response with the INTERVIEW_COMPLETE phrase
	mockResp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []genai.Part{
						genai.Text("This is the summary. INTERVIEW_COMPLETE"),
					},
				},
			},
		},
	}
	mockChat.On("SendMessage", mock.Anything, mock.Anything).Return(mockResp, nil).Once()

	// Test that the next question terminates the interview
	question, hasMore = provider.NextQuestion("Some answer")
	assert.False(t, hasMore)
	assert.Equal(t, "This is the summary.", provider.Summary())
}
