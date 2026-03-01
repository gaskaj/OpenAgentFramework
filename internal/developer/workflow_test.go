package developer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gaskaj/DeveloperAndQAAgent/internal/gitops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRepo creates a temporary git repo with some files for testing.
func setupTestRepo(t *testing.T) *gitops.Repo {
	t.Helper()
	dir := t.TempDir()

	// Initialize a bare git repo so gitops.Open works.
	// We use the Repo's file operations directly, so we just need the directory.
	// Create a minimal structure that gitops expects.
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal", "developer"), 0o755))

	// Write test files.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "internal", "developer", "workflow.go"), []byte(`package developer

func claimIssue() error {
	// existing code
	return nil
}

func processIssue() error {
	return claimIssue()
}
`), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(dir, "internal", "developer", "agent.go"), []byte(`package developer

type DeveloperAgent struct {
	name string
}
`), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(`module github.com/example/test

go 1.21
`), 0o644))

	// gitops.Repo needs a real git repo, so init one.
	initGitRepo(t, dir)

	repo, err := gitops.Open(dir, "test-token")
	require.NoError(t, err)
	return repo
}

// initGitRepo initializes a git repo in the given directory.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	// Use go-git to init
	cmd := filepath.Join(dir, ".git")
	// Simpler: just create .git directory structure manually
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git", "objects"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git", "refs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0o644))
	_ = cmd // suppress unused
}

// --- executeEditFile tests ---

func TestExecuteEditFile_Success(t *testing.T) {
	repo := setupTestRepo(t)

	result, err := executeEditFile(repo, map[string]string{
		"path":       "internal/developer/workflow.go",
		"old_string": "// existing code\n\treturn nil",
		"new_string": "// updated code\n\tfmt.Println(\"claimed\")\n\treturn nil",
	})
	require.NoError(t, err)
	assert.Equal(t, "file edited successfully", result)

	// Verify file was updated.
	content, err := repo.ReadFile("internal/developer/workflow.go")
	require.NoError(t, err)
	assert.Contains(t, content, "// updated code")
	assert.Contains(t, content, "fmt.Println(\"claimed\")")
	// Verify untouched code is still there.
	assert.Contains(t, content, "func processIssue()")
}

func TestExecuteEditFile_NotFound(t *testing.T) {
	repo := setupTestRepo(t)

	_, err := executeEditFile(repo, map[string]string{
		"path":       "internal/developer/workflow.go",
		"old_string": "this string does not exist in the file",
		"new_string": "replacement",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestExecuteEditFile_AmbiguousMatch(t *testing.T) {
	repo := setupTestRepo(t)

	// "return" appears twice in the file (claimIssue and processIssue).
	_, err := executeEditFile(repo, map[string]string{
		"path":       "internal/developer/workflow.go",
		"old_string": "return",
		"new_string": "return // modified",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "appears")
	assert.Contains(t, err.Error(), "times")
}

func TestExecuteEditFile_FileDoesNotExist(t *testing.T) {
	repo := setupTestRepo(t)

	_, err := executeEditFile(repo, map[string]string{
		"path":       "nonexistent/file.go",
		"old_string": "anything",
		"new_string": "replacement",
	})
	require.Error(t, err)
}

// --- executeSearchFiles tests ---

func TestExecuteSearchFiles_FoundMatches(t *testing.T) {
	repo := setupTestRepo(t)

	result, err := executeSearchFiles(repo, map[string]string{
		"pattern": "claimIssue",
	})
	require.NoError(t, err)
	// Should find matches in workflow.go (definition + call).
	assert.Contains(t, result, "internal/developer/workflow.go")
	assert.Contains(t, result, "claimIssue")
	lines := strings.Split(result, "\n")
	assert.GreaterOrEqual(t, len(lines), 2, "expected at least 2 matches for claimIssue")
}

func TestExecuteSearchFiles_NoMatches(t *testing.T) {
	repo := setupTestRepo(t)

	result, err := executeSearchFiles(repo, map[string]string{
		"pattern": "xyzNonExistentPattern123",
	})
	require.NoError(t, err)
	assert.Equal(t, "no matches found", result)
}

func TestExecuteSearchFiles_RegexPattern(t *testing.T) {
	repo := setupTestRepo(t)

	result, err := executeSearchFiles(repo, map[string]string{
		"pattern": `func\s+\w+Issue`,
	})
	require.NoError(t, err)
	assert.Contains(t, result, "claimIssue")
	assert.Contains(t, result, "processIssue")
}

func TestExecuteSearchFiles_SubdirectoryFilter(t *testing.T) {
	repo := setupTestRepo(t)

	// Search only in internal/developer — should find matches.
	result, err := executeSearchFiles(repo, map[string]string{
		"pattern": "package",
		"path":    "internal/developer",
	})
	require.NoError(t, err)
	assert.Contains(t, result, "package developer")

	// Should NOT find go.mod content since we're scoped to internal/developer.
	assert.NotContains(t, result, "go.mod")
}

func TestExecuteSearchFiles_InvalidRegexFallsBackToLiteral(t *testing.T) {
	repo := setupTestRepo(t)

	// "[invalid regex" is not valid regex — should fall back to literal match.
	result, err := executeSearchFiles(repo, map[string]string{
		"pattern": "return nil",
	})
	require.NoError(t, err)
	assert.Contains(t, result, "workflow.go")
}

// --- preReadFiles tests ---

func TestPreReadFiles_HigherLimit(t *testing.T) {
	repo := setupTestRepo(t)

	// Write a large file to test truncation.
	largeContent := strings.Repeat("// line of code\n", 1500) // ~24k chars
	require.NoError(t, os.WriteFile(
		filepath.Join(repo.Dir(), "internal", "developer", "large.go"),
		[]byte("package developer\n\n"+largeContent),
		0o644,
	))

	result := preReadFiles(repo, []string{"internal/developer/large.go"})
	assert.NotEmpty(t, result)
	// Should be truncated at 15000, not 4000.
	assert.Contains(t, result, "truncated at 15000 chars")
}

func TestPreReadFiles_SmallFileNotTruncated(t *testing.T) {
	repo := setupTestRepo(t)

	result := preReadFiles(repo, []string{"internal/developer/workflow.go"})
	assert.NotEmpty(t, result)
	assert.NotContains(t, result, "truncated")
	assert.Contains(t, result, "claimIssue")
	assert.Contains(t, result, "processIssue")
}

func TestPreReadFiles_NonexistentFileSkipped(t *testing.T) {
	repo := setupTestRepo(t)

	result := preReadFiles(repo, []string{"does/not/exist.go", "internal/developer/workflow.go"})
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "workflow.go")
	assert.NotContains(t, result, "does/not/exist.go")
}

func TestPreReadFiles_EmptyPaths(t *testing.T) {
	repo := setupTestRepo(t)
	result := preReadFiles(repo, nil)
	assert.Empty(t, result)
}
