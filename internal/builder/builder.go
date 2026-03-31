package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/mkh/rice-railing/internal/config"
	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/provenance"
	"github.com/mkh/rice-railing/internal/resolution"
)

// Builder generates project-specific toolkit artifacts.
type Builder struct {
	Root         string
	Constitution *constitution.Constitution
	Plan         *resolution.RolloutPlan
	DryRun       bool
	Tracker      *provenance.Tracker
}

// NewBuilder creates a toolkit builder.
func NewBuilder(root string, c *constitution.Constitution, plan *resolution.RolloutPlan, tracker ...*provenance.Tracker) *Builder {
	b := &Builder{
		Root:         root,
		Constitution: c,
		Plan:         plan,
	}
	if len(tracker) > 0 {
		b.Tracker = tracker[0]
	}
	return b
}

// recordGeneration is a nil-safe helper for provenance tracking.
func (b *Builder) recordGeneration(id, artifact string) {
	if b.Tracker != nil {
		b.Tracker.RecordGeneration(id, artifact, "build-toolkit")
	}
}

// Build generates all toolkit artifacts.
func (b *Builder) Build() (*BuildReport, error) {
	report := &BuildReport{}

	dirs := []string{
		config.ToolkitDir,
		config.ProvenanceDir,
		config.PoliciesDir,
		config.RulesDir,
		filepath.Join(config.RulesDir, "semgrep"),
		filepath.Join(config.RulesDir, "ast-grep"),
		filepath.Join(config.RulesDir, "custom"),
		config.CodemodsDir,
		filepath.Join(config.CodemodsDir, "local"),
		filepath.Join(config.CodemodsDir, "generated"),
		config.ScaffoldsDir,
		config.PromptsDir,
		config.ReportsDir,
		config.StateDir,
		config.DocsDir,
		config.AgentDir,
		filepath.Join(config.AgentDir, "profiles"),
		filepath.Join(config.AgentDir, "workflow-packs"),
		filepath.Join(config.AgentDir, "adapters"),
		filepath.Join(config.AgentDir, "prompts"),
		filepath.Join(config.AgentDir, "state"),
		"bin",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(b.Root, dir)
		if b.DryRun {
			report.Actions = append(report.Actions, BuildAction{Type: "mkdir", Path: dir})
			continue
		}
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return nil, fmt.Errorf("creating %s: %w", dir, err)
		}
		report.Actions = append(report.Actions, BuildAction{Type: "mkdir", Path: dir})
	}

	// Generate wrapper scripts
	if b.Constitution.Workflow.GenerateLocalWrappers {
		wrappers := []struct {
			Name    string
			Content string
		}{
			{"bin/rice-rail-check", wrapperCheck},
			{"bin/rice-rail-fix", wrapperFix},
			{"bin/rice-rail-baseline", wrapperBaseline},
			{"bin/rice-rail-cycle", wrapperCycle},
			{"bin/rice-rail-report", wrapperReport},
			{"bin/rice-rail-explain", wrapperExplain},
		}

		for _, w := range wrappers {
			fullPath := filepath.Join(b.Root, w.Name)
			if b.DryRun {
				report.Actions = append(report.Actions, BuildAction{Type: "generate", Path: w.Name})
				continue
			}
			if err := os.WriteFile(fullPath, []byte(w.Content), 0755); err != nil {
				return nil, fmt.Errorf("writing %s: %w", w.Name, err)
			}
			report.Actions = append(report.Actions, BuildAction{Type: "generate", Path: w.Name})
			b.recordGeneration("wrapper-"+filepath.Base(w.Name), w.Name)
		}
	}

	// Generate workflow packs
	packs := []string{"init", "build-toolkit", "baseline", "cycle", "explain"}
	for _, pack := range packs {
		dir := filepath.Join(config.AgentDir, "workflow-packs", pack)
		fullDir := filepath.Join(b.Root, dir)
		if b.DryRun {
			report.Actions = append(report.Actions, BuildAction{Type: "workflow-pack", Path: dir})
			continue
		}
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			return nil, fmt.Errorf("creating workflow pack dir %s: %w", dir, err)
		}

		content, err := renderWorkflowPack(pack, b.Constitution)
		if err != nil {
			return nil, fmt.Errorf("rendering workflow pack %s: %w", pack, err)
		}
		readmePath := filepath.Join(fullDir, "README.md")
		if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("writing workflow pack %s: %w", pack, err)
		}
		report.Actions = append(report.Actions, BuildAction{Type: "workflow-pack", Path: dir})
		b.recordGeneration("workflow-pack-"+pack, dir)
	}

	// Generate toolkit docs
	if !b.DryRun {
		overview := renderToolkitOverview(b.Constitution, b.Plan)
		overviewPath := filepath.Join(b.Root, config.DocsDir, "toolkit-overview.md")
		if err := os.WriteFile(overviewPath, []byte(overview), 0644); err != nil {
			return nil, fmt.Errorf("writing toolkit overview: %w", err)
		}
		report.Actions = append(report.Actions, BuildAction{Type: "doc", Path: config.DocsDir + "/toolkit-overview.md"})
		b.recordGeneration("doc-toolkit-overview", config.DocsDir+"/toolkit-overview.md")

		// Operator guide
		opGuide := RenderOperatorGuide(b.Constitution)
		opGuidePath := filepath.Join(b.Root, config.DocsDir, "operator-guide.md")
		if err := os.WriteFile(opGuidePath, []byte(opGuide), 0644); err != nil {
			return nil, fmt.Errorf("writing operator guide: %w", err)
		}
		report.Actions = append(report.Actions, BuildAction{Type: "doc", Path: config.DocsDir + "/operator-guide.md"})
		b.recordGeneration("doc-operator-guide", config.DocsDir+"/operator-guide.md")

		// Rule catalog
		ruleCatalog := RenderRuleCatalog(b.Constitution, b.Plan)
		ruleCatalogPath := filepath.Join(b.Root, config.DocsDir, "rule-catalog.md")
		if err := os.WriteFile(ruleCatalogPath, []byte(ruleCatalog), 0644); err != nil {
			return nil, fmt.Errorf("writing rule catalog: %w", err)
		}
		report.Actions = append(report.Actions, BuildAction{Type: "doc", Path: config.DocsDir + "/rule-catalog.md"})
		b.recordGeneration("doc-rule-catalog", config.DocsDir+"/rule-catalog.md")
	}

	// Generate CI workflow if enabled
	if b.Constitution.Workflow.GenerateCIIntegration {
		ciDir := filepath.Join(b.Root, ".github", "workflows")
		ciPath := filepath.Join(ciDir, "rice-rail.yml")
		if b.DryRun {
			report.Actions = append(report.Actions, BuildAction{Type: "ci", Path: ".github/workflows/rice-rail.yml"})
		} else {
			ciContent, err := RenderGitHubActionsWorkflow(b.Constitution)
			if err != nil {
				return nil, fmt.Errorf("rendering CI workflow: %w", err)
			}
			if err := os.MkdirAll(ciDir, 0755); err != nil {
				return nil, fmt.Errorf("creating CI dir: %w", err)
			}
			if err := os.WriteFile(ciPath, []byte(ciContent), 0644); err != nil {
				return nil, fmt.Errorf("writing CI workflow: %w", err)
			}
			report.Actions = append(report.Actions, BuildAction{Type: "ci", Path: ".github/workflows/rice-rail.yml"})
		}
	}

	// Generate agent-native skills for all detected CLI agents
	skillActions, err := GenerateAgentSkills(b.Root, b.Constitution, b.DryRun)
	if err != nil {
		return nil, fmt.Errorf("generating agent skills: %w", err)
	}
	report.Actions = append(report.Actions, skillActions...)

	// Save toolkit version state
	if !b.DryRun {
		stateDir := filepath.Join(b.Root, config.StateDir)
		os.MkdirAll(stateDir, 0755)
		versionState := fmt.Sprintf(`{"version": 1, "generated_at": "%s"}`, time.Now().Format(time.RFC3339))
		os.WriteFile(filepath.Join(stateDir, "toolkit-version.json"), []byte(versionState), 0644)
	}

	report.Success = true
	return report, nil
}

// BuildReport describes what the builder did.
type BuildReport struct {
	Success bool          `yaml:"success" json:"success"`
	Actions []BuildAction `yaml:"actions" json:"actions"`
}

// BuildAction is a single builder action.
type BuildAction struct {
	Type string `yaml:"type" json:"type"` // mkdir, generate, workflow-pack, doc
	Path string `yaml:"path" json:"path"`
}

func renderWorkflowPack(name string, c *constitution.Constitution) (string, error) {
	tmpl, err := template.New(name).Parse(workflowPackTemplate)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	data := map[string]any{
		"Name":         name,
		"SafetyMode":   c.Quality.SafetyMode,
		"BlockOn":      c.Quality.BlockOn,
		"AllowAutofix": c.Automation.AllowSafeAutofix,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderToolkitOverview(c *constitution.Constitution, plan *resolution.RolloutPlan) string {
	var buf strings.Builder
	buf.WriteString("# Toolkit Overview\n\n")
	buf.WriteString("Generated by rice-rail.\n\n")
	buf.WriteString("## Constitution Summary\n\n")
	buf.WriteString(fmt.Sprintf("- Architecture: %s\n", c.Architecture.TargetStyle))
	buf.WriteString(fmt.Sprintf("- Safety mode: %s\n", c.Quality.SafetyMode))
	buf.WriteString(fmt.Sprintf("- Blocking checks: %v\n", c.Quality.BlockOn))
	buf.WriteString(fmt.Sprintf("- Safe autofix: %v\n", c.Automation.AllowSafeAutofix))
	buf.WriteString(fmt.Sprintf("- Languages: %v\n", c.Project.Languages))

	if plan != nil && len(plan.Steps) > 0 {
		buf.WriteString("\n## Rollout Plan\n\n")
		for _, s := range plan.Steps {
			buf.WriteString(fmt.Sprintf("%d. [%s] %s — %s\n", s.Order, s.Risk, s.Capability, s.Description))
		}
	}

	return buf.String()
}

const workflowPackTemplate = `# Workflow Pack: {{.Name}}

## Purpose

Agent guidance for the {{.Name}} phase of rice-rail.

## Safety Rules

- Safety mode: {{.SafetyMode}}
- Safe autofix allowed: {{.AllowAutofix}}
{{- range .BlockOn}}
- BLOCKING: {{.}}
{{- end}}

## Instructions

Follow the project constitution in .project-toolkit/constitution.yaml.
Do not modify the constitution during this phase.
Report unresolved issues separately from mechanical fixes.
`
