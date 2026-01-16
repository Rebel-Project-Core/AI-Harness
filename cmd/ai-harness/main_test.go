package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestCLI_Stdin(t *testing.T) {
	// We'll use "go run ." to execute the main package in the current directory
	cmd := exec.Command("go", "run", ".")
	cmd.Stdin = strings.NewReader("Error: fake error log from stdin")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// We expect it to fail because of missing API key, but the output should indicate it tried to analyze
	err := cmd.Run()

	// It should exit with non-zero (1) because of the missing API key error
	if err == nil {
		t.Errorf("Expected error (exit code 1) due to missing API key, but got nil")
	}

	output := stderr.String() + stdout.String()
	if !strings.Contains(output, "Analysis failed: no API key found") {
		t.Errorf("Expected output to contain 'Analysis failed: no API key found', but got:\n%s", output)
	}
}

func TestCLI_CommandExecution(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "echo", "hello world")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Expected successful execution, got error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "hello world") {
		t.Errorf("Expected output to contain 'hello world', but got:\n%s", output)
	}
	if !strings.Contains(output, "Command executed successfully") {
		t.Errorf("Expected output to indicate success, but got:\n%s", output)
	}
}
