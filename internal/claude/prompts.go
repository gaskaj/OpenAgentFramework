package claude

import (
	"fmt"
	"strings"
)

// FormatSystemPrompt builds a system prompt from parts.
func FormatSystemPrompt(parts ...string) string {
	return strings.Join(parts, "\n\n")
}

// FormatIssueContext formats a GitHub issue for inclusion in a prompt.
func FormatIssueContext(number int, title, body string, labels []string) string {
	return fmt.Sprintf(`## GitHub Issue #%d: %s

**Labels:** %s

%s`, number, title, strings.Join(labels, ", "), body)
}

// FormatFileList formats a list of files for inclusion in a prompt.
func FormatFileList(files []string) string {
	if len(files) == 0 {
		return "(no files)"
	}
	var b strings.Builder
	for _, f := range files {
		b.WriteString("- ")
		b.WriteString(f)
		b.WriteString("\n")
	}
	return b.String()
}
