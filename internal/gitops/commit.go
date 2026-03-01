package gitops

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// WriteFile writes content to a file in the repository, creating directories as needed.
func (r *Repo) WriteFile(path, content string) error {
	fullPath := filepath.Join(r.dir, path)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("creating directories for %s: %w", path, err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing file %s: %w", path, err)
	}

	return nil
}

// ReadFile reads a file from the repository.
func (r *Repo) ReadFile(path string) (string, error) {
	fullPath := filepath.Join(r.dir, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", path, err)
	}
	return string(data), nil
}

// StageAll stages all changes in the working directory.
func (r *Repo) StageAll() error {
	if err := r.worktree.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return fmt.Errorf("staging all changes: %w", err)
	}
	return nil
}

// Commit creates a commit with the given message.
func (r *Repo) Commit(message string) error {
	_, err := r.worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "DeveloperAgent",
			Email: "agent@devqaagent.local",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("committing: %w", err)
	}
	return nil
}

// ListFiles lists all files in the given directory relative to the repo root.
func (r *Repo) ListFiles(dir string) ([]string, error) {
	fullPath := filepath.Join(r.dir, dir)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("listing files in %s: %w", dir, err)
	}

	var files []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		files = append(files, name)
	}

	return files, nil
}
