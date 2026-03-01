# DeveloperAndQAAgent

An autonomous development agent framework that uses **GitHub as its source of truth** and **Claude as its AI engine**.

## Overview

DeveloperAndQAAgent provides agent personas — Developer, QA, and Development Manager — that monitor GitHub issues, write code, create pull requests, run tests, and coordinate via GitHub labels and comments.

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
  orchestrator/          Agent pool + health checks
```

### Agent Coordination

Agents coordinate entirely through GitHub:
- **Issues** are the work queue
- **Labels** signal state (`agent:claimed`, `agent:in-progress`, `agent:review`)
- **Assignments** track ownership
- **Comments** enable human-in-the-loop feedback

### Developer Agent State Machine

```
idle → claim → analyze → workspace → implement → commit → PR → review → complete
```

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
