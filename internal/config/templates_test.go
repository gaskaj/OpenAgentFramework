package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateEngine_RenderString(t *testing.T) {
	te := NewTemplateEngine("")

	template := "Hello {{.Name}}, you are a {{.Type}} agent with {{.MaxIterations}} iterations."
	context := map[string]interface{}{
		"Name":          "TestAgent",
		"Type":          "developer",
		"MaxIterations": 25,
	}

	result, err := te.RenderString(template, context)
	require.NoError(t, err)

	expected := "Hello TestAgent, you are a developer agent with 25 iterations."
	assert.Equal(t, expected, result)
}

func TestTemplateEngine_RenderPrompt(t *testing.T) {
	// Create temporary directory for test templates
	tempDir, err := os.MkdirTemp("", "template_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templateContent := `{{define "system"}}
You are a {{.AgentType}} agent.
Your task: {{.IssueContent}}
Max iterations: {{.MaxIterations}}
{{end}}

{{define "analyze"}}
Analyze this issue: {{.IssueContent}}
Profile: {{.ProfileName}}
{{end}}`

	templatePath := filepath.Join(tempDir, "test-agent.tmpl")
	require.NoError(t, os.WriteFile(templatePath, []byte(templateContent), 0644))

	te := NewTemplateEngine(tempDir)

	context := map[string]interface{}{
		"AgentType":       "developer",
		"IssueContent":    "Fix the bug in the login system",
		"MaxIterations":   20,
		"ProfileName":     "web-development",
	}

	// Test system prompt
	systemResult, err := te.RenderPrompt("test-agent", "system", context)
	require.NoError(t, err)

	assert.Contains(t, systemResult, "You are a developer agent")
	assert.Contains(t, systemResult, "Fix the bug in the login system")
	assert.Contains(t, systemResult, "Max iterations: 20")

	// Test analyze prompt
	analyzeResult, err := te.RenderPrompt("test-agent", "analyze", context)
	require.NoError(t, err)

	assert.Contains(t, analyzeResult, "Analyze this issue: Fix the bug in the login system")
	assert.Contains(t, analyzeResult, "Profile: web-development")
}

func TestTemplateEngine_TemplateFunctions(t *testing.T) {
	te := NewTemplateEngine("")

	tests := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "upper function",
			template: "{{upper .Text}}",
			context:  map[string]interface{}{"Text": "hello"},
			expected: "HELLO",
		},
		{
			name:     "lower function",
			template: "{{lower .Text}}",
			context:  map[string]interface{}{"Text": "HELLO"},
			expected: "hello",
		},
		{
			name:     "default function with empty value",
			template: "{{default \"fallback\" .Empty}}",
			context:  map[string]interface{}{"Empty": ""},
			expected: "fallback",
		},
		{
			name:     "default function with value",
			template: "{{default \"fallback\" .Value}}",
			context:  map[string]interface{}{"Value": "actual"},
			expected: "actual",
		},
		{
			name:     "join function",
			template: "{{join \", \" .Items}}",
			context:  map[string]interface{}{"Items": []string{"a", "b", "c"}},
			expected: "a, b, c",
		},
		{
			name:     "printf function",
			template: "{{printf \"Value: %d\" .Number}}",
			context:  map[string]interface{}{"Number": 42},
			expected: "Value: 42",
		},
		{
			name:     "indent function",
			template: "{{indent 4 .Text}}",
			context:  map[string]interface{}{"Text": "line1\nline2"},
			expected: "    line1\n    line2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := te.RenderString(tt.template, tt.context)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPromptContext_ToMap(t *testing.T) {
	ctx := NewPromptContext()
	ctx.IssueContent = "Test issue"
	ctx.Plan = "Test plan"
	ctx.AgentType = "developer"
	ctx.MaxIterations = 25
	ctx.SetVariable("custom", "value")

	contextMap := ctx.ToMap()

	assert.Equal(t, "Test issue", contextMap["IssueContent"])
	assert.Equal(t, "Test plan", contextMap["Plan"])
	assert.Equal(t, "developer", contextMap["AgentType"])
	assert.Equal(t, 25, contextMap["MaxIterations"])
	assert.Equal(t, "value", contextMap["custom"])
}

func TestPromptContext_Variables(t *testing.T) {
	ctx := NewPromptContext()

	// Test setting and getting variables
	ctx.SetVariable("test_key", "test_value")
	assert.Equal(t, "test_value", ctx.GetVariable("test_key"))

	// Test missing variable
	assert.Nil(t, ctx.GetVariable("missing_key"))

	// Test variables in context map
	contextMap := ctx.ToMap()
	assert.Equal(t, "test_value", contextMap["test_key"])
}

func TestTemplateEngine_ComplexTemplate(t *testing.T) {
	te := NewTemplateEngine("")

	template := `
{{- if .ComplexityEstimationEnabled}}
## Complexity Estimation
Budget: {{.MaxIterations}} iterations
{{- end}}

Tools: {{join ", " .ToolsAllowed}}

{{- if .AdditionalInstructions}}

Additional: {{.AdditionalInstructions}}
{{- end}}
`

	context := map[string]interface{}{
		"ComplexityEstimationEnabled": true,
		"MaxIterations":               20,
		"ToolsAllowed":               []string{"read_file", "write_file", "edit_file"},
		"AdditionalInstructions":      "Be extra careful",
	}

	result, err := te.RenderString(template, context)
	require.NoError(t, err)

	assert.Contains(t, result, "## Complexity Estimation")
	assert.Contains(t, result, "Budget: 20 iterations")
	assert.Contains(t, result, "Tools: read_file, write_file, edit_file")
	assert.Contains(t, result, "Additional: Be extra careful")
}