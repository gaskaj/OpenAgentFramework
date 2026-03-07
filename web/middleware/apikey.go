package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gaskaj/OpenAgentFramework/web/store"
)

const apiKeyOrgContextKey contextKey = "apikey_org"

// APIKeyOrgContext holds the org resolved from an API key.
type APIKeyOrgContext struct {
	OrgID  interface{} // uuid.UUID - kept as interface to avoid circular
	KeyID  interface{}
	Scopes []string
}

// RequireAPIKey is middleware that validates API keys for agent ingestion endpoints.
func RequireAPIKey(keyStore *store.PgAPIKeyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error":"invalid authorization header"}`, http.StatusUnauthorized)
				return
			}

			rawKey := parts[1]
			if len(rawKey) < 12 {
				http.Error(w, `{"error":"invalid API key format"}`, http.StatusUnauthorized)
				return
			}

			// Extract prefix (first 8 chars after "oaf_" if present)
			key := rawKey
			if strings.HasPrefix(key, "oaf_") {
				key = key[4:]
			}
			if len(key) < 8 {
				http.Error(w, `{"error":"invalid API key format"}`, http.StatusUnauthorized)
				return
			}
			prefix := key[:8]

			// Look up by prefix
			apiKey, err := keyStore.GetByPrefix(r.Context(), prefix)
			if err != nil || apiKey == nil {
				http.Error(w, `{"error":"invalid API key"}`, http.StatusUnauthorized)
				return
			}

			// Verify full key hash
			hash := sha256.Sum256([]byte(rawKey))
			hashStr := hex.EncodeToString(hash[:])
			if hashStr != apiKey.KeyHash {
				http.Error(w, `{"error":"invalid API key"}`, http.StatusUnauthorized)
				return
			}

			// Update last used (fire-and-forget)
			go keyStore.UpdateLastUsed(context.Background(), apiKey.ID)

			// Set org context
			ctx := context.WithValue(r.Context(), orgContextKey, &OrgContext{
				OrgID: apiKey.OrgID,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
