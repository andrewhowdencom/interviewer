package interview

import "fmt"

// QuestionProvider is an interface for providing questions for an interview.
type QuestionProvider interface {
	// NextQuestion returns the next question in the interview.
	// It returns the question as a string and a boolean indicating if there are more questions.
	NextQuestion(previousAnswer string) (question string, hasMore bool)
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

// Interview encapsulates the logic for running an interview.
type Interview struct {
	Provider QuestionProvider
	UI       InterviewUI
}

// NewInterview creates a new Interview.
func NewInterview(provider QuestionProvider, ui InterviewUI) *Interview {
	return &Interview{
		Provider: provider,
		UI:       ui,
	}
}

// Run executes the interview loop.
func (i *Interview) Run() error {
	var qas []QuestionAndAnswer
	var answer string
	var err error

	for {
		question, hasMore := i.Provider.NextQuestion(answer)
		if !hasMore {
			break
		}

		answer, err = i.UI.Ask(question)
		if err != nil {
			return fmt.Errorf("error asking question: %w", err)
		}

		qas = append(qas, QuestionAndAnswer{
			Question: question,
			Answer:   answer,
		})
	}

	// Note: The original summary logic was here. In a true hexagonal architecture,
	// the core domain should not know about specific provider implementations like Gemini.
	// This will be handled by the adapter layer.
	i.UI.DisplaySummary(qas)
	return nil
}
