package adapters

import (
	"context"
	"testing"

	"github.com/mkh/rice-railing/internal/exec"
)

func TestClaudeCodeAdapter_Name(t *testing.T) {
	a := NewClaudeCodeAdapter(exec.NewRunner(), "/tmp")
	if got := a.Name(); got != "claude-code" {
		t.Errorf("Name() = %q, want %q", got, "claude-code")
	}
}

func TestClaudeCodeAdapter_Capabilities(t *testing.T) {
	a := NewClaudeCodeAdapter(exec.NewRunner(), "/tmp")
	caps := a.Capabilities()
	if !caps.MCP {
		t.Error("expected MCP=true")
	}
	if !caps.LocalTools {
		t.Error("expected LocalTools=true")
	}
	if !caps.StructuredOutput {
		t.Error("expected StructuredOutput=true")
	}
	if !caps.NonInteractive {
		t.Error("expected NonInteractive=true")
	}
	if !caps.PatchPreview {
		t.Error("expected PatchPreview=true")
	}
	if !caps.ScopeConstraints {
		t.Error("expected ScopeConstraints=true")
	}
}

func TestAiderAdapter_Name(t *testing.T) {
	a := NewAiderAdapter(exec.NewRunner(), "/tmp")
	if got := a.Name(); got != "aider" {
		t.Errorf("Name() = %q, want %q", got, "aider")
	}
}

func TestAiderAdapter_Capabilities(t *testing.T) {
	a := NewAiderAdapter(exec.NewRunner(), "/tmp")
	caps := a.Capabilities()
	if caps.MCP {
		t.Error("expected MCP=false")
	}
	if caps.LocalTools {
		t.Error("expected LocalTools=false")
	}
	if caps.StructuredOutput {
		t.Error("expected StructuredOutput=false")
	}
	if !caps.NonInteractive {
		t.Error("expected NonInteractive=true")
	}
	if !caps.PatchPreview {
		t.Error("expected PatchPreview=true")
	}
	if !caps.ScopeConstraints {
		t.Error("expected ScopeConstraints=true")
	}
}

func TestCodexAdapter_Name(t *testing.T) {
	a := NewCodexAdapter(exec.NewRunner(), "/tmp")
	if got := a.Name(); got != "codex" {
		t.Errorf("Name() = %q, want %q", got, "codex")
	}
}

func TestCodexAdapter_Capabilities(t *testing.T) {
	a := NewCodexAdapter(exec.NewRunner(), "/tmp")
	caps := a.Capabilities()
	if caps.MCP {
		t.Error("expected MCP=false")
	}
	if !caps.LocalTools {
		t.Error("expected LocalTools=true")
	}
	if caps.StructuredOutput {
		t.Error("expected StructuredOutput=false")
	}
	if !caps.NonInteractive {
		t.Error("expected NonInteractive=true")
	}
	if caps.PatchPreview {
		t.Error("expected PatchPreview=false")
	}
	if caps.ScopeConstraints {
		t.Error("expected ScopeConstraints=false")
	}
}

func TestClaudeCodeAdapter_BuildPrompt(t *testing.T) {
	a := NewClaudeCodeAdapter(exec.NewRunner(), "/tmp")
	a.workflowPack = "Test workflow instructions"

	input := TaskInput{
		Intent: "Fix the broken tests",
		Files:  []string{"main.go", "main_test.go"},
		Module: "core",
		Constraints: map[string]string{
			"language": "go",
		},
	}

	prompt := a.buildPrompt(input)

	if !contains(prompt, "Test workflow instructions") {
		t.Error("prompt should contain workflow pack instructions")
	}
	if !contains(prompt, "Fix the broken tests") {
		t.Error("prompt should contain intent")
	}
	if !contains(prompt, "main.go") {
		t.Error("prompt should contain file constraints")
	}
	if !contains(prompt, "main_test.go") {
		t.Error("prompt should contain all file constraints")
	}
	if !contains(prompt, "core") {
		t.Error("prompt should contain module constraint")
	}
	if !contains(prompt, "Only modify files in scope") {
		t.Error("prompt should contain scope suffix")
	}
}

func TestAiderAdapter_BuildPrompt(t *testing.T) {
	a := NewAiderAdapter(exec.NewRunner(), "/tmp")

	input := TaskInput{
		Intent: "Refactor handler",
		Module: "api",
	}

	prompt := a.buildPrompt(input)

	if !contains(prompt, "Refactor handler") {
		t.Error("prompt should contain intent")
	}
	if !contains(prompt, "api") {
		t.Error("prompt should contain module")
	}
}

func TestCodexAdapter_BuildPrompt(t *testing.T) {
	a := NewCodexAdapter(exec.NewRunner(), "/tmp")

	input := TaskInput{
		Intent: "Add logging",
		Files:  []string{"server.go"},
	}

	prompt := a.buildPrompt(input)

	if !contains(prompt, "Add logging") {
		t.Error("prompt should contain intent")
	}
	if !contains(prompt, "server.go") {
		t.Error("prompt should contain files")
	}
}

func TestClaudeCodeAdapter_RunTask_DryRun(t *testing.T) {
	runner := exec.NewRunner()
	runner.DryRun = true

	a := NewClaudeCodeAdapter(runner, "/tmp")
	input := TaskInput{
		Intent: "Fix compilation error in main.go",
		Files:  []string{"main.go"},
	}

	result, err := a.RunTask(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("dry-run should succeed")
	}
}

func TestAiderAdapter_RunTask_DryRun(t *testing.T) {
	runner := exec.NewRunner()
	runner.DryRun = true

	a := NewAiderAdapter(runner, "/tmp")
	input := TaskInput{
		Intent: "Add error handling",
		Files:  []string{"handler.go"},
	}

	result, err := a.RunTask(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("dry-run should succeed")
	}
}

func TestCodexAdapter_RunTask_DryRun(t *testing.T) {
	runner := exec.NewRunner()
	runner.DryRun = true

	a := NewCodexAdapter(runner, "/tmp")
	input := TaskInput{
		Intent: "Optimize query",
	}

	result, err := a.RunTask(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("dry-run should succeed")
	}
}

func TestParseClaudeOutput_FilePaths(t *testing.T) {
	r := &exec.Result{
		Stdout: "Modified internal/adapters/claude_code.go\nAlso changed cmd/main.go\nUnresolved: could not fix broken import",
	}

	tr := parseClaudeOutput(r)

	if len(tr.FilesChanged) == 0 {
		t.Error("expected files to be extracted from output")
	}
	if len(tr.Unresolved) == 0 {
		t.Error("expected unresolved items to be detected")
	}
}

func TestParseClaudeOutput_ExitCodeFailure(t *testing.T) {
	r := &exec.Result{
		ExitCode: 1,
		Stdout:   "Something went wrong",
	}

	tr := parseClaudeOutput(r)
	if tr.Success {
		t.Error("expected Success=false for non-zero exit code")
	}
}

func TestLooksLikeFilePath(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"main.go", true},
		{"internal/adapters/claude_code.go", true},
		{"./cmd/root.go", true},
		{"https://example.com/foo.go", false},
		{".go", false},
		{"abc", false},
	}

	for _, tt := range tests {
		if got := looksLikeFilePath(tt.input); got != tt.want {
			t.Errorf("looksLikeFilePath(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestAllAdaptersImplementInterface(t *testing.T) {
	runner := exec.NewRunner()
	var _ AgentAdapter = NewClaudeCodeAdapter(runner, "/tmp")
	var _ AgentAdapter = NewAiderAdapter(runner, "/tmp")
	var _ AgentAdapter = NewCodexAdapter(runner, "/tmp")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
