package runner

import (
	"bytes"
	"testing"
)

func TestPrefixWriter(t *testing.T) {
	var buf bytes.Buffer
	pw := &prefixWriter{
		w:      &buf,
		prefix: "[test] ",
		atBOL:  true,
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"hello\n", "[test] hello\n"},
		{"world", "[test] world"},
		{"\nagain\n", "[test] \n[test] again\n"},
		{"multi\nline\n", "[test] multi\n[test] line\n"},
	}

	for _, tt := range tests {
		buf.Reset()
		pw.atBOL = true
		_, err := pw.Write([]byte(tt.input))
		if err != nil {
			t.Errorf("Write() error = %v", err)
		}
		if buf.String() != tt.expected {
			t.Errorf("Write(%q) = %q, want %q", tt.input, buf.String(), tt.expected)
		}
	}
}

func TestPrefixWriter_Continuous(t *testing.T) {
	var buf bytes.Buffer
	pw := &prefixWriter{
		w:      &buf,
		prefix: "[test] ",
		atBOL:  true,
	}

	pw.Write([]byte("part1"))
	pw.Write([]byte(" part2\n"))
	pw.Write([]byte("line2"))

	expected := "[test] part1 part2\n[test] line2"
	if buf.String() != expected {
		t.Errorf("Continuous Write = %q, want %q", buf.String(), expected)
	}
}
