package baseline

import (
	"context"
	"fmt"

	"github.com/mkh/rice-railing/internal/adapters"
	"github.com/mkh/rice-railing/internal/constitution"
)

// Mode controls baseline behavior.
type Mode int

const (
	ModeReportOnly Mode = iota
	ModeSafeOnly
	ModeSafePlusAI
)

// Runner executes baseline remediation.
type Runner struct {
	Constitution *constitution.Constitution
	Mode         Mode
	Module       string // empty = all
	MaxIter      int
	RuleEngines  []adapters.RuleEngineAdapter
	Fixers       []adapters.RuleEngineAdapter // adapters that support Fix()
}

// NewRunner creates a baseline runner.
func NewRunner(c *constitution.Constitution) *Runner {
	return &Runner{
		Constitution: c,
		Mode:         ModeSafeOnly,
		MaxIter:      10,
	}
}

// Result is the outcome of a baseline run.
type Result struct {
	Iterations    int                  `yaml:"iterations" json:"iterations"`
	Converged     bool                 `yaml:"converged" json:"converged"`
	FilesChanged  []string             `yaml:"files_changed" json:"files_changed"`
	FixesApplied  int                  `yaml:"fixes_applied" json:"fixes_applied"`
	Violations    []adapters.Violation `yaml:"violations" json:"violations"`
	Residual      []adapters.Violation `yaml:"residual" json:"residual"`
	SkippedUnsafe int                  `yaml:"skipped_unsafe" json:"skipped_unsafe"`
	StopReason    string               `yaml:"stop_reason" json:"stop_reason"`
}

// Run executes the baseline loop.
func (r *Runner) Run(ctx context.Context, targets []string) (*Result, error) {
	result := &Result{}

	for i := 0; i < r.MaxIter; i++ {
		result.Iterations = i + 1
		iterationFixes := 0

		// Step 1: Run all fixers (safe only)
		if r.Mode != ModeReportOnly {
			for _, fixer := range r.Fixers {
				fixes, err := fixer.Fix(ctx, targets)
				if err != nil {
					return nil, fmt.Errorf("fixer %s: %w", fixer.Name(), err)
				}
				for _, f := range fixes {
					switch f.Action {
					case "applied":
						iterationFixes++
						result.FixesApplied++
						result.FilesChanged = appendUnique(result.FilesChanged, f.File)
					case "skipped":
						result.SkippedUnsafe++
					}
				}
			}
		}

		// Step 2: Run all checks
		var allViolations []adapters.Violation
		for _, engine := range r.RuleEngines {
			violations, err := engine.Check(ctx, targets)
			if err != nil {
				return nil, fmt.Errorf("checker %s: %w", engine.Name(), err)
			}
			allViolations = append(allViolations, violations...)
		}

		// Step 3: Classify results
		var blocking []adapters.Violation
		for _, v := range allViolations {
			if v.Severity == "BLOCKING" {
				blocking = append(blocking, v)
			}
		}

		result.Violations = allViolations

		// Step 4: Check convergence
		if len(blocking) == 0 {
			result.Converged = true
			result.StopReason = "all blocking checks pass"
			return result, nil
		}

		// In report-only mode, don't loop
		if r.Mode == ModeReportOnly {
			result.Residual = blocking
			result.StopReason = "report-only mode"
			return result, nil
		}

		// Check if we made progress this iteration
		if iterationFixes == 0 && i > 0 {
			result.Residual = blocking
			result.StopReason = "no further safe fixes available"
			return result, nil
		}
	}

	result.StopReason = "max iterations reached"
	return result, nil
}

func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}
