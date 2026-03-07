# API Versioning Strategy

This document describes the comprehensive API versioning framework implemented in the OpenAgent control plane. The framework ensures backward compatibility while allowing for safe evolution of the API.

## Table of Contents

1. [Overview](#overview)
2. [Version Detection](#version-detection)
3. [Supported Versions](#supported-versions)
4. [Deprecation Process](#deprecation-process)
5. [Client Integration](#client-integration)
6. [WebSocket Versioning](#websocket-versioning)
7. [Agent Communication](#agent-communication)
8. [Migration Guides](#migration-guides)

## Overview

The API versioning framework provides:

- **Multiple version detection strategies** with clear precedence
- **Graceful deprecation** with configurable timelines
- **Backward compatibility** for existing integrations
- **Clear migration paths** between versions
- **Comprehensive error handling** for version mismatches

### Design Principles

1. **Backward Compatibility**: Existing clients continue to work during API evolution
2. **Explicit Versioning**: Version requirements are clearly specified and validated
3. **Graceful Degradation**: Deprecated versions receive warnings before removal
4. **Developer Experience**: Clear error messages and migration guidance

## Version Detection

The API supports multiple version detection strategies in order of precedence:

### 1. Accept Header (Recommended)
```http
Accept: application/vnd.openagent.v2+json
```

This is the preferred method as it follows HTTP content negotiation standards.

### 2. API-Version Header
```http
API-Version: v2
```

Simple and explicit version specification.

### 3. Query Parameter
```http
GET /api/v1/agents?api-version=v2
```

Useful for testing and simple integrations.

### 4. URL Path
```http
GET /api/v2/agents
```

Supported but not recommended for new integrations.

### 5. Default Version
If no version is specified, the API uses the configured default version (currently `v1`).

## Supported Versions

Current API versions and their status:

| Version | Status | Default | Deprecated | Sunset Date |
|---------|--------|---------|------------|-------------|
| v1      | Active | ✓       | ✗          | -           |

### Version Configuration

API versions are configured in `configs/api-versions.yaml`:

```yaml
default_version: "v1"
deprecation_warning: true
supported_versions:
  - version: "v1"
    is_default: true
    is_deprecated: false
```

## Deprecation Process

The deprecation process follows a structured timeline:

### Phase 1: Warning (6 months)
- Version marked as deprecated
- Deprecation warnings added to response headers
- Documentation updated with migration guidance
- No breaking changes to functionality

### Phase 2: Sunset (Final 3 months)
- API returns HTTP 410 Gone
- Response includes sunset date and migration guide
- Critical bug fixes only

### Phase 3: Removal
- Version removed from supported list
- Handler code cleanup
- Documentation archival

### Deprecation Headers

When using a deprecated version, the API includes these response headers:

```http
Deprecation: true
Warning: 299 - "API version v0 is deprecated"
Deprecation-Date: 2024-01-15T00:00:00Z
Sunset: Mon, 15 Jul 2024 00:00:00 GMT
Link: </docs/webui/api-versioning.md>; rel="deprecation"; type="text/html"
```

## Client Integration

### Frontend Integration

The frontend automatically includes version headers in API requests:

```typescript
import { createVersionHeaders, CURRENT_API_VERSION } from '@/api/versions';

// Automatic version headers
const headers = createVersionHeaders(CURRENT_API_VERSION);
// Result: { Accept: "application/vnd.openagent.v1+json" }
```

### Version Response Handling

The client automatically parses version information from responses:

```typescript
import { parseVersionResponse } from '@/api/versions';

// Parse version info from response headers
const versionInfo = parseVersionResponse(response.headers);
if (versionInfo.isDeprecated) {
  console.warn(`API version ${versionInfo.version} is deprecated`);
}
```

### Error Handling

The client provides specific error types for version-related issues:

```typescript
import { VersionError, DeprecatedVersionError, SunsetVersionError } from '@/api/versions';

try {
  await apiCall();
} catch (error) {
  if (error instanceof SunsetVersionError) {
    // Show upgrade prompt to user
    showUpgradeModal(error.migrationGuide);
  }
}
```

## WebSocket Versioning

WebSocket connections negotiate version during the handshake:

### Connection Handshake
```javascript
const ws = new WebSocket('/api/v1/orgs/my-org/ws', ['openagent.v1']);
```

### Protocol Version Headers
```http
Sec-WebSocket-Protocol: openagent.v1
API-Version: v1
```

## Agent Communication

Agents specify their preferred API version during registration and event reporting.

### Agent Configuration

```go
reporter, err := reporter.New(reporter.Config{
    ControlPlaneURL: "https://api.openagent.dev",
    APIKey:          "key_abc123",
    AgentName:       "my-agent",
    APIVersion:      "v1", // Specify API version
})
```

### Version Headers

All agent requests include version information:

```http
POST /api/v1/ingest/events/batch
Accept: application/vnd.openagent.v1+json
Authorization: Bearer key_abc123
Content-Type: application/json
```

### Event Payloads

Events include the API version used:

```json
{
  "agent_name": "my-agent",
  "events": [
    {
      "event_type": "agent.heartbeat",
      "severity": "info",
      "timestamp": "2024-01-15T10:00:00Z",
      "api_version": "v1"
    }
  ]
}
```

## Error Responses

### Unsupported Version (400 Bad Request)

```json
{
  "error": "unsupported_api_version",
  "message": "API version \"v999\" is not supported",
  "supported_versions": ["v1"],
  "default_version": "v1"
}
```

### Sunset Version (410 Gone)

```json
{
  "error": "api_version_sunset",
  "message": "API version \"v0\" was sunset on 2024-07-15",
  "sunset_date": "2024-07-15T00:00:00Z",
  "migration_guide": "/docs/webui/api-versioning.md"
}
```

## Migration Guides

### Migrating to v1

This section will be updated when new API versions are introduced.

#### Breaking Changes
- None (v1 is the initial version)

#### New Features
- Complete REST API for agent management
- Event ingestion and querying
- Real-time WebSocket updates
- Organization and user management

#### Migration Steps
1. Update client to specify `Accept: application/vnd.openagent.v1+json`
2. Update WebSocket connections to use protocol `openagent.v1`
3. Update agent configurations to include `APIVersion: "v1"`

## Configuration Management

### Server Configuration

Add versioning configuration to your server config:

```yaml
versioning:
  default_version: "v1"
  deprecation_warning: true
  supported_versions:
    - version: "v1"
      is_default: true
      is_deprecated: false
```

### Environment Variables

Version configuration can be overridden with environment variables:

```bash
API_DEFAULT_VERSION=v1
API_DEPRECATION_WARNING=true
```

## Best Practices

### For API Consumers

1. **Always specify API version explicitly**
2. **Monitor deprecation warnings** in response headers
3. **Handle version errors gracefully**
4. **Subscribe to API change notifications**
5. **Test with multiple versions** during development

### For API Developers

1. **Follow semantic versioning** for breaking changes
2. **Provide comprehensive migration guides**
3. **Maintain backward compatibility** within major versions
4. **Use feature flags** for gradual rollouts
5. **Monitor version usage metrics**

## Testing

### Unit Tests

Version middleware includes comprehensive unit tests:

```bash
go test ./web/middleware -v -run TestVersionMiddleware
```

### Integration Tests

End-to-end version negotiation tests:

```bash
go test ./web/handler -v -run TestVersioning
```

### Frontend Tests

Client-side version handling tests:

```bash
cd frontend && npm test -- versioning.test.ts
```

## Monitoring and Metrics

### Version Usage Metrics

Monitor API version usage to inform deprecation decisions:

- Request count by version
- Client distribution by version
- Error rates by version

### Deprecation Tracking

Track deprecation warnings and client migration progress:

- Deprecated version usage over time
- Client migration completion rates
- Support ticket volume by version

## Support and Migration Assistance

### Getting Help

- **Documentation**: This guide and API reference
- **Examples**: See `examples/` directory
- **Issues**: Create GitHub issues for migration problems
- **Community**: Join our Discord for real-time help

### Migration Support

For complex migrations, we provide:

1. **Migration tools** for automated client updates
2. **Compatibility layers** for gradual migration
3. **Extended support** for enterprise customers
4. **Custom migration assistance** on request

---

*This document is automatically updated when new API versions are released.*