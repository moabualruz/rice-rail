# Contributing to riceRail

Thank you for your interest in contributing! This document provides quick guidelines.

## Getting Started

```bash
# Clone and setup
git clone https://github.com/YOUR_USERNAME/rice-railing.git
cd rice-railing

# Build
make build

# Run tests
make test

# Run linter
make lint

# Format code
make fmt
```

## Ways to Contribute

### Report Bugs

1. Check [existing issues](https://github.com/user/rice-railing/issues)
2. Create new issue with:
   - riceRail version (`rice-rail version`)
   - Steps to reproduce
   - Expected vs actual behavior

### Suggest Features

1. Open a [Discussion](https://github.com/user/rice-railing/discussions)
2. Describe use case and benefits
3. Consider implementation approach

### Submit Code

1. Fork the repository
2. Create feature branch: `git checkout -b feat/my-feature`
3. Write tests
4. Run checks: `make test && make lint`
5. Submit pull request

## Code Style

- Go 1.21+ with idiomatic patterns
- Use [Conventional Commits](https://www.conventionalcommits.org/)
- Keep functions under 30 lines
- Adapter interfaces in `internal/adapters/interfaces.go`
- Commands are thin — delegate to core packages

## Pull Request Checklist

- [ ] Tests pass (`make test`)
- [ ] Vet passes (`go vet ./...`)
- [ ] Build succeeds (`make build`)
- [ ] New adapters implement the correct interface
- [ ] Documentation updated if needed
- [ ] Commit messages follow conventions

## Adding a New Tool Adapter

1. Create `internal/adapters/<tool>.go` implementing `RuleEngineAdapter`, `TypecheckAdapter`, or `TestRunnerAdapter`
2. Register in `internal/adapters/registry.go` under the appropriate language section
3. Add detection pattern in `internal/profiling/detectors.go`
4. Write tests in `internal/adapters/<tool>_test.go`

## Adding a New Agent Adapter

1. Create `internal/adapters/<agent>.go` implementing `AgentAdapter`
2. Register in `internal/adapters/registry.go` agent section
3. Add skill generation in `internal/builder/agent_skills.go` if the agent supports project-level skills

## Code of Conduct

Be respectful and inclusive. We welcome contributors of all backgrounds and experience levels.

## License

By contributing, you agree your contributions will be licensed under [CC BY-NC-SA 4.0](https://creativecommons.org/licenses/by-nc-sa/4.0/).

## Questions?

- Open a [Discussion](https://github.com/user/rice-railing/discussions)
