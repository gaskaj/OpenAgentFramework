package config

import (
	"testing"
	"time"

	"github.com/gaskaj/OpenAgentFramework/pkg/apitypes"
	"github.com/stretchr/testify/assert"
)

func strPtr(s string) *string    { return &s }
func intPtr(i int) *int          { return &i }
func boolPtr(b bool) *bool       { return &b }
func float64Ptr(f float64) *float64 { return &f }

func TestMergeRemoteConfig_NilDoesNothing(t *testing.T) {
	cfg := &Config{}
	ApplyDefaults(cfg)
	original := *cfg

	MergeRemoteConfig(cfg, nil)
	assert.Equal(t, original, *cfg)
}

func TestMergeRemoteConfig_GitHub(t *testing.T) {
	cfg := &Config{}
	ApplyDefaults(cfg)

	remote := &apitypes.RemoteConfig{
		GitHub: &apitypes.GitHubRemoteConfig{
			Owner:        strPtr("testorg"),
			Repo:         strPtr("testrepo"),
			PollInterval: strPtr("1m"),
			WatchLabels:  []string{"custom:label"},
		},
	}

	MergeRemoteConfig(cfg, remote)

	assert.Equal(t, "testorg", cfg.GitHub.Owner)
	assert.Equal(t, "testrepo", cfg.GitHub.Repo)
	assert.Equal(t, 1*time.Minute, cfg.GitHub.PollInterval)
	assert.Equal(t, []string{"custom:label"}, cfg.GitHub.WatchLabels)
}

func TestMergeRemoteConfig_PartialOverride(t *testing.T) {
	cfg := &Config{}
	ApplyDefaults(cfg)
	cfg.GitHub.Owner = "original"
	cfg.GitHub.Repo = "originalrepo"

	remote := &apitypes.RemoteConfig{
		GitHub: &apitypes.GitHubRemoteConfig{
			Owner: strPtr("newowner"),
			// Repo is nil, should not change
		},
	}

	MergeRemoteConfig(cfg, remote)

	assert.Equal(t, "newowner", cfg.GitHub.Owner)
	assert.Equal(t, "originalrepo", cfg.GitHub.Repo)
}

func TestMergeRemoteConfig_Claude(t *testing.T) {
	cfg := &Config{}
	ApplyDefaults(cfg)

	remote := &apitypes.RemoteConfig{
		Claude: &apitypes.ClaudeRemoteConfig{
			Model:     strPtr("claude-opus-4-20250514"),
			MaxTokens: intPtr(4096),
		},
	}

	MergeRemoteConfig(cfg, remote)

	assert.Equal(t, "claude-opus-4-20250514", cfg.Claude.Model)
	assert.Equal(t, 4096, cfg.Claude.MaxTokens)
}

func TestMergeRemoteConfig_Developer(t *testing.T) {
	cfg := &Config{}
	ApplyDefaults(cfg)

	remote := &apitypes.RemoteConfig{
		Agents: &apitypes.AgentsRemoteConfig{
			Developer: &apitypes.DeveloperRemoteConfig{
				Enabled:                  boolPtr(true),
				MaxConcurrent:            intPtr(3),
				AllowPRMerging:           boolPtr(true),
				AllowAutoIssueProcessing: boolPtr(false),
			},
		},
	}

	MergeRemoteConfig(cfg, remote)

	assert.True(t, cfg.Agents.Developer.Enabled)
	assert.Equal(t, 3, cfg.Agents.Developer.MaxConcurrent)
	assert.True(t, cfg.Agents.Developer.AllowPRMerging)
	assert.False(t, cfg.Agents.Developer.AllowAutoIssueProcessing)
}

func TestMergeRemoteConfig_Creativity(t *testing.T) {
	cfg := &Config{}
	ApplyDefaults(cfg)

	remote := &apitypes.RemoteConfig{
		Creativity: &apitypes.CreativityRemoteConfig{
			Enabled:              boolPtr(true),
			IdleThresholdSeconds: intPtr(300),
			MaxPendingSuggestions: intPtr(10),
		},
	}

	MergeRemoteConfig(cfg, remote)

	assert.True(t, cfg.Creativity.Enabled)
	assert.Equal(t, 300, cfg.Creativity.IdleThresholdSeconds)
	assert.Equal(t, 10, cfg.Creativity.MaxPendingSuggestions)
}

func TestMergeRemoteConfig_Decomposition(t *testing.T) {
	cfg := &Config{}
	ApplyDefaults(cfg)

	remote := &apitypes.RemoteConfig{
		Decomposition: &apitypes.DecompositionRemoteConfig{
			Enabled:            boolPtr(true),
			MaxIterationBudget: intPtr(50),
			MaxSubtasks:        intPtr(8),
		},
	}

	MergeRemoteConfig(cfg, remote)

	assert.True(t, cfg.Decomposition.Enabled)
	assert.Equal(t, 50, cfg.Decomposition.MaxIterationBudget)
	assert.Equal(t, 8, cfg.Decomposition.MaxSubtasks)
}

func TestMergeRemoteConfig_Memory(t *testing.T) {
	cfg := &Config{}
	ApplyDefaults(cfg)

	remote := &apitypes.RemoteConfig{
		Memory: &apitypes.MemoryRemoteConfig{
			Enabled:           boolPtr(false),
			MaxEntries:        intPtr(200),
			ExtractOnComplete: boolPtr(false),
		},
	}

	MergeRemoteConfig(cfg, remote)

	assert.False(t, cfg.Memory.Enabled)
	assert.Equal(t, 200, cfg.Memory.MaxEntries)
	assert.False(t, cfg.Memory.ExtractOnComplete)
}

func TestMergeRemoteConfig_ErrorHandling(t *testing.T) {
	cfg := &Config{}
	ApplyDefaults(cfg)

	remote := &apitypes.RemoteConfig{
		ErrorHandling: &apitypes.ErrorHandlingRemoteConfig{
			Retry: &apitypes.RetryRemoteConfig{
				Enabled: boolPtr(false),
				DefaultPolicy: &apitypes.RetryPolicyRemoteConfig{
					MaxAttempts:   intPtr(5),
					BackoffFactor: float64Ptr(3.0),
				},
			},
		},
	}

	MergeRemoteConfig(cfg, remote)

	assert.False(t, cfg.ErrorHandling.Retry.Enabled)
	assert.Equal(t, 5, cfg.ErrorHandling.Retry.DefaultPolicy.MaxAttempts)
	assert.Equal(t, 3.0, cfg.ErrorHandling.Retry.DefaultPolicy.BackoffFactor)
}

func TestMergeRemoteConfig_Shutdown(t *testing.T) {
	cfg := &Config{}
	ApplyDefaults(cfg)

	remote := &apitypes.RemoteConfig{
		Shutdown: &apitypes.ShutdownRemoteConfig{
			Timeout:           strPtr("60s"),
			CleanupWorkspaces: boolPtr(false),
			ResetClaims:       boolPtr(false),
		},
	}

	MergeRemoteConfig(cfg, remote)

	assert.Equal(t, 60*time.Second, cfg.Shutdown.Timeout)
	assert.False(t, cfg.Shutdown.CleanupWorkspaces)
	assert.False(t, cfg.Shutdown.ResetClaims)
}

func TestMergeRemoteConfig_Logging(t *testing.T) {
	cfg := &Config{}
	ApplyDefaults(cfg)

	remote := &apitypes.RemoteConfig{
		Logging: &apitypes.LoggingRemoteConfig{
			Level: strPtr("debug"),
		},
	}

	MergeRemoteConfig(cfg, remote)

	assert.Equal(t, "debug", cfg.Logging.Level)
}

func TestMergeRemoteConfig_MultipleOverrides(t *testing.T) {
	cfg := &Config{}
	ApplyDefaults(cfg)
	cfg.GitHub.Owner = "base"

	// First override (type-level)
	typeConfig := &apitypes.RemoteConfig{
		GitHub: &apitypes.GitHubRemoteConfig{
			Owner: strPtr("type-level"),
			Repo:  strPtr("type-repo"),
		},
		Creativity: &apitypes.CreativityRemoteConfig{
			Enabled: boolPtr(true),
		},
	}
	MergeRemoteConfig(cfg, typeConfig)

	// Second override (agent-level)
	agentConfig := &apitypes.RemoteConfig{
		GitHub: &apitypes.GitHubRemoteConfig{
			Repo: strPtr("agent-repo"),
		},
	}
	MergeRemoteConfig(cfg, agentConfig)

	assert.Equal(t, "type-level", cfg.GitHub.Owner)
	assert.Equal(t, "agent-repo", cfg.GitHub.Repo) // agent override wins
	assert.True(t, cfg.Creativity.Enabled)          // type-level preserved
}
