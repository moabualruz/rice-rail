package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mkh/rice-railing/internal/adapters"
	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/cycle"
	"github.com/mkh/rice-railing/internal/reporting"
)

var (
	cycleFiles   []string
	cycleModule  string
	cycleMaxIter int
)

var cycleCmd = &cobra.Command{
	Use:   "cycle [intent]",
	Short: "Run intent → tool → verify → refine loop",
	Long:  "Execute a constrained development cycle: parse intent, scaffold, transform, fix, check, and refine until done.",
	Args:  cobra.ExactArgs(1),
	RunE:  runCycle,
}

func init() {
	cycleCmd.Flags().StringSliceVar(&cycleFiles, "files", nil, "constrain to specific files")
	cycleCmd.Flags().StringVar(&cycleModule, "module", "", "constrain to a specific module")
	cycleCmd.Flags().IntVar(&cycleMaxIter, "max-iterations", 10, "maximum refinement iterations")
}

func runCycle(cmd *cobra.Command, args []string) error {
	intent := args[0]
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

	rep.Section("Cycle")
	rep.Item("Intent", intent)
	rep.Item("Safety mode", c.Quality.SafetyMode)

	registry := adapters.DiscoverAdaptersWithCustom(cwd, c.Project.Languages, c.Tools.Custom)

	engine := cycle.NewEngine(c)
	engine.MaxIter = cycleMaxIter
	engine.Files = cycleFiles
	engine.Module = cycleModule
	engine.RuleEngines = registry.RuleEngines
	engine.Fixers = registry.Fixers
	if len(registry.Agents) > 0 {
		engine.Agent = registry.Agents[0]
	}

	ctx := context.Background()
	result, err := engine.Run(ctx, intent)
	if err != nil {
		return fmt.Errorf("cycle: %w", err)
	}

	if err := engine.CheckScopeLimits(result); err != nil {
		rep.Status("SCOPE LIMIT", err.Error())
	}

	rep.Item("Iterations", fmt.Sprintf("%d", result.Iterations))
	rep.Item("Success", fmt.Sprintf("%v", result.Success))
	rep.Item("Tools invoked", fmt.Sprintf("%v", result.ToolsInvoked))
	rep.Item("Rules triggered", fmt.Sprintf("%d", len(result.RulesTriggered)))
	rep.Item("Files changed", fmt.Sprintf("%d", len(result.FilesChanged)))
	rep.Item("Residual issues", fmt.Sprintf("%d", len(result.Residual)))
	rep.Item("Unresolved", fmt.Sprintf("%d", len(result.Unresolved)))
	rep.Item("Stop reason", result.StopReason)

	for _, issue := range result.Residual {
		rep.Status(fmt.Sprintf("[%s] %s: %s", issue.Type, issue.File, issue.Message), "RESIDUAL")
	}
	for _, u := range result.Unresolved {
		rep.Status(u, "UNRESOLVED")
	}

	rep.Section("Cycle Complete")

	return nil
}
