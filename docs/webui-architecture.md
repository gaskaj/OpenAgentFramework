# WebUI Architecture

The WebUI adds a fleet management control plane to OpenAgentFramework. It consists of three components: a Go backend API server, a React frontend, and a lightweight reporter library that agents embed to phone home.

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

### Separate Binary
The control plane runs as `cmd/controlplane`, completely separate from `cmd/agentctl`. It shares no imports from `internal/`. Shared types live in `pkg/` which is importable by both.

### Multi-Tenancy
Every database table includes `org_id`. All queries are scoped by organization. Middleware extracts the org from the URL slug and validates membership.

### Agent Authentication
Agents authenticate via API keys (not JWT). Keys are created in the UI, shown once, and stored as SHA-256 hashes. The key prefix (first 8 chars) enables fast lookup.

### Real-Time Events
When events are ingested, they're pushed to a WebSocket hub that fans out to connected browser clients, scoped by organization.

### Reporter Library
`pkg/reporter` is a fire-and-forget client with internal buffering. Events go into a channel, are batched, and flushed periodically. It has zero impact on agent performance.

## Directory Structure

- `cmd/controlplane/` - Server entry point
- `web/server/` - HTTP server lifecycle
- `web/router/` - Chi router and route definitions
- `web/middleware/` - Auth, CORS, logging, rate limiting, tenant scoping
- `web/handler/` - REST API handlers
- `web/store/` - PostgreSQL store interfaces and implementations
- `web/auth/` - JWT, password hashing, OAuth providers
- `web/ws/` - WebSocket hub and client management
- `web/migrate/` - Database migrations
- `web/config/` - Server configuration types
- `pkg/apitypes/` - Shared event and agent types
- `pkg/reporter/` - Agent reporter client library
- `frontend/` - React + TypeScript SPA
- `configs/controlplane.example.yaml` - Server config example

## API Reference

See [webui-api-reference.md](webui-api-reference.md) for the full endpoint catalog.

## Database Schema

See `web/migrate/migrations/001_initial_schema.up.sql` for the complete schema.

Tables: `organizations`, `users`, `user_oauth_links`, `org_members`, `agents`, `agent_events`, `api_keys`, `invitations`, `audit_logs`

## Authentication Flows

### User Authentication (JWT)
1. Login via email/password or OAuth (Google/Azure)
2. Server issues short-lived access token (15min) + refresh token (7 days)
3. Frontend attaches access token to all API requests
4. Auto-refresh on 401 response

### Agent Authentication (API Key)
1. Admin creates API key in UI
2. Agent configures `pkg/reporter` with the key
3. Reporter sends `Authorization: Bearer oaf_<key>` on all requests
4. Server hashes key, looks up by prefix + hash, resolves org context

## Integration with Existing Agents

Add to agent config YAML:
```yaml
controlplane:
  enabled: true
  url: "https://controlplane.example.com"
  api_key: "oaf_..."
  agent_name: "dev-agent-1"
  heartbeat_interval: "30s"
```

The reporter hooks into existing workflow transitions in `internal/developer/workflow.go`. See `internal/config/config.go` for the `ControlPlaneConfig` struct.

## Deployment

See [webui-deployment.md](webui-deployment.md) for Docker and production deployment instructions.
