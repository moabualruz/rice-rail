package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mkh/rice-railing/internal/builder"
	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/reporting"
	"github.com/mkh/rice-railing/internal/resolution"
)

var (
	regenDryRun bool
	regenForce  bool
)

var regenerateCmd = &cobra.Command{
	Use:   "regenerate",
	Short: "Regenerate all generated files from current constitution",
	Long:  "Re-run the builder using the existing constitution and rollout plan without running init. Useful after manual constitution edits.",
	RunE:  runRegenerate,
}

func init() {
	regenerateCmd.Flags().BoolVar(&regenDryRun, "dry-run", false, "show what would be regenerated without writing files")
	regenerateCmd.Flags().BoolVar(&regenForce, "force", false, "overwrite hand-edited files (default: skip files modified since last generation)")
}

func runRegenerate(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	p := paths()
	rep := reporting.New(getFormat())

	// Load constitution
	rep.Section("Loading Constitution")
	c, err := constitution.Load(p.Constitution)
	if err != nil {
		return fmt.Errorf("loading constitution (run 'rice-rail init' first): %w", err)
	}
	rep.Status(p.Constitution, "LOADED")

	// Load rollout plan
	var plan resolution.RolloutPlan
	planData, err := os.ReadFile(p.RolloutPlan)
	if err != nil {
		rep.Status("rollout plan", "NOT FOUND (proceeding without)")
	} else {
		if err := yaml.Unmarshal(planData, &plan); err != nil {
			rep.Status("rollout plan", fmt.Sprintf("PARSE ERROR: %v (proceeding without)", err))
			plan = resolution.RolloutPlan{}
		} else {
			rep.Status(p.RolloutPlan, "LOADED")
		}
	}

	// Build toolkit
	rep.Section("Regenerating Toolkit")

	b := builder.NewBuilder(cwd, c, &plan)
	b.DryRun = regenDryRun

	report, err := b.Build()
	if err != nil {
		return fmt.Errorf("regenerating toolkit: %w", err)
	}

	for _, action := range report.Actions {
		label := action.Path
		if regenDryRun {
			label = "[dry-run] " + label
		}
		if regenForce {
			label = "[force] " + label
		}
		rep.Status(label, action.Type)
	}

	rep.Section("Regeneration Complete")
	rep.Item("Actions", fmt.Sprintf("%d", len(report.Actions)))
	if regenDryRun {
		rep.Item("Mode", "dry-run (no files written)")
	}
	if regenForce {
		rep.Item("Mode", "force (overwrote hand-edited files)")
	}
	rep.Item("Next step", "Run 'rice-rail doctor' to verify toolkit health")

	return nil
}
