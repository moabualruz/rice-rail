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

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run all blocking checks without modifications",
	RunE:  runCheck,
}

func runCheck(cmd *cobra.Command, args []string) error {
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

	rep.Section("Running Checks")
	rep.Item("Safety mode", c.Quality.SafetyMode)

	registry := adapters.DiscoverAdaptersWithCustom(cwd, c.Project.Languages, c.Tools.Custom)
	ctx := context.Background()
	totalViolations := 0
	hasFailure := false

	// Run rule engines
	for _, engine := range registry.RuleEngines {
		violations, err := engine.Check(ctx, []string{"."})
		if err != nil {
			rep.Status(engine.Name(), fmt.Sprintf("ERROR: %v", err))
			continue
		}
		blocking := 0
		for _, v := range violations {
			if v.Severity == "BLOCKING" {
				blocking++
			}
		}
		totalViolations += len(violations)
		if blocking > 0 {
			hasFailure = true
			rep.Status(engine.Name(), fmt.Sprintf("FAIL (%d blocking, %d total)", blocking, len(violations)))
		} else if len(violations) > 0 {
			rep.Status(engine.Name(), fmt.Sprintf("WARN (%d advisory)", len(violations)))
		} else {
			rep.Status(engine.Name(), "PASS")
		}
	}

	// Run typecheckers
	for _, tc := range registry.Typecheckers {
		violations, err := tc.Check(ctx, []string{"."})
		if err != nil {
			rep.Status(tc.Name(), fmt.Sprintf("ERROR: %v", err))
			continue
		}
		totalViolations += len(violations)
		if len(violations) > 0 {
			hasFailure = true
			rep.Status(tc.Name(), fmt.Sprintf("FAIL (%d errors)", len(violations)))
		} else {
			rep.Status(tc.Name(), "PASS")
		}
	}

	// Run test runners
	for _, tr := range registry.TestRunners {
		result, err := tr.Run(ctx, []string{"."})
		if err != nil {
			rep.Status(tr.Name(), fmt.Sprintf("ERROR: %v", err))
			continue
		}
		if result.Failed > 0 {
			hasFailure = true
			rep.Status(tr.Name(), fmt.Sprintf("FAIL (%d/%d passed)", result.Passed, result.Total))
		} else {
			rep.Status(tr.Name(), fmt.Sprintf("PASS (%d tests)", result.Total))
		}
	}

	if len(registry.RuleEngines) == 0 && len(registry.Typecheckers) == 0 && len(registry.TestRunners) == 0 {
		rep.Item("Note", "No tool adapters found on PATH. Install tools and re-run.")
	}

	rep.Section("Check Complete")
	rep.Item("Violations", fmt.Sprintf("%d", totalViolations))

	if hasFailure {
		return fmt.Errorf("blocking checks failed")
	}

	return nil
}
