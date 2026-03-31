package adapters

import (
	"context"
	"fmt"

	"github.com/mkh/rice-railing/internal/exec"
)

// CopilotAdapter implements AgentAdapter by shelling out to GitHub Copilot CLI.
type CopilotAdapter struct {
	runner      *exec.Runner
	repoRoot    string
	packContext string
}

func NewCopilotAdapter(runner *exec.Runner, repoRoot string) *CopilotAdapter {
	return &CopilotAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *CopilotAdapter) Name() string { return "copilot" }

func (a *CopilotAdapter) Capabilities() AgentCapabilities {
	return AgentCapabilities{
		MCP:              true,
		LocalTools:       true,
		StructuredOutput: false,
		NonInteractive:   true,
		PatchPreview:     false,
		ScopeConstraints: true,
	}
}

func (a *CopilotAdapter) LoadWorkflowPack(ctx context.Context, packName string) error {
	content, err := loadPackReadme(a.repoRoot, packName)
	if err != nil {
		return err
	}
	a.packContext = content
	return nil
}

func (a *CopilotAdapter) RunTask(ctx context.Context, input TaskInput) (*TaskResult, error) {
	prompt := buildAgentPrompt(a.packContext, input)

	result, err := a.runner.Run(ctx, "copilot", "-p", prompt, "--allow-all-tools", "--allow-all-paths")
	if err != nil {
		return nil, fmt.Errorf("copilot: %w", err)
	}

	return &TaskResult{
		Success:      result.ExitCode == 0,
		FilesChanged: extractFilePaths(result.Stdout),
		Summary:      truncate(result.Stdout, 2000),
		Unresolved:   extractUnresolved(result.Stdout),
	}, nil
}
