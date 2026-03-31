package cycle

import (
	"context"
	"fmt"

	"github.com/mkh/rice-railing/internal/adapters"
	"github.com/mkh/rice-railing/internal/constitution"
)

// Engine orchestrates the daily intent → tool → verify → refine loop.
type Engine struct {
	Constitution *constitution.Constitution
	RuleEngines  []adapters.RuleEngineAdapter
	Fixers       []adapters.RuleEngineAdapter
	Agent        adapters.AgentAdapter
	MaxIter      int
	Files        []string
	Module       string
}

// NewEngine creates a cycle engine.
func NewEngine(c *constitution.Constitution) *Engine {
	return &Engine{
		Constitution: c,
		MaxIter:      10,
	}
}

// Result is the outcome of a cycle run.
type Result struct {
	Intent         string          `yaml:"intent" json:"intent"`
	Iterations     int             `yaml:"iterations" json:"iterations"`
	Success        bool            `yaml:"success" json:"success"`
	ToolsInvoked   []string        `yaml:"tools_invoked" json:"tools_invoked"`
	RulesTriggered []string        `yaml:"rules_triggered" json:"rules_triggered"`
	FilesChanged   []string        `yaml:"files_changed" json:"files_changed"`
	Residual       []ResidualIssue `yaml:"residual" json:"residual"`
	Unresolved     []string        `yaml:"unresolved" json:"unresolved"`
	StopReason     string          `yaml:"stop_reason" json:"stop_reason"`
}

// ResidualIssue is a problem that couldn't be fixed mechanically.
type ResidualIssue struct {
	Type    string `yaml:"type" json:"type"` // mechanical, structural, semantic, policy
	File    string `yaml:"file" json:"file"`
	Message string `yaml:"message" json:"message"`
	RuleID  string `yaml:"rule_id,omitempty" json:"rule_id,omitempty"`
}

// Run executes the cycle for a given intent.
func (e *Engine) Run(ctx context.Context, intent string) (*Result, error) {
	result := &Result{
		Intent: intent,
	}

	targets := e.Files
	if len(targets) == 0 {
		targets = []string{"."}
	}

	for i := 0; i < e.MaxIter; i++ {
		result.Iterations = i + 1

		// Step 1: Run safe fixes (canonicalization, formatting)
		for _, fixer := range e.Fixers {
			fixes, err := fixer.Fix(ctx, targets)
			if err != nil {
				return nil, fmt.Errorf("fixer %s: %w", fixer.Name(), err)
			}
			result.ToolsInvoked = appendUnique(result.ToolsInvoked, fixer.Name())
			for _, f := range fixes {
				if f.Action == "applied" {
					result.FilesChanged = appendUnique(result.FilesChanged, f.File)
				}
			}
		}

		// Step 2: Run blocking checks
		var allViolations []adapters.Violation
		for _, engine := range e.RuleEngines {
			violations, err := engine.Check(ctx, targets)
			if err != nil {
				return nil, fmt.Errorf("checker %s: %w", engine.Name(), err)
			}
			result.ToolsInvoked = appendUnique(result.ToolsInvoked, engine.Name())
			allViolations = append(allViolations, violations...)
		}

		// Step 3: Classify violations
		var blocking, semantic []adapters.Violation
		for _, v := range allViolations {
			result.RulesTriggered = appendUnique(result.RulesTriggered, v.RuleID)
			switch v.FixKind {
			case "SAFE_AUTOFIX":
				// Already handled by fixers above — skip
				continue
			case "AI_REPAIR", "HUMAN_REVIEW":
				semantic = append(semantic, v)
			default:
				blocking = append(blocking, v)
			}
		}

		// Step 4: If no blocking violations, success
		if len(blocking) == 0 && len(semantic) == 0 {
			result.Success = true
			result.StopReason = "all checks pass"
			return result, nil
		}

		// Step 5: Surface semantic issues
		for _, v := range semantic {
			result.Residual = append(result.Residual, ResidualIssue{
				Type:    classifyIssue(v),
				File:    v.File,
				Message: v.Message,
				RuleID:  v.RuleID,
			})
		}

		// Step 6: If we have an agent, ask it to handle semantic issues
		if e.Agent != nil && len(semantic) > 0 {
			taskResult, err := e.Agent.RunTask(ctx, adapters.TaskInput{
				Intent: intent,
				Files:  targets,
				Module: e.Module,
			})
			if err != nil {
				result.Unresolved = append(result.Unresolved, fmt.Sprintf("agent error: %v", err))
			} else if taskResult != nil {
				result.FilesChanged = append(result.FilesChanged, taskResult.FilesChanged...)
				result.Unresolved = append(result.Unresolved, taskResult.Unresolved...)
			}
			continue // re-check after agent changes
		}

		// No agent available — report what's left
		if len(blocking) > 0 {
			for _, v := range blocking {
				result.Unresolved = append(result.Unresolved, fmt.Sprintf("[%s] %s:%d %s", v.RuleID, v.File, v.Line, v.Message))
			}
		}
		result.StopReason = "residual issues remain without agent"
		return result, nil
	}

	result.StopReason = "max iterations reached"
	return result, nil
}

// CheckScopeLimits verifies the cycle hasn't exceeded constitution limits.
func (e *Engine) CheckScopeLimits(result *Result) error {
	maxFiles := e.Constitution.Quality.MaxChangedFilesPerCycle
	if maxFiles > 0 && len(result.FilesChanged) > maxFiles {
		return fmt.Errorf("scope limit exceeded: %d files changed (max %d)", len(result.FilesChanged), maxFiles)
	}
	return nil
}

func classifyIssue(v adapters.Violation) string {
	switch v.FixKind {
	case "SAFE_AUTOFIX":
		return "mechanical"
	case "CODEMOD":
		return "structural"
	case "AI_REPAIR":
		return "semantic"
	case "HUMAN_REVIEW":
		return "policy"
	default:
		return "mechanical"
	}
}

func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}
