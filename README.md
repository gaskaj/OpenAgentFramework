# OpenAgentFramework

![Coverage](https://img.shields.io/badge/coverage-49.0%25-red)


An autonomous development agent framework that uses **GitHub as its source of truth** and **Claude as its AI engine**.

## Overview

OpenAgentFramework provides agent personas — Developer, QA, and Development Manager — that monitor GitHub issues, write code, create pull requests, run tests, and coordinate via GitHub labels and comments.

## Documentation

| Document | Description |
|----------|-------------|
| [CLAUDE.md](CLAUDE.md) | LLM-first project orientation |
| **Architecture** | |
| [architecture.md](docs/architecture/architecture.md) | System design, packages, data flow |
| [package-reference.md](docs/architecture/package-reference.md) | Per-package API catalog |
| **Guides** | |
| [developer-workflow.md](docs/guides/developer-workflow.md) | State machine, decomposition, creativity |
| [claude-integration.md](docs/guides/claude-integration.md) | Claude client, tools, prompts, SDK patterns |
| [github-integration.md](docs/guides/github-integration.md) | GitHub client, poller, labels, PRs |
| [code-conventions.md](docs/guides/code-conventions.md) | Go patterns, testing, git conventions |
| **Configuration** | |
| [configuration.md](docs/configuration/configuration.md) | Full YAML reference, env vars, defaults |
| [configuration-management.md](docs/configuration/configuration-management.md) | Runtime validation, environment overlays |
| [environment-variables.md](docs/configuration/environment-variables.md) | Environment variable reference |
| **Observability** | |
| [structured-logging.md](docs/observability/structured-logging.md) | Observability, correlation IDs, metrics |
| [error-recovery.md](docs/observability/error-recovery.md) | Error handling, retry, circuit breakers |
| **Testing** | |
| [testing-strategy.md](docs/testing/testing-strategy.md) | Test strategy and patterns |
| [integration-testing.md](docs/testing/integration-testing.md) | Integration test suite, CI pipeline |
| [test-coverage.md](docs/testing/test-coverage.md) | Coverage reporting and quality gates |
| **WebUI / Control Plane** | |
| [webui-architecture.md](docs/webui/webui-architecture.md) | Control plane architecture, multi-tenant design |
| [webui-api-reference.md](docs/webui/webui-api-reference.md) | REST API endpoints |
| [webui-deployment.md](docs/webui/webui-deployment.md) | Docker Compose deployment |

## Architecture

```
cmd/agentctl/            CLI entry point
internal/
  cli/                   Cobra CLI commands
  config/                Configuration loading (Viper)
  ghub/                  GitHub API integration
  claude/                Claude AI integration
  gitops/                Git operations (go-git)
  agent/                 Agent interface + registry
  state/                 Persistent agent work state
  developer/             Developer agent implementation
  creativity/            Autonomous suggestion engine (idle mode)
  errors/                Retry, circuit breakers, error classification
  observability/         Structured logging, correlation IDs, metrics
  orchestrator/          Agent pool + health checks
```

### Agent Coordination

Agents coordinate entirely through GitHub:
- **Issues** are the work queue
- **Labels** signal state (`agent:ready`, `agent:claimed`, `agent:in-progress`, `agent:in-review`, `agent:suggestion`)
- **Assignments** track ownership
- **Comments** enable human-in-the-loop feedback

### Developer Agent Workflow

```
idle → claim → workspace → analyze → [decompose] → implement → commit → PR → review → complete
```

When idle and creativity is enabled, the agent enters `creative_thinking` mode — generating improvement suggestions as GitHub issues. Complex issues are automatically decomposed into subtasks. See [docs/guides/developer-workflow.md](docs/guides/developer-workflow.md) for details.

## Setup

### Prerequisites

- Go 1.22+
- GitHub personal access token (repo scope)
- Anthropic API key

### Configuration

```bash
cp configs/config.example.yaml configs/config.yaml
# Edit configs/config.yaml with your tokens and repo details
```

See [docs/configuration/configuration.md](docs/configuration/configuration.md) for the full YAML reference.

### Build & Run

```bash
make build          # Build the agentctl binary
make test           # Run tests
make run            # Build and run with example config
```

### Commands

```bash
agentctl start --config configs/config.yaml   # Start agent loop
agentctl status --config configs/config.yaml  # Show agent status
```

## License

MIT
