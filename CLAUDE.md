# CLAUDE.md — Project Orientation

OpenAgentFramework is an autonomous Go agent that monitors GitHub issues, uses Claude to analyze and implement solutions, and creates pull requests. GitHub is the coordination layer (issues, labels, PRs); Claude is the AI engine.

## Quick Reference

```bash
# Agent CLI
make build                    # Build → bin/agentctl
make test                     # Run tests with -race
make lint                     # golangci-lint + go vet
make fmt                      # gofmt -s -w .
agentctl start --config configs/config.yaml
agentctl status --config configs/config.yaml

# Control Plane
make build-controlplane       # Build → bin/controlplane
docker compose up             # Start full stack (postgres + controlplane + frontend)

# Frontend Testing
cd frontend && npm test       # Run Vitest unit/component tests
cd frontend && npm run test:e2e  # Run Playwright e2e tests (requires running stack)
```

## Repository Layout

```
cmd/agentctl/                 CLI entry point (main.go → cli.Execute())
cmd/controlplane/             Control plane server entry point
internal/
  agent/                      Agent interface, BaseAgent, Dependencies DI, registry
  orchestrator/               Concurrent agent execution (errgroup)
  developer/                  Developer agent: workflow, decomposition, prompts
  claude/                     Claude API client, conversation manager, tool definitions
  ghub/                       GitHub client interface, issues, PRs, branches, comments, poller
  gitops/                     Git clone, checkout, commit, push (go-git)
  config/                     Config loading (Viper), validation, defaults
  state/                      WorkflowState enum, AgentWorkState, file-based Store
  creativity/                 Idle-mode suggestion engine
  memory/                     Persistent repo memory for Claude efficiency
  integration/                Integration tests: agent communication, handoffs, shared state
  errors/                     Retry with backoff, circuit breakers, error classification
  observability/              Structured logger, correlation IDs, metrics
  cli/                        Cobra commands (start, status)
web/                          Control plane backend (Go + chi + pgx)
  handler/                    HTTP handlers (auth, agents, events, orgs, API keys, audit)
  store/                      PostgreSQL stores (pgx) for all entities
  auth/                       JWT, bcrypt, OAuth providers
  middleware/                 Auth, API key, logging middleware
  router/                     Chi router with full route tree
  ws/                         WebSocket hub for real-time event streaming
  config/                     Server config (Viper)
  migrate/                    SQL migrations (embedded)
pkg/
  apitypes/                   Shared event types between agents and control plane
  reporter/                   Buffered HTTP reporter client for agents
frontend/                     React + TypeScript + Vite control plane UI
  src/pages/                  Page components (Dashboard, Agents, Events, Settings, etc.)
  src/store/                  Zustand state stores (auth, agent, event)
  src/hooks/                  React hooks (useAuth, useAgents, useEvents, useWebSocket)
  src/api/                    Axios API clients
  src/components/             Reusable UI components
  e2e/                        Playwright e2e tests
configs/                      YAML config files (config.remote.yaml, config.example.yaml, controlplane.example.yaml)
docs/                         Detailed documentation (see links below)
```

## Architecture at a Glance

The Orchestrator runs agents concurrently. The Developer agent polls GitHub for `agent:ready` issues, claims them, clones the repo, asks Claude to analyze and implement, then commits and creates a PR. Complex issues are decomposed into subtask issues. Idle agents generate improvement suggestions via the creativity engine.

**Workflow**: `idle → claim → workspace → analyze → [decompose] → implement → commit → PR → validation → review → complete`

## Build & Release Protocol

**MANDATORY**: Every time you build binaries (via `make build`, `make build-all`, `docker compose build`, etc.), you MUST also cut a new release. This applies to all builds, not just explicit release requests.

### Steps

1. **Determine next version**: Query the latest GitHub release tag with `gh release list --limit 1` or `git tag --list 'v*' --sort=-v:refname | head -1`. Increment the **patch** version (e.g., v0.1.0 → v0.1.1, v0.1.5 → v0.1.6).
2. **Build locally**: Run `make build && make build-controlplane` to verify the code compiles.
3. **Rebuild Docker**: Run `docker compose build` (or `docker compose build --no-cache` if needed) to update the running stack.
4. **Commit all changes**: Stage and commit any pending changes.
5. **Tag and push**: Create the new version tag and push it along with the branch:
   ```bash
   git tag v<NEW_VERSION>
   git push origin <branch> --tags
   ```
6. **Verify**: The `.github/workflows/release.yml` workflow will automatically build cross-platform binaries (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) and create a GitHub Release with all artifacts.

### Version injection

Binaries are stamped at build time via `-ldflags` (see `Makefile`). The `internal/version` package exposes `Version`, `Commit`, and `BuildDate`. The release workflow injects the tag name as the version.

### Important

- Never skip the release step when building. Every build = every release.
- Always increment patch from the **latest published tag**, not from a hardcoded value.
- The release workflow builds `.exe` for Windows automatically — you do not need to cross-compile locally.

## Key Patterns and Conventions

- **Dependency injection**: All components receive `agent.Dependencies` (Config, GitHub, Claude, Store, Logger, Metrics, ErrorManager)
- **Interface-based design**: `ghub.Client`, `state.Store`, `agent.Agent` — defined where used, not where implemented
- **Error handling**: `fmt.Errorf("doing X: %w", err)` for wrapping; `errors.Join` for aggregation; `ClassifyError()` for typed retry decisions
- **Error resilience**: All external API calls wrapped with retry (exponential backoff) + circuit breakers via `errors.Manager`
- **Logging**: `log/slog` with structured key-value pairs; child loggers via `.With()`
- **Testing**: `testify` assertions, `t.TempDir()` for file tests, mocks alongside interfaces
- **Context propagation**: All I/O operations take `context.Context` carrying correlation data
- **Builder pattern**: `NewClient().WithObservability().WithErrorHandling()`
- **Tool preferences**: Use `edit_file` over `write_file` for modifications; `search_files` over multiple `read_file` calls

## Tool Definitions

The developer agent gives Claude 6 tools (defined in `internal/claude/tools.go`):

| Tool | Purpose |
|------|---------|
| `read_file` | Read file contents |
| `edit_file` | Search-and-replace in a file (old_string must be unique) |
| `write_file` | Create or overwrite a file |
| `search_files` | Grep across workspace (regex or literal, up to 50 results) |
| `list_files` | List directory contents |
| `run_command` | Execute shell command in workspace |

## Labels Protocol

| Label | Meaning |
|-------|---------|
| `agent:ready` | Issue available for agent to claim |
| `agent:claimed` | Agent has taken ownership |
| `agent:in-progress` | Implementation underway |
| `agent:in-review` | PR created, awaiting human review |
| `agent:failed` | Processing failed |
| `agent:suggestion` | Creativity engine suggestion |
| `agent:suggestion-rejected` | Rejected suggestion (remembered) |
| `agent:subtask` | Child issue from decomposition |
| `agent:epic` | Parent issue that was decomposed |

## Workflow States

From `internal/state/models.go`:

`idle` · `claim` · `workspace` · `analyze` · `decompose` · `implement` · `commit` · `pr` · `validation` · `review` · `complete` · `failed` · `creative_thinking`

## Prompt Constants

From `internal/developer/prompts.go`:

| Constant | Purpose |
|----------|---------|
| `SystemPrompt` | Base system prompt — coding guidelines, tool preferences, efficiency |
| `AnalyzePrompt` | Issue analysis → implementation plan |
| `ImplementPrompt` | Execute plan using tools |
| `ComplexityEstimatePrompt` | Estimate API iterations (decomposition) |
| `DecomposePrompt` | Break complex issue into subtasks |
| `ReactiveDecomposePrompt` | Decompose remaining work after iteration limit |

## Configuration

**Remote mode (recommended for production):** Agents need only `controlplane.url`, `controlplane.api_key`, and `config_mode: "remote"`. All other settings are managed centrally via the control plane. Agent name/type are derived from the API key. See `configs/config.remote.yaml` and [docs/configuration/remote-configuration.md](docs/configuration/remote-configuration.md).

**Local mode (development/testing):** Requires `GITHUB_TOKEN`, `ANTHROPIC_API_KEY` env vars and a full config file. See `configs/config.example.yaml` and [docs/configuration/configuration.md](docs/configuration/configuration.md).

Key config sections: `github`, `claude`, `agents`, `state`, `logging`, `creativity`, `decomposition`, `memory`, `error_handling`, `controlplane`

## Deep-Dive Documentation

- [docs/architecture/architecture.md](docs/architecture/architecture.md) — System design, packages, data flow, agent personas
- [docs/architecture/package-reference.md](docs/architecture/package-reference.md) — Per-package API catalog
- [docs/guides/developer-workflow.md](docs/guides/developer-workflow.md) — State machine, decomposition, creativity, self-building
- [docs/guides/claude-integration.md](docs/guides/claude-integration.md) — Client, conversation loop, tools, prompts, SDK patterns
- [docs/guides/github-integration.md](docs/guides/github-integration.md) — Client interface, poller, labels, branches, PRs
- [docs/guides/code-conventions.md](docs/guides/code-conventions.md) — Error handling, naming, interfaces, logging, testing
- [docs/guides/repository-memory.md](docs/guides/repository-memory.md) — Persistent repo memory system for Claude efficiency
- [docs/guides/release-pipeline.md](docs/guides/release-pipeline.md) — Release workflow, version injection, binary distribution
- [docs/configuration/configuration.md](docs/configuration/configuration.md) — Full YAML reference, env vars, defaults, validation
- [docs/configuration/remote-configuration.md](docs/configuration/remote-configuration.md) — Remote config mode, three-tier merge, control plane management
- [docs/observability/structured-logging.md](docs/observability/structured-logging.md) — Observability, correlation IDs, metrics
- [docs/testing/integration-testing.md](docs/testing/integration-testing.md) — Integration test suite, mock infrastructure, CI pipeline
- [docs/controlplane/controlplane-pages.md](docs/controlplane/controlplane-pages.md) — All control plane pages with screenshots, features, and usage
- [docs/controlplane/controlplane-architecture.md](docs/controlplane/controlplane-architecture.md) — Control plane architecture, multi-tenant design
- [docs/controlplane/controlplane-api-reference.md](docs/controlplane/controlplane-api-reference.md) — REST API endpoints for the control plane
- [docs/controlplane/controlplane-deployment.md](docs/controlplane/controlplane-deployment.md) — Docker Compose deployment, configuration

## Control Plane Testing Requirements

**MANDATORY**: Any change to frontend code, backend API handlers, or shared types (`pkg/apitypes`) MUST pass all existing frontend tests before being committed. This applies to both human and agent-authored changes.

### Running Tests

```bash
cd frontend
npm test              # Vitest unit/component tests (runs during `npm run build`)
npm run test:e2e      # Playwright e2e tests (requires docker compose up)
```

### Test Structure

- **Unit tests** (`src/**/*.test.{ts,tsx}`): Vitest + React Testing Library. Test stores, hooks, components, and utilities. These run as part of `npm run build`.
- **E2e tests** (`e2e/*.spec.ts`): Playwright. Test full user flows including signup, login, and dashboard rendering. Run against the full stack.

### Test Email Format

E2e tests use generated emails in the format:
`WebTesting-YYYYMMDD-HHMMSSUTC-GUID@OpenAgentFramework.com`

Helper: `frontend/src/test/helpers.ts` → `generateTestEmail()`

### Adding New Tests

When modifying UI components or API endpoints:
1. Add or update Vitest tests for affected components/stores
2. Run `cd frontend && npm test` to verify all tests pass
3. If the change affects user-facing flows (auth, navigation, data display), add or update Playwright e2e tests
4. Run `go build ./...` to verify backend compiles
5. Run `go test ./pkg/... ./internal/...` to verify Go tests pass

## Documentation Instructions

- All new features must be documented in the ./docs folder
- Documentation is to give the LLM and a Human context what the code is doing
- Documentation should be optimized for LLM and avoid duplication of content
- Reference other documentation and/or code with file references
- Avoid duplicating code in the instructions and reference the files the documentation is describing
- Documentation MUST be added with the Issue and PR of the new feature

### Control Plane Documentation Requirements

**MANDATORY**: Any change to the control plane (new pages, modified pages, new features, changed layouts) MUST update the control plane documentation:

1. **Update `docs/controlplane/controlplane-pages.md`** — Add or modify the page description, features list, and usage instructions for any changed or new page
2. **Regenerate screenshots** — Run `cd frontend && npx playwright test e2e/screenshots.spec.ts` to capture updated screenshots after UI changes. If adding a new page, add a screenshot step to `frontend/e2e/screenshots.spec.ts`
3. **Update `docs/controlplane/controlplane-api-reference.md`** if new API endpoints are added or existing ones change
4. **Update `docs/controlplane/controlplane-architecture.md`** if architectural patterns change (new stores, hooks, components)

Screenshots are stored in `docs/controlplane/screenshots/` and referenced from `docs/controlplane/controlplane-pages.md`. The Playwright screenshot test at `frontend/e2e/screenshots.spec.ts` automates capture against the running stack (`docker compose up`).
