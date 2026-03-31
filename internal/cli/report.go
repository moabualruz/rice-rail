package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/reporting"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Print toolkit status, constitution summary, and known gaps",
	RunE:  runReport,
}

func runReport(cmd *cobra.Command, args []string) error {
	p := paths()
	rep := reporting.New(getFormat())

	rep.Section("Toolkit Status")

	files := []struct {
		path  string
		label string
	}{
		{p.Constitution, "Constitution"},
		{p.Profile, "Profile"},
		{p.ToolInventory, "Tool Inventory"},
		{p.GapReport, "Gap Report"},
		{p.RolloutPlan, "Rollout Plan"},
	}

	for _, f := range files {
		if _, err := os.Stat(f.path); err == nil {
			rep.Status(f.label, "PRESENT")
		} else {
			rep.Status(f.label, "MISSING")
		}
	}

	c, err := constitution.Load(p.Constitution)
	if err != nil {
		rep.Item("Constitution", "not found — run 'rice-rail init'")
		return nil
	}

	rep.Section("Constitution Summary")
	rep.Item("Architecture", c.Architecture.TargetStyle)
	rep.Item("Safety mode", c.Quality.SafetyMode)
	rep.Item("Languages", fmt.Sprintf("%v", c.Project.Languages))
	rep.Item("Blocking checks", fmt.Sprintf("%v", c.Quality.BlockOn))
	rep.Item("Advisory checks", fmt.Sprintf("%v", c.Quality.AdvisoryOn))
	rep.Item("Safe autofix", fmt.Sprintf("%v", c.Automation.AllowSafeAutofix))
	rep.Item("Unsafe autofix", fmt.Sprintf("%v", c.Automation.AllowUnsafeAutofix))
	rep.Item("Generated codemods", fmt.Sprintf("%v", c.Automation.AllowGeneratedCodemods))
	rep.Item("MCP enabled", fmt.Sprintf("%v", c.MCP.Enabled))

	if c.Architecture.Layering.Enabled {
		rep.Item("Layers", fmt.Sprintf("%v", c.Architecture.Layering.Layers))
	}

	return nil
}
