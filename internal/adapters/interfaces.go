package adapters

import "context"

// ToolDiscoveryAdapter discovers tools available in the repo or system.
type ToolDiscoveryAdapter interface {
	Name() string
	Discover(ctx context.Context, repoRoot string) ([]DiscoveredTool, error)
}

// DiscoveredTool represents a tool found during discovery.
type DiscoveredTool struct {
	Name       string  `yaml:"name" json:"name"`
	Category   string  `yaml:"category" json:"category"`
	Evidence   string  `yaml:"evidence" json:"evidence"`
	Confidence float64 `yaml:"confidence" json:"confidence"`
	Installed  bool    `yaml:"installed" json:"installed"`
	ConfigPath string  `yaml:"config_path,omitempty" json:"config_path,omitempty"`
}

// RuleEngineAdapter runs rules against code and reports violations.
type RuleEngineAdapter interface {
	Name() string
	SupportedLanguages() []string
	Check(ctx context.Context, targets []string) ([]Violation, error)
	Fix(ctx context.Context, targets []string) ([]FixResult, error)
}

// Violation is a rule check result.
type Violation struct {
	RuleID   string `yaml:"rule_id" json:"rule_id"`
	Severity string `yaml:"severity" json:"severity"` // BLOCKING, WARNING, INFO
	File     string `yaml:"file" json:"file"`
	Line     int    `yaml:"line" json:"line"`
	Message  string `yaml:"message" json:"message"`
	FixKind  string `yaml:"fix_kind" json:"fix_kind"` // NONE, SAFE_AUTOFIX, UNSAFE_AUTOFIX, CODEMOD, AI_REPAIR, HUMAN_REVIEW
}

// FixResult reports what a fix operation did.
type FixResult struct {
	RuleID string `yaml:"rule_id" json:"rule_id"`
	File   string `yaml:"file" json:"file"`
	Action string `yaml:"action" json:"action"` // applied, skipped, failed
	Detail string `yaml:"detail,omitempty" json:"detail,omitempty"`
}

// CodemodEngineAdapter executes codemods against code.
type CodemodEngineAdapter interface {
	Name() string
	SupportedLanguages() []string
	Run(ctx context.Context, codemodID string, targets []string, dryRun bool) (*CodemodResult, error)
}

// CodemodResult reports what a codemod did.
type CodemodResult struct {
	CodemodID    string   `yaml:"codemod_id" json:"codemod_id"`
	FilesChanged []string `yaml:"files_changed" json:"files_changed"`
	DryRun       bool     `yaml:"dry_run" json:"dry_run"`
	Summary      string   `yaml:"summary" json:"summary"`
}

// AgentAdapter allows a CLI agent to consume project-owned workflow packs.
type AgentAdapter interface {
	Name() string
	Capabilities() AgentCapabilities
	LoadWorkflowPack(ctx context.Context, packName string) error
	RunTask(ctx context.Context, input TaskInput) (*TaskResult, error)
}

// AgentCapabilities describes what an agent adapter supports.
type AgentCapabilities struct {
	MCP              bool `json:"mcp"`
	LocalTools       bool `json:"local_tools"`
	StructuredOutput bool `json:"structured_output"`
	NonInteractive   bool `json:"non_interactive"`
	PatchPreview     bool `json:"patch_preview"`
	ScopeConstraints bool `json:"scope_constraints"`
}

// TaskInput is the input to an agent task run.
type TaskInput struct {
	Intent      string            `json:"intent"`
	Files       []string          `json:"files,omitempty"`
	Module      string            `json:"module,omitempty"`
	Constraints map[string]string `json:"constraints,omitempty"`
}

// TaskResult is the output of an agent task run.
type TaskResult struct {
	Success      bool     `json:"success"`
	FilesChanged []string `json:"files_changed"`
	Summary      string   `json:"summary"`
	Unresolved   []string `json:"unresolved,omitempty"`
}

// TypecheckAdapter runs type checking.
type TypecheckAdapter interface {
	Name() string
	SupportedLanguages() []string
	Check(ctx context.Context, targets []string) ([]Violation, error)
}

// TestRunnerAdapter runs tests.
type TestRunnerAdapter interface {
	Name() string
	SupportedLanguages() []string
	Run(ctx context.Context, targets []string) (*TestResult, error)
}

// TestResult reports test execution.
type TestResult struct {
	Passed int    `json:"passed"`
	Failed int    `json:"failed"`
	Total  int    `json:"total"`
	Output string `json:"output"`
}

// BuildSystemAdapter interacts with build systems.
type BuildSystemAdapter interface {
	Name() string
	Build(ctx context.Context, targets []string) error
}
