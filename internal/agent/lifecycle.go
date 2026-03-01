package agent

import (
	"context"
	"log/slog"
	"time"
)

// Heartbeat periodically logs that the agent is alive.
// It blocks until the context is cancelled.
func Heartbeat(ctx context.Context, agentType AgentType, interval time.Duration, logger *slog.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("agent shutting down", "type", string(agentType))
			return
		case <-ticker.C:
			logger.Debug("agent heartbeat", "type", string(agentType))
		}
	}
}

// RunWithContext executes a function with a cancellable context derived from the parent.
// Returns when the function completes or the parent context is cancelled.
func RunWithContext(parent context.Context, fn func(ctx context.Context) error) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	return fn(ctx)
}
