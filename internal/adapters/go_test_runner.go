package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mkh/rice-railing/internal/exec"
)

// GoTestAdapter wraps `go test` for running Go tests.
type GoTestAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewGoTestAdapter(runner *exec.Runner, repoRoot string) *GoTestAdapter {
	return &GoTestAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *GoTestAdapter) Name() string                 { return "go-test" }
func (a *GoTestAdapter) SupportedLanguages() []string { return []string{"go"} }

// goTestEvent matches `go test -json` output lines.
type goTestEvent struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Output  string  `json:"Output"`
	Elapsed float64 `json:"Elapsed"`
}

func (a *GoTestAdapter) Run(ctx context.Context, targets []string) (*TestResult, error) {
	args := []string{"test", "-json", "-count=1"}
	if len(targets) > 0 && targets[0] != "." {
		args = append(args, targets...)
	} else {
		args = append(args, "./...")
	}

	result, err := a.runner.Run(ctx, "go", args...)
	if err != nil {
		return nil, fmt.Errorf("go test: %w", err)
	}

	tr := &TestResult{Output: result.Stderr}
	passed := 0
	failed := 0

	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var event goTestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		if event.Test == "" {
			continue // package-level event
		}
		switch event.Action {
		case "pass":
			passed++
		case "fail":
			failed++
		}
	}

	tr.Passed = passed
	tr.Failed = failed
	tr.Total = passed + failed

	return tr, nil
}
