package cycle

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

// mockAgent implements adapters.AgentAdapter for testing.
type mockAgent struct {
	name      string
	taskFn    func(ctx context.Context, input adapters.TaskInput) (*adapters.TaskResult, error)
	taskCalls int
}

func (m *mockAgent) Name() string { return m.name }
func (m *mockAgent) Capabilities() adapters.AgentCapabilities {
	return adapters.AgentCapabilities{NonInteractive: true}
}
func (m *mockAgent) LoadWorkflowPack(ctx context.Context, packName string) error { return nil }
func (m *mockAgent) RunTask(ctx context.Context, input adapters.TaskInput) (*adapters.TaskResult, error) {
	m.taskCalls++
	if m.taskFn != nil {
		return m.taskFn(ctx, input)
	}
	return &adapters.TaskResult{Success: true}, nil
}

func newConstitution() *constitution.Constitution {
	return &constitution.Constitution{
		Version: 1,
		Project: constitution.ProjectInfo{Name: "test"},
	}
}

func TestEmptyAdaptersSuccess(t *testing.T) {
	e := NewEngine(newConstitution())
	// No engines, no fixers, no agent.

	result, err := e.Run(context.Background(), "format code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success with empty adapters")
	}
	if result.Iterations != 1 {
		t.Errorf("expected 1 iteration, got %d", result.Iterations)
	}
	if result.StopReason != "all checks pass" {
		t.Errorf("unexpected stop reason: %s", result.StopReason)
	}
	if result.Intent != "format code" {
		t.Errorf("expected intent 'format code', got %q", result.Intent)
	}
}

func TestSafeAutofixViolationsSkipped(t *testing.T) {
	// SAFE_AUTOFIX violations should be skipped (not blocking, not semantic).
	// If all violations are SAFE_AUTOFIX, the engine should succeed.
	checker := &mockRuleEngine{
		name: "checker",
		checkFn: func(ctx context.Context, targets []string) ([]adapters.Violation, error) {
			return []adapters.Violation{
				{RuleID: "FMT1", Severity: "WARNING", File: "a.go", Message: "formatting", FixKind: "SAFE_AUTOFIX"},
				{RuleID: "FMT2", Severity: "WARNING", File: "b.go", Message: "formatting", FixKind: "SAFE_AUTOFIX"},
			}, nil
		},
	}

	e := NewEngine(newConstitution())
	e.RuleEngines = []adapters.RuleEngineAdapter{checker}

	result, err := e.Run(context.Background(), "check formatting")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success when only SAFE_AUTOFIX violations exist")
	}
	if result.StopReason != "all checks pass" {
		t.Errorf("unexpected stop reason: %s", result.StopReason)
	}
}

func TestSafeAutofixMixedWithBlocking(t *testing.T) {
	// SAFE_AUTOFIX should be filtered out, but other violations remain blocking.
	checker := &mockRuleEngine{
		name: "checker",
		checkFn: func(ctx context.Context, targets []string) ([]adapters.Violation, error) {
			return []adapters.Violation{
				{RuleID: "FMT1", Severity: "WARNING", File: "a.go", Message: "formatting", FixKind: "SAFE_AUTOFIX"},
				{RuleID: "ERR1", Severity: "BLOCKING", File: "b.go", Message: "missing error check", FixKind: "NONE"},
			}, nil
		},
	}

	e := NewEngine(newConstitution())
	e.RuleEngines = []adapters.RuleEngineAdapter{checker}
	// No agent, so blocking issues cause stop.

	result, err := e.Run(context.Background(), "check code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Error("should not succeed with a non-SAFE_AUTOFIX blocking violation")
	}
	if result.StopReason != "residual issues remain without agent" {
		t.Errorf("unexpected stop reason: %s", result.StopReason)
	}
	if len(result.Unresolved) != 1 {
		t.Errorf("expected 1 unresolved issue, got %d", len(result.Unresolved))
	}
}

func TestScopeLimitEnforcement(t *testing.T) {
	c := newConstitution()
	c.Quality.MaxChangedFilesPerCycle = 2

	e := NewEngine(c)

	result := &Result{
		FilesChanged: []string{"a.go", "b.go", "c.go"},
	}
	err := e.CheckScopeLimits(result)
	if err == nil {
		t.Fatal("expected scope limit error")
	}

	// Under limit should pass.
	result.FilesChanged = []string{"a.go"}
	err = e.CheckScopeLimits(result)
	if err != nil {
		t.Fatalf("unexpected error for under-limit: %v", err)
	}

	// Exactly at limit should pass.
	result.FilesChanged = []string{"a.go", "b.go"}
	err = e.CheckScopeLimits(result)
	if err != nil {
		t.Fatalf("unexpected error for at-limit: %v", err)
	}
}

func TestScopeLimitZeroMeansUnlimited(t *testing.T) {
	c := newConstitution()
	c.Quality.MaxChangedFilesPerCycle = 0

	e := NewEngine(c)
	result := &Result{
		FilesChanged: []string{"a.go", "b.go", "c.go", "d.go", "e.go"},
	}
	err := e.CheckScopeLimits(result)
	if err != nil {
		t.Fatalf("zero limit should mean unlimited, got: %v", err)
	}
}

func TestSemanticIssueClassification(t *testing.T) {
	tests := []struct {
		fixKind  string
		wantType string
	}{
		{"SAFE_AUTOFIX", "mechanical"},
		{"CODEMOD", "structural"},
		{"AI_REPAIR", "semantic"},
		{"HUMAN_REVIEW", "policy"},
		{"UNKNOWN", "mechanical"},
		{"", "mechanical"},
	}

	for _, tt := range tests {
		t.Run(tt.fixKind, func(t *testing.T) {
			v := adapters.Violation{FixKind: tt.fixKind}
			got := classifyIssue(v)
			if got != tt.wantType {
				t.Errorf("classifyIssue(%q) = %q, want %q", tt.fixKind, got, tt.wantType)
			}
		})
	}
}

func TestAIRepairCreatesResidualIssue(t *testing.T) {
	checker := &mockRuleEngine{
		name: "checker",
		checkFn: func(ctx context.Context, targets []string) ([]adapters.Violation, error) {
			return []adapters.Violation{
				{RuleID: "SEC1", Severity: "BLOCKING", File: "auth.go", Line: 42, Message: "insecure hash", FixKind: "AI_REPAIR"},
			}, nil
		},
	}

	e := NewEngine(newConstitution())
	e.RuleEngines = []adapters.RuleEngineAdapter{checker}
	// No agent — semantic issues become residual and unresolved.

	result, err := e.Run(context.Background(), "security scan")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Error("should not succeed with semantic violations and no agent")
	}
	if len(result.Residual) != 1 {
		t.Fatalf("expected 1 residual, got %d", len(result.Residual))
	}
	if result.Residual[0].Type != "semantic" {
		t.Errorf("expected type 'semantic', got %q", result.Residual[0].Type)
	}
	if result.Residual[0].RuleID != "SEC1" {
		t.Errorf("expected rule ID 'SEC1', got %q", result.Residual[0].RuleID)
	}
}

func TestHumanReviewCreatesResidualIssue(t *testing.T) {
	checker := &mockRuleEngine{
		name: "checker",
		checkFn: func(ctx context.Context, targets []string) ([]adapters.Violation, error) {
			return []adapters.Violation{
				{RuleID: "POL1", File: "config.go", Message: "needs review", FixKind: "HUMAN_REVIEW"},
			}, nil
		},
	}

	e := NewEngine(newConstitution())
	e.RuleEngines = []adapters.RuleEngineAdapter{checker}

	result, err := e.Run(context.Background(), "policy check")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Residual) != 1 {
		t.Fatalf("expected 1 residual, got %d", len(result.Residual))
	}
	if result.Residual[0].Type != "policy" {
		t.Errorf("expected type 'policy', got %q", result.Residual[0].Type)
	}
}

func TestAgentInvokedForSemanticIssues(t *testing.T) {
	callCount := 0
	checker := &mockRuleEngine{
		name: "checker",
		checkFn: func(ctx context.Context, targets []string) ([]adapters.Violation, error) {
			callCount++
			if callCount <= 1 {
				return []adapters.Violation{
					{RuleID: "AI1", File: "a.go", Message: "needs AI fix", FixKind: "AI_REPAIR"},
				}, nil
			}
			// After agent fixes, no violations.
			return nil, nil
		},
	}

	agent := &mockAgent{
		name: "test-agent",
		taskFn: func(ctx context.Context, input adapters.TaskInput) (*adapters.TaskResult, error) {
			return &adapters.TaskResult{
				Success:      true,
				FilesChanged: []string{"a.go"},
			}, nil
		},
	}

	e := NewEngine(newConstitution())
	e.RuleEngines = []adapters.RuleEngineAdapter{checker}
	e.Agent = agent

	result, err := e.Run(context.Background(), "fix issues")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success after agent fix, stop reason: %s", result.StopReason)
	}
	if agent.taskCalls != 1 {
		t.Errorf("expected agent called once, got %d", agent.taskCalls)
	}
}

func TestMaxIterationsStopCondition(t *testing.T) {
	// Violations never clear, agent always tries but fails.
	checker := &mockRuleEngine{
		name: "checker",
		checkFn: func(ctx context.Context, targets []string) ([]adapters.Violation, error) {
			return []adapters.Violation{
				{RuleID: "AI1", File: "a.go", Message: "unfixable", FixKind: "AI_REPAIR"},
			}, nil
		},
	}
	agent := &mockAgent{
		name: "test-agent",
		taskFn: func(ctx context.Context, input adapters.TaskInput) (*adapters.TaskResult, error) {
			return &adapters.TaskResult{Success: false}, nil
		},
	}

	e := NewEngine(newConstitution())
	e.MaxIter = 3
	e.RuleEngines = []adapters.RuleEngineAdapter{checker}
	e.Agent = agent

	result, err := e.Run(context.Background(), "fix issues")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Iterations != 3 {
		t.Errorf("expected 3 iterations (max), got %d", result.Iterations)
	}
	if result.StopReason != "max iterations reached" {
		t.Errorf("unexpected stop reason: %s", result.StopReason)
	}
}

func TestCheckerError(t *testing.T) {
	checker := &mockRuleEngine{
		name: "broken",
		checkFn: func(ctx context.Context, targets []string) ([]adapters.Violation, error) {
			return nil, fmt.Errorf("lint crashed")
		},
	}

	e := NewEngine(newConstitution())
	e.RuleEngines = []adapters.RuleEngineAdapter{checker}

	_, err := e.Run(context.Background(), "check")
	if err == nil {
		t.Fatal("expected error from broken checker")
	}
}

func TestFixerError(t *testing.T) {
	fixer := &mockRuleEngine{
		name: "broken-fixer",
		fixFn: func(ctx context.Context, targets []string) ([]adapters.FixResult, error) {
			return nil, fmt.Errorf("disk full")
		},
	}

	e := NewEngine(newConstitution())
	e.Fixers = []adapters.RuleEngineAdapter{fixer}

	_, err := e.Run(context.Background(), "fix")
	if err == nil {
		t.Fatal("expected error from broken fixer")
	}
}

func TestDefaultTargetsWhenEmpty(t *testing.T) {
	var receivedTargets []string
	checker := &mockRuleEngine{
		name: "checker",
		checkFn: func(ctx context.Context, targets []string) ([]adapters.Violation, error) {
			receivedTargets = targets
			return nil, nil
		},
	}

	e := NewEngine(newConstitution())
	e.RuleEngines = []adapters.RuleEngineAdapter{checker}
	e.Files = nil // empty

	_, err := e.Run(context.Background(), "check")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(receivedTargets) != 1 || receivedTargets[0] != "." {
		t.Errorf("expected default target [\".\"], got %v", receivedTargets)
	}
}

func TestToolsInvokedTracking(t *testing.T) {
	checker := &mockRuleEngine{name: "eslint"}
	fixer := &mockRuleEngine{name: "prettier"}

	e := NewEngine(newConstitution())
	e.RuleEngines = []adapters.RuleEngineAdapter{checker}
	e.Fixers = []adapters.RuleEngineAdapter{fixer}

	result, err := e.Run(context.Background(), "lint")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantTools := map[string]bool{"eslint": true, "prettier": true}
	for _, tool := range result.ToolsInvoked {
		delete(wantTools, tool)
	}
	if len(wantTools) > 0 {
		t.Errorf("missing tools in ToolsInvoked: %v", wantTools)
	}
}

func TestRulesTriggeredTracking(t *testing.T) {
	checker := &mockRuleEngine{
		name: "checker",
		checkFn: func(ctx context.Context, targets []string) ([]adapters.Violation, error) {
			return []adapters.Violation{
				{RuleID: "R1", File: "a.go", Message: "x", FixKind: "SAFE_AUTOFIX"},
				{RuleID: "R2", File: "b.go", Message: "y", FixKind: "SAFE_AUTOFIX"},
				{RuleID: "R1", File: "c.go", Message: "z", FixKind: "SAFE_AUTOFIX"}, // duplicate rule
			}, nil
		},
	}

	e := NewEngine(newConstitution())
	e.RuleEngines = []adapters.RuleEngineAdapter{checker}

	result, err := e.Run(context.Background(), "lint")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.RulesTriggered) != 2 {
		t.Errorf("expected 2 unique rules triggered, got %d: %v", len(result.RulesTriggered), result.RulesTriggered)
	}
}
