package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestMainOutput(t *testing.T) {
	// os.Stdout is process-wide.
	got := captureStdout(t, main)
	trimmed := strings.TrimSpace(got)
	if trimmed == "" {
		t.Fatal("main() produced no output")
	}
	lines := strings.Split(trimmed, "\n")
	if len(lines) != 4 {
		t.Fatalf("main() output line count = %d, want 4: %q", len(lines), got)
	}
	for _, want := range []string{"metric.cpu: 0", "metric.cpu: 1", "metric.cpu: 2", "metric.cpu: 3"} {
		if !strings.Contains(got, want+"\n") {
			t.Fatalf("main() output missing %q in %q", want, got)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}
