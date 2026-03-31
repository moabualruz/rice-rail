package exec

import (
	"context"
	"testing"
)

func TestRunnerSuccess(t *testing.T) {
	r := NewRunner()
	result, err := r.Run(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success() {
		t.Fatalf("expected success, got exit code %d", result.ExitCode)
	}
	if result.Stdout != "hello\n" {
		t.Fatalf("expected 'hello\\n', got %q", result.Stdout)
	}
}

func TestRunnerFailure(t *testing.T) {
	r := NewRunner()
	result, err := r.Run(context.Background(), "false")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success() {
		t.Fatal("expected failure")
	}
}

func TestRunnerDryRun(t *testing.T) {
	r := NewRunner()
	r.DryRun = true
	result, err := r.Run(context.Background(), "rm", "-rf", "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stdout != "" {
		t.Fatal("dry run should produce no output")
	}
	if result.Command != "rm -rf /" {
		t.Fatalf("expected command string, got %q", result.Command)
	}
}

func TestWhich(t *testing.T) {
	path, found := Which("echo")
	if !found {
		t.Fatal("echo should be on PATH")
	}
	if path == "" {
		t.Fatal("path should not be empty")
	}

	_, found = Which("nonexistent-tool-xyz")
	if found {
		t.Fatal("nonexistent tool should not be found")
	}
}
