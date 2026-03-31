package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mkh/rice-railing/internal/builder"
	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/discovery"
	"github.com/mkh/rice-railing/internal/profiling"
	"github.com/mkh/rice-railing/internal/reporting"
	"github.com/mkh/rice-railing/internal/resolution"
)

var (
	upgradeApply  bool
	upgradeDryRun bool
)

var upgradeToolkitCmd = &cobra.Command{
	Use:   "upgrade-toolkit",
	Short: "Evolve toolkit based on changed constitution or new dependencies",
	Long:  "Re-profile the repository, compare against stored profile, run gap analysis, and optionally rebuild the toolkit.",
	RunE:  runUpgradeToolkit,
}

func init() {
	upgradeToolkitCmd.Flags().BoolVar(&upgradeApply, "apply", false, "also run build-toolkit after updating")
	upgradeToolkitCmd.Flags().BoolVar(&upgradeDryRun, "dry-run", false, "show what would change without writing files")
}

func runUpgradeToolkit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	p := paths()
	rep := reporting.New(getFormat())

	// Load current constitution
	rep.Section("Loading Constitution")
	c, err := constitution.Load(p.Constitution)
	if err != nil {
		return fmt.Errorf("loading constitution (run 'rice-rail init' first): %w", err)
	}
	rep.Status(p.Constitution, "LOADED")

	// Load stored profile for comparison
	rep.Section("Loading Stored Profile")
	var oldProfile profiling.RepoProfile
	oldProfileData, err := os.ReadFile(p.Profile)
	if err != nil {
		rep.Status("stored profile", "NOT FOUND (will create new)")
	} else {
		if err := yaml.Unmarshal(oldProfileData, &oldProfile); err != nil {
			rep.Status("stored profile", fmt.Sprintf("PARSE ERROR: %v", err))
		} else {
			rep.Status(p.Profile, "LOADED")
		}
	}

	// Re-profile the repo
	rep.Section("Re-profiling Repository")
	scanner := profiling.NewScanner(cwd)
	newProfile, err := scanner.Scan()
	if err != nil {
		return fmt.Errorf("scanning repo: %w", err)
	}

	rep.Item("Languages", fmt.Sprintf("%d detected", len(newProfile.Languages)))
	rep.Item("Package managers", fmt.Sprintf("%d detected", len(newProfile.PackageManagers)))
	rep.Item("Build systems", fmt.Sprintf("%d detected", len(newProfile.BuildSystems)))

	// Compare profiles
	rep.Section("Profile Diff")
	diffs := compareProfiles(&oldProfile, newProfile)
	if len(diffs) == 0 {
		rep.Item("Changes", "none detected")
	} else {
		for _, d := range diffs {
			rep.Status(d.description, d.kind)
		}
	}

	// Run gap analysis with new profile
	rep.Section("Gap Analysis")
	inv := discovery.BuildInventory(newProfile)
	gapReport, rolloutPlan := resolution.Resolve(c, inv)

	for _, gap := range gapReport.Gaps {
		rep.Status(fmt.Sprintf("%s (%s)", gap.Capability, gap.Category), gap.Status)
	}
	rep.Item("Gaps found", fmt.Sprintf("%d", countUpgradeGaps(gapReport)))
	rep.Item("Rollout steps", fmt.Sprintf("%d", len(rolloutPlan.Steps)))

	if upgradeDryRun {
		rep.Section("Dry Run Complete")
		rep.Item("Mode", "dry-run (no files written)")
		return nil
	}

	// Save updated artifacts
	rep.Section("Saving Updated Artifacts")

	if err := os.MkdirAll(p.ToolkitDir, 0755); err != nil {
		return fmt.Errorf("creating toolkit dir: %w", err)
	}

	if err := reporting.WriteFile(p.Profile, newProfile); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}
	rep.Status(p.Profile, "SAVED")

	if err := reporting.WriteFile(p.ToolInventory, inv); err != nil {
		return fmt.Errorf("saving tool inventory: %w", err)
	}
	rep.Status(p.ToolInventory, "SAVED")

	if err := reporting.WriteFile(p.GapReport, gapReport); err != nil {
		return fmt.Errorf("saving gap report: %w", err)
	}
	rep.Status(p.GapReport, "SAVED")

	if err := reporting.WriteFile(p.RolloutPlan, rolloutPlan); err != nil {
		return fmt.Errorf("saving rollout plan: %w", err)
	}
	rep.Status(p.RolloutPlan, "SAVED")

	// Optionally run build-toolkit
	if upgradeApply {
		rep.Section("Applying: Building Toolkit")

		b := builder.NewBuilder(cwd, c, rolloutPlan)
		report, err := b.Build()
		if err != nil {
			return fmt.Errorf("building toolkit: %w", err)
		}

		for _, action := range report.Actions {
			rep.Status(action.Path, action.Type)
		}
		rep.Item("Build actions", fmt.Sprintf("%d", len(report.Actions)))
	}

	rep.Section("Upgrade Complete")
	if !upgradeApply {
		rep.Item("Next step", "Run 'rice-rail upgrade-toolkit --apply' to rebuild, or 'rice-rail build-toolkit' manually")
	}

	return nil
}

// profileDiff records a single difference between old and new profiles.
type profileDiff struct {
	kind        string // NEW, REMOVED, CHANGED
	description string
}

func compareProfiles(old, new *profiling.RepoProfile) []profileDiff {
	var diffs []profileDiff

	// Compare languages
	oldLangs := itemSet(old.Languages)
	newLangs := itemSet(new.Languages)
	for name := range newLangs {
		if !oldLangs[name] {
			diffs = append(diffs, profileDiff{kind: "NEW", description: "language: " + name})
		}
	}
	for name := range oldLangs {
		if !newLangs[name] {
			diffs = append(diffs, profileDiff{kind: "REMOVED", description: "language: " + name})
		}
	}

	// Compare package managers
	oldPMs := itemSet(old.PackageManagers)
	newPMs := itemSet(new.PackageManagers)
	for name := range newPMs {
		if !oldPMs[name] {
			diffs = append(diffs, profileDiff{kind: "NEW", description: "package manager: " + name})
		}
	}
	for name := range oldPMs {
		if !newPMs[name] {
			diffs = append(diffs, profileDiff{kind: "REMOVED", description: "package manager: " + name})
		}
	}

	// Compare build systems
	oldBS := itemSet(old.BuildSystems)
	newBS := itemSet(new.BuildSystems)
	for name := range newBS {
		if !oldBS[name] {
			diffs = append(diffs, profileDiff{kind: "NEW", description: "build system: " + name})
		}
	}
	for name := range oldBS {
		if !newBS[name] {
			diffs = append(diffs, profileDiff{kind: "REMOVED", description: "build system: " + name})
		}
	}

	// Compare topology
	if old.RepoTopology != "" && old.RepoTopology != new.RepoTopology {
		diffs = append(diffs, profileDiff{
			kind:        "CHANGED",
			description: fmt.Sprintf("topology: %s -> %s", old.RepoTopology, new.RepoTopology),
		})
	}

	return diffs
}

func itemSet(items []profiling.DetectedItem) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item.Name] = true
	}
	return s
}

func countUpgradeGaps(r *resolution.GapReport) int {
	count := 0
	for _, g := range r.Gaps {
		if g.Status != resolution.PresentReady {
			count++
		}
	}
	return count
}
