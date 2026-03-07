# Integration Testing

## Overview

The integration test suite (`internal/integration/`) validates agent-to-agent communication, workflow handoffs, shared state management, and race condition safety. Tests run real agent instances against mock infrastructure (GitHub API, Claude API, state store) in isolated environments.

For the package API, see [package-reference.md](package-reference.md). For CI pipeline details, see `.github/workflows/integration-tests.yml`.

## Test Suites

| File | What It Tests |
|------|---------------|
| `agent_communication_test.go` | Message protocol compliance, concurrent agent processing, timeout handling, retry mechanisms, state transition consistency |
| `workflow_handoff_test.go` | Developer-to-QA handoffs, epic decomposition handoffs, context preservation across transitions, error handling during handoffs, handoff performance |
| `shared_state_test.go` | Concurrent reads/writes to shared state, resource contention (multiple agents claiming one issue), race condition detection, state consistency under failure, long-running state operations |
| `simple_agent_test.go` | Smoke tests for agent lifecycle, GitHub mock contract, state store mock contract, error simulation and recovery |

## Running Tests

### Locally

```bash
make test-integration                    # Run all integration tests with -race
go test -v -race -timeout=30m ./internal/integration/...   # Direct invocation
make test-race                           # Race-condition tests (3x runs)
```

### Docker-Based

```bash
make docker-test                         # Full test suite in Docker
make docker-test-performance             # Performance tests with resource limits (512m RAM, 2 CPUs)
make docker-test-race                    # Race detection tests in Docker
make docker-test-all                     # All Docker test suites
```

Docker Compose (`docker-compose.test.yml`) starts four services:
- **test-runner**: Runs integration + unit tests via `test.Dockerfile`
- **mock-github-api**: nginx serving fixtures from `test/fixtures/github-api`
- **mock-claude-api**: nginx serving fixtures from `test/fixtures/claude-api`
- **test-state-store**: Redis 7 (memory-only) for state store testing

## CI Pipeline

**File**: `.github/workflows/integration-tests.yml`

Triggers: push/PR to `main`/`develop`, daily at 2 AM UTC.

| Job | Description |
|-----|-------------|
| `integration-tests` | Matrix of 4 suites (agent-communication, workflow-handoffs, shared-state, race-conditions) running in parallel with coverage |
| `performance-tests` | Runs after integration tests pass; generates performance report |
| `coverage-report` | Combines per-suite coverage, generates HTML report, posts PR comment with metrics |
| `notification` | Creates GitHub step summary with results from all jobs |

Environment variables: `INTEGRATION_TEST_MODE=true`, `GO_TEST_TIMEOUT=30m`, `GO_TEST_PARALLEL=4`.

## Test Infrastructure

**Location**: `internal/integration/test_helpers.go`, `internal/integration/mock_services.go`

### TestEnvironment

The `TestEnvironment` struct is the main test harness. `NewTestEnvironment(t)` creates:
- A bare Git repo (with `README.md` + `go.mod`) for clone operations
- A `MockClaudeServer` (httptest) listening on a random port
- A real `claude.Client` pointed at the mock server
- A file-based `state.Store` in `t.TempDir()`
- Structured logger, metrics, and error manager instances

Key methods: `CreateDependencies()`, `CreateDeveloperAgent()`, `CreateOrchestrator()`, `RunWithTimeout()`, `SimulateGitHubIssue()`, `SimulateAgentFailure()`, `SimulateConcurrentAccess()`, assertion helpers (`AssertWorkflowState`, `AssertIssueLabels`, `AssertCommentCreated`, `AssertPRCreated`).

### MockGitHubClient

Thread-safe (`sync.RWMutex`) mock implementing `ghub.Client`. Tracks issues, labels, comments, and PRs in memory. Supports error injection via `SimulateError()` / `ClearError()`.

### MockClaudeServer

HTTP mock for the Claude `/v1/messages` endpoint. Detects tool definitions in requests and returns `tool_use` responses (file writes). Supports a FIFO response queue (`EnqueueResponse()`) for test-specific responses and error simulation (`SetHTTPError()` / `ClearHTTPError()`).

### MockStore

Thread-safe (`sync.RWMutex`) mock implementing `state.Store`. Returns defensive copies of state to avoid race conditions. Supports error injection.
