package discovery

import (
	"github.com/mkh/rice-railing/internal/profiling"
)

// BuildInventory constructs a tool inventory from a repo profile.
func BuildInventory(profile *profiling.RepoProfile) *ToolInventory {
	inv := &ToolInventory{}

	addTooling := func(tools []profiling.DetectedTool, adapterClass string) {
		for _, t := range tools {
			status := "configured"
			if !t.Installed {
				status = "partially_configured"
			}
			inv.Tools = append(inv.Tools, InventoryEntry{
				Name:              t.Name,
				Category:          t.Category,
				AdapterClass:      adapterClass,
				InstalledLocally:  t.Installed,
				ConfiguredInRepo:  t.ConfigFile != "",
				Role:              t.Category,
				Status:            status,
				Evidence:          t.Evidence,
				RecommendedAction: recommendAction(t),
			})
		}
	}

	addTooling(profile.Tooling.Linters, "PolicyRuleAdapter")
	addTooling(profile.Tooling.Formatters, "CanonicalizationAdapter")
	addTooling(profile.Tooling.Typecheckers, "TypecheckAdapter")
	addTooling(profile.Tooling.TestRunners, "TestAdapter")
	addTooling(profile.Tooling.RuleEngines, "PolicyRuleAdapter")
	addTooling(profile.Tooling.Codemods, "StructuralRewriteAdapter")
	addTooling(profile.Tooling.Security, "DeepAnalysisAdapter")

	return inv
}

func recommendAction(t profiling.DetectedTool) string {
	if t.Installed && t.ConfigFile != "" {
		return "ready to use"
	}
	if t.ConfigFile != "" && !t.Installed {
		return "install tool"
	}
	return "configure"
}
