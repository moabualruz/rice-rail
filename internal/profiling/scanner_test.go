package profiling

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create Go files
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "util.go"), []byte("package main"), 0644)

	// Create manifest
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)

	// Create build system
	os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:"), 0644)

	// Create tool config
	os.WriteFile(filepath.Join(dir, ".golangci.yml"), []byte("linters:"), 0644)

	// Create architecture dirs
	os.MkdirAll(filepath.Join(dir, "cmd"), 0755)
	os.MkdirAll(filepath.Join(dir, "internal"), 0755)
	os.MkdirAll(filepath.Join(dir, "pkg"), 0755)

	return dir
}

func TestScanLanguages(t *testing.T) {
	dir := setupTestRepo(t)
	scanner := NewScanner(dir)
	profile, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if len(profile.Languages) == 0 {
		t.Fatal("expected languages to be detected")
	}
	if profile.Languages[0].Name != "go" {
		t.Errorf("expected go, got %s", profile.Languages[0].Name)
	}
}

func TestScanPackageManagers(t *testing.T) {
	dir := setupTestRepo(t)
	scanner := NewScanner(dir)
	profile, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	found := false
	for _, pm := range profile.PackageManagers {
		if pm.Name == "go modules" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected go modules to be detected")
	}
}

func TestScanBuildSystems(t *testing.T) {
	dir := setupTestRepo(t)
	scanner := NewScanner(dir)
	profile, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	found := false
	for _, bs := range profile.BuildSystems {
		if bs.Name == "make" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected make to be detected")
	}
}

func TestScanTooling(t *testing.T) {
	dir := setupTestRepo(t)
	scanner := NewScanner(dir)
	profile, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	found := false
	for _, l := range profile.Tooling.Linters {
		if l.Name == "golangci-lint" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected golangci-lint to be detected from .golangci.yml")
	}
}

func TestScanArchHints(t *testing.T) {
	dir := setupTestRepo(t)
	scanner := NewScanner(dir)
	profile, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if len(profile.ArchHints) == 0 {
		t.Fatal("expected architecture hints from cmd/, internal/, pkg/")
	}

	hints := map[string]bool{}
	for _, h := range profile.ArchHints {
		hints[h.Pattern] = true
	}

	for _, expected := range []string{"cmd/", "internal/", "pkg/"} {
		if !hints[expected] {
			t.Errorf("expected hint for %s", expected)
		}
	}
}

func TestScanTopologySingle(t *testing.T) {
	dir := setupTestRepo(t)
	scanner := NewScanner(dir)
	profile, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if profile.RepoTopology != "single" {
		t.Errorf("expected single, got %s", profile.RepoTopology)
	}
}

func TestScanTopologyMonorepo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "nx.json"), []byte("{}"), 0644)

	scanner := NewScanner(dir)
	profile, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if profile.RepoTopology != "monorepo" {
		t.Errorf("expected monorepo, got %s", profile.RepoTopology)
	}
}

func TestScanSkipsNodeModules(t *testing.T) {
	dir := t.TempDir()
	nmDir := filepath.Join(dir, "node_modules", "pkg")
	os.MkdirAll(nmDir, 0755)
	os.WriteFile(filepath.Join(nmDir, "index.js"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "app.go"), []byte("package main"), 0644)

	scanner := NewScanner(dir)
	profile, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	for _, lang := range profile.Languages {
		if lang.Name == "javascript" {
			t.Error("should not detect JS files inside node_modules")
		}
	}
}
