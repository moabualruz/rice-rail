package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mkh/rice-railing/internal/exec"
)

// JestAdapter wraps Jest for running JS/TS tests.
type JestAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewJestAdapter(runner *exec.Runner, repoRoot string) *JestAdapter {
	return &JestAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *JestAdapter) Name() string                { return "jest" }
func (a *JestAdapter) SupportedLanguages() []string { return []string{"typescript", "javascript"} }

type jestResult struct {
	NumPassedTests int `json:"numPassedTests"`
	NumFailedTests int `json:"numFailedTests"`
	NumTotalTests  int `json:"numTotalTests"`
}

func (a *JestAdapter) Run(ctx context.Context, targets []string) (*TestResult, error) {
	result, err := a.runner.Run(ctx, "npx", "jest", "--json", "--no-coverage")
	if err != nil {
		return nil, fmt.Errorf("jest: %w", err)
	}

	var parsed jestResult
	if err := json.Unmarshal([]byte(result.Stdout), &parsed); err != nil {
		return &TestResult{Output: result.Stderr}, nil
	}

	return &TestResult{
		Passed: parsed.NumPassedTests,
		Failed: parsed.NumFailedTests,
		Total:  parsed.NumTotalTests,
		Output: result.Stderr,
	}, nil
}

// VitestAdapter wraps Vitest for running JS/TS tests.
type VitestAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewVitestAdapter(runner *exec.Runner, repoRoot string) *VitestAdapter {
	return &VitestAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *VitestAdapter) Name() string                { return "vitest" }
func (a *VitestAdapter) SupportedLanguages() []string { return []string{"typescript", "javascript"} }

func (a *VitestAdapter) Run(ctx context.Context, targets []string) (*TestResult, error) {
	result, err := a.runner.Run(ctx, "npx", "vitest", "run", "--reporter=json")
	if err != nil {
		return nil, fmt.Errorf("vitest: %w", err)
	}

	var parsed struct {
		NumPassedTests int `json:"numPassedTests"`
		NumFailedTests int `json:"numFailedTests"`
		NumTotalTests  int `json:"numTotalTests"`
	}
	if err := json.Unmarshal([]byte(result.Stdout), &parsed); err != nil {
		return &TestResult{Output: result.Stdout + result.Stderr}, nil
	}

	return &TestResult{
		Passed: parsed.NumPassedTests,
		Failed: parsed.NumFailedTests,
		Total:  parsed.NumTotalTests,
		Output: result.Stderr,
	}, nil
}

// DetectJSTestRunner checks which test runner is configured in the repo.
func DetectJSTestRunner(repoRoot string) string {
	configs := map[string]string{
		"vitest.config.ts":  "vitest",
		"vitest.config.js":  "vitest",
		"vitest.config.mts": "vitest",
		"jest.config.js":    "jest",
		"jest.config.ts":    "jest",
		"jest.config.json":  "jest",
	}
	for file, runner := range configs {
		if _, err := os.Stat(filepath.Join(repoRoot, file)); err == nil {
			return runner
		}
	}
	return ""
}
