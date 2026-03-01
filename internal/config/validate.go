package config

import (
	"errors"
	"fmt"
)

// Validate checks that all required configuration fields are set.
func Validate(cfg *Config) error {
	var errs []error

	if cfg.GitHub.Token == "" {
		errs = append(errs, fmt.Errorf("github.token is required"))
	}
	if cfg.GitHub.Owner == "" {
		errs = append(errs, fmt.Errorf("github.owner is required"))
	}
	if cfg.GitHub.Repo == "" {
		errs = append(errs, fmt.Errorf("github.repo is required"))
	}
	if cfg.Claude.APIKey == "" {
		errs = append(errs, fmt.Errorf("claude.api_key is required"))
	}
	if cfg.Claude.Model == "" {
		cfg.Claude.Model = "claude-sonnet-4-20250514"
	}
	if cfg.Claude.MaxTokens == 0 {
		cfg.Claude.MaxTokens = 8192
	}
	if cfg.GitHub.PollInterval == 0 {
		cfg.GitHub.PollInterval = 30_000_000_000 // 30s in nanoseconds
	}
	if cfg.State.Backend == "" {
		cfg.State.Backend = "file"
	}
	if cfg.State.Dir == "" {
		cfg.State.Dir = ".agentctl/state"
	}
	if cfg.Agents.Developer.MaxConcurrent == 0 {
		cfg.Agents.Developer.MaxConcurrent = 1
	}
	if cfg.Agents.Developer.WorkspaceDir == "" {
		cfg.Agents.Developer.WorkspaceDir = "./workspaces"
	}

	return errors.Join(errs...)
}
