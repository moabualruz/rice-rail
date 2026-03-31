package interview

import (
	"testing"

	"github.com/mkh/rice-railing/internal/profiling"
)

func TestNonInteractiveInterview(t *testing.T) {
	profile := &profiling.RepoProfile{
		RepoTopology: "single",
		Languages: []profiling.DetectedItem{
			{Name: "go", Confidence: 0.9},
		},
	}

	engine := NewEngine(ModeQuick, profile, nil)
	prompter := &NonInteractivePrompter{}

	transcript, err := engine.Run(prompter)
	if err != nil {
		t.Fatalf("interview failed: %v", err)
	}

	if len(transcript.Answers) == 0 {
		t.Fatal("expected answers")
	}
	if transcript.Mode != "quick" {
		t.Errorf("expected quick mode, got %s", transcript.Mode)
	}
}

func TestSeedOverridesInference(t *testing.T) {
	profile := &profiling.RepoProfile{
		RepoTopology: "single",
	}
	seed := &Seed{
		Answers: map[string]string{
			"arch_target": "hexagonal",
			"safety_mode": "strict",
		},
	}

	engine := NewEngine(ModeQuick, profile, seed)
	prompter := &NonInteractivePrompter{}

	_, err := engine.Run(prompter)
	if err != nil {
		t.Fatalf("interview failed: %v", err)
	}

	if a, ok := engine.Answers["arch_target"]; !ok || a.Value != "hexagonal" {
		t.Error("seed should override arch_target")
	}
	if a, ok := engine.Answers["safety_mode"]; !ok || a.Value != "strict" {
		t.Error("seed should set safety_mode to strict")
	}
}

func TestSkipConditions(t *testing.T) {
	engine := NewEngine(ModeNormal, nil, nil)

	// enforce_layers requires arch_target = layered_monolith|hexagonal|clean|modular_monolith
	// Without any arch_target answer, it should be skipped
	q := Question{
		ID:          "enforce_layers",
		RequireWhen: []string{"arch_target=hexagonal"},
	}

	if !engine.shouldSkip(q) {
		t.Error("should skip enforce_layers when arch_target is not set")
	}

	// Set arch_target
	engine.Answers["arch_target"] = &Answer{QuestionID: "arch_target", Value: "hexagonal"}
	if engine.shouldSkip(q) {
		t.Error("should NOT skip enforce_layers when arch_target=hexagonal")
	}
}

func TestConstitutionBuilder(t *testing.T) {
	profile := &profiling.RepoProfile{
		RepoTopology: "single",
		Languages:    []profiling.DetectedItem{{Name: "go"}},
	}
	answers := map[string]*Answer{
		"arch_target":        {Value: "hexagonal"},
		"safety_mode":        {Value: "strict"},
		"block_on":           {Values: []string{"lint", "tests"}},
		"allow_safe_autofix": {Value: "true"},
		"mcp_enabled":        {Value: "false"},
	}

	c := BuildConstitution(profile, answers)

	if c.Architecture.TargetStyle != "hexagonal" {
		t.Errorf("arch: got %s, want hexagonal", c.Architecture.TargetStyle)
	}
	if c.Quality.SafetyMode != "strict" {
		t.Errorf("safety: got %s, want strict", c.Quality.SafetyMode)
	}
	if len(c.Quality.BlockOn) != 2 {
		t.Errorf("block_on: got %d items, want 2", len(c.Quality.BlockOn))
	}
	if !c.Automation.AllowSafeAutofix {
		t.Error("allow_safe_autofix should be true")
	}
	if c.MCP.Enabled {
		t.Error("MCP should be disabled")
	}
}
