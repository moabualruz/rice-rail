package adapters

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// loadPackReadme reads a workflow pack's README.md for prompt context.
func loadPackReadme(repoRoot, packName string) (string, error) {
	readmePath := filepath.Join(repoRoot, ".agent", "workflow-packs", packName, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return "", fmt.Errorf("loading workflow pack %s: %w", packName, err)
	}
	return string(data), nil
}

// buildAgentPrompt constructs a prompt from workflow pack context and task input.
func buildAgentPrompt(packContext string, input TaskInput) string {
	var parts []string

	if packContext != "" {
		parts = append(parts, "## Workflow Pack Instructions\n\n"+packContext)
	}

	parts = append(parts, fmt.Sprintf("## Task\n\n%s", input.Intent))

	if len(input.Files) > 0 {
		parts = append(parts, fmt.Sprintf("## File Scope\n\nOnly modify these files: %s", strings.Join(input.Files, ", ")))
	}

	if input.Module != "" {
		parts = append(parts, fmt.Sprintf("## Module Scope\n\nOnly work within module: %s", input.Module))
	}

	if len(input.Constraints) > 0 {
		var cs []string
		for k, v := range input.Constraints {
			cs = append(cs, fmt.Sprintf("- %s: %s", k, v))
		}
		parts = append(parts, "## Constraints\n\n"+strings.Join(cs, "\n"))
	}

	parts = append(parts, "## Output Requirements\n\nOnly modify files in scope. Report what you changed and what remains unresolved.")

	return strings.Join(parts, "\n\n")
}

// extractFilePaths finds file paths mentioned in agent output.
var filePathRegex = regexp.MustCompile(`(?m)(?:^|\s)((?:[a-zA-Z0-9_\-]+/)*[a-zA-Z0-9_\-]+\.[a-zA-Z0-9]+)`)

func extractFilePaths(output string) []string {
	matches := filePathRegex.FindAllStringSubmatch(output, -1)
	seen := map[string]bool{}
	var paths []string
	for _, m := range matches {
		p := strings.TrimSpace(m[1])
		if !seen[p] && looksLikeSourceFile(p) {
			seen[p] = true
			paths = append(paths, p)
		}
	}
	return paths
}

func looksLikeSourceFile(path string) bool {
	ext := filepath.Ext(path)
	sourceExts := map[string]bool{
		".go": true, ".rs": true, ".ts": true, ".tsx": true, ".js": true, ".jsx": true,
		".py": true, ".java": true, ".kt": true, ".cs": true, ".rb": true, ".php": true,
		".swift": true, ".c": true, ".cpp": true, ".h": true, ".hpp": true,
		".yaml": true, ".yml": true, ".json": true, ".toml": true, ".md": true,
		".sh": true, ".bash": true,
	}
	return sourceExts[ext]
}

// extractUnresolved finds unresolved items mentioned in agent output.
func extractUnresolved(output string) []string {
	var items []string
	markers := []string{"TODO:", "FIXME:", "UNRESOLVED:", "BLOCKED:", "COULD NOT:", "FAILED:"}
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		for _, marker := range markers {
			if strings.Contains(strings.ToUpper(trimmed), marker) {
				items = append(items, trimmed)
				break
			}
		}
	}
	return items
}

// truncate limits a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
