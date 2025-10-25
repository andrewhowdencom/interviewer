package bbolt

import (
	"encoding/json"
	"fmt"

	"github.com/adrg/xdg"
	"github.com/andrewhowdencom/vox/internal/domain"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

// bboltRepository implements the storage.Repository interface using bbolt.
type bboltRepository struct {
	db *bbolt.DB
}

var (
	interviewsBucket = []byte("interviews")
	transcriptsBucket = []byte("transcripts")
	summariesBucket   = []byte("summaries")
)

// NewRepository creates a new bbolt repository, opening the database file
// at the appropriate XDG data path.
func NewRepository() (*bboltRepository, error) {
	dbPath, err := xdg.DataFile("vox/vox.db")
	if err != nil {
		return nil, fmt.Errorf("could not resolve XDG data path: %w", err)
	}

	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open bbolt database: %w", err)
	}

	// Create buckets if they don't exist
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
		return nil, fmt.Errorf("could not create buckets: %w", err)
	}

	return &bboltRepository{db: db}, nil
}

// Close closes the database connection.
func (r *bboltRepository) Close() error {
	return r.db.Close()
}

// SaveInterview saves all parts of an interview to the database.
func (r *bboltRepository) SaveInterview(interview *domain.Interview, transcript *domain.Transcript, summary *domain.Summary) (string, error) {
	// Generate a new UUID for the interview
	id := uuid.New().String()
	interview.ID = id
	transcript.InterviewID = id
	summary.InterviewID = id

	return id, r.db.Update(func(tx *bbolt.Tx) error {
		// Save Interview metadata
		b := tx.Bucket(interviewsBucket)
		buf, err := json.Marshal(interview)
		if err != nil {
			return fmt.Errorf("could not marshal interview: %w", err)
		}
		if err := b.Put([]byte(id), buf); err != nil {
			return fmt.Errorf("could not save interview: %w", err)
		}

		// Save Transcript
		b = tx.Bucket(transcriptsBucket)
		buf, err = json.Marshal(transcript)
		if err != nil {
			return fmt.Errorf("could not marshal transcript: %w", err)
		}
		if err := b.Put([]byte(id), buf); err != nil {
			return fmt.Errorf("could not save transcript: %w", err)
		}

		// Save Summary
		b = tx.Bucket(summariesBucket)
		buf, err = json.Marshal(summary)
		if err != nil {
			return fmt.Errorf("could not marshal summary: %w", err)
		}
		if err := b.Put([]byte(id), buf); err != nil {
			return fmt.Errorf("could not save summary: %w", err)
		}

		return nil
	})
}

// GetInterview retrieves interview metadata from the database.
func (r *bboltRepository) GetInterview(id string) (*domain.Interview, error) {
	var interview domain.Interview
	err := r.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(interviewsBucket)
		v := b.Get([]byte(id))
		if v == nil {
			return fmt.Errorf("interview not found")
		}
		if err := json.Unmarshal(v, &interview); err != nil {
			return fmt.Errorf("could not unmarshal interview: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &interview, nil
}

// GetTranscript retrieves an interview transcript from the database.
func (r *bboltRepository) GetTranscript(interviewID string) (*domain.Transcript, error) {
	var transcript domain.Transcript
	err := r.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(transcriptsBucket)
		v := b.Get([]byte(interviewID))
		if v == nil {
			return fmt.Errorf("transcript not found")
		}
		if err := json.Unmarshal(v, &transcript); err != nil {
			return fmt.Errorf("could not unmarshal transcript: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &transcript, nil
}

// GetSummary retrieves an interview summary from the database.
func (r *bboltRepository) GetSummary(interviewID string) (*domain.Summary, error) {
	var summary domain.Summary
	err := r.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(summariesBucket)
		v := b.Get([]byte(interviewID))
		if v == nil {
			return fmt.Errorf("summary not found")
		}
		if err := json.Unmarshal(v, &summary); err != nil {
			return fmt.Errorf("could not unmarshal summary: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

// ListInterviews retrieves metadata for all stored interviews.
func (r *bboltRepository) ListInterviews() ([]*domain.Interview, error) {
	var interviews []*domain.Interview
	err := r.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(interviewsBucket)
		return b.ForEach(func(k, v []byte) error {
			var interview domain.Interview
			if err := json.Unmarshal(v, &interview); err != nil {
				// Log or skip corrupted data? For now, we'll return the error.
				return fmt.Errorf("could not unmarshal interview data: %w", err)
			}
			interviews = append(interviews, &interview)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return interviews, nil
}
