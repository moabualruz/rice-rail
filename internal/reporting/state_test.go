package reporting

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSaveAndLoadRunState(t *testing.T) {
	dir := t.TempDir()

	state := RunState{
		Command:   "check",
		Timestamp: time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC),
		Success:   true,
		Duration:  "2.5s",
		Summary:   "all checks passed",
	}

	if err := SaveRunState(dir, state); err != nil {
		t.Fatalf("SaveRunState failed: %v", err)
	}

	path := filepath.Join(dir, "last-run.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file at %s: %v", path, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	var loaded RunState
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshalling JSON: %v", err)
	}

	if loaded.Command != "check" {
		t.Errorf("Command: got %q, want %q", loaded.Command, "check")
	}
	if !loaded.Success {
		t.Error("Success: got false, want true")
	}
	if loaded.Duration != "2.5s" {
		t.Errorf("Duration: got %q, want %q", loaded.Duration, "2.5s")
	}
	if loaded.Summary != "all checks passed" {
		t.Errorf("Summary: got %q, want %q", loaded.Summary, "all checks passed")
	}
}

func TestSaveAndLoadBaselineState(t *testing.T) {
	dir := t.TempDir()

	state := BaselineState{
		LastRun:    time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC),
		Converged:  true,
		Iterations: 3,
		StopReason: "converged",
		Violations: 0,
	}

	if err := SaveBaselineState(dir, state); err != nil {
		t.Fatalf("SaveBaselineState failed: %v", err)
	}

	path := filepath.Join(dir, "baseline-status.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	var loaded BaselineState
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshalling JSON: %v", err)
	}

	if !loaded.Converged {
		t.Error("Converged: got false, want true")
	}
	if loaded.Iterations != 3 {
		t.Errorf("Iterations: got %d, want 3", loaded.Iterations)
	}
	if loaded.StopReason != "converged" {
		t.Errorf("StopReason: got %q, want %q", loaded.StopReason, "converged")
	}
	if loaded.Violations != 0 {
		t.Errorf("Violations: got %d, want 0", loaded.Violations)
	}
}

func TestWriteInterviewLog(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "interview-log.md")

	records := []InterviewRecord{
		{
			Question: "safety_mode",
			Inferred: "balanced",
			Answer:   "strict",
			Source:   "user",
		},
		{
			Question: "allow_autofix",
			Answer:   "true",
			Source:   "inferred",
		},
	}

	if err := WriteInterviewLog(path, records); err != nil {
		t.Fatalf("WriteInterviewLog failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "# Interview Log") {
		t.Error("missing '# Interview Log' header")
	}
	if !strings.Contains(content, "## safety_mode") {
		t.Error("missing '## safety_mode' question header")
	}
	if !strings.Contains(content, "## allow_autofix") {
		t.Error("missing '## allow_autofix' question header")
	}
	if !strings.Contains(content, "**Answer**: strict") {
		t.Error("missing answer for safety_mode")
	}
	if !strings.Contains(content, "**Inferred**: balanced") {
		t.Error("missing inferred value for safety_mode")
	}
	if !strings.Contains(content, "**Source**: user") {
		t.Error("missing source for safety_mode")
	}
}

func TestWriteInterviewLogCreatesDir(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "deep", "nested", "dir", "interview-log.md")

	records := []InterviewRecord{
		{Question: "test_q", Answer: "test_a", Source: "test"},
	}

	if err := WriteInterviewLog(nested, records); err != nil {
		t.Fatalf("WriteInterviewLog failed to create parent dirs: %v", err)
	}

	if _, err := os.Stat(nested); err != nil {
		t.Fatalf("expected file at %s: %v", nested, err)
	}
}
