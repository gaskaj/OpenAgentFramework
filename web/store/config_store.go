package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AgentTypeConfig represents a type-level configuration template.
type AgentTypeConfig struct {
	ID          uuid.UUID       `json:"id"`
	OrgID       uuid.UUID       `json:"org_id"`
	AgentType   string          `json:"agent_type"`
	Config      json.RawMessage `json:"config"`
	Version     int64           `json:"version"`
	Description string          `json:"description,omitempty"`
	CreatedBy   *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// AgentConfigOverride represents per-agent configuration overrides.
type AgentConfigOverride struct {
	ID          uuid.UUID       `json:"id"`
	OrgID       uuid.UUID       `json:"org_id"`
	AgentID     uuid.UUID       `json:"agent_id"`
	Config      json.RawMessage `json:"config"`
	Version     int64           `json:"version"`
	Description string          `json:"description,omitempty"`
	CreatedBy   *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ConfigAuditEntry represents a configuration change audit entry.
type ConfigAuditEntry struct {
	ID             uuid.UUID       `json:"id"`
	OrgID          uuid.UUID       `json:"org_id"`
	TargetType     string          `json:"target_type"`
	TargetID       uuid.UUID       `json:"target_id"`
	ChangedBy      *uuid.UUID      `json:"changed_by,omitempty"`
	PreviousConfig json.RawMessage `json:"previous_config,omitempty"`
	NewConfig      json.RawMessage `json:"new_config"`
	Version        int64           `json:"version"`
	CreatedAt      time.Time       `json:"created_at"`
}

// PgConfigStore implements configuration persistence with PostgreSQL.
type PgConfigStore struct {
	pool    *pgxpool.Pool
	monitor *QueryMonitor
}

// GetAgentTypeConfig retrieves the type-level config for an agent type within an org.
func (s *PgConfigStore) GetAgentTypeConfig(ctx context.Context, orgID uuid.UUID, agentType string) (*AgentTypeConfig, error) {
	c := &AgentTypeConfig{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, org_id, agent_type, config, version, COALESCE(description, ''), created_by, created_at, updated_at
		 FROM agent_type_configs WHERE org_id = $1 AND agent_type = $2`, orgID, agentType,
	).Scan(&c.ID, &c.OrgID, &c.AgentType, &c.Config, &c.Version, &c.Description, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting agent type config: %w", err)
	}
	return c, nil
}

// UpsertAgentTypeConfig creates or updates a type-level config, increments version, and writes audit.
func (s *PgConfigStore) UpsertAgentTypeConfig(ctx context.Context, orgID uuid.UUID, agentType string, config json.RawMessage, description string, userID uuid.UUID) (*AgentTypeConfig, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get previous config for audit
	var prevConfig json.RawMessage
	err = tx.QueryRow(ctx,
		`SELECT config FROM agent_type_configs WHERE org_id = $1 AND agent_type = $2`,
		orgID, agentType,
	).Scan(&prevConfig)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("getting previous config: %w", err)
	}

	c := &AgentTypeConfig{}
	err = tx.QueryRow(ctx,
		`INSERT INTO agent_type_configs (org_id, agent_type, config, version, description, created_by)
		 VALUES ($1, $2, $3, 1, $4, $5)
		 ON CONFLICT (org_id, agent_type) DO UPDATE SET
		   config = EXCLUDED.config,
		   version = agent_type_configs.version + 1,
		   description = EXCLUDED.description,
		   created_by = EXCLUDED.created_by
		 RETURNING id, org_id, agent_type, config, version, COALESCE(description, ''), created_by, created_at, updated_at`,
		orgID, agentType, config, description, userID,
	).Scan(&c.ID, &c.OrgID, &c.AgentType, &c.Config, &c.Version, &c.Description, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("upserting agent type config: %w", err)
	}

	// Write audit entry
	_, err = tx.Exec(ctx,
		`INSERT INTO config_audit_log (org_id, target_type, target_id, changed_by, previous_config, new_config, version)
		 VALUES ($1, 'agent_type', $2, $3, $4, $5, $6)`,
		orgID, c.ID, userID, prevConfig, config, c.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("writing config audit: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return c, nil
}

// GetAgentOverride retrieves the per-agent config override.
func (s *PgConfigStore) GetAgentOverride(ctx context.Context, agentID uuid.UUID) (*AgentConfigOverride, error) {
	c := &AgentConfigOverride{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, org_id, agent_id, config, version, COALESCE(description, ''), created_by, created_at, updated_at
		 FROM agent_config_overrides WHERE agent_id = $1`, agentID,
	).Scan(&c.ID, &c.OrgID, &c.AgentID, &c.Config, &c.Version, &c.Description, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting agent config override: %w", err)
	}
	return c, nil
}

// UpsertAgentOverride creates or updates a per-agent override, increments version, and writes audit.
func (s *PgConfigStore) UpsertAgentOverride(ctx context.Context, agentID uuid.UUID, orgID uuid.UUID, config json.RawMessage, description string, userID uuid.UUID) (*AgentConfigOverride, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get previous config for audit
	var prevConfig json.RawMessage
	err = tx.QueryRow(ctx,
		`SELECT config FROM agent_config_overrides WHERE agent_id = $1`,
		agentID,
	).Scan(&prevConfig)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("getting previous override: %w", err)
	}

	c := &AgentConfigOverride{}
	err = tx.QueryRow(ctx,
		`INSERT INTO agent_config_overrides (org_id, agent_id, config, version, description, created_by)
		 VALUES ($1, $2, $3, 1, $4, $5)
		 ON CONFLICT (agent_id) DO UPDATE SET
		   config = EXCLUDED.config,
		   version = agent_config_overrides.version + 1,
		   description = EXCLUDED.description,
		   created_by = EXCLUDED.created_by
		 RETURNING id, org_id, agent_id, config, version, COALESCE(description, ''), created_by, created_at, updated_at`,
		orgID, agentID, config, description, userID,
	).Scan(&c.ID, &c.OrgID, &c.AgentID, &c.Config, &c.Version, &c.Description, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("upserting agent config override: %w", err)
	}

	// Write audit entry
	_, err = tx.Exec(ctx,
		`INSERT INTO config_audit_log (org_id, target_type, target_id, changed_by, previous_config, new_config, version)
		 VALUES ($1, 'agent', $2, $3, $4, $5, $6)`,
		orgID, c.ID, userID, prevConfig, config, c.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("writing config audit: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return c, nil
}

// DeleteAgentOverride removes a per-agent config override.
func (s *PgConfigStore) DeleteAgentOverride(ctx context.Context, agentID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM agent_config_overrides WHERE agent_id = $1`, agentID)
	if err != nil {
		return fmt.Errorf("deleting agent config override: %w", err)
	}
	return nil
}

// GetMergedConfig returns the merged configuration for a specific agent.
// It merges the agent type config with any per-agent overrides using PostgreSQL jsonb concatenation.
func (s *PgConfigStore) GetMergedConfig(ctx context.Context, orgID uuid.UUID, agentID uuid.UUID, agentType string) (json.RawMessage, int64, error) {
	var merged json.RawMessage
	var version int64
	err := s.pool.QueryRow(ctx,
		`SELECT
			COALESCE(atc.config, '{}') || COALESCE(aco.config, '{}') AS merged_config,
			GREATEST(COALESCE(atc.version, 0), COALESCE(aco.version, 0)) AS version
		 FROM (SELECT $1::uuid AS org_id, $2::uuid AS agent_id, $3::text AS agent_type) params
		 LEFT JOIN agent_type_configs atc ON atc.org_id = params.org_id AND atc.agent_type = params.agent_type
		 LEFT JOIN agent_config_overrides aco ON aco.agent_id = params.agent_id`,
		orgID, agentID, agentType,
	).Scan(&merged, &version)
	if err != nil {
		return nil, 0, fmt.Errorf("getting merged config: %w", err)
	}
	return merged, version, nil
}

// ListConfigAudit returns config audit entries for an org.
func (s *PgConfigStore) ListConfigAudit(ctx context.Context, orgID uuid.UUID, opts ListOpts) ([]ConfigAuditEntry, int, error) {
	var total int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM config_audit_log WHERE org_id = $1`, orgID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting config audit entries: %w", err)
	}

	rows, err := s.pool.Query(ctx,
		`SELECT id, org_id, target_type, target_id, changed_by, previous_config, new_config, version, created_at
		 FROM config_audit_log WHERE org_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`, orgID, opts.Limit, opts.Offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing config audit: %w", err)
	}
	defer rows.Close()

	var entries []ConfigAuditEntry
	for rows.Next() {
		var e ConfigAuditEntry
		if err := rows.Scan(&e.ID, &e.OrgID, &e.TargetType, &e.TargetID, &e.ChangedBy, &e.PreviousConfig, &e.NewConfig, &e.Version, &e.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scanning config audit entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, total, nil
}

// ListAgentTypeConfigs returns all type-level configs for an org.
func (s *PgConfigStore) ListAgentTypeConfigs(ctx context.Context, orgID uuid.UUID) ([]AgentTypeConfig, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, org_id, agent_type, config, version, COALESCE(description, ''), created_by, created_at, updated_at
		 FROM agent_type_configs WHERE org_id = $1
		 ORDER BY agent_type`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing agent type configs: %w", err)
	}
	defer rows.Close()

	var configs []AgentTypeConfig
	for rows.Next() {
		var c AgentTypeConfig
		if err := rows.Scan(&c.ID, &c.OrgID, &c.AgentType, &c.Config, &c.Version, &c.Description, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning agent type config: %w", err)
		}
		configs = append(configs, c)
	}
	return configs, nil
}
