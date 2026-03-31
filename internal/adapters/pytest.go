package adapters

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mkh/rice-railing/internal/exec"
)

// PytestAdapter wraps pytest for running Python tests.
type PytestAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewPytestAdapter(runner *exec.Runner, repoRoot string) *PytestAdapter {
	return &PytestAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *PytestAdapter) Name() string                { return "pytest" }
func (a *PytestAdapter) SupportedLanguages() []string { return []string{"python"} }

func (a *PytestAdapter) Run(ctx context.Context, targets []string) (*TestResult, error) {
	// Use pytest-json-report if available, fall back to exit code
	result, err := a.runner.Run(ctx, "pytest", "--tb=short", "-q")
	if err != nil {
		return nil, fmt.Errorf("pytest: %w", err)
	}

	tr := &TestResult{Output: result.Stdout + result.Stderr}

	// Try to parse JSON report if plugin is installed
	jsonResult, jsonErr := a.runner.Run(ctx, "pytest", "--json-report", "--json-report-file=-", "-q")
	if jsonErr == nil && jsonResult.ExitCode <= 1 {
		var report struct {
			Summary struct {
				Passed int `json:"passed"`
				Failed int `json:"failed"`
				Total  int `json:"total"`
			} `json:"summary"`
		}
		if json.Unmarshal([]byte(jsonResult.Stdout), &report) == nil {
			tr.Passed = report.Summary.Passed
			tr.Failed = report.Summary.Failed
			tr.Total = report.Summary.Total
			return tr, nil
		}
	}

	// Fallback: infer from exit code
	if result.ExitCode == 0 {
		tr.Passed = 1 // at least
	} else if result.ExitCode == 1 {
		tr.Failed = 1
	}

	return tr, nil
}
