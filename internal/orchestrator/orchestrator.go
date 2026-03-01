package orchestrator

import (
	"context"
	"log/slog"

	"github.com/gaskaj/DeveloperAndQAAgent/internal/agent"
	"golang.org/x/sync/errgroup"
)

// Orchestrator manages the lifecycle of multiple agents.
type Orchestrator struct {
	agents []agent.Agent
	logger *slog.Logger
}

// New creates a new Orchestrator with the given agents.
func New(agents []agent.Agent, logger *slog.Logger) *Orchestrator {
	return &Orchestrator{
		agents: agents,
		logger: logger,
	}
}

// Run starts all agents concurrently and blocks until they all stop or the context is cancelled.
func (o *Orchestrator) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	for _, a := range o.agents {
		a := a // capture for goroutine
		o.logger.Info("starting agent", "type", string(a.Type()))

		g.Go(func() error {
			if err := a.Run(ctx); err != nil {
				o.logger.Error("agent stopped with error",
					"type", string(a.Type()),
					"error", err,
				)
				return err
			}
			return nil
		})
	}

	return g.Wait()
}

// Status returns status reports from all managed agents.
func (o *Orchestrator) Status() []agent.StatusReport {
	reports := make([]agent.StatusReport, len(o.agents))
	for i, a := range o.agents {
		reports[i] = a.Status()
	}
	return reports
}
