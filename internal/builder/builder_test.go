package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mkh/rice-railing/internal/config"
	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/provenance"
	"github.com/mkh/rice-railing/internal/resolution"
)

func TestBuildCreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	c := &constitution.Constitution{
		Version: 1,
		Workflow: constitution.WorkflowSpec{
			GenerateLocalWrappers: true,
		},
		Quality: constitution.QualitySpec{
			SafetyMode: "balanced",
		},
		Automation: constitution.AutomationSpec{
			AllowSafeAutofix: true,
		},
	}

	b := NewBuilder(dir, c, &resolution.RolloutPlan{})
	report, err := b.Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if !report.Success {
		t.Fatal("expected success")
	}

	// Check key directories exist
	for _, d := range []string{".project-toolkit", ".agent", "bin"} {
		if _, err := os.Stat(filepath.Join(dir, d)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", d)
		}
	}

	// Check wrapper scripts exist and are executable
	for _, w := range []string{"rice-rail-check", "rice-rail-fix", "rice-rail-baseline"} {
		path := filepath.Join(dir, "bin", w)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			t.Errorf("expected %s to exist", w)
			continue
		}
		if info.Mode()&0111 == 0 {
			t.Errorf("expected %s to be executable", w)
		}
	}
}

func TestBuildDryRun(t *testing.T) {
	dir := t.TempDir()
	c := &constitution.Constitution{
		Workflow: constitution.WorkflowSpec{
			GenerateLocalWrappers: true,
		},
	}

	b := NewBuilder(dir, c, &resolution.RolloutPlan{})
	b.DryRun = true
	report, err := b.Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if !report.Success {
		t.Fatal("expected success")
	}

	// In dry-run, directories should NOT be created
	if _, err := os.Stat(filepath.Join(dir, "bin")); err == nil {
		t.Error("dry-run should not create directories")
	}
}

func TestBuildIdempotent(t *testing.T) {
	dir := t.TempDir()
	c := &constitution.Constitution{
		Workflow: constitution.WorkflowSpec{
			GenerateLocalWrappers: true,
		},
	}

	b := NewBuilder(dir, c, &resolution.RolloutPlan{})

	// Run twice — should not error
	_, err := b.Build()
	if err != nil {
		t.Fatalf("first build: %v", err)
	}
	_, err = b.Build()
	if err != nil {
		t.Fatalf("second build: %v", err)
	}
}

func TestBuildGeneratesCIWorkflow(t *testing.T) {
	dir := t.TempDir()
	c := &constitution.Constitution{
		Version: 1,
		Project: constitution.ProjectInfo{
			Languages: []string{"go"},
		},
		Quality: constitution.QualitySpec{
			SafetyMode: "balanced",
			BlockOn:    []string{"lint", "tests"},
		},
		Workflow: constitution.WorkflowSpec{
			GenerateCIIntegration: true,
		},
		Automation: constitution.AutomationSpec{
			AllowSafeAutofix: true,
		},
	}

	b := NewBuilder(dir, c, &resolution.RolloutPlan{})
	report, err := b.Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if !report.Success {
		t.Fatal("expected success")
	}

	ciPath := filepath.Join(dir, ".github", "workflows", "rice-rail.yml")
	data, err := os.ReadFile(ciPath)
	if err != nil {
		t.Fatalf("expected CI workflow to exist: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "setup-go") {
		t.Error("CI workflow should contain setup-go for Go language")
	}
	if !strings.Contains(content, "rice-rail checks") {
		t.Error("CI workflow should contain workflow name")
	}
}

func TestBuildGeneratesDocs(t *testing.T) {
	dir := t.TempDir()
	c := &constitution.Constitution{
		Version: 1,
		Quality: constitution.QualitySpec{
			SafetyMode: "balanced",
			BlockOn:    []string{"lint"},
		},
		Automation: constitution.AutomationSpec{
			AllowSafeAutofix: true,
		},
	}

	b := NewBuilder(dir, c, &resolution.RolloutPlan{})
	_, err := b.Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	docs := []string{
		filepath.Join(dir, config.DocsDir, "toolkit-overview.md"),
		filepath.Join(dir, config.DocsDir, "operator-guide.md"),
		filepath.Join(dir, config.DocsDir, "rule-catalog.md"),
	}

	for _, docPath := range docs {
		if _, err := os.Stat(docPath); os.IsNotExist(err) {
			t.Errorf("expected doc %s to exist", docPath)
		}
	}

	// Verify overview has content
	data, err := os.ReadFile(docs[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Toolkit Overview") {
		t.Error("toolkit-overview.md should contain heading")
	}
}

func TestBuildGeneratesToolkitVersion(t *testing.T) {
	dir := t.TempDir()
	c := &constitution.Constitution{
		Version: 1,
		Quality: constitution.QualitySpec{
			SafetyMode: "balanced",
		},
		Automation: constitution.AutomationSpec{
			AllowSafeAutofix: true,
		},
	}

	b := NewBuilder(dir, c, &resolution.RolloutPlan{})
	_, err := b.Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	versionPath := filepath.Join(dir, config.StateDir, "toolkit-version.json")
	data, err := os.ReadFile(versionPath)
	if err != nil {
		t.Fatalf("expected toolkit-version.json to exist: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, `"version": 1`) {
		t.Error("toolkit-version.json should contain version field")
	}
	if !strings.Contains(content, `"generated_at"`) {
		t.Error("toolkit-version.json should contain generated_at timestamp")
	}
}

func TestBuildWithTracker(t *testing.T) {
	dir := t.TempDir()
	c := &constitution.Constitution{
		Version: 1,
		Quality: constitution.QualitySpec{
			SafetyMode: "balanced",
		},
		Automation: constitution.AutomationSpec{
			AllowSafeAutofix: true,
		},
		Workflow: constitution.WorkflowSpec{
			GenerateLocalWrappers: true,
		},
	}

	tracker := provenance.NewTracker()
	b := NewBuilder(dir, c, &resolution.RolloutPlan{}, tracker)
	_, err := b.Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if len(tracker.Log.Records) == 0 {
		t.Fatal("expected provenance tracker to have records after build")
	}

	// Verify at least some generation records exist
	hasGenerated := false
	for _, r := range tracker.Log.Records {
		if r.Type == "generated" {
			hasGenerated = true
			break
		}
	}
	if !hasGenerated {
		t.Error("expected at least one 'generated' provenance record")
	}

	// Verify wrapper and doc records are tracked
	hasWrapper := false
	hasDoc := false
	for _, r := range tracker.Log.Records {
		if strings.Contains(r.ID, "wrapper-") {
			hasWrapper = true
		}
		if strings.Contains(r.ID, "doc-") {
			hasDoc = true
		}
	}
	if !hasWrapper {
		t.Error("expected wrapper generation to be tracked")
	}
	if !hasDoc {
		t.Error("expected doc generation to be tracked")
	}
}
