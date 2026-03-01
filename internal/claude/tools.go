package claude

import (
	"github.com/anthropics/anthropic-sdk-go"
)

// DevTools returns the tool definitions available to the developer agent.
func DevTools() []anthropic.ToolUnionParam {
	return []anthropic.ToolUnionParam{
		{OfTool: &anthropic.ToolParam{
			Name:        "read_file",
			Description: anthropic.String("Read the contents of a file at the given path."),
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
			Name:        "write_file",
			Description: anthropic.String("Write content to a file at the given path. Creates parent directories as needed."),
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
