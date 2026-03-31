package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/discovery"
	"github.com/mkh/rice-railing/internal/profiling"
	"github.com/mkh/rice-railing/internal/reporting"
	"github.com/mkh/rice-railing/internal/resolution"
)

var discoverToolsCmd = &cobra.Command{
	Use:   "discover-tools",
	Short: "Rediscover local/system/repo tools and update inventory",
	RunE:  runDiscoverTools,
}

func runDiscoverTools(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	p := paths()
	rep := reporting.New(getFormat())

	rep.Section("Discovering Tools")

	scanner := profiling.NewScanner(cwd)
	profile, err := scanner.Scan()
	if err != nil {
		return fmt.Errorf("scanning repo: %w", err)
	}

	inv := discovery.BuildInventory(profile)

	rep.Item("Languages", fmt.Sprintf("%d", len(profile.Languages)))
	rep.Item("Tools found", fmt.Sprintf("%d", len(inv.Tools)))

	for _, tool := range inv.Tools {
		status := "configured"
		if !tool.InstalledLocally {
			status = "config only (not on PATH)"
		}
		rep.Status(fmt.Sprintf("%s (%s)", tool.Name, tool.Category), status)
	}

	// Save updated inventory
	if err := reporting.WriteFile(p.ToolInventory, inv); err != nil {
		return fmt.Errorf("saving tool inventory: %w", err)
	}
	rep.Status(p.ToolInventory, "SAVED")

	// If constitution exists, also update gap report
	c, loadErr := constitution.Load(p.Constitution)
	if loadErr == nil {
		gapReport, rolloutPlan := resolution.Resolve(c, inv)
		_ = reporting.WriteFile(p.GapReport, gapReport)
		_ = reporting.WriteFile(p.RolloutPlan, rolloutPlan)
		rep.Status(p.GapReport, "UPDATED")
		rep.Status(p.RolloutPlan, "UPDATED")
	}

	rep.Section("Discovery Complete")
	return nil
}
