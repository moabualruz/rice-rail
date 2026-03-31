package builder

import (
	"strings"
	"testing"

	"github.com/mkh/rice-railing/internal/constitution"
)

func TestRenderGitHubActionsGo(t *testing.T) {
	c := &constitution.Constitution{
		Project: constitution.ProjectInfo{
			Languages: []string{"go"},
		},
		Quality: constitution.QualitySpec{
			BlockOn: []string{"lint", "tests", "typecheck"},
		},
	}

	content, err := RenderGitHubActionsWorkflow(c)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(content, "setup-go") {
		t.Error("expected setup-go step for Go project")
	}
	if !strings.Contains(content, "golangci-lint") {
		t.Error("expected golangci-lint step")
	}
	if !strings.Contains(content, "go test") {
		t.Error("expected go test step")
	}
	if strings.Contains(content, "setup-node") {
		t.Error("should not have node setup for Go-only project")
	}
}

func TestRenderGitHubActionsTypeScript(t *testing.T) {
	c := &constitution.Constitution{
		Project: constitution.ProjectInfo{
			Languages: []string{"typescript"},
		},
		Quality: constitution.QualitySpec{
			BlockOn: []string{"lint", "tests", "typecheck"},
		},
	}

	content, err := RenderGitHubActionsWorkflow(c)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(content, "setup-node") {
		t.Error("expected setup-node step")
	}
	if !strings.Contains(content, "tsc --noEmit") {
		t.Error("expected tsc typecheck step")
	}
	if !strings.Contains(content, "npm test") {
		t.Error("expected npm test step")
	}
}

func TestRenderGitHubActionsMultiLang(t *testing.T) {
	c := &constitution.Constitution{
		Project: constitution.ProjectInfo{
			Languages: []string{"go", "python"},
		},
		Quality: constitution.QualitySpec{
			BlockOn: []string{"lint", "tests"},
		},
	}

	content, err := RenderGitHubActionsWorkflow(c)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(content, "setup-go") {
		t.Error("expected setup-go")
	}
	if !strings.Contains(content, "setup-python") {
		t.Error("expected setup-python")
	}
	if !strings.Contains(content, "ruff check") {
		t.Error("expected ruff check for Python")
	}
	if !strings.Contains(content, "pytest") {
		t.Error("expected pytest for Python")
	}
}

func TestRenderGitHubActionsNoChecks(t *testing.T) {
	c := &constitution.Constitution{
		Project: constitution.ProjectInfo{
			Languages: []string{"go"},
		},
		Quality: constitution.QualitySpec{
			BlockOn: []string{},
		},
	}

	content, err := RenderGitHubActionsWorkflow(c)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if strings.Contains(content, "golangci-lint") {
		t.Error("should not have lint step when lint not in BlockOn")
	}
}
