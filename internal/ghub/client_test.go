package ghub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-token", "testowner", "testrepo")
	require.NotNil(t, client)
	assert.Equal(t, "testowner", client.owner)
	assert.Equal(t, "testrepo", client.repo)
	assert.NotNil(t, client.client)
}

func TestNewClient_EmptyToken(t *testing.T) {
	// Should still create client (auth will fail on API calls, not construction).
	client := NewClient("", "owner", "repo")
	require.NotNil(t, client)
}

func TestPROptions(t *testing.T) {
	opts := PROptions{
		Title: "Test PR",
		Body:  "Test body",
		Head:  "feature-branch",
		Base:  "main",
	}
	assert.Equal(t, "Test PR", opts.Title)
	assert.Equal(t, "main", opts.Base)
}
