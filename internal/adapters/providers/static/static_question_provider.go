package static

import (
	"github.com/andrewhowdencom/vox/internal/domain"
	"github.com/andrewhowdencom/vox/internal/domain/interview"
)

// QuestionProvider provides a predefined list of questions for the interview.
type QuestionProvider struct {
	questions    []string
	currentIndex int
}

// New creates a new StaticQuestionProvider.
func New(questions []string) *QuestionProvider {
	return &QuestionProvider{
		questions: questions,
	}
}

// NextQuestion returns the next question from the predefined list.
// It returns the question and a boolean indicating if there are more questions.
func (p *QuestionProvider) NextQuestion(previousAnswer string) (string, bool) {
	if p.currentIndex < len(p.questions) {
		question := p.questions[p.currentIndex]
		p.currentIndex++
		return question, true
	}
	return "", false
}

// Ensure QuestionProvider implements the domain interface.
var _ interview.QuestionProvider = (*QuestionProvider)(nil)
var _ interview.Summarizer = (*QuestionProvider)(nil)

// Summarize returns an empty string, as static interviews do not have summaries.
func (p *QuestionProvider) Summarize(transcript *domain.Transcript) (string, error) {
	return "", nil
}
