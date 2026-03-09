# Remote Configuration

Remote configuration is the **recommended deployment mode** for production agent fleets. Agents carry only a minimal local config and fetch all other settings from the control plane control plane. This enables:

- **Centralized management** — configure all agents from a single UI
- **Consistent fleet settings** — type-level defaults propagate to all agents
- **Live updates** — change settings without restarting agents
- **Per-agent overrides** — customize individual agents while inheriting global defaults
- **Audit trail** — track who changed what and when

## Architecture

Configuration uses a **three-tier merge** strategy:

1. **System defaults** (Go code in `internal/config/defaults.go`)
2. **Agent type config** (stored in `agent_type_configs` table, managed per agent type per org)
3. **Per-agent overrides** (stored in `agent_config_overrides` table, per individual agent)

Each tier overrides the previous. Only non-nil fields in the remote config override the local config.

## Agent Setup

### 1. Provision the Agent

In the control plane, go to **Agents** and click **Create New Agent**. Select the agent type and optionally provide a name. The UI creates the agent (status "offline") and generates a bound API key. Copy the API key — it is shown only once.

### 2. Create the Local Config

The agent only needs a minimal local config file:

```yaml
# config.yaml (or set via environment variables)
controlplane:
  enabled: true
  url: "http://controlplane:8080"
  api_key: "${OAF_API_KEY}"
  config_mode: "remote"
  config_poll_interval: "30s"
```

A template is available at `configs/config.remote.yaml`.

The agent name and type are derived from the API key — no `agent_name` field is needed. All other configuration (GitHub, Claude, creativity, decomposition, etc.) is fetched from the control plane.

### 3. Configure via control plane

Set the global configuration for the agent type in **Configuration** (`/config`). This covers Claude AI, agent settings, creativity, decomposition, memory, logging, shutdown, and error handling.

For per-agent settings (GitHub owner/repo/token, Claude API key), navigate to the agent's detail page and click **Configure** to set overrides.

### 4. Start the Agent

```bash
./bin/agentctl start --config config.yaml
```

On startup, the agent fetches its merged configuration from the control plane, applies it, and begins operating.

## How It Works

### Startup
1. Agent loads local `config.yaml` (with defaults applied; validation skipped for remote mode)
2. Agent calls `GET /api/v1/ingest/config` (agent identity resolved from API key)
3. The control plane merges the agent type config with any per-agent overrides using PostgreSQL `jsonb ||` operator
4. Agent applies the remote config on top of local defaults via `config.MergeRemoteConfig()`
5. Agent validates the merged config and starts

### Polling
After startup, the agent polls the control plane at `config_poll_interval` for updates. The endpoint supports **ETag-based conditional requests** — if the config hasn't changed (304 Not Modified), no processing occurs.

When configuration changes are detected, `MergeRemoteConfig()` applies the new values to the running config.

## control plane Configuration Pages

### Global Type Configuration (`/config`)
- Set defaults for each agent type (e.g., all "developer" agents)
- Structured form with collapsible sections: Claude AI, Agent Settings, Creativity, Decomposition, Memory, Logging, Shutdown, Error Handling
- JSON editor toggle for advanced users
- Version tracking and audit trail
- GitHub and Claude credentials are per-agent, not global

### Per-Agent Override (`/agents/:agentId/config`)
- Override specific fields for a single agent (including GitHub owner/repo/token and Claude API key)
- Inherited values from type config shown as placeholders
- Amber highlighting on overridden fields
- "Preview Merged" button shows the effective final config
- Reset button removes all overrides

## Local Configuration (Development/Testing)

For local development or running without a control plane, use `config_mode: "local"` (the default):

```yaml
controlplane:
  enabled: true
  url: "http://localhost:8080"
  api_key: "oaf_..."
  config_mode: "local"
  heartbeat_interval: "30s"

github:
  token: "${GITHUB_TOKEN}"
  owner: "myorg"
  repo: "myrepo"
  # ... all other settings inline
```

In local mode, the agent reads all configuration from the YAML file and only reports events to the control plane. See `configs/config.example.yaml` for the full reference.

## API Endpoints

### Management API (JWT auth, org-scoped)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/orgs/{orgSlug}/config/types` | List all type configs |
| GET | `/orgs/{orgSlug}/config/types/{agentType}` | Get type config |
| PUT | `/orgs/{orgSlug}/config/types/{agentType}` | Create/update type config |
| GET | `/orgs/{orgSlug}/config/agents/{agentId}` | Get agent override |
| PUT | `/orgs/{orgSlug}/config/agents/{agentId}` | Create/update agent override |
| DELETE | `/orgs/{orgSlug}/config/agents/{agentId}` | Delete agent override |
| GET | `/orgs/{orgSlug}/config/agents/{agentId}/merged` | Preview merged config |
| GET | `/orgs/{orgSlug}/config/audit` | Config change audit trail |

### Agent Ingest API (API key auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/ingest/config` | Fetch merged config (agent identity from API key, supports ETag) |

## Database Schema

See `web/migrate/migrations/002_agent_configs.up.sql`:
- `agent_type_configs` — One config per agent type per org (JSONB)
- `agent_config_overrides` — Per-agent overrides (JSONB), unique per agent
- `config_audit_log` — Tracks all config changes with previous/new values

## Key Files

- `pkg/apitypes/config.go` — Shared `RemoteConfig` types (pointer fields for optional overrides)
- `internal/config/remote.go` — `MergeRemoteConfig()` merge logic
- `internal/config/remote_test.go` — Merge tests
- `pkg/reporter/config_fetcher.go` — Agent-side config polling (no agent_name param needed)
- `web/store/config_store.go` — Database store for config CRUD
- `web/handler/config_handler.go` — HTTP handlers for config API
- `frontend/src/pages/ConfigPage.tsx` — Type-level config editor
- `frontend/src/pages/AgentConfigOverridePage.tsx` — Per-agent override editor
- `frontend/src/store/config-store.ts` — Zustand state store
- `frontend/src/api/config.ts` — API client functions

## Dynamically Hot-Reloadable Fields

These fields can change without agent restart:
- `creativity.*`, `decomposition.*`, `memory.*`
- `logging.level`, `github.poll_interval`
- `error_handling.retry.*`, `error_handling.circuit_breaker.*`

Fields requiring restart: `github.token`, `github.owner`, `github.repo`, `claude.api_key`, `claude.model`, `agents.developer.workspace_dir`
