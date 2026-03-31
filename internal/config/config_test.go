package config

import (
	"strings"
	"testing"
)

func TestDefaultPaths(t *testing.T) {
	p := DefaultPaths()

	fields := map[string]string{
		"ToolkitDir":       p.ToolkitDir,
		"Profile":          p.Profile,
		"Constitution":     p.Constitution,
		"ToolInventory":    p.ToolInventory,
		"CapabilityMatrix": p.CapabilityMatrix,
		"GapReport":        p.GapReport,
		"RolloutPlan":      p.RolloutPlan,
		"InterviewLog":     p.InterviewLog,
	}

	for name, val := range fields {
		if !strings.HasPrefix(val, ".project-toolkit") {
			t.Errorf("%s = %q does not start with %q", name, val, ".project-toolkit")
		}
	}
}

func TestResolvePathsDefault(t *testing.T) {
	p := ResolvePaths("")
	want := DefaultPaths()

	if p.ToolkitDir != want.ToolkitDir {
		t.Errorf("ToolkitDir: got %q, want %q", p.ToolkitDir, want.ToolkitDir)
	}
	if p.Constitution != want.Constitution {
		t.Errorf("Constitution: got %q, want %q", p.Constitution, want.Constitution)
	}
	if p.Profile != want.Profile {
		t.Errorf("Profile: got %q, want %q", p.Profile, want.Profile)
	}
	if p.ToolInventory != want.ToolInventory {
		t.Errorf("ToolInventory: got %q, want %q", p.ToolInventory, want.ToolInventory)
	}
	if p.GapReport != want.GapReport {
		t.Errorf("GapReport: got %q, want %q", p.GapReport, want.GapReport)
	}
	if p.RolloutPlan != want.RolloutPlan {
		t.Errorf("RolloutPlan: got %q, want %q", p.RolloutPlan, want.RolloutPlan)
	}
	if p.InterviewLog != want.InterviewLog {
		t.Errorf("InterviewLog: got %q, want %q", p.InterviewLog, want.InterviewLog)
	}
}

func TestResolvePathsCustom(t *testing.T) {
	p := ResolvePaths("/custom/path/constitution.yaml")

	if p.ToolkitDir != "/custom/path" {
		t.Errorf("ToolkitDir: got %q, want %q", p.ToolkitDir, "/custom/path")
	}
	if p.Profile != "/custom/path/profile.yaml" {
		t.Errorf("Profile: got %q, want %q", p.Profile, "/custom/path/profile.yaml")
	}
	if p.ToolInventory != "/custom/path/tool-inventory.yaml" {
		t.Errorf("ToolInventory: got %q, want %q", p.ToolInventory, "/custom/path/tool-inventory.yaml")
	}
	if p.GapReport != "/custom/path/gap-report.yaml" {
		t.Errorf("GapReport: got %q, want %q", p.GapReport, "/custom/path/gap-report.yaml")
	}
	if p.RolloutPlan != "/custom/path/rollout-plan.yaml" {
		t.Errorf("RolloutPlan: got %q, want %q", p.RolloutPlan, "/custom/path/rollout-plan.yaml")
	}
	if p.InterviewLog != "/custom/path/interview-log.md" {
		t.Errorf("InterviewLog: got %q, want %q", p.InterviewLog, "/custom/path/interview-log.md")
	}
}

func TestResolvePathsPreservesConstitutionPath(t *testing.T) {
	input := "/my/special/dir/constitution.yaml"
	p := ResolvePaths(input)

	if p.Constitution != input {
		t.Errorf("Constitution: got %q, want %q", p.Constitution, input)
	}
}
