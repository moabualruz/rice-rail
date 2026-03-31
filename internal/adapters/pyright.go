package adapters

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mkh/rice-railing/internal/exec"
)

// PyrightAdapter wraps Pyright for Python type checking.
type PyrightAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewPyrightAdapter(runner *exec.Runner, repoRoot string) *PyrightAdapter {
	return &PyrightAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *PyrightAdapter) Name() string                { return "pyright" }
func (a *PyrightAdapter) SupportedLanguages() []string { return []string{"python"} }

func (a *PyrightAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	result, err := a.runner.Run(ctx, "pyright", "--outputjson")
	if err != nil {
		// Fallback to text output parsing
		return a.checkText(ctx)
	}

	if result.ExitCode == 0 {
		return nil, nil
	}

	// Parse text output (pyright --outputjson may not be available in all versions)
	return parsePyrightText(result.Stdout + result.Stderr), nil
}

func (a *PyrightAdapter) checkText(ctx context.Context) ([]Violation, error) {
	result, err := a.runner.Run(ctx, "pyright")
	if err != nil {
		return nil, fmt.Errorf("pyright: %w", err)
	}
	if result.ExitCode == 0 {
		return nil, nil
	}
	return parsePyrightText(result.Stdout + result.Stderr), nil
}

// parsePyrightText parses pyright text output: file.py:line:col - error: message (rule)
func parsePyrightText(output string) []Violation {
	var violations []Violation
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		// Format: file.py:10:5 - error: message (ruleId)
		dashIdx := strings.Index(line, " - ")
		if dashIdx < 0 {
			continue
		}
		loc := line[:dashIdx]
		rest := line[dashIdx+3:]

		parts := strings.SplitN(loc, ":", 3)
		if len(parts) < 2 {
			continue
		}
		file := parts[0]
		lineNum, _ := strconv.Atoi(parts[1])

		severity := "BLOCKING"
		if strings.HasPrefix(rest, "warning") {
			severity = "WARNING"
		} else if strings.HasPrefix(rest, "information") {
			severity = "INFO"
		}

		violations = append(violations, Violation{
			RuleID:   "pyright",
			Severity: severity,
			File:     file,
			Line:     lineNum,
			Message:  rest,
			FixKind:  "NONE",
		})
	}
	return violations
}

// MypyAdapter wraps mypy for Python type checking.
type MypyAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewMypyAdapter(runner *exec.Runner, repoRoot string) *MypyAdapter {
	return &MypyAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *MypyAdapter) Name() string                { return "mypy" }
func (a *MypyAdapter) SupportedLanguages() []string { return []string{"python"} }

func (a *MypyAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	result, err := a.runner.Run(ctx, "mypy", ".")
	if err != nil {
		return nil, fmt.Errorf("mypy: %w", err)
	}
	if result.ExitCode == 0 {
		return nil, nil
	}

	// Parse: file.py:line: error: message [error-code]
	var violations []Violation
	scanner := bufio.NewScanner(strings.NewReader(result.Stdout))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ": ", 3)
		if len(parts) < 3 {
			continue
		}
		locParts := strings.SplitN(parts[0], ":", 2)
		if len(locParts) < 2 {
			continue
		}
		file := locParts[0]
		lineNum, _ := strconv.Atoi(locParts[1])

		severity := "BLOCKING"
		if strings.HasPrefix(parts[1], "note") {
			severity = "INFO"
		}

		violations = append(violations, Violation{
			RuleID:   "mypy",
			Severity: severity,
			File:     file,
			Line:     lineNum,
			Message:  parts[1] + ": " + parts[2],
			FixKind:  "NONE",
		})
	}
	return violations, nil
}
