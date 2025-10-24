package interview

import (
	"bufio"
	"fmt"
	"os"
)

// TerminalUI handles the user interface for the interview in the terminal.
type TerminalUI struct{}

// NewTerminalUI creates a new TerminalUI.
func NewTerminalUI() *TerminalUI {
	return &TerminalUI{}
}

// Ask displays a question to the user in the terminal and returns the answer.
func (t *TerminalUI) Ask(question string) (string, error) {
	fmt.Printf("%s\n> ", question)
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	// Trim the newline character from the answer
	return answer[:len(answer)-1], nil
}

// DisplaySummary displays the interview summary in the terminal.
func (t *TerminalUI) DisplaySummary(qas []QuestionAndAnswer) {
	fmt.Println("\n--- Interview Summary ---")
	for _, qa := range qas {
		fmt.Printf("Q: %s\n", qa.Question)
		fmt.Printf("A: %s\n", qa.Answer)
	}
	fmt.Println("-----------------------")
}
