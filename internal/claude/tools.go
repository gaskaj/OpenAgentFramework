package claude

import (
	"github.com/anthropics/anthropic-sdk-go"
)

// DevTools returns the tool definitions available to the developer agent.
func DevTools() []anthropic.ToolUnionParam {
	return []anthropic.ToolUnionParam{
		{OfTool: &anthropic.ToolParam{
			Name:        "read_file",
			Description: anthropic.String("Read the contents of a file at the given path. For searching across multiple files, prefer search_files instead."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The file path to read, relative to the workspace root.",
					},
				},
				Required: []string{"path"},
			},
		}},
		{OfTool: &anthropic.ToolParam{
			Name:        "edit_file",
			Description: anthropic.String("Edit a file by replacing a specific string with new content. The old_string must appear exactly once in the file. Preferred over write_file for modifying existing files — avoids rewriting the entire file."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The file path to edit, relative to the workspace root.",
					},
					"old_string": map[string]interface{}{
						"type":        "string",
						"description": "The exact string to find and replace. Must appear exactly once in the file. Include enough surrounding context to make it unique.",
					},
					"new_string": map[string]interface{}{
						"type":        "string",
						"description": "The replacement string.",
					},
				},
				Required: []string{"path", "old_string", "new_string"},
			},
		}},
		{OfTool: &anthropic.ToolParam{
			Name:        "write_file",
			Description: anthropic.String("Write content to a file at the given path. Creates parent directories as needed. For modifying existing files, prefer edit_file instead to avoid rewriting the entire file."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The file path to write, relative to the workspace root.",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "The full content to write to the file.",
					},
				},
				Required: []string{"path", "content"},
			},
		}},
		{OfTool: &anthropic.ToolParam{
			Name:        "search_files",
			Description: anthropic.String("Search for a text pattern across all files in the workspace. Returns matching lines with file paths and line numbers. Use this to find where functions, types, or variables are defined or used."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "The search pattern (plain text or Go regexp).",
					},
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Optional subdirectory to restrict the search to, relative to workspace root. Defaults to searching the entire workspace.",
					},
				},
				Required: []string{"pattern"},
			},
		}},
		{OfTool: &anthropic.ToolParam{
			Name:        "list_files",
			Description: anthropic.String("List files and directories at the given path."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The directory path to list, relative to the workspace root.",
					},
				},
				Required: []string{"path"},
			},
		}},
		{OfTool: &anthropic.ToolParam{
			Name:        "run_command",
			Description: anthropic.String("Run a shell command in the workspace directory. Use for building, testing, or other operations."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The shell command to execute.",
					},
				},
				Required: []string{"command"},
			},
		}},
	}
}
