package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mkh/rice-railing/internal/constitution"
)

func TestDiscoverAdaptersEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	reg := DiscoverAdapters(tmpDir, nil)

	if reg == nil {
		t.Fatal("expected non-nil registry")
	}

	// With no languages, no language-specific tools should be found.
	// Only cross-language tools (semgrep, ast-grep, comby) and agents
	// may appear if installed. We verify that the registry is at least
	// structurally valid and language-specific slots are empty or only
	// contain cross-language entries.
	for _, re := range reg.RuleEngines {
		name := re.Name()
		// Should not contain language-specific tools
		switch name {
		case "golangci-lint", "go-vet", "ruff", "clippy", "eslint", "biome":
			t.Errorf("unexpected language-specific rule engine %q with empty languages", name)
		}
	}
}

func TestDiscoverAdaptersWithCustomLinter(t *testing.T) {
	tmpDir := t.TempDir()
	custom := []constitution.CustomTool{
		{
			Name:     "echo-linter",
			Binary:   "echo",
			Role:     "linter",
			CheckCmd: []string{"--check"},
		},
	}

	reg := DiscoverAdaptersWithCustom(tmpDir, nil, custom)

	found := false
	for _, re := range reg.RuleEngines {
		if re.Name() == "echo-linter" {
			found = true
			break
		}
	}
	if !found {
		t.Error("custom linter 'echo-linter' not found in RuleEngines")
	}
}

func TestDiscoverAdaptersWithCustomFormatter(t *testing.T) {
	tmpDir := t.TempDir()
	custom := []constitution.CustomTool{
		{
			Name:     "echo-formatter",
			Binary:   "echo",
			Role:     "formatter",
			CheckCmd: []string{"--check"},
			FixCmd:   []string{"--fix"},
		},
	}

	reg := DiscoverAdaptersWithCustom(tmpDir, nil, custom)

	foundInFixers := false
	for _, f := range reg.Fixers {
		if f.Name() == "echo-formatter" {
			foundInFixers = true
			break
		}
	}
	if !foundInFixers {
		t.Error("custom formatter 'echo-formatter' not found in Fixers")
	}

	foundInRuleEngines := false
	for _, re := range reg.RuleEngines {
		if re.Name() == "echo-formatter" {
			foundInRuleEngines = true
			break
		}
	}
	if !foundInRuleEngines {
		t.Error("custom formatter 'echo-formatter' not found in RuleEngines (check mode)")
	}
}

func TestDiscoverAdaptersWithCustomTestRunner(t *testing.T) {
	tmpDir := t.TempDir()
	custom := []constitution.CustomTool{
		{
			Name:    "echo-tester",
			Binary:  "echo",
			Role:    "test_runner",
			TestCmd: []string{"--test"},
		},
	}

	reg := DiscoverAdaptersWithCustom(tmpDir, nil, custom)

	found := false
	for _, tr := range reg.TestRunners {
		if tr.Name() == "echo-tester" {
			found = true
			break
		}
	}
	if !found {
		t.Error("custom test runner 'echo-tester' not found in TestRunners")
	}
}

func TestDiscoverAdaptersWithCustomTypechecker(t *testing.T) {
	tmpDir := t.TempDir()
	custom := []constitution.CustomTool{
		{
			Name:     "echo-typecheck",
			Binary:   "echo",
			Role:     "typechecker",
			CheckCmd: []string{"--check"},
		},
	}

	reg := DiscoverAdaptersWithCustom(tmpDir, nil, custom)

	found := false
	for _, tc := range reg.Typecheckers {
		if tc.Name() == "echo-typecheck" {
			found = true
			break
		}
	}
	if !found {
		t.Error("custom typechecker 'echo-typecheck' not found in Typecheckers")
	}
}

func TestDiscoverAdaptersCustomMissingBinary(t *testing.T) {
	tmpDir := t.TempDir()
	custom := []constitution.CustomTool{
		{
			Name:     "ghost-tool",
			Binary:   "nonexistent-tool-xyz-12345",
			Role:     "linter",
			CheckCmd: []string{"--check"},
		},
	}

	reg := DiscoverAdaptersWithCustom(tmpDir, nil, custom)

	for _, re := range reg.RuleEngines {
		if re.Name() == "ghost-tool" {
			t.Error("tool with missing binary 'nonexistent-tool-xyz-12345' should not be registered")
		}
	}
	for _, f := range reg.Fixers {
		if f.Name() == "ghost-tool" {
			t.Error("tool with missing binary should not appear in Fixers")
		}
	}
	for _, tr := range reg.TestRunners {
		if tr.Name() == "ghost-tool" {
			t.Error("tool with missing binary should not appear in TestRunners")
		}
	}
	for _, tc := range reg.Typecheckers {
		if tc.Name() == "ghost-tool" {
			t.Error("tool with missing binary should not appear in Typecheckers")
		}
	}
}

func TestDiscoverAdaptersCustomAllRoles(t *testing.T) {
	tmpDir := t.TempDir()
	custom := []constitution.CustomTool{
		{
			Name:     "all-linter",
			Binary:   "echo",
			Role:     "linter",
			CheckCmd: []string{"lint"},
			FixCmd:   []string{"fix"},
		},
		{
			Name:   "all-formatter",
			Binary: "echo",
			Role:   "formatter",
			FixCmd: []string{"fmt"},
		},
		{
			Name:    "all-tester",
			Binary:  "echo",
			Role:    "test_runner",
			TestCmd: []string{"test"},
		},
		{
			Name:     "all-typechecker",
			Binary:   "echo",
			Role:     "typechecker",
			CheckCmd: []string{"typecheck"},
		},
	}

	reg := DiscoverAdaptersWithCustom(tmpDir, nil, custom)

	// Linter should be in RuleEngines and Fixers (has FixCmd)
	assertInRuleEngines(t, reg, "all-linter")
	assertInFixers(t, reg, "all-linter")

	// Formatter should be in both Fixers and RuleEngines
	assertInFixers(t, reg, "all-formatter")
	assertInRuleEngines(t, reg, "all-formatter")

	// Test runner should be in TestRunners
	assertInTestRunners(t, reg, "all-tester")

	// Typechecker should be in Typecheckers
	assertInTypecheckers(t, reg, "all-typechecker")
}

func TestHasConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// No config files exist
	if hasConfig(tmpDir, "biome.json", ".eslintrc") {
		t.Error("expected false when no config files exist")
	}

	// Create one config file
	configPath := filepath.Join(tmpDir, "biome.json")
	if err := os.WriteFile(configPath, []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	if !hasConfig(tmpDir, "biome.json") {
		t.Error("expected true when biome.json exists")
	}

	if !hasConfig(tmpDir, ".eslintrc", "biome.json") {
		t.Error("expected true when one of multiple configs exists")
	}

	if hasConfig(tmpDir, ".eslintrc", ".prettierrc") {
		t.Error("expected false when none of the queried configs exist")
	}
}

// --- helpers ---

func assertInRuleEngines(t *testing.T, reg *Registry, name string) {
	t.Helper()
	for _, re := range reg.RuleEngines {
		if re.Name() == name {
			return
		}
	}
	t.Errorf("%q not found in RuleEngines", name)
}

func assertInFixers(t *testing.T, reg *Registry, name string) {
	t.Helper()
	for _, f := range reg.Fixers {
		if f.Name() == name {
			return
		}
	}
	t.Errorf("%q not found in Fixers", name)
}

func assertInTestRunners(t *testing.T, reg *Registry, name string) {
	t.Helper()
	for _, tr := range reg.TestRunners {
		if tr.Name() == name {
			return
		}
	}
	t.Errorf("%q not found in TestRunners", name)
}

func assertInTypecheckers(t *testing.T, reg *Registry, name string) {
	t.Helper()
	for _, tc := range reg.Typecheckers {
		if tc.Name() == name {
			return
		}
	}
	t.Errorf("%q not found in Typecheckers", name)
}
