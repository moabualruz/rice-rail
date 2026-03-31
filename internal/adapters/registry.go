package adapters

import (
	"os"
	"path/filepath"

	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/exec"
)

// Registry holds all available adapter instances for a project.
type Registry struct {
	RuleEngines  []RuleEngineAdapter
	Fixers       []RuleEngineAdapter // adapters that support Fix()
	Codemods     []CodemodEngineAdapter
	Typecheckers []TypecheckAdapter
	TestRunners  []TestRunnerAdapter
	Agents       []AgentAdapter
}

// DiscoverAdapters probes the system for available tools and returns
// adapters for each tool found on PATH with a config in the repo.
// DiscoverAdaptersWithCustom is like DiscoverAdapters but also registers custom tools from the constitution.
func DiscoverAdaptersWithCustom(repoRoot string, languages []string, customTools []constitution.CustomTool) *Registry {
	r := DiscoverAdapters(repoRoot, languages)
	runner := exec.NewRunner()

	for _, ct := range customTools {
		if _, ok := exec.Which(ct.Binary); !ok {
			continue // tool not installed, skip
		}
		adapter := NewCustomAdapter(runner, ct)

		switch ct.Role {
		case "linter", "rule_engine":
			r.RuleEngines = append(r.RuleEngines, adapter)
			if len(ct.FixCmd) > 0 {
				r.Fixers = append(r.Fixers, adapter)
			}
		case "formatter":
			r.Fixers = append(r.Fixers, adapter)
			r.RuleEngines = append(r.RuleEngines, adapter) // check mode reports unformatted files
		case "typechecker":
			r.Typecheckers = append(r.Typecheckers, &CustomTypecheckAdapter{inner: adapter})
		case "test_runner":
			r.TestRunners = append(r.TestRunners, &CustomTestRunnerAdapter{inner: adapter})
		case "codemod":
			// Custom codemods only run via check/fix, not the codemod engine interface
			if len(ct.FixCmd) > 0 {
				r.Fixers = append(r.Fixers, adapter)
			}
		default:
			// Unknown role — add as rule engine if it has a check command
			if len(ct.CheckCmd) > 0 {
				r.RuleEngines = append(r.RuleEngines, adapter)
			}
		}
	}

	return r
}

func DiscoverAdapters(repoRoot string, languages []string) *Registry {
	r := &Registry{}
	runner := exec.NewRunner()

	langSet := map[string]bool{}
	for _, l := range languages {
		langSet[l] = true
	}

	// === Go ecosystem ===
	if langSet["go"] {
		if _, ok := exec.Which("golangci-lint"); ok {
			adapter := NewGolangciLintAdapter(runner, repoRoot)
			r.RuleEngines = append(r.RuleEngines, adapter)
			r.Fixers = append(r.Fixers, adapter)
		}
		if _, ok := exec.Which("gofmt"); ok {
			r.Fixers = append(r.Fixers, NewGofmtAdapter(runner, repoRoot))
		}
		if _, ok := exec.Which("go"); ok {
			r.RuleEngines = append(r.RuleEngines, NewGoVetAdapter(runner, repoRoot))
			r.TestRunners = append(r.TestRunners, NewGoTestAdapter(runner, repoRoot))
		}
	}

	// === TypeScript / JavaScript ecosystem ===
	if langSet["typescript"] || langSet["javascript"] {
		// Typechecker
		if _, ok := exec.Which("tsc"); ok {
			r.Typecheckers = append(r.Typecheckers, NewTscAdapter(runner, repoRoot))
		}

		// Linters — prefer biome if configured, else eslint
		if hasConfig(repoRoot, "biome.json", "biome.jsonc") {
			adapter := NewBiomeAdapter(runner, repoRoot)
			r.RuleEngines = append(r.RuleEngines, adapter)
			r.Fixers = append(r.Fixers, adapter)
		} else if _, ok := exec.Which("npx"); ok {
			if hasConfig(repoRoot, ".eslintrc", ".eslintrc.js", ".eslintrc.json", ".eslintrc.yml", ".eslintrc.cjs", "eslint.config.js", "eslint.config.mjs", "eslint.config.ts") {
				adapter := NewESLintAdapter(runner, repoRoot)
				r.RuleEngines = append(r.RuleEngines, adapter)
				r.Fixers = append(r.Fixers, adapter)
			}
		}

		// Formatter — prettier (skip if biome handles it)
		if !hasConfig(repoRoot, "biome.json", "biome.jsonc") {
			if hasConfig(repoRoot, ".prettierrc", ".prettierrc.js", ".prettierrc.json", ".prettierrc.yml", ".prettierrc.toml", "prettier.config.js", "prettier.config.mjs") {
				adapter := NewPrettierAdapter(runner, repoRoot)
				r.Fixers = append(r.Fixers, adapter)
			}
		}

		// Test runners
		switch DetectJSTestRunner(repoRoot) {
		case "vitest":
			r.TestRunners = append(r.TestRunners, NewVitestAdapter(runner, repoRoot))
		case "jest":
			r.TestRunners = append(r.TestRunners, NewJestAdapter(runner, repoRoot))
		}

		// Architecture — dependency-cruiser
		if _, ok := exec.Which("depcruise"); ok {
			r.RuleEngines = append(r.RuleEngines, NewDepCruiserAdapter(runner, repoRoot))
		}
	}

	// === Python ecosystem ===
	if langSet["python"] {
		// Linter + formatter
		if _, ok := exec.Which("ruff"); ok {
			adapter := NewRuffAdapter(runner, repoRoot)
			r.RuleEngines = append(r.RuleEngines, adapter)
			r.Fixers = append(r.Fixers, adapter)
		}

		// Typechecker — prefer pyright, fall back to mypy
		if _, ok := exec.Which("pyright"); ok {
			r.Typecheckers = append(r.Typecheckers, NewPyrightAdapter(runner, repoRoot))
		} else if _, ok := exec.Which("mypy"); ok {
			r.Typecheckers = append(r.Typecheckers, NewMypyAdapter(runner, repoRoot))
		}

		// Test runner
		if _, ok := exec.Which("pytest"); ok {
			r.TestRunners = append(r.TestRunners, NewPytestAdapter(runner, repoRoot))
		}
	}

	// === Rust ecosystem ===
	if langSet["rust"] {
		if _, ok := exec.Which("cargo"); ok {
			// Linter
			clippy := NewClippyAdapter(runner, repoRoot)
			r.RuleEngines = append(r.RuleEngines, clippy)
			r.Fixers = append(r.Fixers, clippy)

			// Formatter
			rustfmt := NewRustfmtAdapter(runner, repoRoot)
			r.Fixers = append(r.Fixers, rustfmt)

			// Tests
			r.TestRunners = append(r.TestRunners, NewCargoTestAdapter(runner, repoRoot))
		}
	}

	// === Cross-language tools ===
	if _, ok := exec.Which("semgrep"); ok {
		r.RuleEngines = append(r.RuleEngines, NewSemgrepAdapter(runner, repoRoot))
	}
	if _, ok := exec.Which("ast-grep"); ok {
		adapter := NewAstGrepAdapter(runner, repoRoot)
		r.RuleEngines = append(r.RuleEngines, adapter)
		r.Codemods = append(r.Codemods, adapter)
	}
	if _, ok := exec.Which("comby"); ok {
		r.Codemods = append(r.Codemods, NewCombyAdapter(runner, repoRoot))
	}

	// === Agent adapters ===
	if _, ok := exec.Which("claude"); ok {
		r.Agents = append(r.Agents, NewClaudeCodeAdapter(runner, repoRoot))
	}
	if _, ok := exec.Which("aider"); ok {
		r.Agents = append(r.Agents, NewAiderAdapter(runner, repoRoot))
	}
	if _, ok := exec.Which("codex"); ok {
		r.Agents = append(r.Agents, NewCodexAdapter(runner, repoRoot))
	}
	if _, ok := exec.Which("gemini"); ok {
		r.Agents = append(r.Agents, NewGeminiAdapter(runner, repoRoot))
	}
	if _, ok := exec.Which("opencode"); ok {
		r.Agents = append(r.Agents, NewOpenCodeAdapter(runner, repoRoot))
	}
	if _, ok := exec.Which("copilot"); ok {
		r.Agents = append(r.Agents, NewCopilotAdapter(runner, repoRoot))
	}
	if _, ok := exec.Which("qwen"); ok {
		r.Agents = append(r.Agents, NewQwenAdapter(runner, repoRoot))
	}
	if _, ok := exec.Which("ollama"); ok {
		r.Agents = append(r.Agents, NewOllamaAdapter(runner, repoRoot))
	}

	return r
}

// hasConfig checks if any of the given config files exist in the repo root.
func hasConfig(repoRoot string, names ...string) bool {
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(repoRoot, name)); err == nil {
			return true
		}
	}
	return false
}
