package interview

// StaticQuestionProvider provides a predefined list of questions for the interview.
type StaticQuestionProvider struct {
	questions []string
	currentIndex int
}

// NewStaticQuestionProvider creates a new StaticQuestionProvider with a hardcoded list of questions.
func NewStaticQuestionProvider() *StaticQuestionProvider {
	return &StaticQuestionProvider{
		questions: []string{
			"how are you",
			"what color is the sky",
			"how do you feel about yellow",
		},
	}
}

// NextQuestion returns the next question from the predefined list.
// It returns the question and a boolean indicating if there are more questions.
func (p *StaticQuestionProvider) NextQuestion() (string, bool) {
	if p.currentIndex < len(p.questions) {
		question := p.questions[p.currentIndex]
		p.currentIndex++
		return question, true
	}
	return "", false
}
