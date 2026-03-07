# WebUI Deployment

## Development Setup

### Prerequisites
- Go 1.25+
- Node.js 20+
- PostgreSQL 16+
- Docker + Docker Compose (optional)

### Quick Start with Docker Compose

```bash
docker compose up -d
```

This starts PostgreSQL, the control plane API (port 8080), and the frontend (port 5173).

### Manual Development Setup

1. Start PostgreSQL:
```bash
docker run -d --name oaf-postgres \
  -e POSTGRES_DB=oaf_controlplane \
  -e POSTGRES_USER=oaf \
  -e POSTGRES_PASSWORD=oaf_dev_password \
  -p 5432:5432 \
  postgres:16-alpine
```

2. Build and run the control plane:
```bash
export DB_USER=oaf DB_PASSWORD=oaf_dev_password JWT_SECRET=dev-secret
make build-controlplane
./bin/controlplane --config configs/controlplane.example.yaml
```

3. Run the frontend dev server:
```bash
cd frontend
npm install
npm run dev
```

Frontend runs on http://localhost:5173, proxying API requests to http://localhost:8080.

## Production Deployment

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DB_USER` | Yes | PostgreSQL username |
| `DB_PASSWORD` | Yes | PostgreSQL password |
| `JWT_SECRET` | Yes | Secret for signing JWTs (min 32 chars) |
| `GOOGLE_CLIENT_ID` | No | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | No | Google OAuth client secret |
| `AZURE_TENANT_ID` | No | Azure AD tenant ID |
| `AZURE_CLIENT_ID` | No | Azure AD client ID |
| `AZURE_CLIENT_SECRET` | No | Azure AD client secret |

### Database Migrations

Migrations run automatically on server startup. To run manually:

```bash
./bin/controlplane migrate --config configs/controlplane.yaml
```

### Connecting Agents

1. Register an organization and create an API key in the UI
2. Add control plane config to agent's `config.yaml`:

```yaml
controlplane:
  enabled: true
  url: "https://your-controlplane.example.com"
  api_key: "oaf_your_api_key_here"
  agent_name: "dev-agent-prod-1"
  heartbeat_interval: "30s"
```

3. Restart the agent - it will register and begin reporting events

### Health Check

```
GET /healthz
```

Returns 200 if the server and database connection are healthy.
