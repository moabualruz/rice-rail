package adapters

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mkh/rice-railing/internal/exec"
)

// SemgrepAdapter wraps Semgrep for cross-language rule checking.
type SemgrepAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewSemgrepAdapter(runner *exec.Runner, repoRoot string) *SemgrepAdapter {
	return &SemgrepAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *SemgrepAdapter) Name() string                 { return "semgrep" }
func (a *SemgrepAdapter) SupportedLanguages() []string { return []string{"*"} }

type semgrepOutput struct {
	Results []semgrepResult `json:"results"`
	Errors  []semgrepError  `json:"errors"`
}

type semgrepResult struct {
	CheckID string `json:"check_id"`
	Path    string `json:"path"`
	Start   struct {
		Line int `json:"line"`
		Col  int `json:"col"`
	} `json:"start"`
	Extra struct {
		Message  string `json:"message"`
		Severity string `json:"severity"`
		Metadata struct {
			Category string `json:"category"`
		} `json:"metadata"`
		Fix string `json:"fix"`
	} `json:"extra"`
}

type semgrepError struct {
	Message string `json:"message"`
	Level   string `json:"level"`
}

func (a *SemgrepAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	args := []string{"scan", "--json", "--quiet"}

	target := "."
	if len(targets) > 0 && targets[0] != "." {
		target = targets[0]
	}
	args = append(args, target)

	result, err := a.runner.Run(ctx, "semgrep", args...)
	if err != nil {
		return nil, fmt.Errorf("semgrep: %w", err)
	}

	var output semgrepOutput
	if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil {
		// Semgrep may exit non-zero but still produce JSON
		if result.Stdout == "" {
			if result.ExitCode != 0 {
				return nil, fmt.Errorf("semgrep failed (exit %d): %s", result.ExitCode, result.Stderr)
			}
			return nil, nil
		}
		return nil, nil
	}

	var violations []Violation
	for _, r := range output.Results {
		severity := "WARNING"
		switch r.Extra.Severity {
		case "ERROR":
			severity = "BLOCKING"
		case "WARNING":
			severity = "WARNING"
		case "INFO":
			severity = "INFO"
		}

		fixKind := "NONE"
		if r.Extra.Fix != "" {
			fixKind = "SAFE_AUTOFIX"
		}

		violations = append(violations, Violation{
			RuleID:   r.CheckID,
			Severity: severity,
			File:     r.Path,
			Line:     r.Start.Line,
			Message:  r.Extra.Message,
			FixKind:  fixKind,
		})
	}

	return violations, nil
}

func (a *SemgrepAdapter) Fix(ctx context.Context, targets []string) ([]FixResult, error) {
	args := []string{"scan", "--autofix", "--quiet"}

	target := "."
	if len(targets) > 0 && targets[0] != "." {
		target = targets[0]
	}
	args = append(args, target)

	result, err := a.runner.Run(ctx, "semgrep", args...)
	if err != nil {
		return nil, fmt.Errorf("semgrep autofix: %w", err)
	}

	if result.ExitCode != 0 {
		return []FixResult{{RuleID: "semgrep", Action: "failed", Detail: result.Stderr}}, nil
	}

	return []FixResult{{RuleID: "semgrep", Action: "applied", Detail: "ran semgrep --autofix"}}, nil
}
