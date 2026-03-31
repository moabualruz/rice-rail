package adapters

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mkh/rice-railing/internal/exec"
)

// CodexAdapter implements AgentAdapter by shelling out to the codex CLI.
type CodexAdapter struct {
	runner       *exec.Runner
	repoRoot     string
	workflowPack string
}

// NewCodexAdapter creates a new Codex adapter.
func NewCodexAdapter(runner *exec.Runner, repoRoot string) *CodexAdapter {
	return &CodexAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *CodexAdapter) Name() string { return "codex" }

func (a *CodexAdapter) Capabilities() AgentCapabilities {
	return AgentCapabilities{
		MCP:              false,
		LocalTools:       true,
		StructuredOutput: false,
		NonInteractive:   true,
		PatchPreview:     false,
		ScopeConstraints: false,
	}
}

func (a *CodexAdapter) LoadWorkflowPack(ctx context.Context, packName string) error {
	readmePath := filepath.Join(a.repoRoot, ".agent", "workflow-packs", packName, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("loading workflow pack %q: %w", packName, err)
	}
	a.workflowPack = string(data)
	return nil
}

func (a *CodexAdapter) RunTask(ctx context.Context, input TaskInput) (*TaskResult, error) {
	prompt := a.buildPrompt(input)

	args := []string{
		"--full-auto",
		prompt,
	}

	result, err := a.runner.Run(ctx, "codex", args...)
	if err != nil {
		return nil, fmt.Errorf("codex: %w", err)
	}

	return parseCodexOutput(result), nil
}

func (a *CodexAdapter) buildPrompt(input TaskInput) string {
	var parts []string

	if a.workflowPack != "" {
		parts = append(parts, a.workflowPack)
	}

	parts = append(parts, input.Intent)

	if len(input.Files) > 0 {
		parts = append(parts, "Files: "+strings.Join(input.Files, ", "))
	}

	if input.Module != "" {
		parts = append(parts, "Module: "+input.Module)
	}

	for key, val := range input.Constraints {
		parts = append(parts, fmt.Sprintf("%s: %s", key, val))
	}

	parts = append(parts, "Only modify files in scope. Report what you changed and what remains unresolved.")

	return strings.Join(parts, "\n\n")
}

// parseCodexOutput extracts structured results from codex CLI stdout.
func parseCodexOutput(result *exec.Result) *TaskResult {
	tr := &TaskResult{
		Success: result.Success(),
	}

	filesSet := map[string]bool{}
	var summaryLines []string

	scanner := bufio.NewScanner(strings.NewReader(result.Stdout))
	for scanner.Scan() {
		line := scanner.Text()
		summaryLines = append(summaryLines, line)

		matches := filePathRe.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			path := m[1]
			if looksLikeFilePath(path) {
				filesSet[path] = true
			}
		}
	}

	for f := range filesSet {
		tr.FilesChanged = append(tr.FilesChanged, f)
	}

	start := 0
	if len(summaryLines) > 10 {
		start = len(summaryLines) - 10
	}
	tr.Summary = strings.Join(summaryLines[start:], "\n")

	return tr
}
