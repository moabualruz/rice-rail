package adapters

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/exec"
)

// CustomAdapter wraps any user-defined CLI tool as a rice-rail adapter.
// Reads its configuration from constitution.yaml custom_tools entries.
type CustomAdapter struct {
	runner *exec.Runner
	def    constitution.CustomTool
}

func NewCustomAdapter(runner *exec.Runner, def constitution.CustomTool) *CustomAdapter {
	return &CustomAdapter{runner: runner, def: def}
}

func (a *CustomAdapter) Name() string                { return a.def.Name }
func (a *CustomAdapter) SupportedLanguages() []string { return a.def.Languages }

// Check runs the tool's check command and parses output into violations.
func (a *CustomAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	if len(a.def.CheckCmd) == 0 {
		return nil, nil // no check mode defined
	}

	args := expandTargets(a.def.CheckCmd, targets)
	result, err := a.runner.Run(ctx, a.def.Binary, args...)
	if err != nil {
		return nil, fmt.Errorf("custom tool %s: %w", a.def.Name, err)
	}

	successExit := a.def.SuccessExit // default 0
	if result.ExitCode == successExit {
		return nil, nil
	}

	return a.parseOutput(result.Stdout + result.Stderr), nil
}

// Fix runs the tool's fix command.
func (a *CustomAdapter) Fix(ctx context.Context, targets []string) ([]FixResult, error) {
	if len(a.def.FixCmd) == 0 {
		return nil, nil // no fix mode defined
	}

	args := expandTargets(a.def.FixCmd, targets)
	result, err := a.runner.Run(ctx, a.def.Binary, args...)
	if err != nil {
		return nil, fmt.Errorf("custom tool %s fix: %w", a.def.Name, err)
	}

	if result.ExitCode != 0 {
		return []FixResult{{
			RuleID: a.def.Name,
			Action: "failed",
			Detail: truncate(result.Stderr, 500),
		}}, nil
	}

	return []FixResult{{
		RuleID: a.def.Name,
		Action: "applied",
		Detail: fmt.Sprintf("ran %s %s", a.def.Binary, strings.Join(a.def.FixCmd, " ")),
	}}, nil
}

// RunTest runs the tool's test command (for test_runner role).
func (a *CustomAdapter) RunTest(ctx context.Context, targets []string) (*TestResult, error) {
	if len(a.def.TestCmd) == 0 {
		return nil, nil
	}

	args := expandTargets(a.def.TestCmd, targets)
	result, err := a.runner.Run(ctx, a.def.Binary, args...)
	if err != nil {
		return nil, fmt.Errorf("custom tool %s test: %w", a.def.Name, err)
	}

	tr := &TestResult{Output: result.Stdout + result.Stderr}
	if result.ExitCode == 0 {
		tr.Passed = 1
	} else {
		tr.Failed = 1
	}
	tr.Total = tr.Passed + tr.Failed

	return tr, nil
}

// --- CustomTestRunnerAdapter wraps CustomAdapter to implement TestRunnerAdapter ---

type CustomTestRunnerAdapter struct {
	inner *CustomAdapter
}

func (a *CustomTestRunnerAdapter) Name() string                { return a.inner.Name() }
func (a *CustomTestRunnerAdapter) SupportedLanguages() []string { return a.inner.SupportedLanguages() }
func (a *CustomTestRunnerAdapter) Run(ctx context.Context, targets []string) (*TestResult, error) {
	return a.inner.RunTest(ctx, targets)
}

// --- CustomTypecheckAdapter wraps CustomAdapter to implement TypecheckAdapter ---

type CustomTypecheckAdapter struct {
	inner *CustomAdapter
}

func (a *CustomTypecheckAdapter) Name() string                { return a.inner.Name() }
func (a *CustomTypecheckAdapter) SupportedLanguages() []string { return a.inner.SupportedLanguages() }
func (a *CustomTypecheckAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	return a.inner.Check(ctx, targets)
}

// --- Helpers ---

// expandTargets replaces {targets} placeholder in args with actual target paths.
// If no placeholder, appends targets at the end.
func expandTargets(args []string, targets []string) []string {
	expanded := make([]string, 0, len(args)+len(targets))
	replaced := false
	for _, arg := range args {
		if arg == "{targets}" {
			expanded = append(expanded, targets...)
			replaced = true
		} else if strings.Contains(arg, "{targets}") {
			expanded = append(expanded, strings.ReplaceAll(arg, "{targets}", strings.Join(targets, " ")))
			replaced = true
		} else {
			expanded = append(expanded, arg)
		}
	}
	if !replaced && len(targets) > 0 && targets[0] != "." {
		expanded = append(expanded, targets...)
	}
	return expanded
}

// parseOutput extracts violations from tool output based on configured format.
func (a *CustomAdapter) parseOutput(output string) []Violation {
	switch a.def.OutputFmt {
	case "json":
		return a.parseJSON(output)
	case "sarif":
		return a.parseSARIF(output)
	default:
		return a.parseText(output)
	}
}

// parseText extracts violations from line-based output.
// Supports common formats: file:line:col: message, file:line: message, file(line,col): message
func (a *CustomAdapter) parseText(output string) []Violation {
	var violations []Violation
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Try file:line:col: message
		v := parseColonFormat(line, a.def.Name)
		if v != nil {
			violations = append(violations, *v)
			continue
		}

		// Try file(line,col): message
		v = parseParenFormat(line, a.def.Name)
		if v != nil {
			violations = append(violations, *v)
		}
	}
	return violations
}

func parseColonFormat(line, ruleID string) *Violation {
	// file.go:10:5: message
	// file.go:10: message
	parts := strings.SplitN(line, ": ", 2)
	if len(parts) != 2 {
		return nil
	}
	locParts := strings.Split(parts[0], ":")
	if len(locParts) < 2 {
		return nil
	}
	file := locParts[0]
	lineNum, err := strconv.Atoi(locParts[1])
	if err != nil {
		return nil
	}
	return &Violation{
		RuleID: ruleID, Severity: "BLOCKING", File: file, Line: lineNum,
		Message: parts[1], FixKind: "NONE",
	}
}

func parseParenFormat(line, ruleID string) *Violation {
	// file.ts(10,5): error message
	parenIdx := strings.Index(line, "(")
	closeIdx := strings.Index(line, ")")
	if parenIdx < 0 || closeIdx < 0 || closeIdx <= parenIdx {
		return nil
	}
	file := line[:parenIdx]
	locStr := line[parenIdx+1 : closeIdx]
	msg := strings.TrimPrefix(line[closeIdx+1:], ": ")
	msg = strings.TrimPrefix(msg, " ")

	locParts := strings.Split(locStr, ",")
	lineNum := 0
	if len(locParts) >= 1 {
		lineNum, _ = strconv.Atoi(strings.TrimSpace(locParts[0]))
	}

	return &Violation{
		RuleID: ruleID, Severity: "BLOCKING", File: file, Line: lineNum,
		Message: msg, FixKind: "NONE",
	}
}

// parseJSON tries to extract violations from JSON array output.
// Supports: [{file, line, message, severity}] or {results: [...]} or {diagnostics: [...]}
func (a *CustomAdapter) parseJSON(output string) []Violation {
	// Try array of objects
	var items []struct {
		File     string `json:"file"`
		Path     string `json:"path"`
		Line     int    `json:"line"`
		Message  string `json:"message"`
		Severity string `json:"severity"`
		Rule     string `json:"rule"`
		RuleID   string `json:"ruleId"`
	}
	if json.Unmarshal([]byte(output), &items) == nil {
		return jsonItemsToViolations(items, a.def.Name)
	}

	// Try {results: [...]}
	var wrapped struct {
		Results []struct {
			File     string `json:"file"`
			Path     string `json:"path"`
			Line     int    `json:"line"`
			Message  string `json:"message"`
			Severity string `json:"severity"`
			Rule     string `json:"rule"`
			RuleID   string `json:"ruleId"`
		} `json:"results"`
	}
	if json.Unmarshal([]byte(output), &wrapped) == nil && len(wrapped.Results) > 0 {
		return jsonItemsToViolations(wrapped.Results, a.def.Name)
	}

	// Try {diagnostics: [...]}
	var diag struct {
		Diagnostics []struct {
			File     string `json:"file"`
			Path     string `json:"path"`
			Line     int    `json:"line"`
			Message  string `json:"message"`
			Severity string `json:"severity"`
			Rule     string `json:"rule"`
			RuleID   string `json:"ruleId"`
		} `json:"diagnostics"`
	}
	if json.Unmarshal([]byte(output), &diag) == nil && len(diag.Diagnostics) > 0 {
		return jsonItemsToViolations(diag.Diagnostics, a.def.Name)
	}

	return nil
}

type jsonItem struct {
	File, Path, Message, Severity, Rule, RuleID string
	Line                                         int
}

func jsonItemsToViolations[T any](items []T, defaultRule string) []Violation {
	// Re-marshal and unmarshal to normalize field access
	data, _ := json.Marshal(items)
	var normalized []struct {
		File     string `json:"file"`
		Path     string `json:"path"`
		Line     int    `json:"line"`
		Message  string `json:"message"`
		Severity string `json:"severity"`
		Rule     string `json:"rule"`
		RuleID   string `json:"ruleId"`
	}
	json.Unmarshal(data, &normalized)

	var violations []Violation
	for _, item := range normalized {
		file := item.File
		if file == "" {
			file = item.Path
		}
		rule := item.RuleID
		if rule == "" {
			rule = item.Rule
		}
		if rule == "" {
			rule = defaultRule
		}
		severity := "BLOCKING"
		if item.Severity == "warning" || item.Severity == "warn" {
			severity = "WARNING"
		} else if item.Severity == "info" || item.Severity == "information" {
			severity = "INFO"
		}

		violations = append(violations, Violation{
			RuleID:   rule,
			Severity: severity,
			File:     file,
			Line:     item.Line,
			Message:  item.Message,
			FixKind:  "NONE",
		})
	}
	return violations
}

// parseSARIF extracts violations from SARIF format.
func (a *CustomAdapter) parseSARIF(output string) []Violation {
	var sarif struct {
		Runs []struct {
			Results []struct {
				RuleID  string `json:"ruleId"`
				Level   string `json:"level"` // error, warning, note
				Message struct {
					Text string `json:"text"`
				} `json:"message"`
				Locations []struct {
					PhysicalLocation struct {
						ArtifactLocation struct {
							URI string `json:"uri"`
						} `json:"artifactLocation"`
						Region struct {
							StartLine int `json:"startLine"`
						} `json:"region"`
					} `json:"physicalLocation"`
				} `json:"locations"`
			} `json:"results"`
		} `json:"runs"`
	}

	if err := json.Unmarshal([]byte(output), &sarif); err != nil {
		return nil
	}

	var violations []Violation
	for _, run := range sarif.Runs {
		for _, r := range run.Results {
			severity := "BLOCKING"
			if r.Level == "warning" {
				severity = "WARNING"
			} else if r.Level == "note" {
				severity = "INFO"
			}

			file := ""
			lineNum := 0
			if len(r.Locations) > 0 {
				file = r.Locations[0].PhysicalLocation.ArtifactLocation.URI
				lineNum = r.Locations[0].PhysicalLocation.Region.StartLine
			}

			violations = append(violations, Violation{
				RuleID:   r.RuleID,
				Severity: severity,
				File:     file,
				Line:     lineNum,
				Message:  r.Message.Text,
				FixKind:  "NONE",
			})
		}
	}
	return violations
}
