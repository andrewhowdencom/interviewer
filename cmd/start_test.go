package cmd

import (
	"bytes"
	"errors"
	"testing"
	"github.com/spf13/cobra"
	"github.com/andrewhowdencom/vox/interview"
)

// MockQuestionProvider is a mock implementation of the QuestionProvider interface for testing.
type MockQuestionProvider struct {
	questions []string
	currentIndex int
}

func (p *MockQuestionProvider) NextQuestion(previousAnswer string) (string, bool) {
	if p.currentIndex < len(p.questions) {
		question := p.questions[p.currentIndex]
		p.currentIndex++
		return question, true
	}
	return "", false
}

// MockInterviewUI is a mock implementation of the InterviewUI interface for testing.
type MockInterviewUI struct {
	answers []string
	currentIndex int
	summary bytes.Buffer
}

func (ui *MockInterviewUI) Ask(question string) (string, error) {
	if ui.currentIndex < len(ui.answers) {
		answer := ui.answers[ui.currentIndex]
		ui.currentIndex++
		return answer, nil
	}
	return "", errors.New("not enough answers")
}

func (ui *MockInterviewUI) DisplaySummary(qas []interview.QuestionAndAnswer) {
	for _, qa := range qas {
		ui.summary.WriteString("Q: " + qa.Question + "\n")
		ui.summary.WriteString("A: " + qa.Answer + "\n")
	}
}

func TestRunInterview(t *testing.T) {
	questions := []string{"q1", "q2"}
	answers := []string{"a1", "a2"}

	mockProvider := &MockQuestionProvider{questions: questions}
	mockUI := &MockInterviewUI{answers: answers}

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	runInterview(cmd, mockProvider, mockUI)

	expectedSummary := "Q: q1\nA: a1\nQ: q2\nA: a2\n"
	if mockUI.summary.String() != expectedSummary {
		t.Errorf("expected summary '%s', got '%s'", expectedSummary, mockUI.summary.String())
	}
}
