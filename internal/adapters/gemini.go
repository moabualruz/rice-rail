package adapters

import (
	"context"
	"fmt"

	"github.com/mkh/rice-railing/internal/exec"
)

// GeminiAdapter implements AgentAdapter by shelling out to the Gemini CLI.
type GeminiAdapter struct {
	runner       *exec.Runner
	repoRoot     string
	packContext  string
}

func NewGeminiAdapter(runner *exec.Runner, repoRoot string) *GeminiAdapter {
	return &GeminiAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *GeminiAdapter) Name() string { return "gemini" }

func (a *GeminiAdapter) Capabilities() AgentCapabilities {
	return AgentCapabilities{
		MCP:              true,
		LocalTools:       true,
		StructuredOutput: false,
		NonInteractive:   true,
		PatchPreview:     false,
		ScopeConstraints: false,
	}
}

func (a *GeminiAdapter) LoadWorkflowPack(ctx context.Context, packName string) error {
	content, err := loadPackReadme(a.repoRoot, packName)
	if err != nil {
		return err
	}
	a.packContext = content
	return nil
}

func (a *GeminiAdapter) RunTask(ctx context.Context, input TaskInput) (*TaskResult, error) {
	prompt := buildAgentPrompt(a.packContext, input)

	result, err := a.runner.Run(ctx, "gemini", "-p", prompt)
	if err != nil {
		return nil, fmt.Errorf("gemini: %w", err)
	}

	return &TaskResult{
		Success:      result.ExitCode == 0,
		FilesChanged: extractFilePaths(result.Stdout),
		Summary:      truncate(result.Stdout, 2000),
		Unresolved:   extractUnresolved(result.Stdout),
	}, nil
}
