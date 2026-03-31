package adapters

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mkh/rice-railing/internal/exec"
)

// TscAdapter wraps the TypeScript compiler for type checking.
type TscAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewTscAdapter(runner *exec.Runner, repoRoot string) *TscAdapter {
	return &TscAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *TscAdapter) Name() string                 { return "tsc" }
func (a *TscAdapter) SupportedLanguages() []string { return []string{"typescript", "javascript"} }

func (a *TscAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	result, err := a.runner.Run(ctx, "tsc", "--noEmit", "--pretty", "false")
	if err != nil {
		return nil, fmt.Errorf("tsc: %w", err)
	}

	if result.ExitCode == 0 {
		return nil, nil
	}

	// Parse: file.ts(line,col): error TSxxxx: message
	var violations []Violation
	scanner := bufio.NewScanner(strings.NewReader(result.Stdout))
	for scanner.Scan() {
		line := scanner.Text()
		v := parseTscLine(line)
		if v != nil {
			violations = append(violations, *v)
		}
	}

	return violations, nil
}

func parseTscLine(line string) *Violation {
	// Format: file.ts(line,col): error TSxxxx: message
	parenIdx := strings.Index(line, "(")
	colonIdx := strings.Index(line, "): ")
	if parenIdx < 0 || colonIdx < 0 || colonIdx <= parenIdx {
		return nil
	}

	file := line[:parenIdx]
	locStr := line[parenIdx+1 : colonIdx]
	rest := line[colonIdx+3:]

	locParts := strings.Split(locStr, ",")
	lineNum := 0
	if len(locParts) >= 1 {
		lineNum, _ = strconv.Atoi(locParts[0])
	}

	severity := "BLOCKING"
	if strings.HasPrefix(rest, "warning") {
		severity = "WARNING"
	}

	return &Violation{
		RuleID:   "tsc",
		Severity: severity,
		File:     file,
		Line:     lineNum,
		Message:  rest,
		FixKind:  "NONE",
	}
}
