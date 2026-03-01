package config

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// TemplateEngine handles rendering of prompt templates with context data.
type TemplateEngine struct {
	templatesDir string
	funcMap      template.FuncMap
}

// NewTemplateEngine creates a new template engine.
func NewTemplateEngine(templatesDir string) *TemplateEngine {
	return &TemplateEngine{
		templatesDir: templatesDir,
		funcMap:      createTemplateFuncMap(),
	}
}

// RenderPrompt renders a prompt template with the given context.
func (te *TemplateEngine) RenderPrompt(templateName string, promptType string, context map[string]interface{}) (string, error) {
	templatePath := filepath.Join(te.templatesDir, templateName+".tmpl")
	
	tmpl, err := template.New(templateName).
		Funcs(te.funcMap).
		ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("parsing template %s: %w", templatePath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, promptType, context); err != nil {
		return "", fmt.Errorf("executing template %s.%s: %w", templateName, promptType, err)
	}

	return buf.String(), nil
}

// RenderString renders a template string with the given context.
func (te *TemplateEngine) RenderString(templateStr string, context map[string]interface{}) (string, error) {
	tmpl, err := template.New("inline").
		Funcs(te.funcMap).
		Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("parsing inline template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, context); err != nil {
		return "", fmt.Errorf("executing inline template: %w", err)
	}

	return buf.String(), nil
}

// PromptContext holds the context data for rendering prompts.
type PromptContext struct {
	// Core data
	IssueContent    string
	Plan            string
	AgentType       string
	ProfileName     string
	
	// Behavioral settings
	MaxIterations   int
	TimeoutSeconds  int
	ToolsAllowed    []string
	
	// Prompt components (from hardcoded prompts or profile)
	SystemPrompt             string
	AnalyzePrompt           string
	ImplementPrompt         string
	ComplexityEstimatePrompt string
	DecomposePrompt         string
	ReactiveDecomposePrompt string
	
	// Flags and options
	ComplexityEstimationEnabled bool
	DecompositionEnabled        bool
	
	// Additional context
	AdditionalInstructions string
	InstructionFooter      string
	RuntimeContext         string
	
	// Template variables
	Variables map[string]interface{}
}

// NewPromptContext creates a new prompt context.
func NewPromptContext() *PromptContext {
	return &PromptContext{
		Variables: make(map[string]interface{}),
	}
}

// SetVariable sets a template variable.
func (pc *PromptContext) SetVariable(key string, value interface{}) {
	pc.Variables[key] = value
}

// GetVariable gets a template variable.
func (pc *PromptContext) GetVariable(key string) interface{} {
	return pc.Variables[key]
}

// ToMap converts the prompt context to a map for template rendering.
func (pc *PromptContext) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"IssueContent":                pc.IssueContent,
		"Plan":                       pc.Plan,
		"AgentType":                  pc.AgentType,
		"ProfileName":                pc.ProfileName,
		"MaxIterations":              pc.MaxIterations,
		"TimeoutSeconds":             pc.TimeoutSeconds,
		"ToolsAllowed":               pc.ToolsAllowed,
		"SystemPrompt":               pc.SystemPrompt,
		"AnalyzePrompt":              pc.AnalyzePrompt,
		"ImplementPrompt":            pc.ImplementPrompt,
		"ComplexityEstimatePrompt":   pc.ComplexityEstimatePrompt,
		"DecomposePrompt":            pc.DecomposePrompt,
		"ReactiveDecomposePrompt":    pc.ReactiveDecomposePrompt,
		"ComplexityEstimationEnabled": pc.ComplexityEstimationEnabled,
		"DecompositionEnabled":       pc.DecompositionEnabled,
		"AdditionalInstructions":     pc.AdditionalInstructions,
		"InstructionFooter":          pc.InstructionFooter,
		"RuntimeContext":             pc.RuntimeContext,
	}
	
	// Add template variables
	for key, value := range pc.Variables {
		result[key] = value
	}
	
	return result
}

// createTemplateFuncMap creates the function map for templates.
func createTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		// String manipulation functions
		"upper":    strings.ToUpper,
		"lower":    strings.ToLower,
		"title":    strings.Title,
		"trim":     strings.TrimSpace,
		"contains": strings.Contains,
		"replace":  strings.ReplaceAll,
		"split":    strings.Split,
		"join": func(sep string, items []string) string {
			return strings.Join(items, sep)
		},
		
		// Conditional functions
		"default": func(defaultVal interface{}, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},
		
		// List functions
		"list": func(items ...interface{}) []interface{} {
			return items
		},
		
		// String formatting
		"printf": fmt.Sprintf,
		
		// Utility functions
		"indent": func(spaces int, text string) string {
			prefix := strings.Repeat(" ", spaces)
			lines := strings.Split(text, "\n")
			for i, line := range lines {
				if line != "" {
					lines[i] = prefix + line
				}
			}
			return strings.Join(lines, "\n")
		},
		
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
		
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
	}
}