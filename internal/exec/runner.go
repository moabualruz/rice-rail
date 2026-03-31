package exec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

// Result captures the output of an external command.
type Result struct {
	Command  string
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// Success returns true if the command exited with code 0.
func (r *Result) Success() bool {
	return r.ExitCode == 0
}

// Runner executes external commands with logging, timeouts, and dry-run support.
type Runner struct {
	DryRun  bool
	Verbose bool
	Timeout time.Duration
}

// NewRunner creates a runner with sensible defaults.
func NewRunner() *Runner {
	return &Runner{
		Timeout: 5 * time.Minute,
	}
}

// Run executes a command and captures its output.
func (r *Runner) Run(ctx context.Context, name string, args ...string) (*Result, error) {
	cmdStr := name
	for _, a := range args {
		cmdStr += " " + a
	}

	if r.DryRun {
		return &Result{Command: cmdStr}, nil
	}

	if r.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.Timeout)
		defer cancel()
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	result := &Result{
		Command:  cmdStr,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
			return result, nil
		}
		return result, fmt.Errorf("executing %s: %w", cmdStr, err)
	}

	return result, nil
}

// Which checks if a command is available on PATH.
func Which(name string) (string, bool) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", false
	}
	return path, true
}
