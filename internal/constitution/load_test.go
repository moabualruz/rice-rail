package constitution

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	c := &Constitution{
		Version: 1,
		Project: ProjectInfo{
			Name:      "test-project",
			RepoType:  "single",
			Languages: []string{"go"},
		},
		Quality: QualitySpec{
			SafetyMode: "balanced",
			BlockOn:    []string{"lint", "tests"},
		},
		Automation: AutomationSpec{
			AllowSafeAutofix:    true,
			BaselineModeDefault: "safe_only",
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "constitution.yaml")

	if err := Save(c, path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Version != 1 {
		t.Errorf("version: got %d, want 1", loaded.Version)
	}
	if loaded.Project.Name != "test-project" {
		t.Errorf("name: got %q, want test-project", loaded.Project.Name)
	}
	if loaded.Quality.SafetyMode != "balanced" {
		t.Errorf("safety_mode: got %q, want balanced", loaded.Quality.SafetyMode)
	}
	if !loaded.Automation.AllowSafeAutofix {
		t.Error("allow_safe_autofix should be true")
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := Load("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	os.WriteFile(path, []byte("{{invalid yaml"), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
