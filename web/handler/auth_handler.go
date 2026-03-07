package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/gaskaj/OpenAgentFramework/web/auth"
	"github.com/gaskaj/OpenAgentFramework/web/middleware"
	"github.com/gaskaj/OpenAgentFramework/web/store"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	users      *store.PgUserStore
	orgs       *store.PgOrgStore
	jwtMgr     *auth.JWTManager
	bcryptCost int
	logger     *slog.Logger
	providers  map[string]auth.OAuthProvider
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(users *store.PgUserStore, orgs *store.PgOrgStore, jwtMgr *auth.JWTManager, bcryptCost int, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		users:      users,
		orgs:       orgs,
		jwtMgr:     jwtMgr,
		bcryptCost: bcryptCost,
		logger:     logger,
		providers:  make(map[string]auth.OAuthProvider),
	}
}

// RegisterProvider registers an OAuth provider.
func (h *AuthHandler) RegisterProvider(name string, provider auth.OAuthProvider) {
	h.providers[name] = provider
}

// Routes returns the auth router.
func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.handleRegister)
	r.Post("/login", h.handleLogin)
	r.Post("/refresh", h.handleRefresh)
	r.Get("/oauth/{provider}", h.handleOAuthRedirect)
	r.Get("/oauth/{provider}/callback", h.handleOAuthCallback)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(h.jwtMgr))
		r.Get("/me", h.handleMe)
	})
	return r
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	OrgName     string `json:"org_name"`
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		respondError(w, http.StatusBadRequest, "email, password, and display_name are required")
		return
	}
	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	// Check if user exists
	existing, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		h.logger.Error("checking existing user", "error", err)
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if existing != nil {
		respondError(w, http.StatusConflict, "email already registered")
		return
	}

	// Hash password
	hash, err := auth.HashPassword(req.Password, h.bcryptCost)
	if err != nil {
		h.logger.Error("hashing password", "error", err)
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Create user
	user := &store.User{
		Email:        req.Email,
		DisplayName:  req.DisplayName,
		PasswordHash: hash,
	}
	if err := h.users.Create(r.Context(), user); err != nil {
		h.logger.Error("creating user", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Create default organization
	orgName := req.OrgName
	if orgName == "" {
		orgName = req.DisplayName + "'s Org"
	}
	slug := slugify(orgName)

	org := &store.Organization{
		Name:     orgName,
		Slug:     slug,
		Plan:     "free",
		Settings: json.RawMessage("{}"),
	}
	if err := h.orgs.Create(r.Context(), org); err != nil {
		h.logger.Error("creating default org", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create organization")
		return
	}

	// Add user as owner
	if err := h.orgs.AddMember(r.Context(), org.ID, user.ID, "owner"); err != nil {
		h.logger.Error("adding org member", "error", err)
	}

	// Generate tokens
	orgs := []auth.OrgClaim{{ID: org.ID, Slug: org.Slug, Role: "owner"}}
	accessToken, err := h.jwtMgr.CreateAccessToken(user.ID, user.Email, orgs)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create token")
		return
	}
	refreshToken, err := h.jwtMgr.CreateRefreshToken(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create refresh token")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if user == nil || user.PasswordHash == "" {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Get user's orgs
	orgList, err := h.orgs.ListForUser(r.Context(), user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var orgClaims []auth.OrgClaim
	for _, org := range orgList {
		member, _ := h.orgs.GetMembership(r.Context(), org.ID, user.ID)
		role := "member"
		if member != nil {
			role = member.Role
		}
		orgClaims = append(orgClaims, auth.OrgClaim{ID: org.ID, Slug: org.Slug, Role: role})
	}

	accessToken, err := h.jwtMgr.CreateAccessToken(user.ID, user.Email, orgClaims)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create token")
		return
	}
	refreshToken, err := h.jwtMgr.CreateRefreshToken(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create refresh token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, err := h.jwtMgr.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil || user == nil {
		respondError(w, http.StatusUnauthorized, "user not found")
		return
	}

	orgList, _ := h.orgs.ListForUser(r.Context(), user.ID)
	var orgClaims []auth.OrgClaim
	for _, org := range orgList {
		member, _ := h.orgs.GetMembership(r.Context(), org.ID, user.ID)
		role := "member"
		if member != nil {
			role = member.Role
		}
		orgClaims = append(orgClaims, auth.OrgClaim{ID: org.ID, Slug: org.Slug, Role: role})
	}

	accessToken, err := h.jwtMgr.CreateAccessToken(user.ID, user.Email, orgClaims)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"access_token": accessToken,
	})
}

func (h *AuthHandler) handleMe(w http.ResponseWriter, r *http.Request) {
	authCtx := middleware.GetUserFromContext(r.Context())
	if authCtx == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.users.GetByID(r.Context(), authCtx.UserID)
	if err != nil || user == nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	orgList, _ := h.orgs.ListForUser(r.Context(), user.ID)

	respondJSON(w, http.StatusOK, map[string]any{
		"user": user,
		"orgs": orgList,
	})
}

func (h *AuthHandler) handleOAuthRedirect(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	provider, ok := h.providers[providerName]
	if !ok {
		respondError(w, http.StatusBadRequest, "unsupported OAuth provider")
		return
	}

	// Generate state token
	stateBytes := make([]byte, 32)
	rand.Read(stateBytes)
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Store state in a cookie for CSRF validation
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	url := provider.GetAuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	provider, ok := h.providers[providerName]
	if !ok {
		respondError(w, http.StatusBadRequest, "unsupported OAuth provider")
		return
	}

	// Validate state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		respondError(w, http.StatusBadRequest, "invalid OAuth state")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		respondError(w, http.StatusBadRequest, "missing authorization code")
		return
	}

	oauthUser, err := provider.ExchangeCode(r.Context(), code)
	if err != nil {
		h.logger.Error("OAuth exchange failed", "provider", providerName, "error", err)
		respondError(w, http.StatusInternalServerError, "OAuth authentication failed")
		return
	}

	// Find or create user
	link, err := h.users.GetOAuthLink(r.Context(), oauthUser.Provider, oauthUser.ProviderUID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var user *store.User
	if link != nil {
		user, err = h.users.GetByID(r.Context(), link.UserID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal error")
			return
		}
	} else {
		// Check if user exists by email
		user, err = h.users.GetByEmail(r.Context(), oauthUser.Email)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal error")
			return
		}

		if user == nil {
			// Create new user
			user = &store.User{
				Email:         oauthUser.Email,
				DisplayName:   oauthUser.Name,
				AvatarURL:     oauthUser.AvatarURL,
				EmailVerified: true,
			}
			if err := h.users.Create(r.Context(), user); err != nil {
				respondError(w, http.StatusInternalServerError, "failed to create user")
				return
			}

			// Create default org
			org := &store.Organization{
				Name:     oauthUser.Name + "'s Org",
				Slug:     slugify(oauthUser.Name),
				Plan:     "free",
				Settings: json.RawMessage("{}"),
			}
			if err := h.orgs.Create(r.Context(), org); err == nil {
				h.orgs.AddMember(r.Context(), org.ID, user.ID, "owner")
			}
		}

		// Create OAuth link
		oauthLink := &store.UserOAuthLink{
			UserID:       user.ID,
			Provider:     oauthUser.Provider,
			ProviderUID:  oauthUser.ProviderUID,
			ProviderEmail: oauthUser.Email,
			AccessToken:  oauthUser.AccessToken,
			RefreshToken: oauthUser.RefreshToken,
			TokenExpires: oauthUser.TokenExpires,
		}
		h.users.CreateOAuthLink(r.Context(), oauthLink)
	}

	// Generate tokens
	orgList, _ := h.orgs.ListForUser(r.Context(), user.ID)
	var orgClaims []auth.OrgClaim
	for _, org := range orgList {
		member, _ := h.orgs.GetMembership(r.Context(), org.ID, user.ID)
		role := "member"
		if member != nil {
			role = member.Role
		}
		orgClaims = append(orgClaims, auth.OrgClaim{ID: org.ID, Slug: org.Slug, Role: role})
	}

	accessToken, _ := h.jwtMgr.CreateAccessToken(user.ID, user.Email, orgClaims)
	refreshToken, _ := h.jwtMgr.CreateRefreshToken(user.ID)

	// Redirect to frontend with tokens in fragment
	redirectURL := fmt.Sprintf("/#/oauth/callback?access_token=%s&refresh_token=%s", accessToken, refreshToken)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

var slugRegex = regexp.MustCompile(`[^a-z0-9-]`)

func slugify(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = slugRegex.ReplaceAllString(slug, "")
	if len(slug) > 50 {
		slug = slug[:50]
	}
	// Add random suffix to ensure uniqueness
	suffix := make([]byte, 4)
	rand.Read(suffix)
	return slug + "-" + fmt.Sprintf("%x", suffix)
}

// Helper to build org claims for a user
func buildOrgClaims(orgs *store.PgOrgStore, ctx interface{}, userID uuid.UUID) []auth.OrgClaim {
	return nil // implemented inline in handlers above
}
