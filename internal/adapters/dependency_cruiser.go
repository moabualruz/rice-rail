package adapters

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mkh/rice-railing/internal/exec"
)

// DepCruiserAdapter wraps dependency-cruiser for JS/TS dependency rule checking.
type DepCruiserAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewDepCruiserAdapter(runner *exec.Runner, repoRoot string) *DepCruiserAdapter {
	return &DepCruiserAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *DepCruiserAdapter) Name() string                 { return "dependency-cruiser" }
func (a *DepCruiserAdapter) SupportedLanguages() []string { return []string{"typescript", "javascript"} }

// dependency-cruiser JSON output structures.
type depCruiserOutput struct {
	Summary struct {
		Violations []depCruiserViolation `json:"violations"`
	} `json:"summary"`
}

type depCruiserViolation struct {
	From string `json:"from"`
	To   string `json:"to"`
	Rule struct {
		Name     string `json:"name"`
		Severity string `json:"severity"`
	} `json:"rule"`
}

func (a *DepCruiserAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	args := []string{"--output-type", "json"}

	if len(targets) > 0 && targets[0] != "." {
		args = append(args, targets...)
	} else {
		args = append(args, "src/")
	}

	result, err := a.runner.Run(ctx, "depcruise", args...)
	if err != nil {
		return nil, fmt.Errorf("dependency-cruiser: %w", err)
	}

	if result.Stdout == "" {
		if result.ExitCode != 0 {
			return nil, fmt.Errorf("dependency-cruiser failed (exit %d): %s", result.ExitCode, result.Stderr)
		}
		return nil, nil
	}

	var output depCruiserOutput
	if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil {
		return nil, fmt.Errorf("dependency-cruiser: parse output: %w", err)
	}

	var violations []Violation
	for _, v := range output.Summary.Violations {
		severity := mapDepCruiserSeverity(v.Rule.Severity)

		violations = append(violations, Violation{
			RuleID:   v.Rule.Name,
			Severity: severity,
			File:     v.From,
			Line:     0, // dependency-cruiser reports module-level violations, not line-level
			Message:  fmt.Sprintf("forbidden dependency: %s -> %s", v.From, v.To),
			FixKind:  "HUMAN_REVIEW",
		})
	}

	return violations, nil
}

// Fix returns nil because dependency-cruiser has no autofix capability.
func (a *DepCruiserAdapter) Fix(_ context.Context, _ []string) ([]FixResult, error) {
	return nil, nil
}

func mapDepCruiserSeverity(s string) string {
	switch s {
	case "error":
		return "BLOCKING"
	case "warn":
		return "WARNING"
	case "info":
		return "INFO"
	default:
		return "WARNING"
	}
}
