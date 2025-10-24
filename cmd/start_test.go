package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestStartCmd(t *testing.T) {
	// Create a buffer to capture the command's output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)

	// Set the arguments for the command
	rootCmd.SetArgs([]string{"interviews", "start", "--candidate", "John Doe"})

	// Execute the command
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check that the candidate variable was set correctly
	if candidate != "John Doe" {
		t.Errorf("expected candidate to be 'John Doe', got '%s'", candidate)
	}

	// Check the output of the command
	expectedOutput := "start called"
	if !strings.Contains(buf.String(), expectedOutput) {
		t.Errorf("expected output to contain '%s', got '%s'", expectedOutput, buf.String())
	}
}
