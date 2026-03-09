# Control Plane API Reference

Base URL: `/api/v1`

## Releases

| Method | Path | Description |
|--------|------|-------------|
| GET | `/releases/latest` | Latest GitHub Release info with download URLs (public, cached 5min) |

Returns `{ tag_name, published_at, html_url, assets: [{ name, browser_download_url, size }] }`.

## Authentication

### Public Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/register` | Register with email, password, display name, org name |
| POST | `/auth/login` | Login with email/password, returns JWT |
| POST | `/auth/refresh` | Refresh access token |
| GET | `/auth/oauth/{provider}` | Initiate OAuth flow (google, azure) |
| GET | `/auth/oauth/{provider}/callback` | OAuth callback |
| POST | `/invitations/accept` | Accept invitation by token |

### Protected Endpoints (JWT Required)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/auth/me` | Get current user |

## Organizations

All require JWT auth.

| Method | Path | Description |
|--------|------|-------------|
| POST | `/orgs` | Create organization |
| GET | `/orgs` | List user's organizations |
| GET | `/orgs/{orgSlug}` | Get organization |
| PUT | `/orgs/{orgSlug}` | Update organization |
| DELETE | `/orgs/{orgSlug}` | Delete organization (owner only) |
| GET | `/orgs/{orgSlug}/members` | List members |
| PUT | `/orgs/{orgSlug}/members/{userId}` | Update member role |
| DELETE | `/orgs/{orgSlug}/members/{userId}` | Remove member |

## Agents

| Method | Path | Description |
|--------|------|-------------|
| POST | `/orgs/{orgSlug}/agents` | Register agent (used by agent process) |
| POST | `/orgs/{orgSlug}/agents/provision` | **Provision agent + API key** (used by Control Plane UI) |
| GET | `/orgs/{orgSlug}/agents` | List agents (filter: status, type, search) |
| GET | `/orgs/{orgSlug}/agents/{agentId}` | Get agent details |
| PUT | `/orgs/{orgSlug}/agents/{agentId}` | Update agent |
| DELETE | `/orgs/{orgSlug}/agents/{agentId}` | Deregister agent |

### Provision Endpoint

`POST /orgs/{orgSlug}/agents/provision` creates both an agent (status "offline") and a bound API key in one call. This is the recommended way to set up new agents from the UI.

**Request:**
```json
{
  "agent_type": "developer",
  "name": ""
}
```

If `name` is empty, it auto-generates as `{agent_type}-{XX}` where XX increments per type.

**Response (201):**
```json
{
  "agent": { "id": "...", "name": "developer-01", "agent_type": "developer", "status": "offline", ... },
  "api_key": { "id": "...", "key_prefix": "f0d78f01", "agent_type": "developer", "agent_name": "developer-01" },
  "key": "oaf_f0d78f01..."
}
```

The `key` field is the raw API key, shown only once. Store it securely.

## Events

| Method | Path | Description |
|--------|------|-------------|
| GET | `/orgs/{orgSlug}/events` | Query events (filter: type, severity, agent, since, until) |
| GET | `/orgs/{orgSlug}/events/stats` | Event counts by type |
| GET | `/orgs/{orgSlug}/agents/{agentId}/events` | Agent-specific events |

## API Keys

| Method | Path | Description |
|--------|------|-------------|
| POST | `/orgs/{orgSlug}/apikeys` | Create API key with `agent_type` and optional `name` |
| GET | `/orgs/{orgSlug}/apikeys` | List API keys (includes agent_type, agent_name) |
| DELETE | `/orgs/{orgSlug}/apikeys/{keyId}` | Revoke API key |

Each API key is bound to an `agent_type` and `agent_name`. The agent name is derived from the key — agents do not need `agent_name` in their local config.

## Configuration Management

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

## Invitations

| Method | Path | Description |
|--------|------|-------------|
| POST | `/orgs/{orgSlug}/invitations` | Send invitation |
| GET | `/orgs/{orgSlug}/invitations` | List pending invitations |
| DELETE | `/orgs/{orgSlug}/invitations/{invId}` | Cancel invitation |

## Audit Logs

| Method | Path | Description |
|--------|------|-------------|
| GET | `/orgs/{orgSlug}/audit` | Query audit logs (filter: action, user, resource_type, since, until) |

## Tunnel Management

Manage the built-in ngrok tunnel for public access. All endpoints require JWT auth.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/tunnel` | Get tunnel status (enabled, public_url, error, has_auth_token) |
| POST | `/tunnel` | Toggle tunnel on/off: `{ "enabled": true\|false }` |
| PUT | `/tunnel/token` | Save or clear ngrok authtoken: `{ "auth_token": "..." }` |

Saving a token automatically starts the tunnel. Saving an empty string clears the token and stops the tunnel. See [ngrok Tunnel Guide](ngrok-tunnel.md) for setup instructions.

## Agent Ingestion (API Key Auth)

These endpoints use API key authentication (`Authorization: Bearer oaf_...`). The agent identity (name, type) is derived from the API key — no `agent_name` field is needed in request bodies.

| Method | Path | Description |
|--------|------|-------------|
| POST | `/ingest/register` | Agent self-registration (updates agent to "online") |
| POST | `/ingest/events` | Single event ingestion |
| POST | `/ingest/events/batch` | Batch event ingestion |
| POST | `/ingest/heartbeat` | Agent heartbeat |
| GET | `/ingest/config` | Fetch merged config (supports ETag) |

The `/ingest/config` endpoint resolves the agent from the API key's bound identity. No query parameters are required.

## WebSocket

| Path | Description |
|------|-------------|
| GET `/orgs/{orgSlug}/ws?token=<jwt>` | Real-time event stream |

Events are JSON objects matching the `AgentEvent` type from `pkg/apitypes/events.go`.

## Response Format

### Success
```json
{
  "data": { ... },
  "total": 100,
  "limit": 20,
  "offset": 0
}
```

### Error
```json
{
  "error": "descriptive error message"
}
```

## Pagination

List endpoints accept `limit` (default 20, max 100) and `offset` query parameters.
