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

// AiderAdapter implements AgentAdapter by shelling out to the aider CLI.
type AiderAdapter struct {
	runner       *exec.Runner
	repoRoot     string
	workflowPack string
}

// NewAiderAdapter creates a new Aider adapter.
func NewAiderAdapter(runner *exec.Runner, repoRoot string) *AiderAdapter {
	return &AiderAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *AiderAdapter) Name() string { return "aider" }

func (a *AiderAdapter) Capabilities() AgentCapabilities {
	return AgentCapabilities{
		MCP:              false,
		LocalTools:       false,
		StructuredOutput: false,
		NonInteractive:   true,
		PatchPreview:     true,
		ScopeConstraints: true,
	}
}

func (a *AiderAdapter) LoadWorkflowPack(ctx context.Context, packName string) error {
	readmePath := filepath.Join(a.repoRoot, ".agent", "workflow-packs", packName, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("loading workflow pack %q: %w", packName, err)
	}
	a.workflowPack = string(data)
	return nil
}

func (a *AiderAdapter) RunTask(ctx context.Context, input TaskInput) (*TaskResult, error) {
	prompt := a.buildPrompt(input)

	args := []string{
		"--yes-always",
		"--no-auto-commits",
		"--message", prompt,
	}

	for _, f := range input.Files {
		args = append(args, "--file", f)
	}

	result, err := a.runner.Run(ctx, "aider", args...)
	if err != nil {
		return nil, fmt.Errorf("aider: %w", err)
	}

	return parseAiderOutput(result), nil
}

func (a *AiderAdapter) buildPrompt(input TaskInput) string {
	var parts []string

	if a.workflowPack != "" {
		parts = append(parts, a.workflowPack)
	}

	parts = append(parts, input.Intent)

	if input.Module != "" {
		parts = append(parts, "Constrain changes to module: "+input.Module)
	}

	for key, val := range input.Constraints {
		parts = append(parts, fmt.Sprintf("%s: %s", key, val))
	}

	parts = append(parts, "Only modify files in scope. Report what you changed and what remains unresolved.")

	return strings.Join(parts, "\n\n")
}

// parseAiderOutput extracts structured results from aider CLI stdout.
func parseAiderOutput(result *exec.Result) *TaskResult {
	tr := &TaskResult{
		Success: result.Success(),
	}

	filesSet := map[string]bool{}
	var summaryLines []string

	scanner := bufio.NewScanner(strings.NewReader(result.Stdout))
	for scanner.Scan() {
		line := scanner.Text()
		summaryLines = append(summaryLines, line)

		// Aider reports edited files with patterns like "Applied edit to file.go".
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
