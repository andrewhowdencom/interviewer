package storage

import "github.com/andrewhowdencom/vox/internal/domain"

// Repository defines the interface for storing and retrieving interview data.
type Repository interface {
	SaveInterview(interview *domain.Interview, transcript *domain.Transcript, summary *domain.Summary) (string, error)
	GetInterview(id string) (*domain.Interview, error)
	GetTranscript(interviewID string) (*domain.Transcript, error)
	GetSummary(interviewID string) (*domain.Summary, error)
	ListInterviews() ([]*domain.Interview, error)
	Close() error
}
