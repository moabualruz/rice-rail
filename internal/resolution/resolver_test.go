package resolution

import (
	"testing"

	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/discovery"
)

func TestResolveWithFullCoverage(t *testing.T) {
	c := &constitution.Constitution{
		Quality: constitution.QualitySpec{
			BlockOn: []string{"lint"},
		},
		Automation: constitution.AutomationSpec{
			AllowSafeAutofix: true,
		},
	}

	inv := &discovery.ToolInventory{
		Tools: []discovery.InventoryEntry{
			{Name: "eslint", Category: "linter", InstalledLocally: true, ConfiguredInRepo: true},
			{Name: "prettier", Category: "formatter", InstalledLocally: true, ConfiguredInRepo: true},
		},
	}

	gaps, plan := Resolve(c, inv)

	for _, g := range gaps.Gaps {
		if g.Capability == "linting" && g.Status != PresentReady {
			t.Errorf("linting should be PRESENT_READY, got %s", g.Status)
		}
		if g.Capability == "formatting" && g.Status != PresentReady {
			t.Errorf("formatting should be PRESENT_READY, got %s", g.Status)
		}
	}

	// Plan should only have steps for non-ready caps
	for _, step := range plan.Steps {
		if step.Capability == "linting" || step.Capability == "formatting" {
			t.Errorf("should not have rollout step for ready capability: %s", step.Capability)
		}
	}
}

func TestResolveWithMissingTools(t *testing.T) {
	c := &constitution.Constitution{
		Quality: constitution.QualitySpec{
			BlockOn: []string{"lint", "tests", "typecheck"},
		},
	}

	inv := &discovery.ToolInventory{} // empty

	gaps, plan := Resolve(c, inv)

	if len(gaps.Gaps) == 0 {
		t.Fatal("expected gaps")
	}

	for _, g := range gaps.Gaps {
		if g.Status == PresentReady {
			t.Errorf("no tools installed, %s should not be PRESENT_READY", g.Capability)
		}
	}

	if len(plan.Steps) == 0 {
		t.Fatal("expected rollout steps for missing tools")
	}
}

func TestResolveArchitectureGaps(t *testing.T) {
	c := &constitution.Constitution{
		Architecture: constitution.ArchitectureSpec{
			Layering:         constitution.LayeringSpec{Enabled: true},
			DependencyPolicy: constitution.DependencyPolicy{NoCycles: true},
		},
	}

	inv := &discovery.ToolInventory{}
	gaps, _ := Resolve(c, inv)

	found := map[string]bool{}
	for _, g := range gaps.Gaps {
		found[g.Capability] = true
	}

	if !found["layer_enforcement"] {
		t.Error("expected layer_enforcement gap")
	}
	if !found["cycle_detection"] {
		t.Error("expected cycle_detection gap")
	}
}
