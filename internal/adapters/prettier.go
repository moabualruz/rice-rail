package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/mkh/rice-railing/internal/exec"
)

// PrettierAdapter wraps Prettier for formatting JS/TS/CSS/HTML/MD.
type PrettierAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewPrettierAdapter(runner *exec.Runner, repoRoot string) *PrettierAdapter {
	return &PrettierAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *PrettierAdapter) Name() string                { return "prettier" }
func (a *PrettierAdapter) SupportedLanguages() []string { return []string{"typescript", "javascript"} }

func (a *PrettierAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	result, err := a.runner.Run(ctx, "npx", "prettier", "--check", ".")
	if err != nil {
		return nil, fmt.Errorf("prettier: %w", err)
	}
	if result.ExitCode == 0 {
		return nil, nil
	}

	var violations []Violation
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "Checking") && !strings.HasPrefix(line, "All") {
			violations = append(violations, Violation{
				RuleID:   "prettier",
				Severity: "BLOCKING",
				File:     line,
				Message:  "not formatted",
				FixKind:  "SAFE_AUTOFIX",
			})
		}
	}
	return violations, nil
}

func (a *PrettierAdapter) Fix(ctx context.Context, targets []string) ([]FixResult, error) {
	result, err := a.runner.Run(ctx, "npx", "prettier", "--write", ".")
	if err != nil {
		return nil, fmt.Errorf("prettier fix: %w", err)
	}
	if result.ExitCode != 0 {
		return []FixResult{{RuleID: "prettier", Action: "failed", Detail: result.Stderr}}, nil
	}
	return []FixResult{{RuleID: "prettier", Action: "applied", Detail: "ran prettier --write"}}, nil
}
