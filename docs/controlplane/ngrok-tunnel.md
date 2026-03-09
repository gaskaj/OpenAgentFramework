# Public Access with ngrok

OpenAgentFramework includes built-in ngrok tunnel support, allowing you to expose your control plane to the internet with a single configuration step. This enables agents running on remote machines, CI/CD pipelines, cloud VMs, or colleagues' laptops to connect back to your locally-hosted control plane — no port forwarding, firewall rules, or cloud deployment required.

## Why ngrok?

Running the control plane locally with Docker Compose is the simplest way to get started. But agents need to reach the control plane over HTTP to report events, fetch configuration, and stream logs. If the control plane and agents aren't on the same network, you need a public URL.

ngrok solves this by creating a secure tunnel from a public URL (e.g. `https://your-subdomain.ngrok-free.dev`) to your local control plane. All traffic is encrypted end-to-end via TLS.

**Use cases:**
- Run agents on cloud VMs or CI runners that report back to your local control plane
- Share a live demo of your control plane with teammates
- Connect agents across different networks without VPN
- Quickly test webhook integrations that require a public URL

## Setup

### 1. Get an ngrok Account

Sign up at [dashboard.ngrok.com/signup](https://dashboard.ngrok.com/signup). The free tier is sufficient — it provides one tunnel with a random subdomain.

### 2. Copy Your Authtoken

Go to [dashboard.ngrok.com/get-started/your-authtoken](https://dashboard.ngrok.com/get-started/your-authtoken) and copy your authtoken.

### 3. Start the Control Plane

```bash
docker compose up -d
```

### 4. Configure the Tunnel in the UI

1. Open `http://localhost:5173/settings`
2. Under **Public Tunnel (ngrok)**, paste your authtoken into the input field
3. Click **Save**

The tunnel starts automatically. Within a few seconds, the public URL appears with a green indicator. The URL is something like `https://random-name.ngrok-free.dev`.

### 5. Point Agents at the Public URL

Use the ngrok URL as the `controlplane.url` in your agent configs:

```yaml
controlplane:
  enabled: true
  url: "https://random-name.ngrok-free.dev"
  api_key: "oaf_your_api_key_here"
  config_mode: "remote"
```

Agents can now reach your local control plane from anywhere on the internet.

## How It Works

The ngrok tunnel is managed by the control plane server process:

1. **On boot**: the control plane checks the database for a saved authtoken. If one exists, it starts an ngrok process that tunnels to the frontend container (which proxies `/api/` to the control plane API).
2. **The tunnel URL** is discovered by polling ngrok's local API (`http://127.0.0.1:4040/api/tunnels`).
3. **The Settings page** displays the current tunnel status and URL, with copy-to-clipboard and open-in-browser buttons.
4. **The authtoken** is persisted in the `system_settings` database table — it survives container restarts.

### Architecture

```
Internet                         Docker Compose
  │
  │  https://xyz.ngrok-free.dev
  ▼
┌─────────┐      ┌──────────────┐      ┌───────────────┐
│  ngrok  │─────▶│  Frontend    │─────▶│ Control Plane │
│ (tunnel)│ :80  │  (nginx)     │ :8080│ (Go API)      │
└─────────┘      │  /api/ proxy │      └───────┬───────┘
                 └──────────────┘              │
                                        ┌──────▼──────┐
                                        │ PostgreSQL  │
                                        └─────────────┘
```

The ngrok process runs inside the controlplane container and connects to the frontend container over the Docker network. This means the public URL serves both the React SPA and the API — exactly like accessing `http://localhost:5173` locally.

## Managing the Tunnel

### Toggle On/Off

The Settings page shows a toggle switch next to the tunnel section. You can turn the tunnel off without removing your authtoken — the token stays saved, and you can re-enable the tunnel at any time.

### Change Authtoken

Paste a new authtoken and click Save. The existing tunnel stops and a new one starts with the updated token.

### Clear Authtoken

Click the **Clear** button next to the authtoken input. This removes the token from the database and stops the tunnel. On the next restart, no tunnel will start.

### API Endpoints

The tunnel is managed via these JWT-protected endpoints:

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/tunnel` | Get tunnel status (enabled, public_url, error, has_auth_token) |
| `POST` | `/api/v1/tunnel` | Toggle tunnel on/off: `{ "enabled": true }` |
| `PUT` | `/api/v1/tunnel/token` | Save authtoken: `{ "auth_token": "..." }` |

## ngrok Free Tier Limitations

The free ngrok plan has some constraints to be aware of:

- **Random subdomain** — the URL changes each time the tunnel restarts (paid plans offer custom subdomains)
- **Interstitial page** — first-time visitors see an ngrok warning page they must click through
- **Rate limits** — sufficient for development and small teams
- **Single tunnel** — one active tunnel per account

For production use with stable URLs, consider an [ngrok paid plan](https://ngrok.com/pricing) or deploy the control plane to a cloud provider.

## Troubleshooting

### Tunnel shows "Establishing..." but no URL appears

The ngrok process may have failed to authenticate. Check the controlplane container logs:

```bash
docker logs openagentframework-controlplane-1 2>&1 | grep ngrok
```

Common causes:
- Invalid authtoken — verify it at [dashboard.ngrok.com](https://dashboard.ngrok.com/get-started/your-authtoken)
- ngrok rate limit exceeded — wait a few minutes and toggle off/on

### Tunnel URL works but shows 502 Bad Gateway

The frontend container may not be running. Check:

```bash
docker compose ps
```

All three containers (postgres, controlplane, frontend) should be running.

### Agents can't connect to the tunnel URL

- Verify the agent's `controlplane.url` matches the ngrok URL exactly (including `https://`)
- The ngrok free tier interstitial page can block programmatic access. Add the header `ngrok-skip-browser-warning: true` or use a paid plan for production agents.

### "ngrok is not installed" error

The Dockerfile includes ngrok. If you see this error, rebuild the controlplane image:

```bash
docker compose build --no-cache controlplane
docker compose up -d
```

## File References

| File | Description |
|------|-------------|
| `web/tunnel/manager.go` | ngrok process lifecycle management |
| `web/handler/tunnel_handler.go` | API endpoints for tunnel status/toggle/token |
| `web/store/settings_store.go` | Key-value persistence for authtoken |
| `web/migrate/migrations/004_system_settings.up.sql` | Database migration for settings table |
| `frontend/src/api/tunnel.ts` | Frontend API client |
| `frontend/src/pages/OrgSettingsPage.tsx` | Settings page with tunnel UI |
| `cmd/controlplane/main.go` | Boot-time tunnel initialization |
