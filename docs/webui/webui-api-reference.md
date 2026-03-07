# WebUI API Reference

Base URL: `/api/v1`

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
| POST | `/orgs/{orgSlug}/agents` | Register agent |
| GET | `/orgs/{orgSlug}/agents` | List agents (filter: status, type, search) |
| GET | `/orgs/{orgSlug}/agents/{agentId}` | Get agent details |
| PUT | `/orgs/{orgSlug}/agents/{agentId}` | Update agent |
| DELETE | `/orgs/{orgSlug}/agents/{agentId}` | Deregister agent |

## Events

| Method | Path | Description |
|--------|------|-------------|
| GET | `/orgs/{orgSlug}/events` | Query events (filter: type, severity, agent, since, until) |
| GET | `/orgs/{orgSlug}/events/stats` | Event counts by type |
| GET | `/orgs/{orgSlug}/agents/{agentId}/events` | Agent-specific events |

## API Keys

| Method | Path | Description |
|--------|------|-------------|
| POST | `/orgs/{orgSlug}/apikeys` | Create API key (key shown once in response) |
| GET | `/orgs/{orgSlug}/apikeys` | List API keys (prefix only) |
| DELETE | `/orgs/{orgSlug}/apikeys/{keyId}` | Revoke API key |

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

## Agent Ingestion (API Key Auth)

These endpoints use API key authentication (`Authorization: Bearer oaf_...`).

| Method | Path | Description |
|--------|------|-------------|
| POST | `/ingest/register` | Agent self-registration |
| POST | `/ingest/events` | Single event ingestion |
| POST | `/ingest/events/batch` | Batch event ingestion |
| POST | `/ingest/heartbeat` | Agent heartbeat |

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
