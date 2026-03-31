package adapters

import (
	"context"
	"fmt"

	"github.com/mkh/rice-railing/internal/exec"
)

// OllamaAdapter implements AgentAdapter by shelling out to Ollama for local LLM inference.
// Ollama is a model runner, not a coding agent — it generates text responses
// but cannot edit files or run commands. It's useful for semantic issue analysis,
// code review suggestions, and intent parsing.
type OllamaAdapter struct {
	runner      *exec.Runner
	repoRoot    string
	model       string
	packContext string
}

// NewOllamaAdapter creates an Ollama adapter. Defaults to qwen3-coder:30b if available.
func NewOllamaAdapter(runner *exec.Runner, repoRoot string) *OllamaAdapter {
	model := detectOllamaModel(runner)
	return &OllamaAdapter{runner: runner, repoRoot: repoRoot, model: model}
}

func (a *OllamaAdapter) Name() string { return "ollama" }

func (a *OllamaAdapter) Capabilities() AgentCapabilities {
	return AgentCapabilities{
		MCP:              false,
		LocalTools:       false,
		StructuredOutput: false,
		NonInteractive:   true,
		PatchPreview:     false,
		ScopeConstraints: false,
	}
}

func (a *OllamaAdapter) LoadWorkflowPack(ctx context.Context, packName string) error {
	content, err := loadPackReadme(a.repoRoot, packName)
	if err != nil {
		return err
	}
	a.packContext = content
	return nil
}

func (a *OllamaAdapter) RunTask(ctx context.Context, input TaskInput) (*TaskResult, error) {
	if a.model == "" {
		return nil, fmt.Errorf("ollama: no coding model available (install qwen3-coder or deepseek-coder-v2)")
	}

	prompt := buildAgentPrompt(a.packContext, input)

	result, err := a.runner.Run(ctx, "ollama", "run", a.model, prompt)
	if err != nil {
		return nil, fmt.Errorf("ollama: %w", err)
	}

	return &TaskResult{
		Success:      result.ExitCode == 0,
		FilesChanged: nil, // Ollama can't edit files directly
		Summary:      truncate(result.Stdout, 2000),
		Unresolved:   extractUnresolved(result.Stdout),
	}, nil
}

// detectOllamaModel picks the best available coding model from ollama list.
func detectOllamaModel(runner *exec.Runner) string {
	result, err := runner.Run(context.Background(), "ollama", "list")
	if err != nil {
		return ""
	}

	// Prefer coding models in order of capability
	preferences := []string{
		"qwen3-coder:30b",
		"qwen3-coder",
		"deepseek-coder-v2",
		"qwen2.5-coder:14b",
		"qwen2.5-coder",
		"qwen3.5:27b",
		"qwen3.5:9b",
		"codellama",
	}

	for _, pref := range preferences {
		if containsModel(result.Stdout, pref) {
			return pref
		}
	}

	return ""
}

func containsModel(output, model string) bool {
	for _, line := range splitLines(output) {
		if len(line) > 0 && lineStartsWith(line, model) {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func lineStartsWith(line, prefix string) bool {
	return len(line) >= len(prefix) && line[:len(prefix)] == prefix
}
