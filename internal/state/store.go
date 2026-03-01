package state

import "context"

// Store defines the interface for persisting agent work state.
type Store interface {
	// Save persists an agent's work state.
	Save(ctx context.Context, state *AgentWorkState) error

	// Load retrieves an agent's work state. Returns nil if not found.
	Load(ctx context.Context, agentType string) (*AgentWorkState, error)

	// Delete removes an agent's work state.
	Delete(ctx context.Context, agentType string) error

	// List returns all stored agent work states.
	List(ctx context.Context) ([]*AgentWorkState, error)
}
