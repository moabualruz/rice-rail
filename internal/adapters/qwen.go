package adapters

import (
	"context"
	"fmt"

	"github.com/mkh/rice-railing/internal/exec"
)

// QwenAdapter implements AgentAdapter by shelling out to the Qwen Code CLI.
type QwenAdapter struct {
	runner      *exec.Runner
	repoRoot    string
	packContext string
}

func NewQwenAdapter(runner *exec.Runner, repoRoot string) *QwenAdapter {
	return &QwenAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *QwenAdapter) Name() string { return "qwen" }

func (a *QwenAdapter) Capabilities() AgentCapabilities {
	return AgentCapabilities{
		MCP:              true,
		LocalTools:       true,
		StructuredOutput: false,
		NonInteractive:   true,
		PatchPreview:     false,
		ScopeConstraints: false,
	}
}

func (a *QwenAdapter) LoadWorkflowPack(ctx context.Context, packName string) error {
	content, err := loadPackReadme(a.repoRoot, packName)
	if err != nil {
		return err
	}
	a.packContext = content
	return nil
}

func (a *QwenAdapter) RunTask(ctx context.Context, input TaskInput) (*TaskResult, error) {
	prompt := buildAgentPrompt(a.packContext, input)

	result, err := a.runner.Run(ctx, "qwen", "-p", prompt)
	if err != nil {
		return nil, fmt.Errorf("qwen: %w", err)
	}

	return &TaskResult{
		Success:      result.ExitCode == 0,
		FilesChanged: extractFilePaths(result.Stdout),
		Summary:      truncate(result.Stdout, 2000),
		Unresolved:   extractUnresolved(result.Stdout),
	}, nil
}
