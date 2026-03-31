package interview

import (
	"strconv"
	"strings"

	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/profiling"
)

// BuildConstitution converts interview answers + profile into a constitution draft.
func BuildConstitution(profile *profiling.RepoProfile, answers map[string]*Answer) *constitution.Constitution {
	c := &constitution.Constitution{
		Version: 1,
	}

	// Project info from profile
	if profile != nil {
		c.Project.RepoType = profile.RepoTopology
		for _, lang := range profile.Languages {
			c.Project.Languages = append(c.Project.Languages, lang.Name)
		}
		for _, pm := range profile.PackageManagers {
			c.Project.PackageManagers = append(c.Project.PackageManagers, pm.Name)
		}
		for _, bs := range profile.BuildSystems {
			c.Project.BuildSystems = append(c.Project.BuildSystems, bs.Name)
		}
	}

	// Architecture from answers
	if a, ok := answers["arch_target"]; ok {
		c.Architecture.TargetStyle = a.Value
	}
	if a, ok := answers["enforce_layers"]; ok && a.Value == "true" {
		c.Architecture.Layering.Enabled = true
	}
	if a, ok := answers["no_cycles"]; ok {
		c.Architecture.DependencyPolicy.NoCycles = a.Value == "true"
	}

	// Quality from answers
	if a, ok := answers["safety_mode"]; ok {
		c.Quality.SafetyMode = a.Value
	}
	if a, ok := answers["block_on"]; ok {
		c.Quality.BlockOn = splitValues(a)
	}
	if a, ok := answers["advisory_on"]; ok {
		c.Quality.AdvisoryOn = splitValues(a)
	}
	if a, ok := answers["require_tests"]; ok {
		c.Quality.RequireChangedCodeTests = a.Value == "true"
	}
	if a, ok := answers["max_files_per_cycle"]; ok {
		if v, err := strconv.Atoi(a.Value); err == nil {
			c.Quality.MaxChangedFilesPerCycle = v
		}
	}
	if a, ok := answers["max_lines_per_cycle"]; ok {
		if v, err := strconv.Atoi(a.Value); err == nil {
			c.Quality.MaxChangedLinesPerCycle = v
		}
	}

	// Automation from answers
	if a, ok := answers["allow_safe_autofix"]; ok {
		c.Automation.AllowSafeAutofix = a.Value == "true"
	}
	if a, ok := answers["allow_unsafe_autofix"]; ok {
		c.Automation.AllowUnsafeAutofix = a.Value == "true"
	}
	if a, ok := answers["allow_codemods"]; ok {
		c.Automation.AllowGeneratedCodemods = a.Value == "true"
	}
	if a, ok := answers["separate_baseline"]; ok && a.Value == "true" {
		c.Workflow.SeparateBaselineAndFeatureWork = true
	}
	c.Automation.BaselineModeDefault = "safe_only"
	c.Automation.FeatureModeDefault = "constrained_cycle"

	// Workflow from answers
	if a, ok := answers["generate_ci"]; ok {
		c.Workflow.GenerateCIIntegration = a.Value == "true"
	}
	if a, ok := answers["generate_wrappers"]; ok {
		c.Workflow.GenerateLocalWrappers = a.Value == "true"
	}
	if a, ok := answers["constitution_changes_ack"]; ok {
		c.Workflow.RequireHumanAckForConstitutionChanges = a.Value == "true"
	}

	// MCP from answers
	if a, ok := answers["mcp_enabled"]; ok {
		c.MCP.Enabled = a.Value == "true"
	}
	if a, ok := answers["offline_important"]; ok && a.Value == "true" {
		c.MCP.AccessPolicy = "local_only"
	} else {
		c.MCP.AccessPolicy = "allow_remote_optional"
	}

	return c
}

func splitValues(a *Answer) []string {
	if len(a.Values) > 0 {
		return a.Values
	}
	if a.Value == "" {
		return nil
	}
	return strings.Split(a.Value, ",")
}
