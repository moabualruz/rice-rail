package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/mkh/rice-railing/internal/exec"
)

// AstGrepAdapter wraps ast-grep (sg) for structural code search, rule checking, and codemods.
type AstGrepAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewAstGrepAdapter(runner *exec.Runner, repoRoot string) *AstGrepAdapter {
	return &AstGrepAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *AstGrepAdapter) Name() string                 { return "ast-grep" }
func (a *AstGrepAdapter) SupportedLanguages() []string { return []string{"*"} }

// ast-grep JSON output structures.
type astGrepMatch struct {
	File   string `json:"file"`
	RuleID string `json:"ruleId"`
	Range  struct {
		Start struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"start"`
	} `json:"range"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

func (a *AstGrepAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	args := []string{"scan", "--json"}

	if len(targets) > 0 && targets[0] != "." {
		args = append(args, targets...)
	}

	result, err := a.runner.Run(ctx, "ast-grep", args...)
	if err != nil {
		return nil, fmt.Errorf("ast-grep: %w", err)
	}

	if result.Stdout == "" {
		if result.ExitCode != 0 {
			return nil, fmt.Errorf("ast-grep failed (exit %d): %s", result.ExitCode, result.Stderr)
		}
		return nil, nil
	}

	var matches []astGrepMatch
	if err := json.Unmarshal([]byte(result.Stdout), &matches); err != nil {
		return nil, fmt.Errorf("ast-grep: parse output: %w", err)
	}

	var violations []Violation
	for _, m := range matches {
		severity := mapAstGrepSeverity(m.Severity)
		fixKind := "NONE"

		violations = append(violations, Violation{
			RuleID:   m.RuleID,
			Severity: severity,
			File:     m.File,
			Line:     m.Range.Start.Line,
			Message:  m.Message,
			FixKind:  fixKind,
		})
	}

	return violations, nil
}

func (a *AstGrepAdapter) Fix(ctx context.Context, targets []string) ([]FixResult, error) {
	args := []string{"scan", "--rewrite"}

	if len(targets) > 0 && targets[0] != "." {
		args = append(args, targets...)
	}

	result, err := a.runner.Run(ctx, "ast-grep", args...)
	if err != nil {
		return nil, fmt.Errorf("ast-grep fix: %w", err)
	}

	if result.ExitCode != 0 {
		return []FixResult{{RuleID: "ast-grep", Action: "failed", Detail: result.Stderr}}, nil
	}

	return []FixResult{{RuleID: "ast-grep", Action: "applied", Detail: "ran sg scan --rewrite"}}, nil
}

// Run implements CodemodEngineAdapter. It looks for a rule file at
// .project-toolkit/rules/ast-grep/<codemodID>.yml and applies it.
func (a *AstGrepAdapter) Run(ctx context.Context, codemodID string, targets []string, dryRun bool) (*CodemodResult, error) {
	rulePath := filepath.Join(a.repoRoot, ".project-toolkit", "rules", "ast-grep", codemodID+".yml")

	args := []string{"scan", "--rule", rulePath}
	if !dryRun {
		args = append(args, "--rewrite")
	}

	if len(targets) > 0 && targets[0] != "." {
		args = append(args, targets...)
	}

	result, err := a.runner.Run(ctx, "ast-grep", args...)
	if err != nil {
		return nil, fmt.Errorf("ast-grep codemod %s: %w", codemodID, err)
	}

	if result.ExitCode != 0 && result.Stdout == "" {
		return nil, fmt.Errorf("ast-grep codemod %s failed (exit %d): %s", codemodID, result.ExitCode, result.Stderr)
	}

	cr := &CodemodResult{
		CodemodID: codemodID,
		DryRun:    dryRun,
		Summary:   fmt.Sprintf("ast-grep rule %s applied", codemodID),
	}

	// Parse JSON output to extract changed files.
	if result.Stdout != "" {
		var matches []astGrepMatch
		if err := json.Unmarshal([]byte(result.Stdout), &matches); err == nil {
			seen := map[string]bool{}
			for _, m := range matches {
				if !seen[m.File] {
					cr.FilesChanged = append(cr.FilesChanged, m.File)
					seen[m.File] = true
				}
			}
		}
	}

	return cr, nil
}

func mapAstGrepSeverity(s string) string {
	switch s {
	case "error":
		return "BLOCKING"
	case "warning":
		return "WARNING"
	case "info", "hint":
		return "INFO"
	default:
		return "WARNING"
	}
}
