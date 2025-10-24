package interview

import (
	"context"
	"errors"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGeminiClient is a mock implementation of the GeminiClient interface.
type MockGeminiClient struct {
	mock.Mock
}

func (m *MockGeminiClient) SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	args := m.Called(ctx, parts)
	return args.Get(0).(*genai.GenerateContentResponse), args.Error(1)
}

// MockChatSession is a mock implementation of the ChatSession interface.
type MockChatSession struct {
	mock.Mock
}

func (m *MockChatSession) SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	args := m.Called(ctx, parts)
	return args.Get(0).(*genai.GenerateContentResponse), args.Error(1)
}

func TestGeminiQuestionProvider_NextQuestion(t *testing.T) {
	t.Run("should return the next question from the Gemini API", func(t *testing.T) {
		mockChatSession := new(MockChatSession)
		provider := &GeminiQuestionProvider{
			conversational: mockChatSession,
			maxQuestions:   5,
		}

		expectedResponse := &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{
							genai.Text("What is your greatest strength?"),
						},
					},
				},
			},
		}
		mockChatSession.On("SendMessage", mock.Anything, mock.Anything).Return(expectedResponse, nil)

		question, hasMore := provider.NextQuestion("")
		assert.True(t, hasMore)
		assert.Equal(t, "What is your greatest strength?", question)
		mockChatSession.AssertExpectations(t)
	})

	t.Run("should return false when the interview is complete", func(t *testing.T) {
		mockChatSession := new(MockChatSession)
		provider := &GeminiQuestionProvider{
			conversational: mockChatSession,
			maxQuestions:   5,
		}

		expectedResponse := &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{
							genai.Text("INTERVIEW_COMPLETE"),
						},
					},
				},
			},
		}
		mockChatSession.On("SendMessage", mock.Anything, mock.Anything).Return(expectedResponse, nil)

		_, hasMore := provider.NextQuestion("")
		assert.False(t, hasMore)
		mockChatSession.AssertExpectations(t)
	})

	t.Run("should return false when the max number of questions has been reached", func(t *testing.T) {
		mockChatSession := new(MockChatSession)
		provider := &GeminiQuestionProvider{
			conversational: mockChatSession,
			maxQuestions:   1,
			questionCount:  1,
		}

		expectedResponse := &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{
							genai.Text("Please summarize our conversation."),
						},
					},
				},
			},
		}

		mockChatSession.On("SendMessage", mock.Anything, mock.Anything).Return(expectedResponse, nil)

		_, hasMore := provider.NextQuestion("")
		assert.False(t, hasMore)
	})
}

func TestGeminiQuestionProvider_generateSummary(t *testing.T) {
	t.Run("should generate a summary of the interview", func(t *testing.T) {
		mockChatSession := new(MockChatSession)
		provider := &GeminiQuestionProvider{
			conversational: mockChatSession,
		}

		expectedResponse := &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{
							genai.Text("This is a summary of the interview."),
						},
					},
				},
			},
		}
		mockChatSession.On("SendMessage", mock.Anything, mock.Anything).Return(expectedResponse, nil)

		provider.generateSummary()
		assert.Equal(t, "This is a summary of the interview.", provider.Summary())
		mockChatSession.AssertExpectations(t)
	})

	t.Run("should handle errors when generating a summary", func(t *testing.T) {
		mockChatSession := new(MockChatSession)
		provider := &GeminiQuestionProvider{
			conversational: mockChatSession,
		}

		mockChatSession.On("SendMessage", mock.Anything, mock.Anything).Return(&genai.GenerateContentResponse{}, errors.New("API error"))

		provider.generateSummary()
		assert.Equal(t, "Error generating summary: API error", provider.Summary())
		mockChatSession.AssertExpectations(t)
	})
}
