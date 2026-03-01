package developer

import (
	"fmt"

	"github.com/gaskaj/DeveloperAndQAAgent/internal/config"
)

// Default prompts - used as fallbacks when no profile is configured
const (
	DefaultSystemPrompt = `You are an autonomous developer agent. You write clean, production-quality Go code.

Your workflow:
1. Analyze the GitHub issue requirements carefully.
2. Plan your implementation approach.
3. Write code using the available tools.
4. Test your changes by running "go build ./..." and "go test ./...".
5. Fix any compilation or test errors.
6. Once build and tests pass, respond with a summary of changes (no tool calls) to finish.

Guidelines:
- Write idiomatic Go code following standard conventions.
- Include appropriate error handling.
- Keep functions focused and well-named.
- Add tests for new functionality.
- Do not modify files unrelated to the issue.

Tool preferences:
- Use edit_file (not write_file) to modify existing files. edit_file replaces a specific string,
  so you only touch the lines that change instead of rewriting the entire file.
- Use write_file only for creating NEW files that don't exist yet.
- Use search_files to find where functions, types, or variables are defined or used,
  instead of reading multiple files one-by-one to search manually.
- Use read_file when you need to see the full contents of a specific file.

Efficiency:
- You have a limited iteration budget. Each API round-trip counts as one iteration.
- IMPORTANT: Invoke multiple tools in a single response whenever they are independent.
  For example, read_file("a.go") and read_file("b.go") can be called together in one response.
  This counts as ONE iteration, not two.
- Batch all independent reads together, all independent edit_file calls together, etc.
- Run "go build ./..." and "go test ./..." in a single run_command when possible:
  run_command("go build ./... && go test ./...")
- STOP as soon as build and tests pass — respond with text only (no tool calls) to finish.`

	DefaultAnalyzePrompt = `Analyze the following GitHub issue and create an implementation plan.

%s

Respond with a concise plan listing:
1. Files to create or modify
2. Key design decisions
3. Testing approach`

	DefaultImplementPrompt = `Implement the following plan for this issue.

## Issue
%s

## Plan
%s

Instructions:
1. Use edit_file to modify existing files (NOT write_file). Use write_file only for new files.
2. Use search_files to find references instead of reading files one-by-one.
3. After making changes, verify with: run_command("go build ./... && go test ./...")
4. If build/tests fail, fix the errors and re-run.
5. Once build and tests pass, respond with a summary of your changes (no tool calls) to finish.

Batch independent tool calls in the same response to save iterations.`

	DefaultComplexityEstimatePrompt = `

## Complexity Estimation

After creating your plan, estimate the number of API round-trip iterations needed to implement it.

IMPORTANT: One iteration = one API round-trip. Multiple tool calls in the SAME response count as
a SINGLE iteration. For example, calling read_file("a.go") and read_file("b.go") together in one
response is 1 iteration, not 2. Group independent operations together.

Enumerate each round-trip step-by-step, for example:

1. search_files("functionName") + read_file("go.mod") — find references and deps (1 iteration)
2. read_file("handler.go") + read_file("model.go") — inspect existing code (1 iteration)
3. edit_file("model.go") + edit_file("handler.go") — modify files (1 iteration)
4. write_file("new_test.go") — create new test file (1 iteration)
5. run_command("go build ./... && go test ./...") — verify (1 iteration)
6. (buffer for fixing test failures, unexpected reads) — reserve 2-3 iterations

Note: edit_file modifies a specific section of a file (no need to read first), and search_files
finds references across the codebase in one call. Factor these into your estimate.

Then state the total on its own line in exactly this format:

**Estimated iterations**: <N>

The iteration budget is %d. Include a buffer of 2-3 iterations for retries and fixes.
If your estimate (including buffer) exceeds %d of the budget (i.e., 50%%), answer "no".

At the end of your response, include exactly one of:

**Fits within budget**: yes

**Fits within budget**: no

If the answer is "no", also include a decomposition plan using this format:

## Decomposition Plan

### Subtask 1: <title>
<description of what this subtask should accomplish>

### Subtask 2: <title>
<description of what this subtask should accomplish>

(and so on, up to %d subtasks)
`

	DefaultDecomposePrompt = `The following GitHub issue is too complex to implement in a single pass (iteration budget: %d).

Break it into smaller, independently implementable subtasks.

## Issue
%s

## Plan
%s

Respond with a decomposition plan using this exact format:

## Decomposition Plan

### Subtask 1: <title>
<description of what this subtask should accomplish>

### Subtask 2: <title>
<description of what this subtask should accomplish>

(up to %d subtasks)

Each subtask should be self-contained and result in a working, testable change.`

	DefaultReactiveDecomposePrompt = `The implementation of the following issue ran out of iteration budget before completing.

## Original Issue
%s

## Plan
%s

The agent was partway through implementation when the iteration limit was reached. Break the REMAINING work into smaller subtasks that can each be completed independently.

Respond with a decomposition plan using this exact format:

## Decomposition Plan

### Subtask 1: <title>
<description of what this subtask should accomplish>

### Subtask 2: <title>
<description of what this subtask should accomplish>

(up to %d subtasks)

Each subtask should be self-contained and result in a working, testable change. Focus on what still needs to be done, not what was already completed.`
)

// PromptRenderer handles rendering of prompts with configuration support.
type PromptRenderer struct {
	config         *config.Config
	profile        *config.AgentProfile
	templateEngine *config.TemplateEngine
}

// NewPromptRenderer creates a new prompt renderer.
func NewPromptRenderer(cfg *config.Config, profile *config.AgentProfile) (*PromptRenderer, error) {
	var templateEngine *config.TemplateEngine
	if cfg.Profiles.Enabled && cfg.Profiles.PromptsDir != "" {
		templateEngine = config.NewTemplateEngine(cfg.Profiles.PromptsDir)
	}

	return &PromptRenderer{
		config:         cfg,
		profile:        profile,
		templateEngine: templateEngine,
	}, nil
}

// RenderSystemPrompt renders the system prompt.
func (pr *PromptRenderer) RenderSystemPrompt() (string, error) {
	if pr.profile != nil && pr.profile.Agent.Prompts.System != "" {
		return pr.profile.Agent.Prompts.System, nil
	}
	return DefaultSystemPrompt, nil
}

// RenderAnalyzePrompt renders the analysis prompt.
func (pr *PromptRenderer) RenderAnalyzePrompt(issueContent string, withComplexityEstimation bool) (string, error) {
	basePrompt := DefaultAnalyzePrompt
	if pr.profile != nil && pr.profile.Agent.Prompts.Analyze != "" {
		basePrompt = pr.profile.Agent.Prompts.Analyze
	}

	prompt := fmt.Sprintf(basePrompt, issueContent)

	if withComplexityEstimation {
		complexityPrompt := DefaultComplexityEstimatePrompt
		if pr.profile != nil && pr.profile.Agent.Prompts.ComplexityEstimate != "" {
			complexityPrompt = pr.profile.Agent.Prompts.ComplexityEstimate
		}

		budget := pr.config.Decomposition.MaxIterationBudget
		halfBudget := budget / 2
		maxSubtasks := pr.config.Decomposition.MaxSubtasks

		prompt += fmt.Sprintf(complexityPrompt, budget, halfBudget, maxSubtasks)
	}

	return prompt, nil
}

// RenderImplementPrompt renders the implementation prompt.
func (pr *PromptRenderer) RenderImplementPrompt(issueContent, plan string) (string, error) {
	basePrompt := DefaultImplementPrompt
	if pr.profile != nil && pr.profile.Agent.Prompts.Implement != "" {
		basePrompt = pr.profile.Agent.Prompts.Implement
	}

	return fmt.Sprintf(basePrompt, issueContent, plan), nil
}

// RenderDecomposePrompt renders the decomposition prompt.
func (pr *PromptRenderer) RenderDecomposePrompt(issueContent, plan string) (string, error) {
	basePrompt := DefaultDecomposePrompt
	if pr.profile != nil && pr.profile.Agent.Prompts.Decompose != "" {
		basePrompt = pr.profile.Agent.Prompts.Decompose
	}

	budget := pr.config.Decomposition.MaxIterationBudget
	maxSubtasks := pr.config.Decomposition.MaxSubtasks

	return fmt.Sprintf(basePrompt, budget, issueContent, plan, maxSubtasks), nil
}

// RenderReactiveDecomposePrompt renders the reactive decomposition prompt.
func (pr *PromptRenderer) RenderReactiveDecomposePrompt(issueContent, plan string) (string, error) {
	basePrompt := DefaultReactiveDecomposePrompt
	if pr.profile != nil && pr.profile.Agent.Prompts.ReactiveDecompose != "" {
		basePrompt = pr.profile.Agent.Prompts.ReactiveDecompose
	}

	maxSubtasks := pr.config.Decomposition.MaxSubtasks

	return fmt.Sprintf(basePrompt, issueContent, plan, maxSubtasks), nil
}

// GetBehaviorConfig returns the behavior configuration from the profile or defaults.
func (pr *PromptRenderer) GetBehaviorConfig() config.BehaviorConfig {
	if pr.profile != nil && (pr.profile.Agent.Behavior.MaxIterations > 0 || 
		pr.profile.Agent.Behavior.TimeoutSeconds > 0 ||
		len(pr.profile.Agent.Behavior.ToolsAllowed) > 0) {
		return pr.profile.Agent.Behavior
	}

	// Return reasonable defaults
	return config.BehaviorConfig{
		MaxIterations:  25,
		TimeoutSeconds: 1800,
		RetryCount:     3,
		ToolsAllowed:   []string{"read_file", "write_file", "edit_file", "search_files", "list_files", "run_command"},
	}
}
