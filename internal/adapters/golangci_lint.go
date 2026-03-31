package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mkh/rice-railing/internal/exec"
)

// GolangciLintAdapter wraps golangci-lint for linting and autofixing Go code.
type GolangciLintAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewGolangciLintAdapter(runner *exec.Runner, repoRoot string) *GolangciLintAdapter {
	return &GolangciLintAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *GolangciLintAdapter) Name() string                 { return "golangci-lint" }
func (a *GolangciLintAdapter) SupportedLanguages() []string { return []string{"go"} }

// golangciIssue matches golangci-lint --out-format=json output.
type golangciResult struct {
	Issues []golangciIssue `json:"Issues"`
}

type golangciIssue struct {
	FromLinter string      `json:"FromLinter"`
	Text       string      `json:"Text"`
	Severity   string      `json:"Severity"`
	Pos        golangciPos `json:"Pos"`
}

type golangciPos struct {
	Filename string `json:"Filename"`
	Line     int    `json:"Line"`
	Column   int    `json:"Column"`
}

func (a *GolangciLintAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	args := []string{"run", "--out-format=json", "--issues-exit-code=0"}
	if len(targets) > 0 && targets[0] != "." {
		args = append(args, targets...)
	} else {
		args = append(args, "./...")
	}

	result, err := a.runner.Run(ctx, "golangci-lint", args...)
	if err != nil {
		return nil, fmt.Errorf("golangci-lint: %w", err)
	}

	var parsed golangciResult
	if err := json.Unmarshal([]byte(result.Stdout), &parsed); err != nil {
		// If JSON parse fails, tool may have printed non-JSON errors
		if result.ExitCode != 0 {
			return nil, fmt.Errorf("golangci-lint failed (exit %d): %s", result.ExitCode, result.Stderr)
		}
		return nil, nil // no issues
	}

	var violations []Violation
	for _, issue := range parsed.Issues {
		severity := "WARNING"
		fixKind := "NONE"
		if issue.Severity == "error" {
			severity = "BLOCKING"
		}

		violations = append(violations, Violation{
			RuleID:   issue.FromLinter,
			Severity: severity,
			File:     issue.Pos.Filename,
			Line:     issue.Pos.Line,
			Message:  issue.Text,
			FixKind:  fixKind,
		})
	}

	return violations, nil
}

func (a *GolangciLintAdapter) Fix(ctx context.Context, targets []string) ([]FixResult, error) {
	args := []string{"run", "--fix", "--issues-exit-code=0"}
	if len(targets) > 0 && targets[0] != "." {
		args = append(args, targets...)
	} else {
		args = append(args, "./...")
	}

	result, err := a.runner.Run(ctx, "golangci-lint", args...)
	if err != nil {
		return nil, fmt.Errorf("golangci-lint fix: %w", err)
	}

	// golangci-lint --fix doesn't report what it fixed in structured form,
	// so we just report success/failure
	if result.ExitCode != 0 && strings.Contains(result.Stderr, "error") {
		return []FixResult{{
			RuleID: "golangci-lint",
			Action: "failed",
			Detail: result.Stderr,
		}}, nil
	}

	return []FixResult{{
		RuleID: "golangci-lint",
		Action: "applied",
		Detail: "ran golangci-lint --fix",
	}}, nil
}
