# Release Pipeline

The release pipeline automatically builds cross-platform `agentctl` binaries and publishes them as GitHub Release assets when a version tag is pushed.

## Cutting a Release

```bash
git tag v0.2.0
git push origin v0.2.0
```

This triggers `.github/workflows/release.yml`, which:

1. Builds `agentctl` for 5 platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)
2. Injects the version, commit hash, and build date via `-ldflags`
3. Creates a GitHub Release with auto-generated notes and attaches all binaries

## Binary Naming

Assets follow the pattern `agentctl-{os}-{arch}[.exe]`:

| Asset | Platform |
|-------|----------|
| `agentctl-darwin-arm64` | macOS Apple Silicon |
| `agentctl-darwin-amd64` | macOS Intel |
| `agentctl-linux-amd64` | Linux x86_64 |
| `agentctl-linux-arm64` | Linux ARM64 |
| `agentctl-windows-amd64.exe` | Windows x86_64 |

## Version Injection

The `internal/version` package exposes three variables set at build time:

```go
var (
    Version   = "dev"       // git tag or describe
    Commit    = "unknown"   // short commit hash
    BuildDate = "unknown"   // ISO 8601 UTC timestamp
)
```

The Makefile and release workflow inject these via:

```
-ldflags "-X .../version.Version=v0.2.0 -X .../version.Commit=abc1234 -X .../version.BuildDate=2026-03-09T..."
```

Running `agentctl --version` shows the injected version.

## Local Builds

`make build` automatically injects version from `git describe --tags --always --dirty`. If no tags exist, it falls back to `"dev"`.

## Control Plane Integration

The control plane exposes `GET /api/v1/releases/latest` which proxies the GitHub Releases API with a 5-minute in-memory cache. The Create Agent page uses this endpoint to show platform-specific download links and a copy-paste setup script after provisioning an agent.

## Key Files

| File | Purpose |
|------|---------|
| `internal/version/version.go` | Build-time version variables |
| `.github/workflows/release.yml` | CI/CD release workflow |
| `web/handler/release_handler.go` | GitHub Releases proxy endpoint |
| `frontend/src/components/DownloadInstructions.tsx` | Download UI component |
| `frontend/src/api/releases.ts` | Frontend API client for releases |
