package state

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileStore implements Store using JSON files on disk.
type FileStore struct {
	dir string
	mu  sync.Mutex
}

// NewFileStore creates a new FileStore at the given directory.
func NewFileStore(dir string) (*FileStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating state dir %s: %w", dir, err)
	}
	return &FileStore{dir: dir}, nil
}

func (fs *FileStore) path(agentType string) string {
	return filepath.Join(fs.dir, agentType+".json")
}

// Save persists an agent's work state to a JSON file.
func (fs *FileStore) Save(_ context.Context, state *AgentWorkState) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	if err := os.WriteFile(fs.path(state.AgentType), data, 0o644); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	return nil
}

// Load reads an agent's work state from its JSON file.
func (fs *FileStore) Load(_ context.Context, agentType string) (*AgentWorkState, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := os.ReadFile(fs.path(agentType))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	var state AgentWorkState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshaling state: %w", err)
	}

	return &state, nil
}

// Delete removes an agent's state file.
func (fs *FileStore) Delete(_ context.Context, agentType string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	err := os.Remove(fs.path(agentType))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing state file: %w", err)
	}

	return nil
}

// List returns all stored agent work states.
func (fs *FileStore) List(_ context.Context) ([]*AgentWorkState, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	entries, err := os.ReadDir(fs.dir)
	if err != nil {
		return nil, fmt.Errorf("listing state dir: %w", err)
	}

	var states []*AgentWorkState
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(fs.dir, entry.Name()))
		if err != nil {
			continue
		}

		var s AgentWorkState
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		states = append(states, &s)
	}

	return states, nil
}
