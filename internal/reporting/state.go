package reporting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RunState tracks the last execution state of a command.
type RunState struct {
	Command   string    `json:"command"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	Duration  string    `json:"duration"`
	Summary   string    `json:"summary"`
}

// BaselineState tracks the baseline remediation status.
type BaselineState struct {
	LastRun    time.Time `json:"last_run"`
	Converged  bool      `json:"converged"`
	Iterations int       `json:"iterations"`
	StopReason string    `json:"stop_reason"`
	Violations int       `json:"violations_remaining"`
}

// SaveRunState writes last-run.json to the state directory.
func SaveRunState(stateDir string, state RunState) error {
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("creating state dir: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(stateDir, "last-run.json"), data, 0644)
}

// SaveBaselineState writes baseline-status.json to the state directory.
func SaveBaselineState(stateDir string, state BaselineState) error {
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("creating state dir: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(stateDir, "baseline-status.json"), data, 0644)
}

// WriteInterviewLog writes a human-readable markdown interview log.
func WriteInterviewLog(path string, records []InterviewRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	var content string
	content += "# Interview Log\n\n"
	content += fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339))

	for _, r := range records {
		content += fmt.Sprintf("## %s\n\n", r.Question)
		if r.Inferred != "" {
			content += fmt.Sprintf("- **Inferred**: %s\n", r.Inferred)
		}
		content += fmt.Sprintf("- **Answer**: %s\n", r.Answer)
		content += fmt.Sprintf("- **Source**: %s\n\n", r.Source)
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// InterviewRecord is a single Q&A for the interview log.
type InterviewRecord struct {
	Question string
	Inferred string
	Answer   string
	Source   string
}
