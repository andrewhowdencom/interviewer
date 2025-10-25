package cli

import (
	"bytes"
	"testing"
	"time"

	"github.com/andrewhowdencom/vox/internal/domain"
	"github.com/andrewhowdencom/vox/internal/domain/storage"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryViewCmd(t *testing.T) {
	// Create a mock repository
	mockRepo := new(MockRepository)

	// Create a sample interview
	interview := &domain.Interview{
		ID:        "1",
		UserID:    "user1",
		ProjectID: "project1",
		CreatedAt: time.Now(),
	}
	summary := &domain.Summary{
		Text: "This is a summary.",
	}
	transcript := &domain.Transcript{
		Entries: []struct {
			Question string `json:"question"`
			Answer   string `json:"answer"`
		}{
			{Question: "Q1", Answer: "A1"},
		},
	}

	// Set up the expected response from the mock repository
	mockRepo.On("GetInterview", "1").Return(interview, nil)
	mockRepo.On("GetSummary", "1").Return(summary, nil)
	mockRepo.On("GetTranscript", "1").Return(transcript, nil)
	mockRepo.On("Close").Return(nil)

	// Create the view command with the mock repository
	cmd := newRepositoryViewCmd(func() (storage.Repository, error) {
		return mockRepo, nil
	})
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"1"})

	// Execute the command
	err := cmd.Execute()

	// Assert that the command executed successfully
	assert.NoError(t, err)

	// Assert that the output contains the expected summary
	output := b.String()
	assert.Contains(t, output, "Interview ID: 1")
	assert.Contains(t, output, "User: user1")
	assert.Contains(t, output, "Project: project1")
	assert.Contains(t, output, "--- Summary ---")
	assert.Contains(t, output, "This is a summary.")

	// Execute the command with the --full flag
	b.Reset()
	cmd.SetArgs([]string{"1", "--full"})
	err = cmd.Execute()

	// Assert that the command executed successfully
	assert.NoError(t, err)

	// Assert that the output contains the expected transcript
	output = b.String()
	assert.Contains(t, output, "--- Transcript ---")
	assert.Contains(t, output, "Q: Q1")
	assert.Contains(t, output, "A: A1")
}
