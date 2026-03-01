package ghub

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"
)

// CreateBranch creates a new branch from the given reference.
func (c *GitHubClient) CreateBranch(ctx context.Context, name string, fromRef string) error {
	// Get the SHA of the source reference.
	ref, _, err := c.client.Git.GetRef(ctx, c.owner, c.repo, "refs/heads/"+fromRef)
	if err != nil {
		return fmt.Errorf("getting ref %q: %w", fromRef, err)
	}

	// Create the new branch reference.
	newRef := &github.Reference{
		Ref:    github.Ptr("refs/heads/" + name),
		Object: ref.Object,
	}

	_, _, err = c.client.Git.CreateRef(ctx, c.owner, c.repo, newRef)
	if err != nil {
		return fmt.Errorf("creating branch %q: %w", name, err)
	}

	return nil
}
