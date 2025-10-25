package interview

import "github.com/andrewhowdencom/vox/internal/domain"

// Summarizer defines the interface for generating a summary from an interview transcript.
type Summarizer interface {
	Summarize(transcript *domain.Transcript) (string, error)
}
