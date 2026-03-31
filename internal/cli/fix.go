package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mkh/rice-railing/internal/adapters"
	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/reporting"
)

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Run safe autofixes and canonicalizers only",
	RunE:  runFix,
}

func runFix(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	p := paths()
	rep := reporting.New(getFormat())

	c, err := constitution.Load(p.Constitution)
	if err != nil {
		return fmt.Errorf("loading constitution (run 'rice-rail init' first): %w", err)
	}

	rep.Section("Running Safe Fixes")

	if !c.Automation.AllowSafeAutofix {
		rep.Item("Status", "Safe autofix is disabled in constitution")
		return nil
	}

	registry := adapters.DiscoverAdaptersWithCustom(cwd, c.Project.Languages, c.Tools.Custom)
	ctx := context.Background()
	totalApplied := 0

	for _, fixer := range registry.Fixers {
		fixes, err := fixer.Fix(ctx, []string{"."})
		if err != nil {
			rep.Status(fixer.Name(), fmt.Sprintf("ERROR: %v", err))
			continue
		}
		for _, f := range fixes {
			rep.Status(fixer.Name(), f.Action+": "+f.Detail)
			if f.Action == "applied" {
				totalApplied++
			}
		}
	}

	if len(registry.Fixers) == 0 {
		rep.Item("Note", "No fixer adapters found on PATH. Install tools and re-run.")
	}

	rep.Section("Fix Complete")
	rep.Item("Fixes applied", fmt.Sprintf("%d", totalApplied))

	return nil
}
