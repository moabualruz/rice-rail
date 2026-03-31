package adapters

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mkh/rice-railing/internal/exec"
)

// BiomeAdapter wraps Biome for linting and formatting JS/TS/JSON/CSS.
type BiomeAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewBiomeAdapter(runner *exec.Runner, repoRoot string) *BiomeAdapter {
	return &BiomeAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *BiomeAdapter) Name() string                { return "biome" }
func (a *BiomeAdapter) SupportedLanguages() []string { return []string{"typescript", "javascript"} }

type biomeDiagnostic struct {
	Category    string `json:"category"`
	Severity    string `json:"severity"` // error, warning, information
	Description string `json:"description"`
	Location    struct {
		Path struct {
			File string `json:"file"`
		} `json:"path"`
		Span struct {
			Start int `json:"start"`
		} `json:"span"`
	} `json:"location"`
}

type biomeOutput struct {
	Diagnostics []biomeDiagnostic `json:"diagnostics"`
}

func (a *BiomeAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	result, err := a.runner.Run(ctx, "npx", "biome", "check", "--reporter=json", ".")
	if err != nil {
		return nil, fmt.Errorf("biome: %w", err)
	}

	var output biomeOutput
	if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil {
		if result.ExitCode == 0 {
			return nil, nil
		}
		return nil, nil
	}

	var violations []Violation
	for _, d := range output.Diagnostics {
		severity := "WARNING"
		if d.Severity == "error" {
			severity = "BLOCKING"
		}
		violations = append(violations, Violation{
			RuleID:   d.Category,
			Severity: severity,
			File:     d.Location.Path.File,
			Line:     0,
			Message:  d.Description,
			FixKind:  "SAFE_AUTOFIX",
		})
	}
	return violations, nil
}

func (a *BiomeAdapter) Fix(ctx context.Context, targets []string) ([]FixResult, error) {
	result, err := a.runner.Run(ctx, "npx", "biome", "check", "--write", ".")
	if err != nil {
		return nil, fmt.Errorf("biome fix: %w", err)
	}
	if result.ExitCode != 0 {
		return []FixResult{{RuleID: "biome", Action: "failed", Detail: result.Stderr}}, nil
	}
	return []FixResult{{RuleID: "biome", Action: "applied", Detail: "ran biome check --write"}}, nil
}
