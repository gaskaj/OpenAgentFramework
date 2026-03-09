package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgSettingsStore provides access to system-wide key-value settings.
type PgSettingsStore struct {
	pool    *pgxpool.Pool
	monitor *QueryMonitor
}

// Get retrieves a setting value by key. Returns empty string if not found.
func (s *PgSettingsStore) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := s.pool.QueryRow(ctx, `SELECT value FROM system_settings WHERE key = $1`, key).Scan(&value)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("getting setting %q: %w", key, err)
	}
	return value, nil
}

// Set creates or updates a setting.
func (s *PgSettingsStore) Set(ctx context.Context, key, value string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO system_settings (key, value) VALUES ($1, $2)
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`,
		key, value,
	)
	if err != nil {
		return fmt.Errorf("setting %q: %w", key, err)
	}
	return nil
}

// Delete removes a setting.
func (s *PgSettingsStore) Delete(ctx context.Context, key string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM system_settings WHERE key = $1`, key)
	if err != nil {
		return fmt.Errorf("deleting setting %q: %w", key, err)
	}
	return nil
}
