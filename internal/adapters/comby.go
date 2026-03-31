package adapters

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/mkh/rice-railing/internal/exec"
)

// CombyAdapter wraps Comby for structural code transformations.
type CombyAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewCombyAdapter(runner *exec.Runner, repoRoot string) *CombyAdapter {
	return &CombyAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *CombyAdapter) Name() string                 { return "comby" }
func (a *CombyAdapter) SupportedLanguages() []string { return []string{"*"} }

// combyTemplate is the TOML schema for a Comby codemod definition.
type combyTemplate struct {
	Match   string `toml:"match"`
	Rewrite string `toml:"rewrite"`
}

// Run implements CodemodEngineAdapter. It reads a TOML template from
// .project-toolkit/codemods/local/<codemodID>.toml containing match/rewrite
// fields and invokes comby.
func (a *CombyAdapter) Run(ctx context.Context, codemodID string, targets []string, dryRun bool) (*CodemodResult, error) {
	templatePath := filepath.Join(a.repoRoot, ".project-toolkit", "codemods", "local", codemodID+".toml")

	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("comby: read template %s: %w", templatePath, err)
	}

	var tmpl combyTemplate
	if err := toml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("comby: parse template %s: %w", templatePath, err)
	}

	if tmpl.Match == "" || tmpl.Rewrite == "" {
		return nil, fmt.Errorf("comby: template %s missing match or rewrite field", templatePath)
	}

	target := "."
	if len(targets) > 0 && targets[0] != "." {
		target = targets[0]
	}

	args := []string{tmpl.Match, tmpl.Rewrite, "-d", target}
	if dryRun {
		args = append(args, "-diff")
	}

	result, err := a.runner.Run(ctx, "comby", args...)
	if err != nil {
		return nil, fmt.Errorf("comby codemod %s: %w", codemodID, err)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("comby codemod %s failed (exit %d): %s", codemodID, result.ExitCode, result.Stderr)
	}

	cr := &CodemodResult{
		CodemodID: codemodID,
		DryRun:    dryRun,
		Summary:   fmt.Sprintf("comby codemod %s applied", codemodID),
	}

	// Parse diff output to extract changed files.
	if result.Stdout != "" {
		cr.FilesChanged = parseCombyChangedFiles(result.Stdout)
	}

	return cr, nil
}

// parseCombyChangedFiles extracts file paths from comby diff output.
// Comby's -diff output uses unified diff format with "--- a/file" lines.
func parseCombyChangedFiles(output string) []string {
	seen := map[string]bool{}
	var files []string

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "--- a/") {
			path := strings.TrimPrefix(line, "--- a/")
			if !seen[path] {
				seen[path] = true
				files = append(files, path)
			}
		}
	}

	return files
}
