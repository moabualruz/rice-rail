<p align="center">
  <img src="branding/banner.png" alt="riceRail" width="500">
</p>

<p align="center">
  <strong>Terminal-first, agent-agnostic project convergence toolkit</strong>
</p>

<p align="center">
  <a href="#quickstart">Quickstart</a> · <a href="#how-it-works">How It Works</a> · <a href="#commands">Commands</a> · <a href="#supported-tools">Supported Tools</a> · <a href="#contributing">Contributing</a>
</p>

---

**riceRail** enters any software repository, profiles it, asks adaptive questions about your engineering standards, discovers tooling gaps, generates a project-specific toolkit, and runs a deterministic **intent → tool → verify → refine** cycle. Humans define goals. Tools enforce them. LLMs handle only what tools can't.

## Quickstart

```bash
# Build from source
git clone https://github.com/user/rice-railing.git
cd rice-railing
make build

# Initialize on any repo
cd /path/to/your/repo
rice-rail init

# Generate project-specific tooling
rice-rail build-toolkit

# Daily work
rice-rail cycle "add input validation to the API handler"
```

## How It Works

```
    ╲╲╲╲╲╲╲╲╲
     ╲╲╲╲╲╲╲╲──────╮
      ╲╲╲╲╲╲╲──────┤
       ╲╲╲╲╲╲──────┤   ╭──────────────╮
        ╲╲╲╲╲──────┼───┤  rice-rail   │
       ╱╱╱╱╱╱──────┤   ╰──────────────╯
      ╱╱╱╱╱╱╱──────┤
     ╱╱╱╱╱╱╱╱──────╯
    ╱╱╱╱╱╱╱╱╱
```

Many inputs (repo state, tools, standards, architecture) converge through a single constitution into focused, controlled output.

### 1. Profile & Interview

```bash
rice-rail init
```

Scans your repo — languages, package managers, build systems, CI configs, tool configs, architecture hints — then asks adaptive questions about your engineering standards. Produces:

| Artifact | Purpose |
|----------|---------|
| `constitution.yaml` | Your project's engineering doctrine |
| `profile.yaml` | Inferred repo profile with confidence scores |
| `tool-inventory.yaml` | Tools discovered in repo and on PATH |
| `gap-report.yaml` | What's missing vs. what's needed |
| `rollout-plan.yaml` | Ordered steps to close gaps |
| `interview-log.md` | Full Q&A transcript |

### 2. Build Toolkit

```bash
rice-rail build-toolkit
```

Generates project-specific tooling from the constitution:

- Wrapper scripts (`bin/rice-rail-check`, `bin/rice-rail-fix`, etc.)
- Rule directories (Semgrep, ast-grep, custom)
- Codemod templates
- Workflow packs for CLI agents
- Agent-native skills (Claude Code, Gemini, Copilot, Qwen, OpenCode)
- CI workflow (GitHub Actions)
- Operator guide, rule catalog, toolkit overview
- Provenance tracking

### 3. Daily Workflow

```bash
rice-rail check                           # Run all blocking checks
rice-rail fix                             # Safe autofixes only
rice-rail baseline                        # Normalize legacy code
rice-rail cycle "refactor auth to ports"  # Full intent→tool→verify loop
rice-rail explain <rule-id>               # Why does this rule exist?
```

The cycle engine applies deterministic transforms first (formatting, lint fixes, canonicalization), then runs checks, then surfaces only **residual semantic issues** to an agent. Tools handle the mechanical. Agents handle the ambiguous.

## Core Principles

| Principle | What It Means |
|-----------|---------------|
| **Repo owns doctrine** | Constitution is version-controlled YAML, not prompts |
| **Agent-agnostic** | Claude Code, Codex, Gemini, Copilot, Qwen, OpenCode, Ollama, Aider are pluggable adapters |
| **Tools before LLMs** | Deterministic transforms run first. LLMs only handle what tools can't |
| **Safe by default** | Unsafe rewrites need explicit flags. Baseline and feature work separated |
| **Inspectable** | All outputs are YAML/JSON/Markdown. `rice-rail explain` traces any rule to its origin |
| **Provenance** | Every decision tracked — who inferred it, who confirmed it, what generated it |

## Commands

### Core

| Command | Purpose |
|---------|---------|
| `rice-rail init` | Profile repo, adaptive interview, generate constitution |
| `rice-rail build-toolkit` | Generate wrappers, rules, skills, CI, docs from constitution |
| `rice-rail check` | Run all blocking checks (lint, test, typecheck, architecture) |
| `rice-rail fix` | Run safe autofixes only (formatting, imports, safe lint fixes) |
| `rice-rail baseline` | Normalize codebase to policy compliance |
| `rice-rail cycle "<intent>"` | Full intent → tool → verify → refine loop |
| `rice-rail report` | Show toolkit status and constitution summary |
| `rice-rail explain <id>` | Explain any rule, tool, or artifact origin |

### Maintenance

| Command | Purpose |
|---------|---------|
| `rice-rail doctor` | Diagnose toolkit health (14 checks) |
| `rice-rail discover-tools` | Rediscover tools and update inventory |
| `rice-rail regenerate` | Regenerate files from current constitution |
| `rice-rail upgrade-toolkit` | Re-profile repo, detect drift, update gaps |
| `rice-rail add-skill <name>` | Add a custom workflow pack |
| `rice-rail add-mcp <server>` | Register an MCP server |
| `rice-rail version` | Print version |

### Flags

All commands support `--json` for structured output, `--verbose` for debug info, and `--config` to override the constitution path.

## Supported Tools

### Per-Ecosystem

| Ecosystem | Lint | Format | Typecheck | Test | Architecture |
|-----------|------|--------|-----------|------|-------------|
| **Go** | golangci-lint | gofmt | go vet | go test | — |
| **TypeScript/JS** | ESLint, Biome | Prettier, Biome | tsc | Jest, Vitest | dependency-cruiser |
| **Python** | Ruff | Ruff | Pyright, mypy | pytest | — |
| **Rust** | Clippy | rustfmt | — | cargo test | — |

### Cross-Language

| Tool | Capability |
|------|-----------|
| Semgrep | Policy rules, security, forbidden patterns |
| ast-grep | Structural search/rewrite, codemods |
| Comby | Structural search/replace |

### Custom Tools

Any CLI tool can be integrated via `constitution.yaml`:

```yaml
tool_preferences:
  custom:
    - name: hadolint
      binary: hadolint
      role: linter
      check_cmd: ["--format", "json", "{targets}"]
      output_format: json

    - name: shellcheck
      binary: shellcheck
      role: linter
      check_cmd: ["--format", "json", "{targets}"]
      output_format: json

    - name: custom-tests
      binary: make
      role: test_runner
      test_cmd: ["test"]
```

Supports `text`, `json`, and `sarif` output parsing.

## Agent Adapters

riceRail is agent-agnostic. The repo owns the doctrine. Agents consume it.

| Agent | Binary | Status |
|-------|--------|--------|
| Claude Code | `claude` | Full adapter — prompt, tools, structured output |
| Codex | `codex` | Full adapter — full-auto mode |
| Gemini | `gemini` | Full adapter — prompt mode |
| GitHub Copilot | `copilot` | Full adapter — prompt + all tools |
| OpenCode | `opencode` | Full adapter — run mode |
| Qwen Code | `qwen` | Full adapter — prompt mode |
| Ollama | `ollama` | Local LLM — auto-detects best coding model |
| Aider | `aider` | Full adapter — non-interactive mode |

`build-toolkit` generates agent-native skills for all detected agents:

- **Claude Code** → `.claude/skills/rice-rail-*/SKILL.md`
- **Gemini** → `.gemini/skills/rice-rail-*/SKILL.md`
- **Copilot** → `.github/instructions/rice-rail-*.instructions.md` + `AGENTS.md`
- **OpenCode** → `AGENTS.md`
- **Qwen** → `.qwen/skills/rice-rail-*/SKILL.md`

## Project Constitution

The constitution is a typed YAML schema that captures your engineering doctrine:

```yaml
version: 1
project:
  name: my-service
  repo_type: single
  languages: [go, typescript]

architecture:
  target_style: hexagonal
  layering:
    enabled: true
    forbidden_dependencies:
      - from: domain
        to: infrastructure

quality:
  safety_mode: strict
  block_on: [lint, tests, typecheck, architecture]
  max_changed_files_per_cycle: 10

automation:
  allow_safe_autofix: true
  allow_unsafe_autofix: false
```

See [constitution schema](internal/constitution/types.go) for the full typed definition.

## Directory Layout

After `rice-rail init` + `rice-rail build-toolkit`:

```
.project-toolkit/
  constitution.yaml          # Engineering doctrine
  profile.yaml               # Inferred repo profile
  tool-inventory.yaml        # Discovered tools
  gap-report.yaml            # Capability gaps
  rollout-plan.yaml          # Steps to close gaps
  interview-log.md           # Q&A transcript
  provenance/                # Audit trail (decisions.json, inferred-evidence.json)
  rules/{semgrep,ast-grep,custom}/
  codemods/{local,generated}/
  docs/{toolkit-overview,operator-guide,rule-catalog}.md
  state/{toolkit-version,baseline-status,last-run}.json

bin/                         # Wrapper scripts
.agent/workflow-packs/       # Agent-neutral workflow packs
.claude/skills/              # Claude Code native skills
.gemini/skills/              # Gemini native skills
.github/instructions/        # Copilot instructions
.github/workflows/           # Generated CI workflow
.qwen/skills/                # Qwen native skills
AGENTS.md                    # OpenCode/Copilot instructions
```

## Build

```bash
make build          # Build bin/rice-rail
make test           # Run all tests (148 tests)
make lint           # golangci-lint
make fmt            # gofmt + goimports
make install        # Install to $GOPATH/bin
```

Requires Go 1.21+. No runtime dependencies — tool adapters are optional and discovered on PATH.

## License

[CC BY-NC-SA 4.0](LICENSE.md) — Free for non-commercial use. Commercial license available.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on submitting code, adding tool adapters, and adding agent adapters.
