package agent

import (
	"fmt"
	"sync"
)

// AgentFactory creates an Agent given its dependencies.
type AgentFactory func(deps Dependencies) (Agent, error)

// Registry maps agent types to their factory functions.
type Registry struct {
	mu        sync.RWMutex
	factories map[AgentType]AgentFactory
}

// NewRegistry creates a new agent registry.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[AgentType]AgentFactory),
	}
}

// Register adds a factory for the given agent type.
func (r *Registry) Register(t AgentType, factory AgentFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[t] = factory
}

// Create instantiates an agent of the given type.
func (r *Registry) Create(t AgentType, deps Dependencies) (Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, ok := r.factories[t]
	if !ok {
		return nil, fmt.Errorf("unknown agent type: %s", t)
	}
	return factory(deps)
}

// Types returns all registered agent types.
func (r *Registry) Types() []AgentType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]AgentType, 0, len(r.factories))
	for t := range r.factories {
		types = append(types, t)
	}
	return types
}
