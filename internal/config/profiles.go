package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// AgentProfile represents a complete agent configuration profile.
type AgentProfile struct {
	Name        string                 `yaml:"name" mapstructure:"name"`
	Description string                 `yaml:"description" mapstructure:"description"`
	Extends     string                 `yaml:"extends,omitempty" mapstructure:"extends"`
	Agent       AgentConfig            `yaml:"agent" mapstructure:"agent"`
	Claude      ClaudeProfileConfig    `yaml:"claude,omitempty" mapstructure:"claude"`
	Environment map[string]interface{} `yaml:"environment,omitempty" mapstructure:"environment"`
}

// AgentConfig holds agent-specific configuration within a profile.
type AgentConfig struct {
	Prompts   PromptsConfig           `yaml:"prompts,omitempty" mapstructure:"prompts"`
	Behavior  BehaviorConfig          `yaml:"behavior,omitempty" mapstructure:"behavior"`
	Templates map[string]string       `yaml:"templates,omitempty" mapstructure:"templates"`
}

// PromptsConfig holds prompt templates for different agent operations.
type PromptsConfig struct {
	System             string `yaml:"system,omitempty" mapstructure:"system"`
	Analyze            string `yaml:"analyze,omitempty" mapstructure:"analyze"`
	Implement          string `yaml:"implement,omitempty" mapstructure:"implement"`
	ComplexityEstimate string `yaml:"complexity_estimate,omitempty" mapstructure:"complexity_estimate"`
	Decompose          string `yaml:"decompose,omitempty" mapstructure:"decompose"`
	ReactiveDecompose  string `yaml:"reactive_decompose,omitempty" mapstructure:"reactive_decompose"`
}

// BehaviorConfig holds behavioral parameters for agents.
type BehaviorConfig struct {
	MaxIterations  int             `yaml:"max_iterations,omitempty" mapstructure:"max_iterations"`
	TimeoutSeconds int             `yaml:"timeout_seconds,omitempty" mapstructure:"timeout_seconds"`
	RetryCount     int             `yaml:"retry_count,omitempty" mapstructure:"retry_count"`
	ToolsAllowed   []string        `yaml:"tools_allowed,omitempty" mapstructure:"tools_allowed"`
	FilePatterns   FilePatternsConfig `yaml:"file_patterns,omitempty" mapstructure:"file_patterns"`
}

// FilePatternsConfig holds file inclusion/exclusion patterns.
type FilePatternsConfig struct {
	Include []string `yaml:"include,omitempty" mapstructure:"include"`
	Exclude []string `yaml:"exclude,omitempty" mapstructure:"exclude"`
}

// ClaudeProfileConfig holds Claude-specific overrides for a profile.
type ClaudeProfileConfig struct {
	Model       string  `yaml:"model,omitempty" mapstructure:"model"`
	MaxTokens   int     `yaml:"max_tokens,omitempty" mapstructure:"max_tokens"`
	Temperature float64 `yaml:"temperature,omitempty" mapstructure:"temperature"`
}

// ProfileManager manages loading and resolving agent profiles.
type ProfileManager struct {
	profilesDir string
	cache       map[string]*AgentProfile
}

// NewProfileManager creates a new profile manager.
func NewProfileManager(profilesDir string) *ProfileManager {
	return &ProfileManager{
		profilesDir: profilesDir,
		cache:       make(map[string]*AgentProfile),
	}
}

// LoadProfile loads and resolves a profile by name, handling inheritance.
func (pm *ProfileManager) LoadProfile(profileName string) (*AgentProfile, error) {
	// Check cache first
	if profile, exists := pm.cache[profileName]; exists {
		return profile, nil
	}

	profile, err := pm.loadProfileFile(profileName)
	if err != nil {
		return nil, fmt.Errorf("loading profile %s: %w", profileName, err)
	}

	// Resolve inheritance if needed
	if profile.Extends != "" {
		parent, err := pm.LoadProfile(profile.Extends)
		if err != nil {
			return nil, fmt.Errorf("loading parent profile %s: %w", profile.Extends, err)
		}
		profile = pm.mergeProfiles(parent, profile)
	}

	// Process templates
	if err := pm.processTemplates(profile); err != nil {
		return nil, fmt.Errorf("processing templates for profile %s: %w", profileName, err)
	}

	// Cache the resolved profile
	pm.cache[profileName] = profile

	return profile, nil
}

// GetAgentConfig returns the agent configuration from a profile.
func (pm *ProfileManager) GetAgentConfig(profile *AgentProfile) AgentConfig {
	return profile.Agent
}

// ApplyToBaseConfig applies profile settings to the base configuration.
func (pm *ProfileManager) ApplyToBaseConfig(baseConfig *Config, profile *AgentProfile, environment string) error {
	// Apply Claude overrides
	if profile.Claude.Model != "" {
		baseConfig.Claude.Model = profile.Claude.Model
	}
	if profile.Claude.MaxTokens > 0 {
		baseConfig.Claude.MaxTokens = profile.Claude.MaxTokens
	}

	// Apply environment-specific overrides
	if envConfig, exists := profile.Environment[environment]; exists {
		if envMap, ok := envConfig.(map[string]interface{}); ok {
			return pm.applyEnvironmentOverrides(baseConfig, envMap)
		}
	}

	return nil
}

// ListProfiles returns a list of available profile names.
func (pm *ProfileManager) ListProfiles() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(pm.profilesDir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("listing profile files: %w", err)
	}

	var profiles []string
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".yaml")
		profiles = append(profiles, name)
	}

	return profiles, nil
}

func (pm *ProfileManager) loadProfileFile(profileName string) (*AgentProfile, error) {
	profilePath := filepath.Join(pm.profilesDir, profileName+".yaml")
	
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("reading profile file %s: %w", profilePath, err)
	}

	var profile AgentProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("parsing profile YAML: %w", err)
	}

	return &profile, nil
}

func (pm *ProfileManager) mergeProfiles(parent, child *AgentProfile) *AgentProfile {
	merged := *child // Start with child

	// Merge prompts
	if merged.Agent.Prompts.System == "" && parent.Agent.Prompts.System != "" {
		merged.Agent.Prompts.System = parent.Agent.Prompts.System
	}
	if merged.Agent.Prompts.Analyze == "" && parent.Agent.Prompts.Analyze != "" {
		merged.Agent.Prompts.Analyze = parent.Agent.Prompts.Analyze
	}
	if merged.Agent.Prompts.Implement == "" && parent.Agent.Prompts.Implement != "" {
		merged.Agent.Prompts.Implement = parent.Agent.Prompts.Implement
	}
	if merged.Agent.Prompts.ComplexityEstimate == "" && parent.Agent.Prompts.ComplexityEstimate != "" {
		merged.Agent.Prompts.ComplexityEstimate = parent.Agent.Prompts.ComplexityEstimate
	}
	if merged.Agent.Prompts.Decompose == "" && parent.Agent.Prompts.Decompose != "" {
		merged.Agent.Prompts.Decompose = parent.Agent.Prompts.Decompose
	}
	if merged.Agent.Prompts.ReactiveDecompose == "" && parent.Agent.Prompts.ReactiveDecompose != "" {
		merged.Agent.Prompts.ReactiveDecompose = parent.Agent.Prompts.ReactiveDecompose
	}

	// Merge behavior
	if merged.Agent.Behavior.MaxIterations == 0 {
		merged.Agent.Behavior.MaxIterations = parent.Agent.Behavior.MaxIterations
	}
	if merged.Agent.Behavior.TimeoutSeconds == 0 {
		merged.Agent.Behavior.TimeoutSeconds = parent.Agent.Behavior.TimeoutSeconds
	}
	if merged.Agent.Behavior.RetryCount == 0 {
		merged.Agent.Behavior.RetryCount = parent.Agent.Behavior.RetryCount
	}
	if len(merged.Agent.Behavior.ToolsAllowed) == 0 {
		merged.Agent.Behavior.ToolsAllowed = parent.Agent.Behavior.ToolsAllowed
	}
	if len(merged.Agent.Behavior.FilePatterns.Include) == 0 {
		merged.Agent.Behavior.FilePatterns.Include = parent.Agent.Behavior.FilePatterns.Include
	}
	if len(merged.Agent.Behavior.FilePatterns.Exclude) == 0 {
		merged.Agent.Behavior.FilePatterns.Exclude = parent.Agent.Behavior.FilePatterns.Exclude
	}

	// Merge templates
	if merged.Agent.Templates == nil {
		merged.Agent.Templates = make(map[string]string)
	}
	for key, value := range parent.Agent.Templates {
		if _, exists := merged.Agent.Templates[key]; !exists {
			merged.Agent.Templates[key] = value
		}
	}

	// Merge Claude config
	if merged.Claude.Model == "" && parent.Claude.Model != "" {
		merged.Claude.Model = parent.Claude.Model
	}
	if merged.Claude.MaxTokens == 0 {
		merged.Claude.MaxTokens = parent.Claude.MaxTokens
	}
	if merged.Claude.Temperature == 0 {
		merged.Claude.Temperature = parent.Claude.Temperature
	}

	return &merged
}

func (pm *ProfileManager) processTemplates(profile *AgentProfile) error {
	templateData := map[string]interface{}{
		"Profile": profile,
	}
	
	// Add template variables
	for key, value := range profile.Agent.Templates {
		templateData[key] = value
	}

	// Process prompts
	prompts := &profile.Agent.Prompts
	if err := pm.processTemplate(&prompts.System, templateData); err != nil {
		return fmt.Errorf("processing system prompt template: %w", err)
	}
	if err := pm.processTemplate(&prompts.Analyze, templateData); err != nil {
		return fmt.Errorf("processing analyze prompt template: %w", err)
	}
	if err := pm.processTemplate(&prompts.Implement, templateData); err != nil {
		return fmt.Errorf("processing implement prompt template: %w", err)
	}
	if err := pm.processTemplate(&prompts.ComplexityEstimate, templateData); err != nil {
		return fmt.Errorf("processing complexity estimate prompt template: %w", err)
	}
	if err := pm.processTemplate(&prompts.Decompose, templateData); err != nil {
		return fmt.Errorf("processing decompose prompt template: %w", err)
	}
	if err := pm.processTemplate(&prompts.ReactiveDecompose, templateData); err != nil {
		return fmt.Errorf("processing reactive decompose prompt template: %w", err)
	}

	return nil
}

func (pm *ProfileManager) processTemplate(content *string, data map[string]interface{}) error {
	if *content == "" {
		return nil
	}

	tmpl, err := template.New("prompt").Parse(*content)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	*content = buf.String()
	return nil
}

func (pm *ProfileManager) applyEnvironmentOverrides(baseConfig *Config, envOverrides map[string]interface{}) error {
	v := viper.New()
	
	// Convert base config to map
	configMap := make(map[string]interface{})
	v.Set("config", baseConfig)
	if err := v.UnmarshalKey("config", &configMap); err != nil {
		return fmt.Errorf("converting base config to map: %w", err)
	}

	// Apply overrides
	for key, value := range envOverrides {
		v.Set(key, value)
	}

	// Unmarshal back to config struct
	if err := v.Unmarshal(baseConfig); err != nil {
		return fmt.Errorf("unmarshaling overridden config: %w", err)
	}

	return nil
}