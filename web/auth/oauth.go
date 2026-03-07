package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	webconfig "github.com/gaskaj/OpenAgentFramework/web/config"
)

// OAuthUser represents a user profile from an OAuth provider.
type OAuthUser struct {
	Provider     string
	ProviderUID  string
	Email        string
	Name         string
	AvatarURL    string
	AccessToken  string
	RefreshToken string
	TokenExpires *time.Time
}

// OAuthProvider handles OAuth authentication flows.
type OAuthProvider interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*OAuthUser, error)
}

// GoogleProvider implements OAuth for Google.
type GoogleProvider struct {
	config *oauth2.Config
}

// NewGoogleProvider creates a new Google OAuth provider.
func NewGoogleProvider(cfg webconfig.OAuthProviderConfig) *GoogleProvider {
	return &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (p *GoogleProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *GoogleProvider) ExchangeCode(ctx context.Context, code string) (*OAuthUser, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging Google code: %w", err)
	}

	client := p.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("fetching Google user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading Google user info: %w", err)
	}

	var info struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parsing Google user info: %w", err)
	}

	var expires *time.Time
	if !token.Expiry.IsZero() {
		expires = &token.Expiry
	}

	return &OAuthUser{
		Provider:     "google",
		ProviderUID:  info.ID,
		Email:        info.Email,
		Name:         info.Name,
		AvatarURL:    info.Picture,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenExpires: expires,
	}, nil
}

// AzureProvider implements OAuth for Azure AD.
type AzureProvider struct {
	config *oauth2.Config
}

// NewAzureProvider creates a new Azure AD OAuth provider.
func NewAzureProvider(cfg webconfig.OAuthProviderConfig) *AzureProvider {
	return &AzureProvider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"openid", "email", "profile", "User.Read"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", cfg.TenantID),
				TokenURL: fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", cfg.TenantID),
			},
		},
	}
}

func (p *AzureProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *AzureProvider) ExchangeCode(ctx context.Context, code string) (*OAuthUser, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging Azure code: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://graph.microsoft.com/v1.0/me", nil)
	if err != nil {
		return nil, fmt.Errorf("creating Azure graph request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching Azure user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading Azure user info: %w", err)
	}

	var info struct {
		ID          string `json:"id"`
		Mail        string `json:"mail"`
		DisplayName string `json:"displayName"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parsing Azure user info: %w", err)
	}

	var expires *time.Time
	if !token.Expiry.IsZero() {
		expires = &token.Expiry
	}

	return &OAuthUser{
		Provider:     "azure",
		ProviderUID:  info.ID,
		Email:        info.Mail,
		Name:         info.DisplayName,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenExpires: expires,
	}, nil
}
