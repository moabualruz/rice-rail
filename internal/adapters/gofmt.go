package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/mkh/rice-railing/internal/exec"
)

// GofmtAdapter wraps gofmt for formatting Go code.
type GofmtAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewGofmtAdapter(runner *exec.Runner, repoRoot string) *GofmtAdapter {
	return &GofmtAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *GofmtAdapter) Name() string                 { return "gofmt" }
func (a *GofmtAdapter) SupportedLanguages() []string { return []string{"go"} }

func (a *GofmtAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	// gofmt -l lists files that differ from formatted form
	args := []string{"-l"}
	if len(targets) > 0 && targets[0] != "." {
		args = append(args, targets...)
	} else {
		args = append(args, ".")
	}

	result, err := a.runner.Run(ctx, "gofmt", args...)
	if err != nil {
		return nil, fmt.Errorf("gofmt: %w", err)
	}

	var violations []Violation
	for _, file := range strings.Split(strings.TrimSpace(result.Stdout), "\n") {
		if file == "" {
			continue
		}
		violations = append(violations, Violation{
			RuleID:   "gofmt",
			Severity: "BLOCKING",
			File:     file,
			Line:     0,
			Message:  "file not formatted",
			FixKind:  "SAFE_AUTOFIX",
		})
	}

	return violations, nil
}

func (a *GofmtAdapter) Fix(ctx context.Context, targets []string) ([]FixResult, error) {
	// gofmt -w writes formatted files in place
	args := []string{"-w"}
	if len(targets) > 0 && targets[0] != "." {
		args = append(args, targets...)
	} else {
		args = append(args, ".")
	}

	result, err := a.runner.Run(ctx, "gofmt", args...)
	if err != nil {
		return nil, fmt.Errorf("gofmt fix: %w", err)
	}

	if result.ExitCode != 0 {
		return []FixResult{{RuleID: "gofmt", Action: "failed", Detail: result.Stderr}}, nil
	}

	return []FixResult{{RuleID: "gofmt", Action: "applied", Detail: "formatted all Go files"}}, nil
}
