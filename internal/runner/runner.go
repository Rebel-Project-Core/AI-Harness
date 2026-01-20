package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Result holds the execution result
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// prefixWriter prepends a prefix to each line
type prefixWriter struct {
	w      io.Writer
	prefix string
	atBOL  bool
}

func (pw *prefixWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		if pw.atBOL {
			_, err = fmt.Fprint(pw.w, pw.prefix)
			if err != nil {
				return n, err
			}
			pw.atBOL = false
		}
		_, err = pw.w.Write([]byte{b})
		if err != nil {
			return n, err
		}
		n++
		if b == '\n' {
			pw.atBOL = true
		}
	}
	return n, nil
}

// Run executes a command and returns the output
func Run(command string, args []string, env []string) (*Result, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = append(os.Environ(), env...)

	prefix := fmt.Sprintf("[%s] ", command)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &prefixWriter{w: &stdoutBuf, prefix: prefix, atBOL: true}
	cmd.Stderr = &prefixWriter{w: &stderrBuf, prefix: prefix, atBOL: true}

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
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		ExitCode: exitCode,
	}, nil
}
