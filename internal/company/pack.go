package company

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/mkh/rice-railing/internal/constitution"
)

// Pack represents a company-level reusable doctrine pack.
type Pack struct {
	Name        string       `yaml:"name"`
	Version     string       `yaml:"version"`
	Description string       `yaml:"description"`
	Doctrine    DoctrinePack `yaml:"doctrine"`
}

// DoctrinePack holds optional doctrine sections that override constitution defaults.
type DoctrinePack struct {
	Architecture *constitution.ArchitectureSpec `yaml:"architecture,omitempty"`
	Quality      *constitution.QualitySpec      `yaml:"quality,omitempty"`
	Automation   *constitution.AutomationSpec   `yaml:"automation,omitempty"`
	Tools        *constitution.ToolPreferences  `yaml:"tool_preferences,omitempty"`
}

// Load reads a doctrine pack from a YAML file.
func Load(path string) (*Pack, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading pack %s: %w", path, err)
	}

	var p Pack
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing pack %s: %w", path, err)
	}

	return &p, nil
}

// Apply merges pack values into the constitution as defaults.
// Existing constitution values take precedence over pack values.
func Apply(pack *Pack, c *constitution.Constitution) {
	if pack.Doctrine.Architecture != nil {
		applyArchitecture(pack.Doctrine.Architecture, &c.Architecture)
	}
	if pack.Doctrine.Quality != nil {
		applyQuality(pack.Doctrine.Quality, &c.Quality)
	}
	if pack.Doctrine.Automation != nil {
		applyAutomation(pack.Doctrine.Automation, &c.Automation)
	}
	if pack.Doctrine.Tools != nil {
		applyTools(pack.Doctrine.Tools, &c.Tools)
	}
}

func applyArchitecture(pack *constitution.ArchitectureSpec, c *constitution.ArchitectureSpec) {
	if c.TargetStyle == "" {
		c.TargetStyle = pack.TargetStyle
	}
	if len(c.BoundedContexts) == 0 {
		c.BoundedContexts = pack.BoundedContexts
	}
	if len(c.Modules) == 0 {
		c.Modules = pack.Modules
	}
	if len(c.Layering.Layers) == 0 {
		c.Layering = pack.Layering
	}
}

func applyQuality(pack *constitution.QualitySpec, c *constitution.QualitySpec) {
	if c.SafetyMode == "" {
		c.SafetyMode = pack.SafetyMode
	}
	if len(c.BlockOn) == 0 {
		c.BlockOn = pack.BlockOn
	}
	if len(c.AdvisoryOn) == 0 {
		c.AdvisoryOn = pack.AdvisoryOn
	}
	if c.MaxChangedFilesPerCycle == 0 {
		c.MaxChangedFilesPerCycle = pack.MaxChangedFilesPerCycle
	}
	if c.MaxChangedLinesPerCycle == 0 {
		c.MaxChangedLinesPerCycle = pack.MaxChangedLinesPerCycle
	}
}

func applyAutomation(pack *constitution.AutomationSpec, c *constitution.AutomationSpec) {
	if c.BaselineModeDefault == "" {
		c.BaselineModeDefault = pack.BaselineModeDefault
	}
	if c.FeatureModeDefault == "" {
		c.FeatureModeDefault = pack.FeatureModeDefault
	}
}

func applyTools(pack *constitution.ToolPreferences, c *constitution.ToolPreferences) {
	if len(c.Linters) == 0 {
		c.Linters = pack.Linters
	}
	if len(c.Formatters) == 0 {
		c.Formatters = pack.Formatters
	}
	if len(c.Typecheckers) == 0 {
		c.Typecheckers = pack.Typecheckers
	}
	if len(c.TestRunners) == 0 {
		c.TestRunners = pack.TestRunners
	}
	if len(c.CodemodEngines) == 0 {
		c.CodemodEngines = pack.CodemodEngines
	}
	if len(c.RuleEngines) == 0 {
		c.RuleEngines = pack.RuleEngines
	}
}
