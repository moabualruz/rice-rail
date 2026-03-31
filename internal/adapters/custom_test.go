package adapters

import (
	"context"
	"testing"

	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/exec"
)

func TestExpandTargets(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		targets []string
		want    int
	}{
		{"placeholder", []string{"--check", "{targets}"}, []string{"src/", "lib/"}, 3},
		{"inline", []string{"--path={targets}"}, []string{"src/"}, 1},
		{"no placeholder", []string{"--check"}, []string{"src/"}, 2},
		{"dot target", []string{"--check"}, []string{"."}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandTargets(tt.args, tt.targets)
			if len(got) != tt.want {
				t.Errorf("expandTargets(%v, %v) = %v (len %d), want len %d", tt.args, tt.targets, got, len(got), tt.want)
			}
		})
	}
}

func TestParseColonFormat(t *testing.T) {
	tests := []struct {
		line string
		file string
		line_ int
		msg  string
	}{
		{"src/main.go:10:5: unused variable", "src/main.go", 10, "unused variable"},
		{"app.py:42: missing return", "app.py", 42, "missing return"},
	}
	for _, tt := range tests {
		v := parseColonFormat(tt.line, "test")
		if v == nil {
			t.Errorf("parseColonFormat(%q) returned nil", tt.line)
			continue
		}
		if v.File != tt.file {
			t.Errorf("file: got %q, want %q", v.File, tt.file)
		}
		if v.Line != tt.line_ {
			t.Errorf("line: got %d, want %d", v.Line, tt.line_)
		}
	}
}

func TestParseParenFormat(t *testing.T) {
	v := parseParenFormat("src/app.ts(15,3): error TS2304: Cannot find name", "test")
	if v == nil {
		t.Fatal("expected violation")
	}
	if v.File != "src/app.ts" {
		t.Errorf("file: got %q", v.File)
	}
	if v.Line != 15 {
		t.Errorf("line: got %d", v.Line)
	}
}

func TestCustomAdapterName(t *testing.T) {
	def := constitution.CustomTool{
		Name:   "my-linter",
		Binary: "my-lint",
		Role:   "linter",
	}
	a := NewCustomAdapter(nil, def)
	if a.Name() != "my-linter" {
		t.Errorf("name: got %q", a.Name())
	}
}

func TestParseJSONOutput(t *testing.T) {
	a := &CustomAdapter{def: constitution.CustomTool{Name: "test", OutputFmt: "json"}}

	// Array format
	output := `[{"file":"a.go","line":10,"message":"bad","severity":"error"}]`
	violations := a.parseJSON(output)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].File != "a.go" {
		t.Errorf("file: got %q", violations[0].File)
	}
	if violations[0].Severity != "BLOCKING" {
		t.Errorf("severity: got %q", violations[0].Severity)
	}

	// Wrapped format
	output2 := `{"results":[{"path":"b.py","line":5,"message":"warn","severity":"warning"}]}`
	violations2 := a.parseJSON(output2)
	if len(violations2) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations2))
	}
	if violations2[0].Severity != "WARNING" {
		t.Errorf("severity: got %q", violations2[0].Severity)
	}
}

func TestParseSARIF(t *testing.T) {
	a := &CustomAdapter{def: constitution.CustomTool{Name: "test", OutputFmt: "sarif"}}
	output := `{"runs":[{"results":[{"ruleId":"R1","level":"error","message":{"text":"bad code"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"src/main.go"},"region":{"startLine":42}}}]}]}]}`

	violations := a.parseSARIF(output)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].RuleID != "R1" {
		t.Errorf("ruleID: got %q", violations[0].RuleID)
	}
	if violations[0].File != "src/main.go" {
		t.Errorf("file: got %q", violations[0].File)
	}
	if violations[0].Line != 42 {
		t.Errorf("line: got %d", violations[0].Line)
	}
}

// --- Integration tests using real binaries ---

func TestCustomAdapterCheckWithEcho(t *testing.T) {
	runner := exec.NewRunner()
	def := constitution.CustomTool{
		Name:     "echo-checker",
		Binary:   "echo",
		Role:     "linter",
		CheckCmd: []string{"hello world"},
	}
	adapter := NewCustomAdapter(runner, def)

	violations, err := adapter.Check(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations (exit 0), got %d", len(violations))
	}
}

func TestCustomAdapterCheckFailure(t *testing.T) {
	runner := exec.NewRunner()
	def := constitution.CustomTool{
		Name:     "false-checker",
		Binary:   "false",
		Role:     "linter",
		CheckCmd: []string{},
	}
	adapter := NewCustomAdapter(runner, def)

	violations, err := adapter.Check(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "false" exits with 1 and produces no output, so violations may be empty
	// but the key assertion is that err is nil (non-zero exit is not an error)
	// and the code path that parses output is exercised.
	_ = violations
}

func TestCustomAdapterCheckFailureWithOutput(t *testing.T) {
	// Use sh -c to produce output on stderr and exit non-zero.
	runner := exec.NewRunner()
	def := constitution.CustomTool{
		Name:     "failing-linter",
		Binary:   "sh",
		Role:     "linter",
		CheckCmd: []string{"-c", "echo 'main.go:10: bad code' && exit 1"},
	}
	adapter := NewCustomAdapter(runner, def)

	violations, err := adapter.Check(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].File != "main.go" {
		t.Errorf("file: got %q, want %q", violations[0].File, "main.go")
	}
	if violations[0].Line != 10 {
		t.Errorf("line: got %d, want 10", violations[0].Line)
	}
	if violations[0].Message != "bad code" {
		t.Errorf("message: got %q, want %q", violations[0].Message, "bad code")
	}
}

func TestCustomAdapterFixWithEcho(t *testing.T) {
	runner := exec.NewRunner()
	def := constitution.CustomTool{
		Name:   "echo-fixer",
		Binary: "echo",
		Role:   "formatter",
		FixCmd: []string{"fixing"},
	}
	adapter := NewCustomAdapter(runner, def)

	results, err := adapter.Fix(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 fix result, got %d", len(results))
	}
	if results[0].Action != "applied" {
		t.Errorf("action: got %q, want %q", results[0].Action, "applied")
	}
	if results[0].RuleID != "echo-fixer" {
		t.Errorf("ruleID: got %q, want %q", results[0].RuleID, "echo-fixer")
	}
}

func TestCustomAdapterFixNoCmd(t *testing.T) {
	runner := exec.NewRunner()
	def := constitution.CustomTool{
		Name:   "no-fix",
		Binary: "echo",
		Role:   "linter",
		// FixCmd intentionally empty
	}
	adapter := NewCustomAdapter(runner, def)

	results, err := adapter.Fix(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil fix results, got %v", results)
	}
}

func TestCustomTestRunnerAdapter(t *testing.T) {
	runner := exec.NewRunner()
	def := constitution.CustomTool{
		Name:    "echo-test-runner",
		Binary:  "echo",
		Role:    "test_runner",
		TestCmd: []string{"tests pass"},
	}
	inner := NewCustomAdapter(runner, def)
	adapter := &CustomTestRunnerAdapter{inner: inner}

	if adapter.Name() != "echo-test-runner" {
		t.Errorf("name: got %q", adapter.Name())
	}

	result, err := adapter.Run(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil TestResult")
	}
	if result.Passed != 1 {
		t.Errorf("passed: got %d, want 1", result.Passed)
	}
	if result.Failed != 0 {
		t.Errorf("failed: got %d, want 0", result.Failed)
	}
	if result.Total != 1 {
		t.Errorf("total: got %d, want 1", result.Total)
	}
	if result.Output == "" {
		t.Error("expected non-empty output")
	}
}

func TestCustomTypecheckAdapter(t *testing.T) {
	runner := exec.NewRunner()
	def := constitution.CustomTool{
		Name:     "echo-typechecker",
		Binary:   "echo",
		Role:     "typechecker",
		CheckCmd: []string{"all good"},
	}
	inner := NewCustomAdapter(runner, def)
	adapter := &CustomTypecheckAdapter{inner: inner}

	if adapter.Name() != "echo-typechecker" {
		t.Errorf("name: got %q", adapter.Name())
	}
	if langs := adapter.SupportedLanguages(); langs != nil {
		t.Errorf("expected nil languages, got %v", langs)
	}

	violations, err := adapter.Check(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations (exit 0), got %d", len(violations))
	}
}

func TestParseTextMultiLine(t *testing.T) {
	a := &CustomAdapter{def: constitution.CustomTool{Name: "multi-lint", OutputFmt: "text"}}

	output := `src/main.go:10:5: unused variable x
lib/util.go:22: missing return statement
not-a-match-line
src/app.ts(5,3): error TS2304
`
	violations := a.parseText(output)

	// Should match: line 1 (colon format), line 2 (colon format), line 4 (paren format)
	// Line 3 matches neither format.
	if len(violations) != 3 {
		t.Fatalf("expected 3 violations, got %d", len(violations))
	}

	// Check first violation (colon with col)
	if violations[0].File != "src/main.go" {
		t.Errorf("v[0].File: got %q", violations[0].File)
	}
	if violations[0].Line != 10 {
		t.Errorf("v[0].Line: got %d", violations[0].Line)
	}
	if violations[0].Message != "unused variable x" {
		t.Errorf("v[0].Message: got %q", violations[0].Message)
	}

	// Check second violation (colon without col)
	if violations[1].File != "lib/util.go" {
		t.Errorf("v[1].File: got %q", violations[1].File)
	}
	if violations[1].Line != 22 {
		t.Errorf("v[1].Line: got %d", violations[1].Line)
	}

	// Check third violation (paren format)
	if violations[2].File != "src/app.ts" {
		t.Errorf("v[2].File: got %q", violations[2].File)
	}
	if violations[2].Line != 5 {
		t.Errorf("v[2].Line: got %d", violations[2].Line)
	}

	// All should have the default rule ID
	for i, v := range violations {
		if v.RuleID != "multi-lint" {
			t.Errorf("v[%d].RuleID: got %q, want %q", i, v.RuleID, "multi-lint")
		}
		if v.Severity != "BLOCKING" {
			t.Errorf("v[%d].Severity: got %q, want %q", i, v.Severity, "BLOCKING")
		}
	}
}

func TestParseJSONDiagnostics(t *testing.T) {
	a := &CustomAdapter{def: constitution.CustomTool{Name: "diag-tool", OutputFmt: "json"}}

	output := `{"diagnostics":[
		{"file":"a.go","line":1,"message":"unused import","severity":"warning","ruleId":"W001"},
		{"path":"b.go","line":5,"message":"type error","severity":"error","rule":"E001"},
		{"file":"c.go","line":10,"message":"info hint","severity":"info"}
	]}`

	violations := a.parseJSON(output)
	if len(violations) != 3 {
		t.Fatalf("expected 3 violations, got %d", len(violations))
	}

	// Check severity mapping
	if violations[0].Severity != "WARNING" {
		t.Errorf("v[0].Severity: got %q, want WARNING", violations[0].Severity)
	}
	if violations[0].RuleID != "W001" {
		t.Errorf("v[0].RuleID: got %q, want W001", violations[0].RuleID)
	}
	if violations[0].File != "a.go" {
		t.Errorf("v[0].File: got %q", violations[0].File)
	}

	if violations[1].Severity != "BLOCKING" {
		t.Errorf("v[1].Severity: got %q, want BLOCKING", violations[1].Severity)
	}
	if violations[1].RuleID != "E001" {
		t.Errorf("v[1].RuleID: got %q, want E001", violations[1].RuleID)
	}
	// "path" field should be used when "file" is empty
	if violations[1].File != "b.go" {
		t.Errorf("v[1].File: got %q, want b.go", violations[1].File)
	}

	if violations[2].Severity != "INFO" {
		t.Errorf("v[2].Severity: got %q, want INFO", violations[2].Severity)
	}
	// No ruleId or rule → falls back to tool name
	if violations[2].RuleID != "diag-tool" {
		t.Errorf("v[2].RuleID: got %q, want diag-tool", violations[2].RuleID)
	}
}

func TestParseSARIFMultipleRuns(t *testing.T) {
	a := &CustomAdapter{def: constitution.CustomTool{Name: "sarif-multi", OutputFmt: "sarif"}}

	output := `{
		"runs": [
			{
				"results": [
					{
						"ruleId": "R1",
						"level": "error",
						"message": {"text": "null deref"},
						"locations": [{"physicalLocation": {"artifactLocation": {"uri": "src/a.go"}, "region": {"startLine": 10}}}]
					},
					{
						"ruleId": "R2",
						"level": "warning",
						"message": {"text": "unused var"},
						"locations": [{"physicalLocation": {"artifactLocation": {"uri": "src/b.go"}, "region": {"startLine": 20}}}]
					}
				]
			},
			{
				"results": [
					{
						"ruleId": "R3",
						"level": "note",
						"message": {"text": "consider refactoring"},
						"locations": [{"physicalLocation": {"artifactLocation": {"uri": "src/c.go"}, "region": {"startLine": 30}}}]
					}
				]
			}
		]
	}`

	violations := a.parseSARIF(output)
	if len(violations) != 3 {
		t.Fatalf("expected 3 violations from 2 runs, got %d", len(violations))
	}

	// Run 1, result 1
	if violations[0].RuleID != "R1" {
		t.Errorf("v[0].RuleID: got %q", violations[0].RuleID)
	}
	if violations[0].Severity != "BLOCKING" {
		t.Errorf("v[0].Severity: got %q", violations[0].Severity)
	}
	if violations[0].File != "src/a.go" {
		t.Errorf("v[0].File: got %q", violations[0].File)
	}
	if violations[0].Line != 10 {
		t.Errorf("v[0].Line: got %d", violations[0].Line)
	}

	// Run 1, result 2
	if violations[1].Severity != "WARNING" {
		t.Errorf("v[1].Severity: got %q", violations[1].Severity)
	}

	// Run 2, result 1
	if violations[2].RuleID != "R3" {
		t.Errorf("v[2].RuleID: got %q", violations[2].RuleID)
	}
	if violations[2].Severity != "INFO" {
		t.Errorf("v[2].Severity: got %q", violations[2].Severity)
	}
	if violations[2].Line != 30 {
		t.Errorf("v[2].Line: got %d", violations[2].Line)
	}
}

func TestExpandTargetsEmpty(t *testing.T) {
	// No targets, no placeholder
	got := expandTargets([]string{"--check", "--verbose"}, nil)
	if len(got) != 2 {
		t.Errorf("expected 2 args with nil targets, got %d: %v", len(got), got)
	}

	// Empty slice targets
	got2 := expandTargets([]string{"--check", "--verbose"}, []string{})
	if len(got2) != 2 {
		t.Errorf("expected 2 args with empty targets, got %d: %v", len(got2), got2)
	}

	// With placeholder but empty targets
	got3 := expandTargets([]string{"--check", "{targets}"}, []string{})
	if len(got3) != 1 {
		t.Errorf("expected 1 arg (placeholder replaced with nothing), got %d: %v", len(got3), got3)
	}
	if got3[0] != "--check" {
		t.Errorf("expected '--check', got %q", got3[0])
	}
}
