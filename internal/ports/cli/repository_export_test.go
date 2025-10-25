package cli

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/andrewhowdencom/vox/internal/domain"
	"github.com/andrewhowdencom/vox/internal/domain/storage"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryExportCmd(t *testing.T) {
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

	// Create the export command with the mock repository
	cmd := newRepositoryExportCmd(func() (storage.Repository, error) {
		return mockRepo, nil
	})
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"1"})

	// Execute the command
	err := cmd.Execute()

	// Assert that the command executed successfully
	assert.NoError(t, err)

	// Assert that the output contains the expected JSON
	type output struct {
		domain.Interview
		domain.Transcript
		domain.Summary
	}
	var out output
	err = json.Unmarshal(b.Bytes(), &out)
	assert.NoError(t, err)
	assert.Equal(t, "1", out.Interview.ID)
	assert.Equal(t, "user1", out.Interview.UserID)
	assert.Equal(t, "project1", out.Interview.ProjectID)
	assert.Equal(t, "This is a summary.", out.Summary.Text)

	// Execute the command with the --format=text flag
	b.Reset()
	cmd.SetArgs([]string{"1", "--format=text"})
	err = cmd.Execute()

	// Assert that the command executed successfully
	assert.NoError(t, err)

	// Assert that the output contains the expected text
	outputStr := b.String()
	assert.Contains(t, outputStr, "Interview ID: 1")
	assert.Contains(t, outputStr, "User: user1")
	assert.Contains(t, outputStr, "Project: project1")
	assert.Contains(t, outputStr, "--- Transcript ---")
	assert.Contains(t, outputStr, "Q: Q1")
	assert.Contains(t, outputStr, "A: A1")
	assert.Contains(t, outputStr, "--- Summary ---")
	assert.Contains(t, outputStr, "This is a summary.")
}
