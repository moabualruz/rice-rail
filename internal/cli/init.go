package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mkh/rice-railing/internal/config"
	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/discovery"
	"github.com/mkh/rice-railing/internal/interview"
	"github.com/mkh/rice-railing/internal/profiling"
	"github.com/mkh/rice-railing/internal/provenance"
	"github.com/mkh/rice-railing/internal/reporting"
	"github.com/mkh/rice-railing/internal/resolution"
)

var (
	initQuick          bool
	initStrict         bool
	initSeed           string
	initNonInteractive bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Profile repo, run adaptive interview, generate constitution",
	Long:  "Analyze the current repository, ask adaptive questions, and generate a project constitution, tool inventory, gap report, and rollout plan.",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initQuick, "quick", false, "minimal questions, accept inferred defaults")
	initCmd.Flags().BoolVar(&initStrict, "strict", false, "ask all question groups")
	initCmd.Flags().StringVar(&initSeed, "seed", "", "path to profile seed file")
	initCmd.Flags().BoolVar(&initNonInteractive, "non-interactive", false, "accept all defaults without prompting")
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	p := paths()
	rep := reporting.New(getFormat())
	tracker := provenance.NewTracker()

	// Phase 1: Profile the repository
	rep.Section("Profiling Repository")
	scanner := profiling.NewScanner(cwd)
	profile, err := scanner.Scan()
	if err != nil {
		return fmt.Errorf("scanning repo: %w", err)
	}

	rep.Item("Languages", fmt.Sprintf("%d detected", len(profile.Languages)))
	for _, lang := range profile.Languages {
		rep.Status(fmt.Sprintf("%s (%s)", lang.Name, lang.Evidence), "OK")
		tracker.RecordInference("lang-"+lang.Name, "profiler", lang.Name, fmt.Sprintf("%.0f%%", lang.Confidence*100))
	}
	rep.Item("Package managers", fmt.Sprintf("%d detected", len(profile.PackageManagers)))
	rep.Item("Build systems", fmt.Sprintf("%d detected", len(profile.BuildSystems)))
	rep.Item("CI", fmt.Sprintf("%d detected", len(profile.CI)))
	rep.Item("Topology", profile.RepoTopology)
	rep.Item("Architecture hints", fmt.Sprintf("%d found", len(profile.ArchHints)))

	tracker.RecordInference("topology", "profiler", profile.RepoTopology, "high")
	for _, tool := range profile.Tooling.Linters {
		tracker.RecordInference("tool-"+tool.Name, "profiler", tool.Name+" (linter)", "high")
	}
	for _, tool := range profile.Tooling.Formatters {
		tracker.RecordInference("tool-"+tool.Name, "profiler", tool.Name+" (formatter)", "high")
	}

	// Save profile
	if err := os.MkdirAll(p.ToolkitDir, 0755); err != nil {
		return fmt.Errorf("creating toolkit dir: %w", err)
	}
	if err := reporting.WriteFile(p.Profile, profile); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}
	rep.Status(p.Profile, "SAVED")

	// Phase 2: Adaptive interview
	rep.Section("Adaptive Interview")

	mode := interview.ModeNormal
	if initQuick {
		mode = interview.ModeQuick
	} else if initStrict {
		mode = interview.ModeStrict
	}

	var seed *interview.Seed
	if initSeed != "" {
		seed, err = interview.LoadSeed(initSeed)
		if err != nil {
			return fmt.Errorf("loading seed: %w", err)
		}
		rep.Item("Seed loaded", initSeed)
	}

	engine := interview.NewEngine(mode, profile, seed)

	var prompter interview.Prompter
	if initNonInteractive {
		prompter = &interview.NonInteractivePrompter{}
	} else {
		prompter = &interview.RecordingPrompter{
			Inner: &interview.TerminalPrompter{},
		}
	}

	transcript, err := engine.Run(prompter)
	if err != nil {
		return fmt.Errorf("interview: %w", err)
	}

	rep.Item("Questions answered", fmt.Sprintf("%d", len(transcript.Answers)))
	rep.Item("Mode", transcript.Mode)

	for _, ans := range transcript.Answers {
		tracker.RecordUserDecision(ans.QuestionID, ans.Value, "source: "+ans.Source)
	}

	// Save interview log
	var interviewRecords []reporting.InterviewRecord
	for _, ans := range transcript.Answers {
		interviewRecords = append(interviewRecords, reporting.InterviewRecord{
			Question: ans.QuestionID,
			Answer:   ans.Value,
			Source:   ans.Source,
		})
	}
	if err := reporting.WriteInterviewLog(p.InterviewLog, interviewRecords); err != nil {
		rep.Status("interview log", fmt.Sprintf("WARN: %v", err))
	} else {
		rep.Status(p.InterviewLog, "SAVED")
	}

	// Phase 3: Generate constitution
	rep.Section("Generating Constitution")

	c := interview.BuildConstitution(profile, engine.Answers)
	if c.Project.Name == "" {
		c.Project.Name = filepath.Base(cwd)
	}
	if err := constitution.Save(c, p.Constitution); err != nil {
		return fmt.Errorf("saving constitution: %w", err)
	}
	rep.Status(p.Constitution, "SAVED")
	tracker.RecordGeneration("constitution", p.Constitution, "init")

	// Phase 4: Build tool inventory
	rep.Section("Tool Inventory")

	inv := discovery.BuildInventory(profile)
	if err := reporting.WriteFile(p.ToolInventory, inv); err != nil {
		return fmt.Errorf("saving tool inventory: %w", err)
	}
	rep.Item("Tools discovered", fmt.Sprintf("%d", len(inv.Tools)))
	rep.Status(p.ToolInventory, "SAVED")
	tracker.RecordGeneration("tool-inventory", p.ToolInventory, "init")

	// Phase 5: Gap analysis
	rep.Section("Gap Analysis")

	gapReport, rolloutPlan := resolution.Resolve(c, inv)
	if err := reporting.WriteFile(p.GapReport, gapReport); err != nil {
		return fmt.Errorf("saving gap report: %w", err)
	}
	if err := reporting.WriteFile(p.RolloutPlan, rolloutPlan); err != nil {
		return fmt.Errorf("saving rollout plan: %w", err)
	}

	for _, gap := range gapReport.Gaps {
		rep.Status(fmt.Sprintf("%s (%s)", gap.Capability, gap.Category), gap.Status)
	}
	rep.Item("Gaps found", fmt.Sprintf("%d", countGaps(gapReport)))
	rep.Item("Rollout steps", fmt.Sprintf("%d", len(rolloutPlan.Steps)))
	rep.Status(p.GapReport, "SAVED")
	rep.Status(p.RolloutPlan, "SAVED")
	tracker.RecordGeneration("gap-report", p.GapReport, "init")
	tracker.RecordGeneration("rollout-plan", p.RolloutPlan, "init")

	// Save provenance
	if err := tracker.Save(filepath.Join(cwd, config.ProvenanceDir)); err != nil {
		return fmt.Errorf("saving provenance: %w", err)
	}

	rep.Section("Init Complete")
	rep.Item("Next step", "Run 'rice-rail build-toolkit' to generate project tooling")

	return nil
}

func getFormat() reporting.Format {
	if jsonOut {
		return reporting.FormatJSON
	}
	return reporting.FormatText
}

func countGaps(r *resolution.GapReport) int {
	count := 0
	for _, g := range r.Gaps {
		if g.Status != resolution.PresentReady {
			count++
		}
	}
	return count
}
