package interview

// QuestionProvider is an interface for providing questions for an interview.
type QuestionProvider interface {
	// NextQuestion returns the next question in the interview.
	// It returns the question as a string and a boolean indicating if there are more questions.
	NextQuestion() (question string, hasMore bool)
}

// InterviewUI is an interface for the user interface of the interview.
type InterviewUI interface {
	// Ask asks a question to the user and returns the answer.
	Ask(question string) (answer string, err error)
	// DisplaySummary displays the summary of the interview.
	DisplaySummary(qas []QuestionAndAnswer)
}

// QuestionAndAnswer holds a question and its corresponding answer.
type QuestionAndAnswer struct {
	Question string
	Answer   string
}
