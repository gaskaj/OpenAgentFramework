package middleware

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"
)

// APIVersion represents a supported API version.
type APIVersion struct {
	Version      string    `yaml:"version"`
	IsDefault    bool      `yaml:"is_default"`
	IsDeprecated bool      `yaml:"is_deprecated"`
	DeprecatedAt *time.Time `yaml:"deprecated_at,omitempty"`
	SunsetAt     *time.Time `yaml:"sunset_at,omitempty"`
}

// VersionConfig holds the versioning configuration.
type VersionConfig struct {
	DefaultVersion     string       `yaml:"default_version"`
	SupportedVersions  []APIVersion `yaml:"supported_versions"`
	DeprecationWarning bool         `yaml:"deprecation_warning"`
}

// VersionMiddleware handles API version negotiation and validation.
type VersionMiddleware struct {
	config            VersionConfig
	supportedVersions map[string]APIVersion
	defaultVersion    string
}

// NewVersionMiddleware creates a new version middleware with the given configuration.
func NewVersionMiddleware(config VersionConfig) *VersionMiddleware {
	vm := &VersionMiddleware{
		config:            config,
		supportedVersions: make(map[string]APIVersion),
		defaultVersion:    config.DefaultVersion,
	}

	// Build lookup map for supported versions
	for _, version := range config.SupportedVersions {
		vm.supportedVersions[version.Version] = version
		if version.IsDefault {
			vm.defaultVersion = version.Version
		}
	}

	// If no default version is set, use the first supported version
	if vm.defaultVersion == "" && len(config.SupportedVersions) > 0 {
		vm.defaultVersion = config.SupportedVersions[0].Version
	}

	return vm
}

// versionContextKey is used for storing version info in request context.
const versionContextKey = "api_version"

// VersionContext holds version information for the current request.
type VersionContext struct {
	Version        string
	IsDeprecated   bool
	DeprecatedAt   *time.Time
	SunsetAt       *time.Time
	RequestedBy    string // How the version was specified (header, query, path, default)
}

// GetVersionFromContext retrieves the API version context from the request context.
func GetVersionFromContext(ctx context.Context) *VersionContext {
	if val := ctx.Value(versionContextKey); val != nil {
		if versionCtx, ok := val.(*VersionContext); ok {
			return versionCtx
		}
	}
	return nil
}

// VersionHandler returns a middleware that handles API version negotiation.
func (vm *VersionMiddleware) VersionHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			version, requestedBy := vm.extractVersion(r)
			
			// Validate version
			apiVersion, ok := vm.supportedVersions[version]
			if !ok {
				vm.respondUnsupportedVersion(w, version)
				return
			}

			// Check if version is sunset
			if apiVersion.SunsetAt != nil && time.Now().After(*apiVersion.SunsetAt) {
				vm.respondSunsetVersion(w, version, *apiVersion.SunsetAt)
				return
			}

			// Create version context
			versionCtx := &VersionContext{
				Version:      version,
				IsDeprecated: apiVersion.IsDeprecated,
				DeprecatedAt: apiVersion.DeprecatedAt,
				SunsetAt:     apiVersion.SunsetAt,
				RequestedBy:  requestedBy,
			}

			// Add version info to response headers
			w.Header().Set("API-Version", version)
			if apiVersion.IsDeprecated && vm.config.DeprecationWarning {
				vm.addDeprecationWarning(w, apiVersion)
			}

			// Add version context to request
			ctx := context.WithValue(r.Context(), versionContextKey, versionCtx)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// extractVersion extracts the API version from the request using various strategies.
func (vm *VersionMiddleware) extractVersion(r *http.Request) (version, requestedBy string) {
	// 1. Try Accept header: Accept: application/vnd.openagent.v2+json
	if accept := r.Header.Get("Accept"); accept != "" {
		if v := vm.ParseAcceptHeader(accept); v != "" {
			return v, "header"
		}
	}

	// 2. Try custom API-Version header
	if v := r.Header.Get("API-Version"); v != "" {
		return v, "header"
	}

	// 3. Try query parameter: ?api-version=v2
	if v := r.URL.Query().Get("api-version"); v != "" {
		return v, "query"
	}

	// 4. Try URL path: /api/v2/...
	if v := vm.ParsePathVersion(r.URL.Path); v != "" {
		return v, "path"
	}

	// 5. Default to configured default version
	return vm.defaultVersion, "default"
}

var acceptHeaderRegex = regexp.MustCompile(`application/vnd\.openagent\.(v\d+)\+json`)

// ParseAcceptHeader extracts version from Accept header.
func (vm *VersionMiddleware) ParseAcceptHeader(accept string) string {
	matches := acceptHeaderRegex.FindStringSubmatch(accept)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

var pathVersionRegex = regexp.MustCompile(`^/api/(v\d+)/`)

// ParsePathVersion extracts version from URL path.
func (vm *VersionMiddleware) ParsePathVersion(path string) string {
	matches := pathVersionRegex.FindStringSubmatch(path)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// addDeprecationWarning adds deprecation warning headers to the response.
func (vm *VersionMiddleware) addDeprecationWarning(w http.ResponseWriter, version APIVersion) {
	w.Header().Set("Deprecation", "true")
	w.Header().Set("Warning", fmt.Sprintf(`299 - "API version %s is deprecated"`, version.Version))
	
	if version.DeprecatedAt != nil {
		w.Header().Set("Deprecation-Date", version.DeprecatedAt.Format(time.RFC3339))
	}
	
	if version.SunsetAt != nil {
		w.Header().Set("Sunset", version.SunsetAt.Format(time.RFC1123))
		w.Header().Set("Link", `</docs/controlplane/api-versioning.md>; rel="deprecation"; type="text/html"`)
	}
}

// respondUnsupportedVersion responds with a 400 error for unsupported versions.
func (vm *VersionMiddleware) respondUnsupportedVersion(w http.ResponseWriter, version string) {
	supportedVersions := make([]string, 0, len(vm.supportedVersions))
	for v := range vm.supportedVersions {
		supportedVersions = append(supportedVersions, v)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, `{
		"error": "unsupported_api_version",
		"message": "API version %q is not supported",
		"supported_versions": %v,
		"default_version": "%s"
	}`, version, supportedVersions, vm.defaultVersion)
}

// respondSunsetVersion responds with a 410 error for sunset versions.
func (vm *VersionMiddleware) respondSunsetVersion(w http.ResponseWriter, version string, sunsetAt time.Time) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Sunset", sunsetAt.Format(time.RFC1123))
	w.Header().Set("Link", `</docs/controlplane/api-versioning.md>; rel="successor-version"; type="text/html"`)
	w.WriteHeader(http.StatusGone)
	fmt.Fprintf(w, `{
		"error": "api_version_sunset",
		"message": "API version %q was sunset on %s",
		"sunset_date": "%s",
		"migration_guide": "/docs/controlplane/api-versioning.md"
	}`, version, sunsetAt.Format("2006-01-02"), sunsetAt.Format(time.RFC3339))
}

// DefaultVersionConfig returns a sensible default version configuration.
func DefaultVersionConfig() VersionConfig {
	return VersionConfig{
		DefaultVersion:     "v1",
		DeprecationWarning: true,
		SupportedVersions: []APIVersion{
			{
				Version:   "v1",
				IsDefault: true,
			},
		},
	}
}

// VersionHandlerForConfig is a convenience function that creates a version middleware
// from a config and returns the handler function.
func VersionHandlerForConfig(config VersionConfig) func(http.Handler) http.Handler {
	return NewVersionMiddleware(config).VersionHandler()
}