package adapters

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mkh/rice-railing/internal/exec"
)

// ClippyAdapter wraps cargo clippy for Rust linting.
type ClippyAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewClippyAdapter(runner *exec.Runner, repoRoot string) *ClippyAdapter {
	return &ClippyAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *ClippyAdapter) Name() string                { return "clippy" }
func (a *ClippyAdapter) SupportedLanguages() []string { return []string{"rust"} }

func (a *ClippyAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	result, err := a.runner.Run(ctx, "cargo", "clippy", "--message-format=json", "--", "-D", "warnings")
	if err != nil {
		return nil, fmt.Errorf("clippy: %w", err)
	}

	var violations []Violation
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var msg struct {
			Reason  string `json:"reason"`
			Message *struct {
				Code    *struct{ Code string } `json:"code"`
				Level   string                 `json:"level"`
				Message string                 `json:"message"`
				Spans   []struct {
					FileName string `json:"file_name"`
					LineStart int   `json:"line_start"`
				} `json:"spans"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &msg) != nil || msg.Message == nil {
			continue
		}
		if msg.Reason != "compiler-message" {
			continue
		}

		severity := "WARNING"
		if msg.Message.Level == "error" {
			severity = "BLOCKING"
		}

		file := ""
		lineNum := 0
		if len(msg.Message.Spans) > 0 {
			file = msg.Message.Spans[0].FileName
			lineNum = msg.Message.Spans[0].LineStart
		}

		ruleID := "clippy"
		if msg.Message.Code != nil {
			ruleID = msg.Message.Code.Code
		}

		violations = append(violations, Violation{
			RuleID:   ruleID,
			Severity: severity,
			File:     file,
			Line:     lineNum,
			Message:  msg.Message.Message,
			FixKind:  "NONE",
		})
	}
	return violations, nil
}

func (a *ClippyAdapter) Fix(ctx context.Context, targets []string) ([]FixResult, error) {
	result, err := a.runner.Run(ctx, "cargo", "clippy", "--fix", "--allow-dirty", "--allow-staged")
	if err != nil {
		return nil, fmt.Errorf("clippy fix: %w", err)
	}
	if result.ExitCode != 0 {
		return []FixResult{{RuleID: "clippy", Action: "failed", Detail: result.Stderr}}, nil
	}
	return []FixResult{{RuleID: "clippy", Action: "applied", Detail: "ran cargo clippy --fix"}}, nil
}

// RustfmtAdapter wraps rustfmt for Rust formatting.
type RustfmtAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewRustfmtAdapter(runner *exec.Runner, repoRoot string) *RustfmtAdapter {
	return &RustfmtAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *RustfmtAdapter) Name() string                { return "rustfmt" }
func (a *RustfmtAdapter) SupportedLanguages() []string { return []string{"rust"} }

func (a *RustfmtAdapter) Check(ctx context.Context, targets []string) ([]Violation, error) {
	result, err := a.runner.Run(ctx, "cargo", "fmt", "--check")
	if err != nil {
		return nil, fmt.Errorf("rustfmt: %w", err)
	}
	if result.ExitCode == 0 {
		return nil, nil
	}

	var violations []Violation
	scanner := bufio.NewScanner(strings.NewReader(result.Stdout + result.Stderr))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Diff in ") {
			file := strings.TrimPrefix(line, "Diff in ")
			file = strings.TrimSuffix(file, ":")
			violations = append(violations, Violation{
				RuleID: "rustfmt", Severity: "BLOCKING", File: file,
				Message: "not formatted", FixKind: "SAFE_AUTOFIX",
			})
		}
	}
	return violations, nil
}

func (a *RustfmtAdapter) Fix(ctx context.Context, targets []string) ([]FixResult, error) {
	result, err := a.runner.Run(ctx, "cargo", "fmt")
	if err != nil {
		return nil, fmt.Errorf("rustfmt fix: %w", err)
	}
	if result.ExitCode != 0 {
		return []FixResult{{RuleID: "rustfmt", Action: "failed", Detail: result.Stderr}}, nil
	}
	return []FixResult{{RuleID: "rustfmt", Action: "applied", Detail: "ran cargo fmt"}}, nil
}

// CargoTestAdapter wraps cargo test for Rust testing.
type CargoTestAdapter struct {
	runner   *exec.Runner
	repoRoot string
}

func NewCargoTestAdapter(runner *exec.Runner, repoRoot string) *CargoTestAdapter {
	return &CargoTestAdapter{runner: runner, repoRoot: repoRoot}
}

func (a *CargoTestAdapter) Name() string                { return "cargo-test" }
func (a *CargoTestAdapter) SupportedLanguages() []string { return []string{"rust"} }

func (a *CargoTestAdapter) Run(ctx context.Context, targets []string) (*TestResult, error) {
	result, err := a.runner.Run(ctx, "cargo", "test", "--", "--format=terse")
	if err != nil {
		return nil, fmt.Errorf("cargo test: %w", err)
	}

	tr := &TestResult{Output: result.Stdout + result.Stderr}
	// Parse: "test result: ok. X passed; Y failed"
	for _, line := range strings.Split(result.Stdout+result.Stderr, "\n") {
		if strings.Contains(line, "test result:") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "passed;" && i > 0 {
					tr.Passed, _ = strconv.Atoi(parts[i-1])
				}
				if p == "failed;" && i > 0 {
					tr.Failed, _ = strconv.Atoi(parts[i-1])
				}
			}
			tr.Total = tr.Passed + tr.Failed
		}
	}
	return tr, nil
}
