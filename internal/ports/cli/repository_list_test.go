package cli

import (
	"bytes"
	"testing"
	"time"

	"github.com/andrewhowdencom/vox/internal/domain"
	"github.com/andrewhowdencom/vox/internal/domain/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveInterview(interview *domain.Interview, transcript *domain.Transcript, summary *domain.Summary) (string, error) {
	args := m.Called(interview, transcript, summary)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) GetInterview(id string) (*domain.Interview, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.Interview), args.Error(1)
}

func (m *MockRepository) GetTranscript(interviewID string) (*domain.Transcript, error) {
	args := m.Called(interviewID)
	return args.Get(0).(*domain.Transcript), args.Error(1)
}

func (m *MockRepository) GetSummary(interviewID string) (*domain.Summary, error) {
	args := m.Called(interviewID)
	return args.Get(0).(*domain.Summary), args.Error(1)
}

func (m *MockRepository) ListInterviews() ([]*domain.Interview, error) {
	args := m.Called()
	return args.Get(0).([]*domain.Interview), args.Error(1)
}

func (m *MockRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestRepositoryListCmd(t *testing.T) {
	// Create a mock repository
	mockRepo := new(MockRepository)

	// Create a sample list of interviews
	interviews := []*domain.Interview{
		{
			ID:        "1",
			UserID:    "user1",
			ProjectID: "project1",
			CreatedAt: time.Now(),
		},
		{
			ID:        "2",
			UserID:    "user2",
			ProjectID: "project2",
			CreatedAt: time.Now(),
		},
	}

	// Set up the expected response from the mock repository
	mockRepo.On("ListInterviews").Return(interviews, nil)
	mockRepo.On("Close").Return(nil)

	// Create the list command with the mock repository
	cmd := newRepositoryListCmd(func() (storage.Repository, error) {
		return mockRepo, nil
	})
	b := bytes.NewBufferString("")
	cmd.SetOut(b)

	// Execute the command
	err := cmd.Execute()

	// Assert that the command executed successfully
	assert.NoError(t, err)

	// Assert that the output contains the expected table
	output := b.String()
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "User")
	assert.Contains(t, output, "Project")
	assert.Contains(t, output, "Created At")
	assert.Contains(t, output, "user1")
	assert.Contains(t, output, "project1")
	assert.Contains(t, output, "user2")
	assert.Contains(t, output, "project2")
}
