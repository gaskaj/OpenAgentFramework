# Control Plane Architecture

The control plane adds fleet management to OpenAgentFramework. It consists of three components: a Go backend API server, a React frontend, and a lightweight reporter library that agents embed to phone home.

## System Overview

```
+------------------+       +-----------------------+       +------------+
|  Agent (remote)  | --->  |  Control Plane API    | <---> | PostgreSQL |
|  pkg/reporter    | HTTP  |  cmd/controlplane     |       +------------+
+------------------+       |  web/                 |
                           +-----------+-----------+
+------------------+                   |
|  Agent (remote)  | -----> ...        | WebSocket + REST
+------------------+                   |
                           +-----------+-----------+
                           |  React Frontend       |
                           |  frontend/            |
                           +-----------------------+
```

## Key Design Decisions

### Centralized Configuration (Remote Mode)

The recommended deployment model is **remote configuration**: agents carry only a minimal local config (`controlplane.url`, `controlplane.api_key`, `config_mode: "remote"`) and fetch all other settings from the control plane. This enables centralized fleet management, consistent configuration across agents, and live updates without restarting agents. See [remote-configuration.md](../configuration/remote-configuration.md).

Local configuration (`config_mode: "local"`) is supported for development and testing but is not recommended for production fleets.

### Separate Binary
The control plane runs as `cmd/controlplane`, completely separate from `cmd/agentctl`. It shares no imports from `internal/`. Shared types live in `pkg/` which is importable by both.

### Multi-Tenancy
Every database table includes `org_id`. All queries are scoped by organization. Middleware extracts the org from the URL slug and validates membership.

### API Key-Bound Agent Identity
Agents authenticate via API keys (not JWT). Each API key carries `agent_type` and `agent_name`, eliminating the need for `agent_name` in local config. Keys are created in the UI (either standalone or via the "Create New Agent" flow), shown once, and stored as SHA-256 hashes. The key prefix (first 8 chars) enables fast lookup.

When an agent connects with its API key, the control plane middleware extracts the agent identity from the key and sets it in the request context. All ingest endpoints (events, heartbeat, config, register) use this identity automatically.

### Agent Provisioning
The UI provides a **Create New Agent** flow (`POST /orgs/{orgSlug}/agents/provision`) that creates both an agent record (status "offline") and a bound API key in one step. The agent appears in the fleet list immediately, before the agent process connects. This gives operators visibility into planned capacity.

### Real-Time Events
When events are ingested, they're pushed to a WebSocket hub that fans out to connected browser clients, scoped by organization.

### Reporter Library
`pkg/reporter` is a fire-and-forget client with internal buffering. Events go into a channel, are batched, and flushed periodically. It has zero impact on agent performance. The reporter does not require `agent_name` — identity is derived from the API key.

## Directory Structure

- `cmd/controlplane/` - Server entry point
- `web/server/` - HTTP server lifecycle
- `web/router/` - Chi router and route definitions
- `web/middleware/` - Auth, API key (with agent identity), logging, tenant scoping
- `web/handler/` - REST API handlers (includes `HandleProvision` for agent+key creation)
- `web/store/` - PostgreSQL store interfaces and implementations
- `web/auth/` - JWT, password hashing, OAuth providers
- `web/ws/` - WebSocket hub and client management
- `web/migrate/` - Database migrations (001: schema, 002: config tables, 003: API key agent identity)
- `web/config/` - Server configuration types
- `pkg/apitypes/` - Shared event and agent types
- `pkg/reporter/` - Agent reporter client library + config fetcher
- `frontend/` - React + TypeScript SPA
- `configs/config.remote.yaml` - Minimal remote config template
- `configs/controlplane.example.yaml` - Server config example

## API Reference

See [controlplane-api-reference.md](controlplane-api-reference.md) for the full endpoint catalog.

## Database Schema

See `web/migrate/migrations/` for the complete schema:
- `001_initial_schema.up.sql` — Core tables
- `002_agent_configs.up.sql` — Configuration management tables
- `003_apikey_agent_identity.up.sql` — API key agent_type/agent_name columns

Tables: `organizations`, `users`, `user_oauth_links`, `org_members`, `agents`, `agent_events`, `api_keys`, `invitations`, `audit_logs`, `agent_type_configs`, `agent_config_overrides`, `config_audit_log`

## Authentication Flows

### User Authentication (JWT)
1. Login via email/password or OAuth (Google/Azure)
2. Server issues short-lived access token (15min) + refresh token (7 days)
3. Frontend attaches access token to all API requests
4. Auto-refresh on 401 response

### Agent Authentication (API Key)
1. Admin provisions agent via "Create New Agent" in the UI — gets agent + API key
2. Agent configures `controlplane.api_key` in its minimal local config
3. Reporter sends `Authorization: Bearer oaf_<key>` on all requests
4. Server hashes key, looks up by prefix + hash, resolves org context and agent identity

## Integration with Agents (Remote Configuration)

The recommended minimal agent config:

```yaml
# configs/config.remote.yaml
controlplane:
  enabled: true
  url: "http://controlplane:8080"
  api_key: "${OAF_API_KEY}"
  config_mode: "remote"
  config_poll_interval: "30s"
```

The agent name and type are derived from the API key. All other configuration (GitHub, Claude, creativity, decomposition, etc.) is fetched from the control plane on startup and polled for updates.

For local development or testing without a control plane, use `config_mode: "local"` with a full `config.yaml` containing all settings. See `configs/config.example.yaml`.

## Deployment

See [controlplane-deployment.md](controlplane-deployment.md) for Docker and production deployment instructions.
