package store

import (
	"encoding/json"
	"net/netip"
	"time"

	"github.com/google/uuid"
)

// Organization represents a tenant in the system.
type Organization struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	Plan      string          `json:"plan"`
	Settings  json.RawMessage `json:"settings"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// User represents a registered user.
type User struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"display_name"`
	PasswordHash  string    `json:"-"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserOAuthLink represents an OAuth provider link for a user.
type UserOAuthLink struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	Provider      string     `json:"provider"`
	ProviderUID   string     `json:"provider_uid"`
	ProviderEmail string     `json:"provider_email,omitempty"`
	AccessToken   string     `json:"-"`
	RefreshToken  string     `json:"-"`
	TokenExpires  *time.Time `json:"token_expires,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// OrgMember represents a user's membership in an organization.
type OrgMember struct {
	ID       uuid.UUID `json:"id"`
	OrgID    uuid.UUID `json:"org_id"`
	UserID   uuid.UUID `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	// Joined fields from user
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// Agent represents a registered agent.
type Agent struct {
	ID             uuid.UUID       `json:"id"`
	OrgID          uuid.UUID       `json:"org_id"`
	Name           string          `json:"name"`
	AgentType      string          `json:"agent_type"`
	Description    string          `json:"description,omitempty"`
	GitHubOwner    string          `json:"github_owner,omitempty"`
	GitHubRepo     string          `json:"github_repo,omitempty"`
	ConfigSnapshot json.RawMessage `json:"config_snapshot,omitempty"`
	LastHeartbeat  *time.Time      `json:"last_heartbeat,omitempty"`
	Status         string          `json:"status"`
	Version        string          `json:"version,omitempty"`
	Hostname       string          `json:"hostname,omitempty"`
	Tags           []string        `json:"tags"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// AgentEvent represents a recorded agent event.
type AgentEvent struct {
	ID            uuid.UUID       `json:"id"`
	OrgID         uuid.UUID       `json:"org_id"`
	AgentID       uuid.UUID       `json:"agent_id"`
	EventType     string          `json:"event_type"`
	Severity      string          `json:"severity"`
	Payload       json.RawMessage `json:"payload"`
	IssueNumber   *int            `json:"issue_number,omitempty"`
	PRNumber      *int            `json:"pr_number,omitempty"`
	WorkflowState string          `json:"workflow_state,omitempty"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	// Joined fields
	AgentName string `json:"agent_name,omitempty"`
}

// APIKey represents an API key for agent authentication.
type APIKey struct {
	ID        uuid.UUID  `json:"id"`
	OrgID     uuid.UUID  `json:"org_id"`
	CreatedBy uuid.UUID  `json:"created_by"`
	Name      string     `json:"name"`
	KeyHash   string     `json:"-"`
	KeyPrefix string     `json:"key_prefix"`
	Scopes    []string   `json:"scopes"`
	AgentType string     `json:"agent_type"`
	AgentName string     `json:"agent_name"`
	LastUsed  *time.Time `json:"last_used,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Revoked   bool       `json:"revoked"`
	CreatedAt time.Time  `json:"created_at"`
}

// Invitation represents a pending organization invitation.
type Invitation struct {
	ID        uuid.UUID `json:"id"`
	OrgID     uuid.UUID `json:"org_id"`
	InvitedBy uuid.UUID `json:"invited_by"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Token     string    `json:"-"`
	Accepted  bool      `json:"accepted"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// AuditLog represents an audit log entry.
type AuditLog struct {
	ID           uuid.UUID       `json:"id"`
	OrgID        uuid.UUID       `json:"org_id"`
	UserID       *uuid.UUID      `json:"user_id,omitempty"`
	Action       string          `json:"action"`
	ResourceType string          `json:"resource_type"`
	ResourceID   *uuid.UUID      `json:"resource_id,omitempty"`
	Details      json.RawMessage `json:"details"`
	IPAddress    *netip.Addr     `json:"ip_address,omitempty"`
	UserAgent    string          `json:"user_agent,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	// Joined fields
	UserEmail string `json:"user_email,omitempty"`
}

// ListOpts holds pagination and sorting options.
type ListOpts struct {
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	OrderBy  string `json:"order_by"`
	OrderDir string `json:"order_dir"`
}

// DefaultListOpts returns list options with sensible defaults.
func DefaultListOpts() ListOpts {
	return ListOpts{
		Limit:    20,
		Offset:   0,
		OrderBy:  "created_at",
		OrderDir: "DESC",
	}
}

// EventFilter holds filtering options for event queries.
type EventFilter struct {
	OrgID     uuid.UUID  `json:"org_id"`
	AgentID   *uuid.UUID `json:"agent_id,omitempty"`
	EventType string     `json:"event_type,omitempty"`
	Severity  string     `json:"severity,omitempty"`
	Since     *time.Time `json:"since,omitempty"`
	Until     *time.Time `json:"until,omitempty"`
	Search    string     `json:"search,omitempty"`
	ListOpts
}

// AuditFilter holds filtering options for audit log queries.
type AuditFilter struct {
	OrgID        uuid.UUID  `json:"org_id"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	Action       string     `json:"action,omitempty"`
	ResourceType string     `json:"resource_type,omitempty"`
	Since        *time.Time `json:"since,omitempty"`
	Until        *time.Time `json:"until,omitempty"`
	ListOpts
}
