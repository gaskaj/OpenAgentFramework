package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileManager_LoadProfile(t *testing.T) {
	// Create temporary directory for test profiles
	tempDir, err := os.MkdirTemp("", "profile_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test profile file
	profileContent := `name: "test-profile"
description: "Test profile"
agent:
  prompts:
    system: "You are a test agent"
  behavior:
    max_iterations: 10
    timeout_seconds: 300
  templates:
    test_var: "test_value"
claude:
  model: "claude-test"
  max_tokens: 1000
`
	profilePath := filepath.Join(tempDir, "test-profile.yaml")
	require.NoError(t, os.WriteFile(profilePath, []byte(profileContent), 0644))

	pm := NewProfileManager(tempDir)

	profile, err := pm.LoadProfile("test-profile")
	require.NoError(t, err)

	assert.Equal(t, "test-profile", profile.Name)
	assert.Equal(t, "Test profile", profile.Description)
	assert.Equal(t, "You are a test agent", profile.Agent.Prompts.System)
	assert.Equal(t, 10, profile.Agent.Behavior.MaxIterations)
	assert.Equal(t, 300, profile.Agent.Behavior.TimeoutSeconds)
	assert.Equal(t, "test_value", profile.Agent.Templates["test_var"])
	assert.Equal(t, "claude-test", profile.Claude.Model)
	assert.Equal(t, 1000, profile.Claude.MaxTokens)
}

func TestProfileManager_LoadProfileWithInheritance(t *testing.T) {
	// Create temporary directory for test profiles
	tempDir, err := os.MkdirTemp("", "profile_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create parent profile
	parentContent := `name: "parent"
description: "Parent profile"
agent:
  prompts:
    system: "Parent system prompt"
    analyze: "Parent analyze prompt"
  behavior:
    max_iterations: 20
    timeout_seconds: 600
  templates:
    parent_var: "parent_value"
    shared_var: "parent_shared"
claude:
  model: "claude-parent"
  max_tokens: 2000
`
	parentPath := filepath.Join(tempDir, "parent.yaml")
	require.NoError(t, os.WriteFile(parentPath, []byte(parentContent), 0644))

	// Create child profile that extends parent
	childContent := `name: "child"
description: "Child profile"
extends: "parent"
agent:
  prompts:
    system: "Child system prompt"  # Override
  behavior:
    max_iterations: 15  # Override
  templates:
    child_var: "child_value"
    shared_var: "child_shared"  # Override
claude:
  max_tokens: 1500  # Override
`
	childPath := filepath.Join(tempDir, "child.yaml")
	require.NoError(t, os.WriteFile(childPath, []byte(childContent), 0644))

	pm := NewProfileManager(tempDir)

	profile, err := pm.LoadProfile("child")
	require.NoError(t, err)

	// Child values should override parent
	assert.Equal(t, "child", profile.Name)
	assert.Equal(t, "Child profile", profile.Description)
	assert.Equal(t, "Child system prompt", profile.Agent.Prompts.System)
	assert.Equal(t, "Parent analyze prompt", profile.Agent.Prompts.Analyze) // Inherited
	assert.Equal(t, 15, profile.Agent.Behavior.MaxIterations)              // Overridden
	assert.Equal(t, 600, profile.Agent.Behavior.TimeoutSeconds)            // Inherited
	assert.Equal(t, "child_value", profile.Agent.Templates["child_var"])
	assert.Equal(t, "parent_value", profile.Agent.Templates["parent_var"]) // Inherited
	assert.Equal(t, "child_shared", profile.Agent.Templates["shared_var"]) // Overridden
	assert.Equal(t, "claude-parent", profile.Claude.Model)                 // Inherited
	assert.Equal(t, 1500, profile.Claude.MaxTokens)                       // Overridden
}

func TestProfileManager_LoadProfileWithTemplates(t *testing.T) {
	// Create temporary directory for test profiles
	tempDir, err := os.MkdirTemp("", "profile_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create profile with template
	profileContent := `name: "template-test"
description: "Template test profile"
agent:
  prompts:
    system: "You are a {{.TestType}} agent for {{.Profile.Name}}"
  templates:
    TestType: "specialized"
`
	profilePath := filepath.Join(tempDir, "template-test.yaml")
	require.NoError(t, os.WriteFile(profilePath, []byte(profileContent), 0644))

	pm := NewProfileManager(tempDir)

	profile, err := pm.LoadProfile("template-test")
	require.NoError(t, err)

	expected := "You are a specialized agent for template-test"
	assert.Equal(t, expected, profile.Agent.Prompts.System)
}

func TestProfileManager_ListProfiles(t *testing.T) {
	// Create temporary directory for test profiles
	tempDir, err := os.MkdirTemp("", "profile_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test profile files
	profiles := []string{"profile1", "profile2", "profile3"}
	for _, name := range profiles {
		content := `name: "` + name + `"`
		path := filepath.Join(tempDir, name+".yaml")
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	}

	pm := NewProfileManager(tempDir)

	listed, err := pm.ListProfiles()
	require.NoError(t, err)

	assert.Len(t, listed, 3)
	for _, expected := range profiles {
		assert.Contains(t, listed, expected)
	}
}

func TestProfileManager_ApplyToBaseConfig(t *testing.T) {
	pm := NewProfileManager("")

	baseConfig := &Config{
		Claude: ClaudeConfig{
			Model:     "claude-base",
			MaxTokens: 4000,
		},
	}

	profile := &AgentProfile{
		Claude: ClaudeProfileConfig{
			Model:     "claude-override",
			MaxTokens: 8000,
		},
		Environment: map[string]interface{}{
			"test": map[string]interface{}{
				"claude.max_tokens": 9000,
			},
		},
	}

	err := pm.ApplyToBaseConfig(baseConfig, profile, "test")
	require.NoError(t, err)

	// Profile overrides should be applied first, then environment overrides
	assert.Equal(t, "claude-override", baseConfig.Claude.Model)
	assert.Equal(t, 9000, baseConfig.Claude.MaxTokens) // Environment override should apply
}

func TestAgentConfig_GetBehaviorDefaults(t *testing.T) {
	agentConfig := AgentConfig{
		Behavior: BehaviorConfig{
			MaxIterations:  25,
			TimeoutSeconds: 1800,
			ToolsAllowed:   []string{"read_file", "write_file"},
		},
	}

	assert.Equal(t, 25, agentConfig.Behavior.MaxIterations)
	assert.Equal(t, 1800, agentConfig.Behavior.TimeoutSeconds)
	assert.Contains(t, agentConfig.Behavior.ToolsAllowed, "read_file")
	assert.Contains(t, agentConfig.Behavior.ToolsAllowed, "write_file")
}