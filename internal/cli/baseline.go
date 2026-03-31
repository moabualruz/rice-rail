package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/mkh/rice-railing/internal/adapters"
	"github.com/mkh/rice-railing/internal/baseline"
	"github.com/mkh/rice-railing/internal/config"
	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/reporting"
)

var (
	baselineReportOnly bool
	baselineWithAI     bool
	baselineModule     string
)

var baselineCmd = &cobra.Command{
	Use:   "baseline",
	Short: "Normalize existing codebase to policy-compliant baseline",
	Long:  "Run baseline remediation: canonicalize, fix, check, and optionally apply AI-assisted repairs until convergence.",
	RunE:  runBaseline,
}

func init() {
	baselineCmd.Flags().BoolVar(&baselineReportOnly, "report-only", false, "report violations without modifying files")
	baselineCmd.Flags().BoolVar(&baselineWithAI, "with-ai", false, "enable AI-assisted repairs for residual issues")
	baselineCmd.Flags().StringVar(&baselineModule, "module", "", "target a specific module")
}

func runBaseline(cmd *cobra.Command, args []string) error {
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

	rep.Section("Baseline Remediation")

	mode := baseline.ModeSafeOnly
	if baselineReportOnly {
		mode = baseline.ModeReportOnly
		rep.Item("Mode", "report-only")
	} else if baselineWithAI {
		mode = baseline.ModeSafePlusAI
		rep.Item("Mode", "safe + AI repair")
	} else {
		rep.Item("Mode", "safe-only")
	}

	registry := adapters.DiscoverAdaptersWithCustom(cwd, c.Project.Languages, c.Tools.Custom)

	runner := baseline.NewRunner(c)
	runner.Mode = mode
	runner.Module = baselineModule
	runner.RuleEngines = registry.RuleEngines
	runner.Fixers = registry.Fixers

	ctx := context.Background()
	result, err := runner.Run(ctx, []string{"."})
	if err != nil {
		return fmt.Errorf("baseline: %w", err)
	}

	rep.Item("Iterations", fmt.Sprintf("%d", result.Iterations))
	rep.Item("Converged", fmt.Sprintf("%v", result.Converged))
	rep.Item("Files changed", fmt.Sprintf("%d", len(result.FilesChanged)))
	rep.Item("Fixes applied", fmt.Sprintf("%d", result.FixesApplied))
	rep.Item("Skipped unsafe", fmt.Sprintf("%d", result.SkippedUnsafe))
	rep.Item("Residual violations", fmt.Sprintf("%d", len(result.Residual)))
	rep.Item("Stop reason", result.StopReason)

	for _, v := range result.Residual {
		rep.Status(fmt.Sprintf("%s:%d %s", v.File, v.Line, v.Message), v.RuleID)
	}

	// Save baseline state
	stateDir := filepath.Join(cwd, config.StateDir)
	reporting.SaveBaselineState(stateDir, reporting.BaselineState{
		LastRun:    time.Now(),
		Converged:  result.Converged,
		Iterations: result.Iterations,
		StopReason: result.StopReason,
		Violations: len(result.Residual),
	})
	reporting.SaveRunState(stateDir, reporting.RunState{
		Command:   "baseline",
		Timestamp: time.Now(),
		Success:   result.Converged,
		Summary:   result.StopReason,
	})

	rep.Section("Baseline Complete")

	return nil
}
