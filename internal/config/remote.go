package config

import (
	"time"

	"github.com/gaskaj/OpenAgentFramework/pkg/apitypes"
)

// MergeRemoteConfig applies a RemoteConfig on top of the current Config.
// Only non-nil fields in the RemoteConfig override the local config.
func MergeRemoteConfig(cfg *Config, remote *apitypes.RemoteConfig) {
	if remote == nil {
		return
	}

	if remote.GitHub != nil {
		mergeGitHub(&cfg.GitHub, remote.GitHub)
	}
	if remote.Claude != nil {
		mergeClaude(&cfg.Claude, remote.Claude)
	}
	if remote.Agents != nil {
		mergeAgents(&cfg.Agents, remote.Agents)
	}
	if remote.Creativity != nil {
		mergeCreativity(&cfg.Creativity, remote.Creativity)
	}
	if remote.Decomposition != nil {
		mergeDecomposition(&cfg.Decomposition, remote.Decomposition)
	}
	if remote.Memory != nil {
		mergeMemory(&cfg.Memory, remote.Memory)
	}
	if remote.ErrorHandling != nil {
		mergeErrorHandling(&cfg.ErrorHandling, remote.ErrorHandling)
	}
	if remote.Logging != nil {
		mergeLogging(&cfg.Logging, remote.Logging)
	}
	if remote.Shutdown != nil {
		mergeShutdown(&cfg.Shutdown, remote.Shutdown)
	}
}

func mergeGitHub(cfg *GitHubConfig, remote *apitypes.GitHubRemoteConfig) {
	if remote.Token != nil {
		cfg.Token = *remote.Token
	}
	if remote.Owner != nil {
		cfg.Owner = *remote.Owner
	}
	if remote.Repo != nil {
		cfg.Repo = *remote.Repo
	}
	if remote.PollInterval != nil {
		if d, err := time.ParseDuration(*remote.PollInterval); err == nil {
			cfg.PollInterval = d
		}
	}
	if remote.WatchLabels != nil {
		cfg.WatchLabels = remote.WatchLabels
	}
}

func mergeClaude(cfg *ClaudeConfig, remote *apitypes.ClaudeRemoteConfig) {
	if remote.APIKey != nil {
		cfg.APIKey = *remote.APIKey
	}
	if remote.Model != nil {
		cfg.Model = *remote.Model
	}
	if remote.MaxTokens != nil {
		cfg.MaxTokens = *remote.MaxTokens
	}
}

func mergeAgents(cfg *AgentsConfig, remote *apitypes.AgentsRemoteConfig) {
	if remote.Developer != nil {
		mergeDeveloper(&cfg.Developer, remote.Developer)
	}
}

func mergeDeveloper(cfg *DeveloperAgentConfig, remote *apitypes.DeveloperRemoteConfig) {
	if remote.Enabled != nil {
		cfg.Enabled = *remote.Enabled
	}
	if remote.MaxConcurrent != nil {
		cfg.MaxConcurrent = *remote.MaxConcurrent
	}
	if remote.WorkspaceDir != nil {
		cfg.WorkspaceDir = *remote.WorkspaceDir
	}
	if remote.AllowPRMerging != nil {
		cfg.AllowPRMerging = *remote.AllowPRMerging
	}
	if remote.AllowAutoIssueProcessing != nil {
		cfg.AllowAutoIssueProcessing = *remote.AllowAutoIssueProcessing
	}
}

func mergeCreativity(cfg *CreativityConfig, remote *apitypes.CreativityRemoteConfig) {
	if remote.Enabled != nil {
		cfg.Enabled = *remote.Enabled
	}
	if remote.IdleThresholdSeconds != nil {
		cfg.IdleThresholdSeconds = *remote.IdleThresholdSeconds
	}
	if remote.SuggestionCooldownSeconds != nil {
		cfg.SuggestionCooldownSeconds = *remote.SuggestionCooldownSeconds
	}
	if remote.MaxPendingSuggestions != nil {
		cfg.MaxPendingSuggestions = *remote.MaxPendingSuggestions
	}
	if remote.MaxRejectionHistory != nil {
		cfg.MaxRejectionHistory = *remote.MaxRejectionHistory
	}
}

func mergeDecomposition(cfg *DecompositionConfig, remote *apitypes.DecompositionRemoteConfig) {
	if remote.Enabled != nil {
		cfg.Enabled = *remote.Enabled
	}
	if remote.MaxIterationBudget != nil {
		cfg.MaxIterationBudget = *remote.MaxIterationBudget
	}
	if remote.MaxSubtasks != nil {
		cfg.MaxSubtasks = *remote.MaxSubtasks
	}
}

func mergeMemory(cfg *MemoryConfig, remote *apitypes.MemoryRemoteConfig) {
	if remote.Enabled != nil {
		cfg.Enabled = *remote.Enabled
	}
	if remote.MaxEntries != nil {
		cfg.MaxEntries = *remote.MaxEntries
	}
	if remote.MaxPromptSize != nil {
		cfg.MaxPromptSize = *remote.MaxPromptSize
	}
	if remote.ExtractOnComplete != nil {
		cfg.ExtractOnComplete = *remote.ExtractOnComplete
	}
}

func mergeErrorHandling(cfg *ErrorHandlingConfig, remote *apitypes.ErrorHandlingRemoteConfig) {
	if remote.Retry != nil {
		mergeRetry(&cfg.Retry, remote.Retry)
	}
	if remote.CircuitBreaker != nil {
		mergeCircuitBreaker(&cfg.CircuitBreaker, remote.CircuitBreaker)
	}
}

func mergeRetry(cfg *RetryConfig, remote *apitypes.RetryRemoteConfig) {
	if remote.Enabled != nil {
		cfg.Enabled = *remote.Enabled
	}
	if remote.DefaultPolicy != nil {
		mergeRetryPolicy(&cfg.DefaultPolicy, remote.DefaultPolicy)
	}
	if remote.Policies != nil {
		if cfg.Policies == nil {
			cfg.Policies = make(map[string]RetryPolicyConfig)
		}
		for name, rp := range remote.Policies {
			p := cfg.Policies[name]
			mergeRetryPolicy(&p, &rp)
			cfg.Policies[name] = p
		}
	}
}

func mergeRetryPolicy(cfg *RetryPolicyConfig, remote *apitypes.RetryPolicyRemoteConfig) {
	if remote.MaxAttempts != nil {
		cfg.MaxAttempts = *remote.MaxAttempts
	}
	if remote.BaseDelay != nil {
		if d, err := time.ParseDuration(*remote.BaseDelay); err == nil {
			cfg.BaseDelay = d
		}
	}
	if remote.MaxDelay != nil {
		if d, err := time.ParseDuration(*remote.MaxDelay); err == nil {
			cfg.MaxDelay = d
		}
	}
	if remote.BackoffFactor != nil {
		cfg.BackoffFactor = *remote.BackoffFactor
	}
	if remote.JitterFactor != nil {
		cfg.JitterFactor = *remote.JitterFactor
	}
	if remote.RetryableErrors != nil {
		cfg.RetryableErrors = remote.RetryableErrors
	}
}

func mergeCircuitBreaker(cfg *CircuitBreakerGroupConfig, remote *apitypes.CircuitBreakerGroupRemoteConfig) {
	if remote.Enabled != nil {
		cfg.Enabled = *remote.Enabled
	}
	if remote.DefaultConfig != nil {
		mergeCBSpec(&cfg.DefaultConfig, remote.DefaultConfig)
	}
	if remote.Breakers != nil {
		if cfg.Breakers == nil {
			cfg.Breakers = make(map[string]CircuitBreakerConfigSpec)
		}
		for name, rb := range remote.Breakers {
			b := cfg.Breakers[name]
			mergeCBSpec(&b, &rb)
			cfg.Breakers[name] = b
		}
	}
}

func mergeCBSpec(cfg *CircuitBreakerConfigSpec, remote *apitypes.CircuitBreakerRemoteConfig) {
	if remote.MaxFailures != nil {
		cfg.MaxFailures = *remote.MaxFailures
	}
	if remote.Timeout != nil {
		if d, err := time.ParseDuration(*remote.Timeout); err == nil {
			cfg.Timeout = d
		}
	}
	if remote.MaxRequests != nil {
		cfg.MaxRequests = *remote.MaxRequests
	}
	if remote.FailureRatio != nil {
		cfg.FailureRatio = *remote.FailureRatio
	}
	if remote.MinRequests != nil {
		cfg.MinRequests = *remote.MinRequests
	}
}

func mergeLogging(cfg *LoggingConfig, remote *apitypes.LoggingRemoteConfig) {
	if remote.Level != nil {
		cfg.Level = *remote.Level
	}
}

func mergeShutdown(cfg *ShutdownConfig, remote *apitypes.ShutdownRemoteConfig) {
	if remote.Timeout != nil {
		if d, err := time.ParseDuration(*remote.Timeout); err == nil {
			cfg.Timeout = d
		}
	}
	if remote.CleanupWorkspaces != nil {
		cfg.CleanupWorkspaces = *remote.CleanupWorkspaces
	}
	if remote.ResetClaims != nil {
		cfg.ResetClaims = *remote.ResetClaims
	}
}
