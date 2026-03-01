package ghub

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"
)

// CreateComment creates a comment on an issue or pull request.
func (c *GitHubClient) CreateComment(ctx context.Context, number int, body string) error {
	_, _, err := c.client.Issues.CreateComment(ctx, c.owner, c.repo, number, &github.IssueComment{
		Body: github.Ptr(body),
	})
	if err != nil {
		return fmt.Errorf("creating comment on #%d: %w", number, err)
	}
	return nil
}

// ListComments returns all comments on an issue or pull request.
func (c *GitHubClient) ListComments(ctx context.Context, number int) ([]*github.IssueComment, error) {
	comments, _, err := c.client.Issues.ListComments(ctx, c.owner, c.repo, number, &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("listing comments on #%d: %w", number, err)
	}
	return comments, nil
}
