package runner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// Result holds the execution result
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Run executes a command and returns the output
func Run(command string, args []string, env []string) (*Result, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = append(os.Environ(), env...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to run command: %w", err)
		}
	}

	return &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}
