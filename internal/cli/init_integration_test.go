//go:build integration

package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mkh/rice-railing/internal/config"
	"github.com/mkh/rice-railing/internal/constitution"
)

// fixtureDir returns the absolute path to a testdata fixture.
func fixtureDir(t *testing.T, name string) string {
	t.Helper()
	// Resolve relative to this test file's location via the repo root.
	dir, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", name))
	if err != nil {
		t.Fatalf("resolving fixture dir: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatalf("fixture dir does not exist: %s", dir)
	}
	return dir
}

// runInitNonInteractive changes to the fixture dir, runs the init flow
// programmatically via the cobra command, and returns to the original dir.
func runInitNonInteractive(t *testing.T, dir string) {
	t.Helper()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting cwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(orig)
	})

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to fixture: %v", err)
	}

	// Clean up any previous run
	toolkitDir := filepath.Join(dir, config.ToolkitDir)
	_ = os.RemoveAll(toolkitDir)
	t.Cleanup(func() {
		_ = os.RemoveAll(toolkitDir)
	})

	rootCmd.SetArgs([]string{"init", "--non-interactive"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rice-rail init --non-interactive failed: %v", err)
	}
}

// assertFileExists checks that a file exists at the given path.
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", path)
	}
}

// loadConstitution loads and returns the constitution from the toolkit dir.
func loadConstitution(t *testing.T, dir string) *constitution.Constitution {
	t.Helper()
	p := filepath.Join(dir, config.ToolkitDir, "constitution.yaml")
	c, err := constitution.Load(p)
	if err != nil {
		t.Fatalf("loading constitution: %v", err)
	}
	return c
}

// assertLanguageDetected checks that the constitution includes the expected language.
func assertLanguageDetected(t *testing.T, c *constitution.Constitution, lang string) {
	t.Helper()
	for _, l := range c.Project.Languages {
		if l == lang {
			return
		}
	}
	t.Errorf("expected language %q in constitution, got %v", lang, c.Project.Languages)
}

func TestInitGoAPI(t *testing.T) {
	dir := fixtureDir(t, "go-api")
	runInitNonInteractive(t, dir)

	p := config.DefaultPaths()
	assertFileExists(t, filepath.Join(dir, p.Constitution))
	assertFileExists(t, filepath.Join(dir, p.Profile))
	assertFileExists(t, filepath.Join(dir, p.ToolInventory))
	assertFileExists(t, filepath.Join(dir, p.GapReport))
	assertFileExists(t, filepath.Join(dir, p.RolloutPlan))

	c := loadConstitution(t, dir)
	assertLanguageDetected(t, c, "go")
}

func TestInitTSMonorepo(t *testing.T) {
	dir := fixtureDir(t, "ts-monorepo")
	runInitNonInteractive(t, dir)

	p := config.DefaultPaths()
	assertFileExists(t, filepath.Join(dir, p.Constitution))
	assertFileExists(t, filepath.Join(dir, p.Profile))
	assertFileExists(t, filepath.Join(dir, p.ToolInventory))
	assertFileExists(t, filepath.Join(dir, p.GapReport))
	assertFileExists(t, filepath.Join(dir, p.RolloutPlan))

	c := loadConstitution(t, dir)
	assertLanguageDetected(t, c, "typescript")
}

func TestInitPythonLib(t *testing.T) {
	dir := fixtureDir(t, "python-lib")
	runInitNonInteractive(t, dir)

	p := config.DefaultPaths()
	assertFileExists(t, filepath.Join(dir, p.Constitution))
	assertFileExists(t, filepath.Join(dir, p.Profile))
	assertFileExists(t, filepath.Join(dir, p.ToolInventory))
	assertFileExists(t, filepath.Join(dir, p.GapReport))
	assertFileExists(t, filepath.Join(dir, p.RolloutPlan))

	c := loadConstitution(t, dir)
	assertLanguageDetected(t, c, "python")
}
