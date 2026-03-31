package builder

import (
	"strings"
	"testing"

	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/resolution"
)

func TestRenderOperatorGuide(t *testing.T) {
	c := testConstitution()
	c.Quality.AdvisoryOn = []string{"complexity", "coverage"}
	c.Automation.AllowUnsafeAutofix = false

	output := RenderOperatorGuide(c)

	// Safety mode
	if !strings.Contains(output, "Safety mode") {
		t.Error("operator guide should contain safety mode heading")
	}
	if !strings.Contains(output, "balanced") {
		t.Error("operator guide should contain the actual safety mode value")
	}

	// Blocking checks
	if !strings.Contains(output, "Blocking Checks") {
		t.Error("operator guide should contain blocking checks section")
	}
	for _, check := range []string{"lint", "tests", "typecheck"} {
		if !strings.Contains(output, check) {
			t.Errorf("operator guide should list blocking check %q", check)
		}
	}

	// Quick reference table
	if !strings.Contains(output, "Quick Reference") {
		t.Error("operator guide should contain quick reference section")
	}
	if !strings.Contains(output, "rice-rail check") {
		t.Error("operator guide should contain rice-rail check in quick reference")
	}
	if !strings.Contains(output, "rice-rail fix") {
		t.Error("operator guide should contain rice-rail fix in quick reference")
	}

	// Advisory checks
	if !strings.Contains(output, "Advisory Checks") {
		t.Error("operator guide should contain advisory checks section")
	}
	if !strings.Contains(output, "complexity") {
		t.Error("operator guide should list advisory check 'complexity'")
	}

	// Scope limits
	if !strings.Contains(output, "Max files per cycle: 10") {
		t.Error("operator guide should contain max files per cycle")
	}
}

func TestRenderRuleCatalog(t *testing.T) {
	c := testConstitution()
	c.Quality.AdvisoryOn = []string{"complexity", "coverage"}

	output := RenderRuleCatalog(c, nil)

	// Blocking rules table
	if !strings.Contains(output, "Blocking Rules") {
		t.Error("rule catalog should contain blocking rules section")
	}
	if !strings.Contains(output, "| lint | BLOCKING | Active |") {
		t.Error("rule catalog should contain lint as blocking rule")
	}
	if !strings.Contains(output, "| tests | BLOCKING | Active |") {
		t.Error("rule catalog should contain tests as blocking rule")
	}

	// Advisory rules table
	if !strings.Contains(output, "Advisory Rules") {
		t.Error("rule catalog should contain advisory rules section")
	}
	if !strings.Contains(output, "| complexity | WARNING | Active |") {
		t.Error("rule catalog should contain complexity as advisory rule")
	}
	if !strings.Contains(output, "| coverage | WARNING | Active |") {
		t.Error("rule catalog should contain coverage as advisory rule")
	}
}

func TestRenderRuleCatalogWithArchitecture(t *testing.T) {
	c := testConstitution()
	c.Architecture.Layering = constitution.LayeringSpec{
		Enabled: true,
		Layers:  []string{"domain", "application", "infrastructure"},
		ForbiddenDependencies: []constitution.ForbiddenDependency{
			{From: "domain", To: "infrastructure"},
			{From: "application", To: "infrastructure"},
		},
	}
	c.Architecture.DependencyPolicy.NoCycles = true
	c.Architecture.DependencyPolicy.RestrictCrossContextImports = true

	output := RenderRuleCatalog(c, nil)

	if !strings.Contains(output, "Architecture Rules") {
		t.Error("rule catalog should contain architecture rules section when layering enabled")
	}
	if !strings.Contains(output, "No import cycles") {
		t.Error("rule catalog should contain no-cycles rule")
	}
	if !strings.Contains(output, "Restrict cross-context imports") {
		t.Error("rule catalog should contain cross-context imports rule")
	}
	if !strings.Contains(output, "No domain → infrastructure") {
		t.Error("rule catalog should contain forbidden dependency rule")
	}
	if !strings.Contains(output, "No application → infrastructure") {
		t.Error("rule catalog should contain second forbidden dependency rule")
	}
}

func TestRenderRuleCatalogWithRolloutPlan(t *testing.T) {
	c := testConstitution()
	plan := &resolution.RolloutPlan{
		Steps: []resolution.RolloutStep{
			{
				Order:       1,
				Capability:  "formatting",
				Action:      "enable-tool",
				Risk:        "low",
				Description: "Enable gofmt for Go formatting",
			},
			{
				Order:       2,
				Capability:  "linting",
				Action:      "configure-rules",
				Risk:        "medium",
				Description: "Configure golangci-lint with project rules",
			},
		},
	}

	output := RenderRuleCatalog(c, plan)

	if !strings.Contains(output, "Pending Rollout") {
		t.Error("rule catalog should contain pending rollout section when plan has steps")
	}
	if !strings.Contains(output, "formatting") {
		t.Error("rule catalog should contain rollout step capability")
	}
	if !strings.Contains(output, "enable-tool") {
		t.Error("rule catalog should contain rollout step action")
	}
	if !strings.Contains(output, "low") {
		t.Error("rule catalog should contain rollout step risk")
	}
	if !strings.Contains(output, "linting") {
		t.Error("rule catalog should contain second rollout step")
	}
}

func TestRenderRuleCatalogWithoutRolloutPlan(t *testing.T) {
	c := testConstitution()

	output := RenderRuleCatalog(c, nil)

	if strings.Contains(output, "Pending Rollout") {
		t.Error("rule catalog should not contain pending rollout section when plan is nil")
	}
}

func TestRenderRuleCatalogEmptyRolloutPlan(t *testing.T) {
	c := testConstitution()
	plan := &resolution.RolloutPlan{}

	output := RenderRuleCatalog(c, plan)

	if strings.Contains(output, "Pending Rollout") {
		t.Error("rule catalog should not contain pending rollout section when plan has no steps")
	}
}
