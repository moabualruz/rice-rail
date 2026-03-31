package profiling

import (
	"os"
	"path/filepath"
	"testing"
)

func containsItem(items []DetectedItem, name string) bool {
	for _, it := range items {
		if it.Name == name {
			return true
		}
	}
	return false
}

func containsTool(tools []DetectedTool, name string) bool {
	for _, t := range tools {
		if t.Name == name {
			return true
		}
	}
	return false
}

func containsHint(hints []ArchHint, pattern string) bool {
	for _, h := range hints {
		if h.Pattern == pattern {
			return true
		}
	}
	return false
}

func TestScanGoFixture(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n\ngo 1.22"), 0644)
	os.WriteFile(filepath.Join(dir, ".golangci.yml"), []byte("linters:\n  enable:\n    - errcheck"), 0644)
	os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:\n\tgo build ./..."), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}"), 0644)

	os.MkdirAll(filepath.Join(dir, "cmd", "app"), 0755)
	os.WriteFile(filepath.Join(dir, "cmd", "app", "main.go"), []byte("package main"), 0644)

	os.MkdirAll(filepath.Join(dir, "internal", "core"), 0755)
	os.WriteFile(filepath.Join(dir, "internal", "core", "core.go"), []byte("package core"), 0644)

	scanner := NewScanner(dir)
	profile, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if !containsItem(profile.Languages, "go") {
		t.Error("expected languages to contain 'go'")
	}
	if !containsItem(profile.PackageManagers, "go modules") {
		t.Error("expected package_managers to contain 'go modules'")
	}
	if !containsItem(profile.BuildSystems, "make") {
		t.Error("expected build_systems to contain 'make'")
	}
	if profile.RepoTopology != "single" {
		t.Errorf("expected topology 'single', got %q", profile.RepoTopology)
	}
	if !containsHint(profile.ArchHints, "cmd/") {
		t.Error("expected arch hints to include 'cmd/'")
	}
	if !containsHint(profile.ArchHints, "internal/") {
		t.Error("expected arch hints to include 'internal/'")
	}
}

func TestScanTSMonorepo(t *testing.T) {
	dir := t.TempDir()

	pkgJSON := `{"name": "root", "private": true, "workspaces": ["packages/*"]}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644)
	os.WriteFile(filepath.Join(dir, "turbo.json"), []byte(`{"pipeline": {}}`), 0644)
	os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte(`{"compilerOptions": {}}`), 0644)
	os.WriteFile(filepath.Join(dir, "biome.json"), []byte(`{"linter": {}}`), 0644)

	os.MkdirAll(filepath.Join(dir, "packages", "a"), 0755)
	os.WriteFile(filepath.Join(dir, "packages", "a", "package.json"), []byte(`{"name": "a"}`), 0644)
	os.WriteFile(filepath.Join(dir, "packages", "a", "index.ts"), []byte("export const a = 1;"), 0644)

	os.MkdirAll(filepath.Join(dir, "packages", "b"), 0755)
	os.WriteFile(filepath.Join(dir, "packages", "b", "package.json"), []byte(`{"name": "b"}`), 0644)
	os.WriteFile(filepath.Join(dir, "packages", "b", "index.ts"), []byte("export const b = 2;"), 0644)

	scanner := NewScanner(dir)
	profile, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if profile.RepoTopology != "monorepo" {
		t.Errorf("expected topology 'monorepo', got %q", profile.RepoTopology)
	}
	if !containsTool(profile.Tooling.Linters, "biome") {
		t.Error("expected tooling to contain 'biome'")
	}
}

func TestScanPythonProject(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[build-system]\nrequires = [\"setuptools\"]"), 0644)
	os.WriteFile(filepath.Join(dir, "ruff.toml"), []byte("[lint]\nselect = [\"E\", \"F\"]"), 0644)
	os.WriteFile(filepath.Join(dir, "conftest.py"), []byte("import pytest"), 0644)

	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(filepath.Join(dir, "src", "app.py"), []byte("def main(): pass"), 0644)

	scanner := NewScanner(dir)
	profile, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if !containsItem(profile.Languages, "python") {
		t.Error("expected languages to contain 'python'")
	}

	hasRuff := containsTool(profile.Tooling.Linters, "ruff")
	if !hasRuff {
		t.Error("expected tooling to contain 'ruff'")
	}

	hasPytest := containsTool(profile.Tooling.TestRunners, "pytest")
	if !hasPytest {
		t.Error("expected tooling to contain 'pytest'")
	}
}

func TestScanRustProject(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"test\"\nversion = \"0.1.0\""), 0644)
	os.WriteFile(filepath.Join(dir, "clippy.toml"), []byte("cognitive-complexity-threshold = 30"), 0644)

	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(filepath.Join(dir, "src", "main.rs"), []byte("fn main() {}"), 0644)

	scanner := NewScanner(dir)
	profile, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if !containsItem(profile.Languages, "rust") {
		t.Error("expected languages to contain 'rust'")
	}
	if !containsTool(profile.Tooling.Linters, "clippy") {
		t.Error("expected tooling to contain 'clippy'")
	}
}

func TestScanEmptyRepo(t *testing.T) {
	dir := t.TempDir()

	scanner := NewScanner(dir)
	profile, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan should not crash on empty repo: %v", err)
	}

	if len(profile.Languages) != 0 {
		t.Errorf("expected no languages, got %d", len(profile.Languages))
	}
	if len(profile.PackageManagers) != 0 {
		t.Errorf("expected no package managers, got %d", len(profile.PackageManagers))
	}
	if len(profile.BuildSystems) != 0 {
		t.Errorf("expected no build systems, got %d", len(profile.BuildSystems))
	}
	if profile.RepoTopology != "single" {
		t.Errorf("expected topology 'single', got %q", profile.RepoTopology)
	}
}

func TestScanMultiLanguage(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test"}`), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "index.ts"), []byte("export default {}"), 0644)

	scanner := NewScanner(dir)
	profile, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	hasGo := containsItem(profile.Languages, "go")
	hasTS := containsItem(profile.Languages, "typescript")

	if !hasGo {
		t.Error("expected languages to contain 'go'")
	}
	if !hasTS {
		t.Error("expected languages to contain 'typescript'")
	}
}
