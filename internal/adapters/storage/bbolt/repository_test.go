package bbolt

import (
	"os"
	"testing"
	"time"

	"github.com/andrewhowdencom/vox/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestBoltRepository_SaveAndGetInterview(t *testing.T) {
	// Create a temporary database file for testing
	f, err := os.CreateTemp("", "test.db")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	// Create a new repository
	repo, err := NewTestRepository(f.Name())
	require.NoError(t, err)
	defer repo.Close()

	// Create a sample interview
	interview := &domain.Interview{
		UserID:    "test-user",
		ProjectID: "test-project",
		CreatedAt: time.Now(),
	}
	transcript := &domain.Transcript{
		Entries: []struct {
			Question string `json:"question"`
			Answer   string `json:"answer"`
		}{
			{Question: "Q1", Answer: "A1"},
		},
	}
	summary := &domain.Summary{
		Text: "This is a summary.",
	}

	// Save the interview
	id, err := repo.SaveInterview(interview, transcript, summary)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	// Get the interview back
	retrievedInterview, err := repo.GetInterview(id)
	require.NoError(t, err)
	assert.Equal(t, id, retrievedInterview.ID)
	assert.Equal(t, interview.UserID, retrievedInterview.UserID)
	assert.Equal(t, interview.ProjectID, retrievedInterview.ProjectID)

	// Get the transcript back
	retrievedTranscript, err := repo.GetTranscript(id)
	require.NoError(t, err)
	assert.Equal(t, id, retrievedTranscript.InterviewID)
	assert.Equal(t, transcript.Entries, retrievedTranscript.Entries)

	// Get the summary back
	retrievedSummary, err := repo.GetSummary(id)
	require.NoError(t, err)
	assert.Equal(t, id, retrievedSummary.InterviewID)
	assert.Equal(t, summary.Text, retrievedSummary.Text)
}

// NewTestRepository creates a new repository using a temporary file path.
func NewTestRepository(path string) (*bboltRepository, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(interviewsBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(transcriptsBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(summariesBucket); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &bboltRepository{db: db}, nil
}
