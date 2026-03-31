package adapters

import (
	"context"
	"fmt"

	"github.com/mkh/rice-railing/internal/exec"
)

// OpenCodeAdapter implements AgentAdapter by shelling out to the OpenCode CLI.
type OpenCodeAdapter struct {
	runner      *exec.Runner
	repoRoot    string
	packContext string
}

func NewOpenCodeAdapter(runner *exec.Runner, repoRoot string) *OpenCodeAdapter {
	return &OpenCodeAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *OpenCodeAdapter) Name() string { return "opencode" }

func (a *OpenCodeAdapter) Capabilities() AgentCapabilities {
	return AgentCapabilities{
		MCP:              true,
		LocalTools:       true,
		StructuredOutput: false,
		NonInteractive:   true,
		PatchPreview:     false,
		ScopeConstraints: false,
	}
}

func (a *OpenCodeAdapter) LoadWorkflowPack(ctx context.Context, packName string) error {
	content, err := loadPackReadme(a.repoRoot, packName)
	if err != nil {
		return err
	}
	a.packContext = content
	return nil
}

func (a *OpenCodeAdapter) RunTask(ctx context.Context, input TaskInput) (*TaskResult, error) {
	prompt := buildAgentPrompt(a.packContext, input)

	result, err := a.runner.Run(ctx, "opencode", "run", prompt)
	if err != nil {
		return nil, fmt.Errorf("opencode: %w", err)
	}

	return &TaskResult{
		Success:      result.ExitCode == 0,
		FilesChanged: extractFilePaths(result.Stdout),
		Summary:      truncate(result.Stdout, 2000),
		Unresolved:   extractUnresolved(result.Stdout),
	}, nil
}
