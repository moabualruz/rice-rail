package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/exec"
)

// AgentSkillSet defines a set of skills to generate for a specific agent.
type AgentSkillSet struct {
	Agent    string
	SkillDir string // relative to repo root
	Generate func(root string, c *constitution.Constitution) error
}

// GenerateAgentSkills generates project-level skills for all detected CLI agents.
func GenerateAgentSkills(root string, c *constitution.Constitution, dryRun bool) ([]BuildAction, error) {
	var actions []BuildAction

	agents := detectAgentsForSkills()

	for _, agent := range agents {
		if dryRun {
			actions = append(actions, BuildAction{Type: "agent-skills", Path: agent.SkillDir})
			continue
		}
		if err := agent.Generate(root, c); err != nil {
			return nil, fmt.Errorf("generating %s skills: %w", agent.Agent, err)
		}
		actions = append(actions, BuildAction{Type: "agent-skills", Path: agent.SkillDir})
	}

	return actions, nil
}

func detectAgentsForSkills() []AgentSkillSet {
	var agents []AgentSkillSet

	if _, ok := exec.Which("claude"); ok {
		agents = append(agents, AgentSkillSet{
			Agent:    "claude-code",
			SkillDir: ".claude/skills",
			Generate: generateClaudeSkills,
		})
	}

	if _, ok := exec.Which("gemini"); ok {
		agents = append(agents, AgentSkillSet{
			Agent:    "gemini",
			SkillDir: ".gemini/skills",
			Generate: generateGeminiSkills,
		})
	}

	if _, ok := exec.Which("copilot"); ok {
		agents = append(agents, AgentSkillSet{
			Agent:    "copilot",
			SkillDir: ".github/instructions",
			Generate: generateCopilotInstructions,
		})
	}

	if _, ok := exec.Which("opencode"); ok {
		agents = append(agents, AgentSkillSet{
			Agent:    "opencode",
			SkillDir: "AGENTS.md",
			Generate: generateOpenCodeAgents,
		})
	}

	if _, ok := exec.Which("qwen"); ok {
		agents = append(agents, AgentSkillSet{
			Agent:    "qwen",
			SkillDir: ".qwen/skills",
			Generate: generateQwenSkills,
		})
	}

	return agents
}

// --- Skill definitions shared across agents ---

type skillDef struct {
	Name        string
	Description string
	Body        string
}

func riceRailSkills(c *constitution.Constitution) []skillDef {
	safetyBlock := fmt.Sprintf("Safety mode: %s. Blocking checks: %s.",
		c.Quality.SafetyMode, strings.Join(c.Quality.BlockOn, ", "))

	return []skillDef{
		{
			Name:        "rice-rail-check",
			Description: "Run all blocking checks via rice-rail without modifying code",
			Body: fmt.Sprintf(`Run rice-rail check on this project.

## What to do
1. Run: rice-rail check
2. Report results: which checks passed, which failed, violation count
3. For any failures, show the specific violations with file:line

## Constitution
%s
Do NOT modify any code. Only report findings.`, safetyBlock),
		},
		{
			Name:        "rice-rail-fix",
			Description: "Run safe autofixes via rice-rail (formatting, imports, safe lint fixes)",
			Body: fmt.Sprintf(`Run rice-rail fix on this project.

## What to do
1. Run: rice-rail fix
2. Report what was fixed
3. Run: rice-rail check to verify fixes didn't break anything

## Constitution
%s
Only SAFE autofixes are allowed. Do not apply unsafe changes.`, safetyBlock),
		},
		{
			Name:        "rice-rail-baseline",
			Description: "Normalize codebase to policy compliance via rice-rail baseline remediation",
			Body: fmt.Sprintf(`Run rice-rail baseline on this project.

## What to do
1. First run: rice-rail baseline --report-only (to see what needs fixing)
2. Show the report to the user
3. If user approves, run: rice-rail baseline
4. Report convergence status and any residual violations

## Constitution
%s
Baseline separates safe fixes from unsafe ones. Never apply unsafe changes without explicit approval.`, safetyBlock),
		},
		{
			Name:        "rice-rail-cycle",
			Description: "Run the daily intent→tool→verify→refine development cycle via rice-rail",
			Body: fmt.Sprintf(`Run a rice-rail development cycle.

## What to do
1. Ask the user for their intent (what they want to accomplish)
2. Run: rice-rail cycle "<intent>" --max-iterations 5
3. Report: tools invoked, rules triggered, files changed, residual issues
4. If there are unresolved semantic issues, help resolve them

## Constitution
%s
Max files per cycle: %d. Max lines per cycle: %d.
Apply deterministic transforms first. Only use LLM judgment for semantic issues.`,
				safetyBlock, c.Quality.MaxChangedFilesPerCycle, c.Quality.MaxChangedLinesPerCycle),
		},
		{
			Name:        "rice-rail-explain",
			Description: "Explain why a rice-rail rule, tool, or artifact exists and how to modify it",
			Body: `Explain a rice-rail rule or artifact.

## What to do
1. Ask the user what they want explained (rule ID, tool, or artifact path)
2. Run: rice-rail explain <target>
3. Show the full explanation: what it checks, why it exists, how it's enforced, where it came from
4. If the user wants to change it, explain how to edit the constitution or waive the rule`,
		},
		{
			Name:        "rice-rail-report",
			Description: "Show rice-rail toolkit status, constitution summary, and known capability gaps",
			Body: `Show the rice-rail project report.

## What to do
1. Run: rice-rail report
2. Present the toolkit status, constitution summary, and any gaps
3. If there are gaps, suggest running rice-rail upgrade-toolkit`,
		},
		{
			Name:        "rice-rail-doctor",
			Description: "Diagnose rice-rail toolkit health and fix any issues found",
			Body: `Run rice-rail doctor to check toolkit health.

## What to do
1. Run: rice-rail doctor
2. Report all PASS/FAIL/WARN results
3. For any failures, suggest specific fixes
4. If toolkit needs rebuilding, suggest: rice-rail regenerate`,
		},
	}
}

// --- Claude Code skills ---

func generateClaudeSkills(root string, c *constitution.Constitution) error {
	skills := riceRailSkills(c)
	for _, s := range skills {
		dir := filepath.Join(root, ".claude", "skills", s.Name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		content := fmt.Sprintf(`---
name: %s
description: "%s"
---

%s
`, s.Name, s.Description, s.Body)
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

// --- Gemini skills ---

func generateGeminiSkills(root string, c *constitution.Constitution) error {
	skills := riceRailSkills(c)
	for _, s := range skills {
		dir := filepath.Join(root, ".gemini", "skills", s.Name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		// Gemini uses same SKILL.md format
		content := fmt.Sprintf(`---
name: %s
description: "%s"
---

%s
`, s.Name, s.Description, s.Body)
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

// --- Copilot instructions ---

func generateCopilotInstructions(root string, c *constitution.Constitution) error {
	dir := filepath.Join(root, ".github", "instructions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	skills := riceRailSkills(c)
	for _, s := range skills {
		content := fmt.Sprintf(`---
applyTo: "**/*"
---

# %s

%s
`, s.Name, s.Body)
		filename := s.Name + ".instructions.md"
		if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644); err != nil {
			return err
		}
	}

	// Also generate AGENTS.md for Copilot
	return generateAgentsMD(root, c)
}

// --- OpenCode AGENTS.md ---

func generateOpenCodeAgents(root string, c *constitution.Constitution) error {
	return generateAgentsMD(root, c)
}

func generateAgentsMD(root string, c *constitution.Constitution) error {
	var b strings.Builder
	b.WriteString("# AGENTS.md\n\n")
	b.WriteString("Generated by rice-rail. Project engineering doctrine for AI coding agents.\n\n")

	b.WriteString("## Project Constitution\n\n")
	if c.Architecture.TargetStyle != "" {
		b.WriteString(fmt.Sprintf("- Architecture: %s\n", c.Architecture.TargetStyle))
	}
	b.WriteString(fmt.Sprintf("- Safety mode: %s\n", c.Quality.SafetyMode))
	b.WriteString(fmt.Sprintf("- Languages: %s\n", strings.Join(c.Project.Languages, ", ")))
	b.WriteString(fmt.Sprintf("- Blocking checks: %s\n", strings.Join(c.Quality.BlockOn, ", ")))
	b.WriteString(fmt.Sprintf("- Safe autofix: %v\n", c.Automation.AllowSafeAutofix))
	b.WriteString(fmt.Sprintf("- Max files per change: %d\n", c.Quality.MaxChangedFilesPerCycle))
	b.WriteString(fmt.Sprintf("- Max lines per change: %d\n\n", c.Quality.MaxChangedLinesPerCycle))

	b.WriteString("## Rules\n\n")
	b.WriteString("- Run `rice-rail check` before claiming any change is complete\n")
	b.WriteString("- Run `rice-rail fix` after making changes to auto-format\n")
	b.WriteString("- Do not modify `.project-toolkit/constitution.yaml` during normal work\n")
	b.WriteString("- Respect scope limits: do not change more files than the constitution allows\n")
	b.WriteString("- Report unresolved issues separately from mechanical fixes\n\n")

	if c.Architecture.Layering.Enabled {
		b.WriteString("## Architecture Constraints\n\n")
		for _, fd := range c.Architecture.Layering.ForbiddenDependencies {
			b.WriteString(fmt.Sprintf("- FORBIDDEN: %s must not depend on %s\n", fd.From, fd.To))
		}
		if c.Architecture.DependencyPolicy.NoCycles {
			b.WriteString("- No import cycles between modules\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("## Available Commands\n\n")
	b.WriteString("| Command | Purpose |\n")
	b.WriteString("|---------|--------|\n")
	b.WriteString("| `rice-rail check` | Run all blocking checks |\n")
	b.WriteString("| `rice-rail fix` | Run safe autofixes |\n")
	b.WriteString("| `rice-rail baseline` | Normalize to policy |\n")
	b.WriteString("| `rice-rail cycle \"<intent>\"` | Full dev cycle |\n")
	b.WriteString("| `rice-rail explain <id>` | Explain any rule |\n")
	b.WriteString("| `rice-rail report` | Show toolkit status |\n")
	b.WriteString("| `rice-rail doctor` | Check toolkit health |\n")

	return os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(b.String()), 0644)
}

// --- Qwen skills ---

func generateQwenSkills(root string, c *constitution.Constitution) error {
	skills := riceRailSkills(c)
	for _, s := range skills {
		dir := filepath.Join(root, ".qwen", "skills", s.Name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		// Qwen uses same SKILL.md format as Claude Code
		content := fmt.Sprintf(`---
name: %s
description: "%s"
---

%s
`, s.Name, s.Description, s.Body)
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}
