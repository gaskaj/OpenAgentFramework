package store

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	webconfig "github.com/gaskaj/OpenAgentFramework/web/config"
)

// PostgresStore wraps a pgxpool.Pool and provides access to all sub-stores.
type PostgresStore struct {
	Pool         *pgxpool.Pool
	Logger       *slog.Logger
	Orgs         *PgOrgStore
	Users        *PgUserStore
	Agents       *PgAgentStore
	Events       *PgEventStore
	APIKeys      *PgAPIKeyStore
	Invitations  *PgInvitationStore
	AuditLogs    *PgAuditStore
}

// NewPostgresStore creates a new PostgresStore with a connection pool.
func NewPostgresStore(ctx context.Context, cfg webconfig.DatabaseConfig, logger *slog.Logger) (*PostgresStore, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parsing database config: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.MaxIdleConns)
	poolCfg.MaxConnLifetime = cfg.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	s := &PostgresStore{
		Pool:   pool,
		Logger: logger,
	}
	s.Orgs = &PgOrgStore{pool: pool}
	s.Users = &PgUserStore{pool: pool}
	s.Agents = &PgAgentStore{pool: pool}
	s.Events = &PgEventStore{pool: pool}
	s.APIKeys = &PgAPIKeyStore{pool: pool}
	s.Invitations = &PgInvitationStore{pool: pool}
	s.AuditLogs = &PgAuditStore{pool: pool}

	return s, nil
}

// Close closes the connection pool.
func (s *PostgresStore) Close() {
	s.Pool.Close()
}

// Ping checks the database connection.
func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.Pool.Ping(ctx)
}
