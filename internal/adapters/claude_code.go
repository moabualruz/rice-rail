package adapters

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mkh/rice-railing/internal/exec"
)

// ClaudeCodeAdapter implements AgentAdapter by shelling out to the claude CLI.
type ClaudeCodeAdapter struct {
	runner       *exec.Runner
	repoRoot     string
	workflowPack string // loaded workflow pack instructions
}

// NewClaudeCodeAdapter creates a new Claude Code adapter.
func NewClaudeCodeAdapter(runner *exec.Runner, repoRoot string) *ClaudeCodeAdapter {
	return &ClaudeCodeAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *ClaudeCodeAdapter) Name() string { return "claude-code" }

func (a *ClaudeCodeAdapter) Capabilities() AgentCapabilities {
	return AgentCapabilities{
		MCP:              true,
		LocalTools:       true,
		StructuredOutput: true,
		NonInteractive:   true,
		PatchPreview:     true,
		ScopeConstraints: true,
	}
}

func (a *ClaudeCodeAdapter) LoadWorkflowPack(ctx context.Context, packName string) error {
	readmePath := filepath.Join(a.repoRoot, ".agent", "workflow-packs", packName, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("loading workflow pack %q: %w", packName, err)
	}
	a.workflowPack = string(data)
	return nil
}

func (a *ClaudeCodeAdapter) RunTask(ctx context.Context, input TaskInput) (*TaskResult, error) {
	prompt := a.buildPrompt(input)

	args := []string{
		"-p", prompt,
		"--allowedTools", "Edit,Write,Bash,Read,Grep,Glob",
	}

	result, err := a.runner.Run(ctx, "claude", args...)
	if err != nil {
		return nil, fmt.Errorf("claude-code: %w", err)
	}

	return parseClaudeOutput(result), nil
}

// buildPrompt assembles the full prompt from workflow pack, intent, and constraints.
func (a *ClaudeCodeAdapter) buildPrompt(input TaskInput) string {
	var parts []string

	if a.workflowPack != "" {
		parts = append(parts, "## Workflow Instructions\n"+a.workflowPack)
	}

	parts = append(parts, "## Task\n"+input.Intent)

	if len(input.Files) > 0 {
		parts = append(parts, "## File Scope\nOnly modify these files:\n- "+strings.Join(input.Files, "\n- "))
	}

	if input.Module != "" {
		parts = append(parts, "## Module Scope\nConstrain changes to module: "+input.Module)
	}

	for key, val := range input.Constraints {
		parts = append(parts, fmt.Sprintf("## Constraint: %s\n%s", key, val))
	}

	parts = append(parts, "Only modify files in scope. Report what you changed and what remains unresolved.")

	return strings.Join(parts, "\n\n")
}

// filePathRe matches common file paths in output (e.g., src/main.go, ./internal/foo.go).
var filePathRe = regexp.MustCompile(`(?:^|\s)((?:\./)?(?:[\w./-]+/)?[\w.-]+\.[\w]+)`)

// parseClaudeOutput extracts structured results from claude CLI stdout.
func parseClaudeOutput(result *exec.Result) *TaskResult {
	tr := &TaskResult{
		Success: result.Success(),
	}

	var summaryLines []string
	var unresolved []string
	filesSet := map[string]bool{}

	scanner := bufio.NewScanner(strings.NewReader(result.Stdout))
	for scanner.Scan() {
		line := scanner.Text()

		// Collect file paths mentioned in the output.
		matches := filePathRe.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			path := m[1]
			if looksLikeFilePath(path) {
				filesSet[path] = true
			}
		}

		// Detect unresolved items.
		lower := strings.ToLower(line)
		if strings.Contains(lower, "unresolved") || strings.Contains(lower, "could not") || strings.Contains(lower, "failed to") {
			unresolved = append(unresolved, strings.TrimSpace(line))
		}

		summaryLines = append(summaryLines, line)
	}

	for f := range filesSet {
		tr.FilesChanged = append(tr.FilesChanged, f)
	}

	// Use last 10 lines as summary (agent output tends to summarise at the end).
	start := 0
	if len(summaryLines) > 10 {
		start = len(summaryLines) - 10
	}
	tr.Summary = strings.Join(summaryLines[start:], "\n")
	tr.Unresolved = unresolved

	return tr
}

// looksLikeFilePath applies heuristics to reject false-positive path matches.
func looksLikeFilePath(s string) bool {
	// Must contain a dot for an extension.
	if !strings.Contains(s, ".") {
		return false
	}
	// Reject URLs.
	if strings.HasPrefix(s, "http") {
		return false
	}
	// Reject very short strings.
	if len(s) < 4 {
		return false
	}
	return true
}
