package interview

// QuestionCatalog returns all interview questions grouped by category.
// Questions are evaluated dynamically — skip/require conditions are checked at runtime.
func QuestionCatalog() []Question {
	return []Question{
		// --- Repo Strategy ---
		{
			ID:    "repo_purpose",
			Group: "repo_strategy",
			Text:  "What is the primary purpose of this repository?",
			Type:  "choice",
			Options: []Option{
				{Value: "application", Label: "Application", Description: "Deployable app or service"},
				{Value: "library", Label: "Library", Description: "Reusable package/module"},
				{Value: "cli", Label: "CLI tool", Description: "Command-line utility"},
				{Value: "monorepo", Label: "Monorepo", Description: "Multiple projects in one repo"},
				{Value: "infrastructure", Label: "Infrastructure", Description: "IaC, configs, deployment"},
				{Value: "other", Label: "Other"},
			},
		},

		// --- Architecture Doctrine ---
		{
			ID:    "arch_target",
			Group: "architecture",
			Text:  "What architecture style are you targeting?",
			Type:  "choice",
			Options: []Option{
				{Value: "layered_monolith", Label: "Layered monolith", Description: "Traditional layers (handler/service/repo)"},
				{Value: "hexagonal", Label: "Hexagonal / Ports & Adapters"},
				{Value: "clean", Label: "Clean Architecture", Description: "Use cases, entities, adapters"},
				{Value: "modular_monolith", Label: "Modular monolith", Description: "Independent modules, shared deploy"},
				{Value: "microservices", Label: "Microservices"},
				{Value: "simple", Label: "Simple / flat", Description: "No formal architecture"},
				{Value: "custom", Label: "Custom"},
			},
		},
		{
			ID:          "enforce_layers",
			Group:       "architecture",
			Text:        "Should dependency direction between layers be enforced?",
			Type:        "confirm",
			Default:     "true",
			RequireWhen: []string{"arch_target=layered_monolith", "arch_target=hexagonal", "arch_target=clean", "arch_target=modular_monolith"},
		},
		{
			ID:          "no_cycles",
			Group:       "architecture",
			Text:        "Should import cycles between modules be forbidden?",
			Type:        "confirm",
			Default:     "true",
			RequireWhen: []string{"arch_target=modular_monolith", "arch_target=microservices"},
		},

		// --- Quality & Safety ---
		{
			ID:    "safety_mode",
			Group: "quality",
			Text:  "What safety level for automated changes?",
			Type:  "choice",
			Options: []Option{
				{Value: "strict", Label: "Strict", Description: "Only safe, verified changes. Block on all failures."},
				{Value: "balanced", Label: "Balanced", Description: "Safe changes auto-applied, some warnings advisory."},
				{Value: "aggressive", Label: "Aggressive", Description: "Apply more changes, tolerate some risk."},
			},
			Default: "balanced",
		},
		{
			ID:    "block_on",
			Group: "quality",
			Text:  "Which checks should BLOCK (fail the pipeline)?",
			Type:  "multi_choice",
			Options: []Option{
				{Value: "lint", Label: "Linting"},
				{Value: "tests", Label: "Tests"},
				{Value: "typecheck", Label: "Type checking"},
				{Value: "forbidden_imports", Label: "Forbidden imports"},
				{Value: "architecture", Label: "Architecture rules"},
				{Value: "security", Label: "Security checks"},
			},
			Default: "lint,tests,typecheck",
		},
		{
			ID:    "advisory_on",
			Group: "quality",
			Text:  "Which checks should be ADVISORY (warn but don't block)?",
			Type:  "multi_choice",
			Options: []Option{
				{Value: "complexity", Label: "Complexity"},
				{Value: "duplication", Label: "Duplication"},
				{Value: "long_functions", Label: "Long functions"},
				{Value: "naming", Label: "Naming conventions"},
				{Value: "coverage", Label: "Test coverage"},
			},
			Default: "complexity,duplication,long_functions",
		},
		{
			ID:      "require_tests",
			Group:   "quality",
			Text:    "Require tests for all changed code?",
			Type:    "confirm",
			Default: "false",
		},

		// --- Automation ---
		{
			ID:      "allow_safe_autofix",
			Group:   "automation",
			Text:    "Allow safe autofixes (formatting, import ordering, etc.)?",
			Type:    "confirm",
			Default: "true",
		},
		{
			ID:      "allow_unsafe_autofix",
			Group:   "automation",
			Text:    "Allow unsafe autofixes (require explicit flag)?",
			Type:    "confirm",
			Default: "false",
		},
		{
			ID:      "allow_codemods",
			Group:   "automation",
			Text:    "Allow generated codemods for project-specific transforms?",
			Type:    "confirm",
			Default: "true",
		},
		{
			ID:      "separate_baseline",
			Group:   "automation",
			Text:    "Keep baseline remediation separate from feature work?",
			Type:    "confirm",
			Default: "true",
		},

		// --- Legacy & Migration ---
		{
			ID:    "baseline_appetite",
			Group: "legacy",
			Text:  "How should existing code be handled?",
			Type:  "choice",
			Options: []Option{
				{Value: "now", Label: "Remediate now", Description: "Run baseline before feature work"},
				{Value: "incremental", Label: "Incremental", Description: "Fix code as you touch it"},
				{Value: "later", Label: "Later", Description: "Focus on new code only"},
				{Value: "never", Label: "Never", Description: "Existing code is exempt"},
			},
			Default: "incremental",
		},

		// --- Tooling Preferences ---
		{
			ID:      "generate_ci",
			Group:   "workflow",
			Text:    "Generate CI integration for checks?",
			Type:    "confirm",
			Default: "true",
		},
		{
			ID:      "generate_wrappers",
			Group:   "workflow",
			Text:    "Generate local wrapper commands (bin/pstk-*)?",
			Type:    "confirm",
			Default: "true",
		},
		{
			ID:      "constitution_changes_ack",
			Group:   "workflow",
			Text:    "Require human acknowledgment for constitution changes?",
			Type:    "confirm",
			Default: "true",
		},

		// --- MCP & Connectivity ---
		{
			ID:      "mcp_enabled",
			Group:   "connectivity",
			Text:    "Enable optional MCP integrations?",
			Type:    "confirm",
			Default: "false",
		},
		{
			ID:      "offline_important",
			Group:   "connectivity",
			Text:    "Is offline/local-only operation important?",
			Type:    "confirm",
			Default: "true",
		},

		// --- Scope Limits ---
		{
			ID:      "max_files_per_cycle",
			Group:   "scope",
			Text:    "Max files changed per cycle (0 = unlimited)?",
			Type:    "number",
			Default: "20",
		},
		{
			ID:      "max_lines_per_cycle",
			Group:   "scope",
			Text:    "Max lines changed per cycle (0 = unlimited)?",
			Type:    "number",
			Default: "500",
		},
	}
}
