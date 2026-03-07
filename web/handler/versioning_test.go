package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gaskaj/OpenAgentFramework/web/middleware"
)

func TestVersionMiddleware_VersionNegotiation(t *testing.T) {
	// Set up version configuration
	deprecatedAt := time.Now().Add(-30 * 24 * time.Hour) // 30 days ago
	sunsetAt := time.Now().Add(30 * 24 * time.Hour)      // 30 days from now
	
	config := middleware.VersionConfig{
		DefaultVersion:     "v1",
		DeprecationWarning: true,
		SupportedVersions: []middleware.APIVersion{
			{
				Version:   "v1",
				IsDefault: true,
			},
			{
				Version:      "v0",
				IsDeprecated: true,
				DeprecatedAt: &deprecatedAt,
				SunsetAt:     &sunsetAt,
			},
		},
	}

	vm := middleware.NewVersionMiddleware(config)
	
	// Test handler that extracts version from context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		versionCtx := middleware.GetVersionFromContext(r.Context())
		require.NotNil(t, versionCtx)
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"version":"` + versionCtx.Version + `","deprecated":` + 
			fmt.Sprintf("%t", versionCtx.IsDeprecated) + `}`))
	})

	handler := vm.VersionHandler()(testHandler)

	tests := []struct {
		name               string
		headers            map[string]string
		query              string
		expectedVersion    string
		expectedStatus     int
		expectedDeprecated bool
		expectWarning      bool
	}{
		{
			name:            "default version when no version specified",
			expectedVersion: "v1",
			expectedStatus:  http.StatusOK,
		},
		{
			name: "version from Accept header",
			headers: map[string]string{
				"Accept": "application/vnd.openagent.v1+json",
			},
			expectedVersion: "v1",
			expectedStatus:  http.StatusOK,
		},
		{
			name: "version from API-Version header",
			headers: map[string]string{
				"API-Version": "v1",
			},
			expectedVersion: "v1",
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "version from query parameter",
			query:           "api-version=v1",
			expectedVersion: "v1",
			expectedStatus:  http.StatusOK,
		},
		{
			name: "deprecated version with warnings",
			headers: map[string]string{
				"Accept": "application/vnd.openagent.v0+json",
			},
			expectedVersion:    "v0",
			expectedStatus:     http.StatusOK,
			expectedDeprecated: true,
			expectWarning:      true,
		},
		{
			name: "unsupported version",
			headers: map[string]string{
				"Accept": "application/vnd.openagent.v999+json",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.query != "" {
				req.URL.RawQuery = tt.query
			}
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedVersion, w.Header().Get("API-Version"))
				
				if tt.expectWarning {
					assert.Equal(t, "true", w.Header().Get("Deprecation"))
					assert.Contains(t, w.Header().Get("Warning"), "deprecated")
					assert.NotEmpty(t, w.Header().Get("Deprecation-Date"))
					assert.NotEmpty(t, w.Header().Get("Sunset"))
				}
			}
		})
	}
}

func TestVersionMiddleware_SunsetVersion(t *testing.T) {
	// Set up version configuration with sunset version
	sunsetAt := time.Now().Add(-24 * time.Hour) // 1 day ago (already sunset)
	
	config := middleware.VersionConfig{
		DefaultVersion:     "v1",
		DeprecationWarning: true,
		SupportedVersions: []middleware.APIVersion{
			{
				Version:   "v1",
				IsDefault: true,
			},
			{
				Version:      "v0",
				IsDeprecated: true,
				SunsetAt:     &sunsetAt,
			},
		},
	}

	vm := middleware.NewVersionMiddleware(config)
	handler := vm.VersionHandler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept", "application/vnd.openagent.v0+json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusGone, w.Code)
	assert.NotEmpty(t, w.Header().Get("Sunset"))
	assert.Contains(t, w.Header().Get("Link"), "rel=\"successor-version\"")
}

func TestVersionMiddleware_VersionParsing(t *testing.T) {
	config := middleware.DefaultVersionConfig()
	vm := middleware.NewVersionMiddleware(config)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid accept header",
			input:    "application/vnd.openagent.v2+json",
			expected: "v2",
		},
		{
			name:     "accept header with multiple types",
			input:    "application/json, application/vnd.openagent.v1+json, text/html",
			expected: "v1",
		},
		{
			name:     "invalid accept header",
			input:    "application/json",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.ParseAcceptHeader(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionMiddleware_PathVersionParsing(t *testing.T) {
	config := middleware.DefaultVersionConfig()
	vm := middleware.NewVersionMiddleware(config)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "version in path",
			path:     "/api/v2/agents",
			expected: "v2",
		},
		{
			name:     "no version in path",
			path:     "/api/agents",
			expected: "",
		},
		{
			name:     "version-like but not matching pattern",
			path:     "/api/version2/agents",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.ParsePathVersion(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionMiddleware_DefaultConfiguration(t *testing.T) {
	config := middleware.DefaultVersionConfig()
	
	assert.Equal(t, "v1", config.DefaultVersion)
	assert.True(t, config.DeprecationWarning)
	assert.Len(t, config.SupportedVersions, 1)
	assert.Equal(t, "v1", config.SupportedVersions[0].Version)
	assert.True(t, config.SupportedVersions[0].IsDefault)
	assert.False(t, config.SupportedVersions[0].IsDeprecated)
}

func TestVersionMiddleware_ContextRetrieval(t *testing.T) {
	config := middleware.DefaultVersionConfig()
	vm := middleware.NewVersionMiddleware(config)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		versionCtx := middleware.GetVersionFromContext(r.Context())
		require.NotNil(t, versionCtx)
		
		assert.Equal(t, "v1", versionCtx.Version)
		assert.Equal(t, "default", versionCtx.RequestedBy)
		assert.False(t, versionCtx.IsDeprecated)
		
		w.WriteHeader(http.StatusOK)
	})

	handler := vm.VersionHandler()(testHandler)
	
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVersionMiddleware_MultipleVersionStrategies(t *testing.T) {
	config := middleware.VersionConfig{
		DefaultVersion: "v1",
		SupportedVersions: []middleware.APIVersion{
			{Version: "v1", IsDefault: true},
			{Version: "v2"},
		},
	}
	vm := middleware.NewVersionMiddleware(config)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		versionCtx := middleware.GetVersionFromContext(r.Context())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"version":"` + versionCtx.Version + `","requested_by":"` + versionCtx.RequestedBy + `"}`))
	})

	handler := vm.VersionHandler()(testHandler)

	// Test priority: Accept header should win over query parameter
	req := httptest.NewRequest(http.MethodGet, "/test?api-version=v1", nil)
	req.Header.Set("Accept", "application/vnd.openagent.v2+json")
	
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "v2", w.Header().Get("API-Version"))
	assert.Contains(t, w.Body.String(), `"version":"v2"`)
	assert.Contains(t, w.Body.String(), `"requested_by":"header"`)
}