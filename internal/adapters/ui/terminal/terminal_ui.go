package terminal

import (
	"bufio"
	"fmt"
	"os"

	"github.com/andrewhowdencom/vox/internal/domain/interview"
)

// UI handles the user interface for the interview in the terminal.
type UI struct{}

// New creates a new TerminalUI.
func New() *UI {
	return &UI{}
}

// Ask displays a question to the user in the terminal and returns the answer.
func (t *UI) Ask(question string) (string, error) {
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
func (t *UI) DisplaySummary(summary string) {
	if summary != "" {
		fmt.Println("\n--- Interview Summary ---")
		fmt.Println(summary)
		fmt.Println("-----------------------")
	}
}

// Ensure UI implements the domain interface.
var _ interview.InterviewUI = (*UI)(nil)
