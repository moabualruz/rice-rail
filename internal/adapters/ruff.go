package adapters

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mkh/rice-railing/internal/exec"
)

// RuffAdapter wraps Ruff for linting and formatting Python code.
type RuffAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewRuffAdapter(runner *exec.Runner, repoRoot string) *RuffAdapter {
	return &RuffAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *RuffAdapter) Name() string                 { return "ruff" }
func (a *RuffAdapter) SupportedLanguages() []string { return []string{"python"} }

type ruffDiagnostic struct {
	Code     string    `json:"code"`
	Message  string    `json:"message"`
	Fix      *struct{} `json:"fix"` // non-nil if fixable
	Location struct {
		Row    int `json:"row"`
		Column int `json:"column"`
	} `json:"location"`
	Filename string `json:"filename"`
}

func (a *RuffAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	args := []string{"check", "--output-format=json", "--exit-zero"}
	if len(targets) > 0 && targets[0] != "." {
		args = append(args, targets...)
	} else {
		args = append(args, ".")
	}

	result, err := a.runner.Run(ctx, "ruff", args...)
	if err != nil {
		return nil, fmt.Errorf("ruff: %w", err)
	}

	var diagnostics []ruffDiagnostic
	if err := json.Unmarshal([]byte(result.Stdout), &diagnostics); err != nil {
		if result.ExitCode != 0 {
			return nil, fmt.Errorf("ruff failed (exit %d): %s", result.ExitCode, result.Stderr)
		}
		return nil, nil
	}

	var violations []Violation
	for _, d := range diagnostics {
		fixKind := "NONE"
		if d.Fix != nil {
			fixKind = "SAFE_AUTOFIX"
		}
		violations = append(violations, Violation{
			RuleID:   d.Code,
			Severity: "BLOCKING",
			File:     d.Filename,
			Line:     d.Location.Row,
			Message:  d.Message,
			FixKind:  fixKind,
		})
	}

	return violations, nil
}

func (a *RuffAdapter) Fix(ctx context.Context, targets []string) ([]FixResult, error) {
	args := []string{"check", "--fix", "--exit-zero"}
	if len(targets) > 0 && targets[0] != "." {
		args = append(args, targets...)
	} else {
		args = append(args, ".")
	}

	result, err := a.runner.Run(ctx, "ruff", args...)
	if err != nil {
		return nil, fmt.Errorf("ruff fix: %w", err)
	}

	if result.ExitCode != 0 {
		return []FixResult{{RuleID: "ruff", Action: "failed", Detail: result.Stderr}}, nil
	}

	// Also run ruff format
	fmtArgs := []string{"format"}
	if len(targets) > 0 && targets[0] != "." {
		fmtArgs = append(fmtArgs, targets...)
	} else {
		fmtArgs = append(fmtArgs, ".")
	}

	a.runner.Run(ctx, "ruff", fmtArgs...)

	return []FixResult{{RuleID: "ruff", Action: "applied", Detail: "ran ruff check --fix + ruff format"}}, nil
}
