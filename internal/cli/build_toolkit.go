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

var buildDryRun bool

var buildToolkitCmd = &cobra.Command{
	Use:   "build-toolkit",
	Short: "Generate project-specific tooling from constitution",
	Long:  "Build wrapper commands, rules, codemods, workflow packs, and configs based on the project constitution and rollout plan.",
	RunE:  runBuildToolkit,
}

func init() {
	buildToolkitCmd.Flags().BoolVar(&buildDryRun, "dry-run", false, "show what would be generated without writing files")
}

func runBuildToolkit(cmd *cobra.Command, args []string) error {
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

	// Load rollout plan
	var plan resolution.RolloutPlan
	planData, err := os.ReadFile(p.RolloutPlan)
	if err != nil {
		rep.Status("rollout plan", "NOT FOUND (proceeding without)")
	} else {
		if err := yaml.Unmarshal(planData, &plan); err != nil {
			rep.Status("rollout plan", fmt.Sprintf("PARSE ERROR: %v (proceeding without)", err))
			plan = resolution.RolloutPlan{}
		}
	}

	rep.Section("Building Toolkit")

	b := builder.NewBuilder(cwd, c, &plan)
	b.DryRun = buildDryRun

	report, err := b.Build()
	if err != nil {
		return fmt.Errorf("building toolkit: %w", err)
	}

	for _, action := range report.Actions {
		label := action.Path
		if buildDryRun {
			label = "[dry-run] " + label
		}
		rep.Status(label, action.Type)
	}

	rep.Section("Build Complete")
	rep.Item("Actions", fmt.Sprintf("%d", len(report.Actions)))
	if buildDryRun {
		rep.Item("Mode", "dry-run (no files written)")
	}
	rep.Item("Next step", "Run 'rice-rail check' to validate, or 'rice-rail baseline' to remediate")

	return nil
}
