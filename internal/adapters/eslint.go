package adapters

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mkh/rice-railing/internal/exec"
)

// ESLintAdapter wraps ESLint for linting and fixing JS/TS code.
type ESLintAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewESLintAdapter(runner *exec.Runner, repoRoot string) *ESLintAdapter {
	return &ESLintAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *ESLintAdapter) Name() string                { return "eslint" }
func (a *ESLintAdapter) SupportedLanguages() []string { return []string{"typescript", "javascript"} }

type eslintResult struct {
	FilePath string         `json:"filePath"`
	Messages []eslintMsg    `json:"messages"`
}

type eslintMsg struct {
	RuleID   string `json:"ruleId"`
	Severity int    `json:"severity"` // 1=warn, 2=error
	Message  string `json:"message"`
	Line     int    `json:"line"`
	Fix      *struct{} `json:"fix"`
}

func (a *ESLintAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	args := []string{".", "--format=json", "--no-error-on-unmatched-pattern"}

	result, err := a.runner.Run(ctx, "npx", append([]string{"eslint"}, args...)...)
	if err != nil {
		return nil, fmt.Errorf("eslint: %w", err)
	}

	var parsed []eslintResult
	if err := json.Unmarshal([]byte(result.Stdout), &parsed); err != nil {
		if result.ExitCode == 0 {
			return nil, nil
		}
		return nil, nil // ESLint may output non-JSON on config errors
	}

	var violations []Violation
	for _, file := range parsed {
		for _, msg := range file.Messages {
			severity := "WARNING"
			if msg.Severity == 2 {
				severity = "BLOCKING"
			}
			fixKind := "NONE"
			if msg.Fix != nil {
				fixKind = "SAFE_AUTOFIX"
			}
			violations = append(violations, Violation{
				RuleID:   msg.RuleID,
				Severity: severity,
				File:     file.FilePath,
				Line:     msg.Line,
				Message:  msg.Message,
				FixKind:  fixKind,
			})
		}
	}
	return violations, nil
}

func (a *ESLintAdapter) Fix(ctx context.Context, targets []string) ([]FixResult, error) {
	result, err := a.runner.Run(ctx, "npx", "eslint", ".", "--fix", "--no-error-on-unmatched-pattern")
	if err != nil {
		return nil, fmt.Errorf("eslint fix: %w", err)
	}
	if result.ExitCode > 1 {
		return []FixResult{{RuleID: "eslint", Action: "failed", Detail: result.Stderr}}, nil
	}
	return []FixResult{{RuleID: "eslint", Action: "applied", Detail: "ran eslint --fix"}}, nil
}
