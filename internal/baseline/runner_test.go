package baseline

import (
	"context"
	"fmt"
	"testing"

	"github.com/mkh/rice-railing/internal/adapters"
	"github.com/mkh/rice-railing/internal/constitution"
)

// mockRuleEngine is a configurable mock implementing adapters.RuleEngineAdapter.
type mockRuleEngine struct {
	name       string
	languages  []string
	checkFn    func(ctx context.Context, targets []string) ([]adapters.Violation, error)
	fixFn      func(ctx context.Context, targets []string) ([]adapters.FixResult, error)
	checkCalls int
	fixCalls   int
}

func (m *mockRuleEngine) Name() string                 { return m.name }
func (m *mockRuleEngine) SupportedLanguages() []string { return m.languages }

func (m *mockRuleEngine) Check(ctx context.Context, targets []string) ([]adapters.Violation, error) {
	m.checkCalls++
	if m.checkFn != nil {
		return m.checkFn(ctx, targets)
	}
	return nil, nil
}

func (m *mockRuleEngine) Fix(ctx context.Context, targets []string) ([]adapters.FixResult, error) {
	m.fixCalls++
	if m.fixFn != nil {
		return m.fixFn(ctx, targets)
	}
	return nil, nil
}

func newConstitution() *constitution.Constitution {
	return &constitution.Constitution{
		Version: 1,
		Project: constitution.ProjectInfo{Name: "test"},
	}
}

func TestConvergenceWithNoViolations(t *testing.T) {
	r := NewRunner(newConstitution())
	r.RuleEngines = []adapters.RuleEngineAdapter{
		&mockRuleEngine{name: "empty-checker"},
	}
	r.Fixers = []adapters.RuleEngineAdapter{
		&mockRuleEngine{name: "empty-fixer"},
	}

	result, err := r.Run(context.Background(), []string{"."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Converged {
		t.Error("expected convergence with no violations")
	}
	if result.Iterations != 1 {
		t.Errorf("expected 1 iteration, got %d", result.Iterations)
	}
	if result.StopReason != "all blocking checks pass" {
		t.Errorf("unexpected stop reason: %s", result.StopReason)
	}
	if result.FixesApplied != 0 {
		t.Errorf("expected 0 fixes applied, got %d", result.FixesApplied)
	}
}

func TestConvergenceWithEmptyAdapters(t *testing.T) {
	r := NewRunner(newConstitution())
	// No engines or fixers at all.

	result, err := r.Run(context.Background(), []string{"."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Converged {
		t.Error("expected convergence with empty adapters")
	}
	if result.Iterations != 1 {
		t.Errorf("expected 1 iteration, got %d", result.Iterations)
	}
}

func TestReportOnlyModeDoesNotFix(t *testing.T) {
	fixer := &mockRuleEngine{
		name: "fixer",
		fixFn: func(ctx context.Context, targets []string) ([]adapters.FixResult, error) {
			t.Error("Fix should not be called in report-only mode")
			return nil, nil
		},
	}
	checker := &mockRuleEngine{
		name: "checker",
		checkFn: func(ctx context.Context, targets []string) ([]adapters.Violation, error) {
			return []adapters.Violation{
				{RuleID: "R1", Severity: "BLOCKING", File: "a.go", Message: "bad"},
			}, nil
		},
	}

	r := NewRunner(newConstitution())
	r.Mode = ModeReportOnly
	r.RuleEngines = []adapters.RuleEngineAdapter{checker}
	r.Fixers = []adapters.RuleEngineAdapter{fixer}

	result, err := r.Run(context.Background(), []string{"."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Converged {
		t.Error("should not converge with blocking violations")
	}
	if result.StopReason != "report-only mode" {
		t.Errorf("unexpected stop reason: %s", result.StopReason)
	}
	if fixer.fixCalls != 0 {
		t.Errorf("fixer was called %d times in report-only mode", fixer.fixCalls)
	}
	if len(result.Residual) != 1 {
		t.Errorf("expected 1 residual violation, got %d", len(result.Residual))
	}
	if result.FixesApplied != 0 {
		t.Errorf("expected 0 fixes in report-only, got %d", result.FixesApplied)
	}
}

func TestPerIterationProgressDetection(t *testing.T) {
	// The runner should use per-iteration fix count, not cumulative.
	// Iteration 1: fixer applies 1 fix, checker still reports blocking.
	// Iteration 2: fixer applies 0 fixes, checker still reports blocking.
	// Runner should stop after iteration 2 because no progress was made THIS iteration.
	iteration := 0
	fixer := &mockRuleEngine{
		name: "fixer",
		fixFn: func(ctx context.Context, targets []string) ([]adapters.FixResult, error) {
			iteration++
			if iteration == 1 {
				return []adapters.FixResult{
					{RuleID: "R1", File: "a.go", Action: "applied"},
				}, nil
			}
			return nil, nil
		},
	}
	checker := &mockRuleEngine{
		name: "checker",
		checkFn: func(ctx context.Context, targets []string) ([]adapters.Violation, error) {
			return []adapters.Violation{
				{RuleID: "R1", Severity: "BLOCKING", File: "a.go", Message: "still bad"},
			}, nil
		},
	}

	r := NewRunner(newConstitution())
	r.Mode = ModeSafeOnly
	r.RuleEngines = []adapters.RuleEngineAdapter{checker}
	r.Fixers = []adapters.RuleEngineAdapter{fixer}

	result, err := r.Run(context.Background(), []string{"."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Converged {
		t.Error("should not converge when blocking violations remain")
	}
	if result.Iterations != 2 {
		t.Errorf("expected 2 iterations (stop when no progress), got %d", result.Iterations)
	}
	if result.StopReason != "no further safe fixes available" {
		t.Errorf("unexpected stop reason: %s", result.StopReason)
	}
	if result.FixesApplied != 1 {
		t.Errorf("expected 1 cumulative fix, got %d", result.FixesApplied)
	}
	if len(result.Residual) != 1 {
		t.Errorf("expected 1 residual violation, got %d", len(result.Residual))
	}
}

func TestMaxIterationsStopCondition(t *testing.T) {
	// Fixer always makes progress so we never hit the "no progress" stop,
	// but we cap at MaxIter.
	fixCount := 0
	fixer := &mockRuleEngine{
		name: "fixer",
		fixFn: func(ctx context.Context, targets []string) ([]adapters.FixResult, error) {
			fixCount++
			return []adapters.FixResult{
				{RuleID: "R1", File: fmt.Sprintf("file%d.go", fixCount), Action: "applied"},
			}, nil
		},
	}
	checker := &mockRuleEngine{
		name: "checker",
		checkFn: func(ctx context.Context, targets []string) ([]adapters.Violation, error) {
			return []adapters.Violation{
				{RuleID: "R1", Severity: "BLOCKING", File: "a.go", Message: "infinite"},
			}, nil
		},
	}

	r := NewRunner(newConstitution())
	r.MaxIter = 3
	r.Mode = ModeSafeOnly
	r.RuleEngines = []adapters.RuleEngineAdapter{checker}
	r.Fixers = []adapters.RuleEngineAdapter{fixer}

	result, err := r.Run(context.Background(), []string{"."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Converged {
		t.Error("should not converge when always blocked")
	}
	if result.Iterations != 3 {
		t.Errorf("expected 3 iterations (max), got %d", result.Iterations)
	}
	if result.StopReason != "max iterations reached" {
		t.Errorf("unexpected stop reason: %s", result.StopReason)
	}
	if result.FixesApplied != 3 {
		t.Errorf("expected 3 fixes applied, got %d", result.FixesApplied)
	}
}

func TestFilesChangedDeduplication(t *testing.T) {
	fixer := &mockRuleEngine{
		name: "fixer",
		fixFn: func(ctx context.Context, targets []string) ([]adapters.FixResult, error) {
			return []adapters.FixResult{
				{RuleID: "R1", File: "a.go", Action: "applied"},
				{RuleID: "R2", File: "a.go", Action: "applied"},
			}, nil
		},
	}
	checker := &mockRuleEngine{
		name: "checker",
		checkFn: func(ctx context.Context, targets []string) ([]adapters.Violation, error) {
			return nil, nil // converges immediately
		},
	}

	r := NewRunner(newConstitution())
	r.RuleEngines = []adapters.RuleEngineAdapter{checker}
	r.Fixers = []adapters.RuleEngineAdapter{fixer}

	result, err := r.Run(context.Background(), []string{"."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.FilesChanged) != 1 {
		t.Errorf("expected 1 unique file, got %d: %v", len(result.FilesChanged), result.FilesChanged)
	}
}

func TestSkippedUnsafeCount(t *testing.T) {
	fixer := &mockRuleEngine{
		name: "fixer",
		fixFn: func(ctx context.Context, targets []string) ([]adapters.FixResult, error) {
			return []adapters.FixResult{
				{RuleID: "R1", File: "a.go", Action: "applied"},
				{RuleID: "R2", File: "b.go", Action: "skipped"},
				{RuleID: "R3", File: "c.go", Action: "skipped"},
			}, nil
		},
	}
	checker := &mockRuleEngine{name: "checker"} // no violations

	r := NewRunner(newConstitution())
	r.RuleEngines = []adapters.RuleEngineAdapter{checker}
	r.Fixers = []adapters.RuleEngineAdapter{fixer}

	result, err := r.Run(context.Background(), []string{"."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SkippedUnsafe != 2 {
		t.Errorf("expected 2 skipped unsafe, got %d", result.SkippedUnsafe)
	}
}

func TestCheckerError(t *testing.T) {
	checker := &mockRuleEngine{
		name: "broken-checker",
		checkFn: func(ctx context.Context, targets []string) ([]adapters.Violation, error) {
			return nil, fmt.Errorf("tool not found")
		},
	}

	r := NewRunner(newConstitution())
	r.RuleEngines = []adapters.RuleEngineAdapter{checker}

	_, err := r.Run(context.Background(), []string{"."})
	if err == nil {
		t.Fatal("expected error from broken checker")
	}
}

func TestFixerError(t *testing.T) {
	fixer := &mockRuleEngine{
		name: "broken-fixer",
		fixFn: func(ctx context.Context, targets []string) ([]adapters.FixResult, error) {
			return nil, fmt.Errorf("permission denied")
		},
	}

	r := NewRunner(newConstitution())
	r.Fixers = []adapters.RuleEngineAdapter{fixer}

	_, err := r.Run(context.Background(), []string{"."})
	if err == nil {
		t.Fatal("expected error from broken fixer")
	}
}
