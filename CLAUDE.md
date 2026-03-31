# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

rice-railing is a terminal-first, agent-agnostic project convergence toolkit. CLI binary: `rice-rail`. It profiles repositories, generates project constitutions (engineering doctrine as typed YAML), builds project-specific tooling (rules, codemods, wrappers), and runs deterministic intent → tool → verify → refine cycles. Agents (Claude Code, Aider, Codex) are pluggable adapters — the repo owns the doctrine.

## Build & Run

```bash
make build          # builds bin/rice-rail
make run ARGS="init" # build + run with args
make test           # all tests verbose
make test-short     # skip slow tests
make lint           # golangci-lint
make fmt            # gofmt + goimports
make install        # copy to $GOPATH/bin
```

Single test: `go test ./internal/exec/ -run TestRunnerSuccess -v`

## Commands

| Command | Purpose |
|---------|---------|
| `rice-rail init` | Profile repo, adaptive interview, generate constitution + inventory + gap report |
| `rice-rail build-toolkit` | Generate wrappers, workflow packs, CI workflow, docs from constitution |
| `rice-rail check` | Run blocking checks via real tool adapters (golangci-lint, go vet, go test, tsc, ruff, semgrep) |
| `rice-rail fix` | Run safe autofixes (gofmt, golangci-lint --fix, ruff --fix) |
| `rice-rail baseline` | Convergence loop: fix → check → repeat until clean or stuck |
| `rice-rail cycle "<intent>"` | Intent → transform → fix → check → refine loop with scope limits |
| `rice-rail report` | Show toolkit status and constitution summary |
| `rice-rail explain <id>` | Trace any rule/artifact to its origin and rationale |
| `rice-rail doctor` | Diagnose toolkit health: check wrappers, packs, tool availability |
| `rice-rail regenerate` | Regenerate all files from current constitution (supports --dry-run) |
| `rice-rail upgrade-toolkit` | Re-profile repo, detect changes, update gap analysis (supports --apply) |
| `rice-rail add-skill <name>` | Create a custom workflow pack |
| `rice-rail add-mcp <server>` | Register an MCP server in the constitution |
| `rice-rail version` | Print version |

## Architecture

```
cmd/rice-rail/         → thin main, calls cli.Execute()
internal/
  cli/                 → cobra commands (one file per subcommand), paths() resolves --config
  config/              → Paths resolver (respects --config flag), runtime config, constants
  constitution/        → typed schema for project doctrine + load/save
  exec/                → execution layer: shell out, capture output, dry-run, timeout
  reporting/           → output formatting (text/JSON/YAML), file writing
  adapters/            → interfaces + implementations: golangci-lint, gofmt, go vet, go test, tsc, ruff, semgrep + registry auto-discovery
  provenance/          → tracker + audit trail: records inferences, user decisions, generated artifacts
  profiling/           → repo analysis: language/tool/topology detection with confidence scores
  interview/           → adaptive question engine: branching, seeds, quick/strict modes, Charm huh TUI
  discovery/           → tool inventory builder from profile data
  resolution/          → capability gap classification (PRESENT_READY → DEFERRED) + rollout plan
  builder/             → toolkit generation: dirs, wrappers, workflow packs, CI workflows, docs, provenance
  baseline/            → safe remediation loop: fix → check → converge with per-iteration progress
  cycle/               → intent→tool→verify→refine engine with scope limits
  company/             → company pack support: reusable doctrine bundles merged into constitution
```

Key design rules:
- Agent-independent core vs agent adapter layer. Repo owns doctrine, agents consume it.
- All external tool calls go through `exec.Runner` (logging, timeout, dry-run).
- All config files use typed Go structs with `yaml`/`json` struct tags.
- Constitution schema is the single source of truth for project policy.
- Adapter interfaces in `internal/adapters/interfaces.go`; registry auto-discovers tools on PATH.
- Commands are thin: parse flags, call `paths()` for config resolution, delegate to core packages.
- `--config` flag overrides constitution path; all other artifact paths derive from its directory.
- Provenance tracker records why every artifact and decision exists.

## Config Precedence

command flags > repo config > constitution > company pack > global defaults

## Output Conventions

All commands support `--json` for structured output. Default is human-readable text. Generated artifacts go to `.project-toolkit/`. Agent workflow packs go to `.agent/`. CI workflow goes to `.github/workflows/rice-rail.yml`.
