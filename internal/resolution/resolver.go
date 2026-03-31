package resolution

import (
	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/discovery"
)

// Status constants for capability classification.
const (
	PresentReady        = "PRESENT_READY"
	PresentNeedsConfig  = "PRESENT_NEEDS_CONFIG"
	PresentNeedsWrapper = "PRESENT_NEEDS_WRAPPER"
	NeedsProjectRule    = "NEEDS_PROJECT_RULE"
	NeedsProjectCodemod = "NEEDS_PROJECT_CODEMOD"
	NeedsNewTool        = "NEEDS_NEW_TOOL"
	AdvisoryOnly        = "ADVISORY_ONLY"
	Deferred            = "DEFERRED"
)

// GapReport describes what's missing between desired and current state.
type GapReport struct {
	Gaps []Gap `yaml:"gaps" json:"gaps"`
}

// Gap is a single capability gap.
type Gap struct {
	Capability string   `yaml:"capability" json:"capability"`
	Category   string   `yaml:"category" json:"category"`
	Status     string   `yaml:"status" json:"status"`
	Reasoning  string   `yaml:"reasoning" json:"reasoning"`
	Candidates []string `yaml:"candidates,omitempty" json:"candidates,omitempty"`
	Strategy   string   `yaml:"strategy" json:"strategy"`
	Priority   int      `yaml:"priority" json:"priority"`
	Risk       string   `yaml:"risk" json:"risk"`
}

// RolloutPlan is the ordered list of actions to close gaps.
type RolloutPlan struct {
	Steps []RolloutStep `yaml:"steps" json:"steps"`
}

// RolloutStep is a single action in the rollout plan.
type RolloutStep struct {
	Order       int    `yaml:"order" json:"order"`
	Capability  string `yaml:"capability" json:"capability"`
	Action      string `yaml:"action" json:"action"`
	Tool        string `yaml:"tool,omitempty" json:"tool,omitempty"`
	Risk        string `yaml:"risk" json:"risk"`
	Description string `yaml:"description" json:"description"`
}

// DesiredCapability defines what the constitution expects.
type DesiredCapability struct {
	Name       string
	Category   string
	Priority   int
	RequiredBy []string // constitution fields that require this
}

// Resolve compares the constitution's requirements against the tool inventory
// and produces a gap report and rollout plan.
func Resolve(c *constitution.Constitution, inv *discovery.ToolInventory) (*GapReport, *RolloutPlan) {
	desired := deriveDesiredCapabilities(c)
	toolMap := buildToolMap(inv)

	report := &GapReport{}
	plan := &RolloutPlan{}
	order := 1

	for _, cap := range desired {
		gap := classifyGap(cap, toolMap)
		report.Gaps = append(report.Gaps, gap)

		if gap.Status != PresentReady {
			plan.Steps = append(plan.Steps, RolloutStep{
				Order:       order,
				Capability:  gap.Capability,
				Action:      gap.Strategy,
				Tool:        firstCandidate(gap.Candidates),
				Risk:        gap.Risk,
				Description: gap.Reasoning,
			})
			order++
		}
	}

	return report, plan
}

func deriveDesiredCapabilities(c *constitution.Constitution) []DesiredCapability {
	var caps []DesiredCapability

	// Always need formatting/canonicalization
	caps = append(caps, DesiredCapability{
		Name: "formatting", Category: "canonicalization", Priority: 1,
		RequiredBy: []string{"automation.allow_safe_autofix"},
	})

	// Blocking checks become required capabilities
	for _, check := range c.Quality.BlockOn {
		p := 1
		cat := check
		switch check {
		case "lint":
			caps = append(caps, DesiredCapability{Name: "linting", Category: "policy", Priority: p})
		case "tests":
			caps = append(caps, DesiredCapability{Name: "test_execution", Category: "testing", Priority: p})
		case "typecheck":
			caps = append(caps, DesiredCapability{Name: "type_checking", Category: "verification", Priority: p})
		case "forbidden_imports":
			caps = append(caps, DesiredCapability{Name: "import_rules", Category: "architecture", Priority: p})
		case "architecture":
			caps = append(caps, DesiredCapability{Name: "architecture_rules", Category: "architecture", Priority: p})
		case "security":
			caps = append(caps, DesiredCapability{Name: "security_scanning", Category: "security", Priority: 2})
		default:
			caps = append(caps, DesiredCapability{Name: cat, Category: "custom", Priority: 2})
		}
	}

	// Architecture enforcement
	if c.Architecture.Layering.Enabled {
		caps = append(caps, DesiredCapability{
			Name: "layer_enforcement", Category: "architecture", Priority: 1,
		})
	}
	if c.Architecture.DependencyPolicy.NoCycles {
		caps = append(caps, DesiredCapability{
			Name: "cycle_detection", Category: "architecture", Priority: 2,
		})
	}

	// Codemod support
	if c.Automation.AllowGeneratedCodemods {
		caps = append(caps, DesiredCapability{
			Name: "structural_rewrite", Category: "codemod", Priority: 3,
		})
	}

	return caps
}

func buildToolMap(inv *discovery.ToolInventory) map[string][]discovery.InventoryEntry {
	m := map[string][]discovery.InventoryEntry{}
	for _, t := range inv.Tools {
		m[t.Category] = append(m[t.Category], t)
	}
	return m
}

func classifyGap(cap DesiredCapability, toolMap map[string][]discovery.InventoryEntry) Gap {
	gap := Gap{
		Capability: cap.Name,
		Category:   cap.Category,
		Priority:   cap.Priority,
	}

	// Map capability categories to tool categories
	toolCat := mapCapToToolCategory(cap.Category)
	tools, found := toolMap[toolCat]

	if !found || len(tools) == 0 {
		gap.Status = NeedsNewTool
		gap.Reasoning = "no tool found for " + cap.Name
		gap.Candidates = suggestTools(cap)
		gap.Strategy = "install and configure recommended tool"
		gap.Risk = "medium"
		return gap
	}

	// Check if any tool is fully ready
	for _, t := range tools {
		if t.InstalledLocally && t.ConfiguredInRepo {
			gap.Status = PresentReady
			gap.Reasoning = t.Name + " is installed and configured"
			gap.Risk = "low"
			return gap
		}
	}

	// Config exists but binary missing — needs install
	for _, t := range tools {
		if t.ConfiguredInRepo && !t.InstalledLocally {
			gap.Status = PresentNeedsWrapper
			gap.Reasoning = t.Name + " is configured in repo but not installed locally"
			gap.Strategy = "install " + t.Name
			gap.Risk = "low"
			return gap
		}
	}

	// Binary exists but no config — needs config
	for _, t := range tools {
		if t.InstalledLocally && !t.ConfiguredInRepo {
			gap.Status = PresentNeedsConfig
			gap.Reasoning = t.Name + " is installed but has no config file"
			gap.Strategy = "generate config for " + t.Name
			gap.Risk = "low"
			return gap
		}
	}

	gap.Status = PresentNeedsWrapper
	gap.Reasoning = "tools found but need wrapper integration"
	gap.Strategy = "generate wrapper command"
	gap.Risk = "low"
	return gap
}

func mapCapToToolCategory(category string) string {
	switch category {
	case "canonicalization":
		return "formatter"
	case "policy":
		return "linter"
	case "verification":
		return "typechecker"
	case "testing":
		return "test_runner"
	case "architecture":
		return "rule_engine"
	case "security":
		return "security"
	case "codemod":
		return "codemod"
	default:
		return category
	}
}

func suggestTools(cap DesiredCapability) []string {
	switch cap.Name {
	case "formatting":
		return []string{"prettier", "gofmt", "rustfmt", "black", "ruff"}
	case "linting":
		return []string{"eslint", "golangci-lint", "ruff", "clippy"}
	case "type_checking":
		return []string{"typescript", "pyright", "mypy"}
	case "test_execution":
		return []string{"go test", "jest", "vitest", "pytest", "cargo test"}
	case "architecture_rules", "layer_enforcement", "import_rules":
		return []string{"dependency-cruiser", "semgrep", "archunit"}
	case "cycle_detection":
		return []string{"dependency-cruiser", "madge"}
	case "security_scanning":
		return []string{"semgrep", "trivy", "snyk"}
	case "structural_rewrite":
		return []string{"ast-grep", "comby", "openrewrite", "jscodeshift"}
	default:
		return nil
	}
}

func firstCandidate(candidates []string) string {
	if len(candidates) > 0 {
		return candidates[0]
	}
	return ""
}
