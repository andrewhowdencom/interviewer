package interview

// StaticQuestionProvider provides a predefined list of questions for the interview.
type StaticQuestionProvider struct {
	questions []string
	currentIndex int
}

// NewStaticQuestionProvider creates a new StaticQuestionProvider.
func NewStaticQuestionProvider(questions []string) *StaticQuestionProvider {
	return &StaticQuestionProvider{
		questions: questions,
	}
}

// NextQuestion returns the next question from the predefined list.
// It returns the question and a boolean indicating if there are more questions.
func (p *StaticQuestionProvider) NextQuestion(previousAnswer string) (string, bool) {
	if p.currentIndex < len(p.questions) {
		question := p.questions[p.currentIndex]
		p.currentIndex++
		return question, true
	}
	return "", false
}
