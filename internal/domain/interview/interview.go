package interview

import (
	"fmt"
	"time"

	"github.com/andrewhowdencom/vox/internal/domain"
	"github.com/andrewhowdencom/vox/internal/domain/storage"
)

// QuestionProvider is an interface for providing questions for an interview.
type QuestionProvider interface {
	Summarizer
	// NextQuestion returns the next question in the interview.
	// It returns the question as a string and a boolean indicating if there are more questions.
	NextQuestion(previousAnswer string) (question string, hasMore bool)
}

// InterviewUI is an interface for the user interface of the interview.
type InterviewUI interface {
	// Ask asks a question to the user and returns the answer.
	Ask(question string) (answer string, err error)
	// DisplaySummary displays the summary of the interview.
	DisplaySummary(summary string)
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
	Repo     storage.Repository
}

// NewInterview creates a new Interview.
func NewInterview(provider QuestionProvider, ui InterviewUI, repo storage.Repository) *Interview {
	return &Interview{
		Provider: provider,
		UI:       ui,
		Repo:     repo,
	}
}

// Run executes the interview loop.
func (i *Interview) Run(userID, projectID string) error {
	var transcriptEntries []struct {
		Question string `json:"question"`
		Answer   string `json:"answer"`
	}
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

		transcriptEntries = append(transcriptEntries, struct {
			Question string `json:"question"`
			Answer   string `json:"answer"`
		}{
			Question: question,
			Answer:   answer,
		})
	}

	// Create the transcript
	transcript := &domain.Transcript{
		Entries: transcriptEntries,
	}

	// Generate the summary
	summaryText, err := i.Provider.Summarize(transcript)
	if err != nil {
		return fmt.Errorf("could not generate summary: %w", err)
	}
	summary := &domain.Summary{
		Text: summaryText,
	}

	// Create the interview metadata
	interview := &domain.Interview{
		UserID:    userID,
		ProjectID: projectID,
		CreatedAt: time.Now(),
	}

	// Save the interview
	interviewID, err := i.Repo.SaveInterview(interview, transcript, summary)
	if err != nil {
		return fmt.Errorf("could not save interview: %w", err)
	}
	transcript.InterviewID = interviewID
	summary.InterviewID = interviewID

	// Display the summary to the user
	i.UI.DisplaySummary(summary.Text)
	return nil
}
