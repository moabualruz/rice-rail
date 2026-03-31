package config

import (
	"path/filepath"

	"github.com/spf13/viper"
)

// ToolkitDir is the project-local toolkit directory.
const ToolkitDir = ".project-toolkit"

// AgentDir is the agent workflow pack directory.
const AgentDir = ".agent"

// Default relative paths for generated toolkit artifacts.
const (
	defaultProfilePath          = ToolkitDir + "/profile.yaml"
	defaultConstitutionPath     = ToolkitDir + "/constitution.yaml"
	defaultToolInventoryPath    = ToolkitDir + "/tool-inventory.yaml"
	defaultCapabilityMatrixPath = ToolkitDir + "/capability-matrix.yaml"
	defaultGapReportPath        = ToolkitDir + "/gap-report.yaml"
	defaultRolloutPlanPath      = ToolkitDir + "/rollout-plan.yaml"
	defaultInterviewLogPath     = ToolkitDir + "/interview-log.md"
)

// Subdirectory paths (always relative to ToolkitDir).
const (
	ProvenanceDir = ToolkitDir + "/provenance"
	PoliciesDir   = ToolkitDir + "/policies"
	RulesDir      = ToolkitDir + "/rules"
	CodemodsDir   = ToolkitDir + "/codemods"
	ScaffoldsDir  = ToolkitDir + "/scaffolds"
	PromptsDir    = ToolkitDir + "/prompts"
	ReportsDir    = ToolkitDir + "/reports"
	StateDir      = ToolkitDir + "/state"
	DocsDir       = ToolkitDir + "/docs"
)

// Paths resolves all artifact paths, respecting --config overrides.
// If --config was set, it determines the toolkit root directory.
type Paths struct {
	ToolkitDir       string
	Profile          string
	Constitution     string
	ToolInventory    string
	CapabilityMatrix string
	GapReport        string
	RolloutPlan      string
	InterviewLog     string
}

// ResolvePaths returns artifact paths based on the --config flag or defaults.
func ResolvePaths(configFlag string) Paths {
	if configFlag != "" {
		// If --config points to a file, derive toolkit dir from its parent
		dir := filepath.Dir(configFlag)
		return Paths{
			ToolkitDir:       dir,
			Profile:          filepath.Join(dir, "profile.yaml"),
			Constitution:     configFlag,
			ToolInventory:    filepath.Join(dir, "tool-inventory.yaml"),
			CapabilityMatrix: filepath.Join(dir, "capability-matrix.yaml"),
			GapReport:        filepath.Join(dir, "gap-report.yaml"),
			RolloutPlan:      filepath.Join(dir, "rollout-plan.yaml"),
			InterviewLog:     filepath.Join(dir, "interview-log.md"),
		}
	}
	return DefaultPaths()
}

// DefaultPaths returns the standard relative paths.
func DefaultPaths() Paths {
	return Paths{
		ToolkitDir:       ToolkitDir,
		Profile:          defaultProfilePath,
		Constitution:     defaultConstitutionPath,
		ToolInventory:    defaultToolInventoryPath,
		CapabilityMatrix: defaultCapabilityMatrixPath,
		GapReport:        defaultGapReportPath,
		RolloutPlan:      defaultRolloutPlanPath,
		InterviewLog:     defaultInterviewLogPath,
	}
}

// AppConfig holds runtime configuration merged from all layers.
type AppConfig struct {
	Verbose bool
	JSON    bool
	DryRun  bool
}

// FromViper reads runtime config from viper.
func FromViper() AppConfig {
	return AppConfig{
		Verbose: viper.GetBool("verbose"),
		JSON:    viper.GetBool("json"),
		DryRun:  viper.GetBool("dry-run"),
	}
}
