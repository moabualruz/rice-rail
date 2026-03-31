package adapters

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mkh/rice-railing/internal/exec"
)

// GoVetAdapter wraps `go vet` for static analysis of Go code.
type GoVetAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewGoVetAdapter(runner *exec.Runner, repoRoot string) *GoVetAdapter {
	return &GoVetAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *GoVetAdapter) Name() string                 { return "go-vet" }
func (a *GoVetAdapter) SupportedLanguages() []string { return []string{"go"} }

func (a *GoVetAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	args := []string{"vet"}
	if len(targets) > 0 && targets[0] != "." {
		args = append(args, targets...)
	} else {
		args = append(args, "./...")
	}

	result, err := a.runner.Run(ctx, "go", args...)
	if err != nil {
		return nil, fmt.Errorf("go vet: %w", err)
	}

	if result.ExitCode == 0 {
		return nil, nil
	}

	// Parse stderr: file.go:line:col: message
	var violations []Violation
	scanner := bufio.NewScanner(strings.NewReader(result.Stderr))
	for scanner.Scan() {
		line := scanner.Text()
		v := parseGoVetLine(line)
		if v != nil {
			violations = append(violations, *v)
		}
	}

	return violations, nil
}

func (a *GoVetAdapter) Fix(ctx context.Context, targets []string) ([]FixResult, error) {
	// go vet has no --fix mode
	return nil, nil
}

func parseGoVetLine(line string) *Violation {
	// Format: path/file.go:line:col: message
	// or: path/file.go:line: message
	parts := strings.SplitN(line, ": ", 2)
	if len(parts) != 2 {
		return nil
	}

	loc := parts[0]
	msg := parts[1]

	// Split location into file:line[:col]
	locParts := strings.Split(loc, ":")
	if len(locParts) < 2 {
		return nil
	}

	file := locParts[0]
	lineNum, err := strconv.Atoi(locParts[1])
	if err != nil {
		return nil
	}

	return &Violation{
		RuleID:   "go-vet",
		Severity: "BLOCKING",
		File:     file,
		Line:     lineNum,
		Message:  msg,
		FixKind:  "NONE",
	}
}
