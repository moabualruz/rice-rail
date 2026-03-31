package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mkh/rice-railing/internal/constitution"
)

func testConstitution() *constitution.Constitution {
	return &constitution.Constitution{
		Version: 1,
		Project: constitution.ProjectInfo{
			Name:      "test-project",
			Languages: []string{"go", "typescript"},
		},
		Architecture: constitution.ArchitectureSpec{
			TargetStyle: "hexagonal",
		},
		Quality: constitution.QualitySpec{
			SafetyMode:              "balanced",
			BlockOn:                 []string{"lint", "tests", "typecheck"},
			AdvisoryOn:              []string{"complexity", "coverage"},
			MaxChangedFilesPerCycle: 10,
			MaxChangedLinesPerCycle: 500,
		},
		Automation: constitution.AutomationSpec{
			AllowSafeAutofix: true,
		},
		Workflow: constitution.WorkflowSpec{
			GenerateLocalWrappers:  true,
			GenerateCIIntegration:  true,
		},
	}
}

func TestRiceRailSkills(t *testing.T) {
	c := testConstitution()
	skills := riceRailSkills(c)

	if len(skills) != 7 {
		t.Fatalf("expected 7 skills, got %d", len(skills))
	}

	expectedNames := []string{
		"rice-rail-check",
		"rice-rail-fix",
		"rice-rail-baseline",
		"rice-rail-cycle",
		"rice-rail-explain",
		"rice-rail-report",
		"rice-rail-doctor",
	}

	for i, s := range skills {
		if s.Name == "" {
			t.Errorf("skill %d has empty Name", i)
		}
		if s.Description == "" {
			t.Errorf("skill %d (%s) has empty Description", i, s.Name)
		}
		if s.Body == "" {
			t.Errorf("skill %d (%s) has empty Body", i, s.Name)
		}
	}

	for i, expected := range expectedNames {
		if skills[i].Name != expected {
			t.Errorf("skill %d: expected name %q, got %q", i, expected, skills[i].Name)
		}
	}

	// Verify constitution data is embedded in skill bodies
	checkSkill := skills[0] // rice-rail-check
	if !strings.Contains(checkSkill.Body, "balanced") {
		t.Error("rice-rail-check body should contain safety mode")
	}
	if !strings.Contains(checkSkill.Body, "lint") {
		t.Error("rice-rail-check body should contain blocking checks")
	}
}

func TestGenerateClaudeSkills(t *testing.T) {
	dir := t.TempDir()
	c := testConstitution()

	if err := generateClaudeSkills(dir, c); err != nil {
		t.Fatalf("generateClaudeSkills failed: %v", err)
	}

	skillPath := filepath.Join(dir, ".claude", "skills", "rice-rail-check", "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("expected %s to exist: %v", skillPath, err)
	}

	content := string(data)
	if !strings.Contains(content, "---") {
		t.Error("SKILL.md should contain frontmatter delimiters")
	}
	if !strings.Contains(content, "name: rice-rail-check") {
		t.Error("SKILL.md should contain name in frontmatter")
	}
	if !strings.Contains(content, "description:") {
		t.Error("SKILL.md should contain description in frontmatter")
	}

	// Verify all 7 skills are generated
	skills := riceRailSkills(c)
	for _, s := range skills {
		p := filepath.Join(dir, ".claude", "skills", s.Name, "SKILL.md")
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected skill file %s to exist", p)
		}
	}
}

func TestGenerateGeminiSkills(t *testing.T) {
	dir := t.TempDir()
	c := testConstitution()

	if err := generateGeminiSkills(dir, c); err != nil {
		t.Fatalf("generateGeminiSkills failed: %v", err)
	}

	skillPath := filepath.Join(dir, ".gemini", "skills", "rice-rail-check", "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("expected %s to exist: %v", skillPath, err)
	}

	content := string(data)
	if !strings.Contains(content, "name: rice-rail-check") {
		t.Error("Gemini SKILL.md should contain name in frontmatter")
	}
	if !strings.Contains(content, "description:") {
		t.Error("Gemini SKILL.md should contain description in frontmatter")
	}

	// Verify all skills generated
	skills := riceRailSkills(c)
	for _, s := range skills {
		p := filepath.Join(dir, ".gemini", "skills", s.Name, "SKILL.md")
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected Gemini skill file %s to exist", p)
		}
	}
}

func TestGenerateCopilotInstructions(t *testing.T) {
	dir := t.TempDir()
	c := testConstitution()

	if err := generateCopilotInstructions(dir, c); err != nil {
		t.Fatalf("generateCopilotInstructions failed: %v", err)
	}

	instrPath := filepath.Join(dir, ".github", "instructions", "rice-rail-check.instructions.md")
	data, err := os.ReadFile(instrPath)
	if err != nil {
		t.Fatalf("expected %s to exist: %v", instrPath, err)
	}

	content := string(data)
	if !strings.Contains(content, "applyTo:") {
		t.Error("Copilot instruction should contain applyTo frontmatter")
	}
	if !strings.Contains(content, `applyTo: "**/*"`) {
		t.Error("Copilot instruction should apply to all files")
	}

	// Verify all instruction files generated
	skills := riceRailSkills(c)
	for _, s := range skills {
		p := filepath.Join(dir, ".github", "instructions", s.Name+".instructions.md")
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected Copilot instruction file %s to exist", p)
		}
	}

	// Copilot also generates AGENTS.md
	agentsPath := filepath.Join(dir, "AGENTS.md")
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		t.Error("expected AGENTS.md to be generated by Copilot instructions")
	}
}

func TestGenerateAgentsMD(t *testing.T) {
	dir := t.TempDir()
	c := testConstitution()

	if err := generateAgentsMD(dir, c); err != nil {
		t.Fatalf("generateAgentsMD failed: %v", err)
	}

	agentsPath := filepath.Join(dir, "AGENTS.md")
	data, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("expected AGENTS.md to exist: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "Safety mode: balanced") {
		t.Error("AGENTS.md should contain safety mode")
	}
	if !strings.Contains(content, "Blocking checks: lint, tests, typecheck") {
		t.Error("AGENTS.md should contain blocking checks")
	}
	if !strings.Contains(content, "Languages: go, typescript") {
		t.Error("AGENTS.md should contain languages")
	}
	if !strings.Contains(content, "Architecture: hexagonal") {
		t.Error("AGENTS.md should contain architecture style")
	}
	if !strings.Contains(content, "rice-rail check") {
		t.Error("AGENTS.md should contain available commands")
	}
}

func TestGenerateAgentsMDWithLayering(t *testing.T) {
	dir := t.TempDir()
	c := testConstitution()
	c.Architecture.Layering = constitution.LayeringSpec{
		Enabled: true,
		Layers:  []string{"domain", "application", "infrastructure"},
		ForbiddenDependencies: []constitution.ForbiddenDependency{
			{From: "domain", To: "infrastructure"},
		},
	}
	c.Architecture.DependencyPolicy.NoCycles = true

	if err := generateAgentsMD(dir, c); err != nil {
		t.Fatalf("generateAgentsMD failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "Architecture Constraints") {
		t.Error("AGENTS.md should contain architecture constraints section when layering enabled")
	}
	if !strings.Contains(content, "FORBIDDEN: domain must not depend on infrastructure") {
		t.Error("AGENTS.md should contain forbidden dependency details")
	}
	if !strings.Contains(content, "No import cycles") {
		t.Error("AGENTS.md should contain no-cycles constraint")
	}
}

func TestGenerateQwenSkills(t *testing.T) {
	dir := t.TempDir()
	c := testConstitution()

	if err := generateQwenSkills(dir, c); err != nil {
		t.Fatalf("generateQwenSkills failed: %v", err)
	}

	skillPath := filepath.Join(dir, ".qwen", "skills", "rice-rail-check", "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("expected %s to exist: %v", skillPath, err)
	}

	content := string(data)
	if !strings.Contains(content, "name: rice-rail-check") {
		t.Error("Qwen SKILL.md should contain name in frontmatter")
	}
	if !strings.Contains(content, "description:") {
		t.Error("Qwen SKILL.md should contain description in frontmatter")
	}

	// Verify all skills generated
	skills := riceRailSkills(c)
	for _, s := range skills {
		p := filepath.Join(dir, ".qwen", "skills", s.Name, "SKILL.md")
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected Qwen skill file %s to exist", p)
		}
	}
}
