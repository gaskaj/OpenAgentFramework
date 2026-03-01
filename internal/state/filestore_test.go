package state

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileStore(dir)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	ws := &AgentWorkState{
		AgentType:   "developer",
		IssueNumber: 42,
		IssueTitle:  "Test Issue",
		State:       StateImplement,
		BranchName:  "agent/issue-42",
		UpdatedAt:   now,
		CreatedAt:   now,
	}

	err = store.Save(ctx, ws)
	require.NoError(t, err)

	loaded, err := store.Load(ctx, "developer")
	require.NoError(t, err)
	require.NotNil(t, loaded)

	assert.Equal(t, "developer", loaded.AgentType)
	assert.Equal(t, 42, loaded.IssueNumber)
	assert.Equal(t, "Test Issue", loaded.IssueTitle)
	assert.Equal(t, StateImplement, loaded.State)
	assert.Equal(t, "agent/issue-42", loaded.BranchName)
}

func TestFileStore_LoadNotFound(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileStore(dir)
	require.NoError(t, err)

	loaded, err := store.Load(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, loaded)
}

func TestFileStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileStore(dir)
	require.NoError(t, err)

	ctx := context.Background()

	ws := &AgentWorkState{
		AgentType: "developer",
		State:     StateIdle,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	err = store.Save(ctx, ws)
	require.NoError(t, err)

	err = store.Delete(ctx, "developer")
	require.NoError(t, err)

	loaded, err := store.Load(ctx, "developer")
	require.NoError(t, err)
	assert.Nil(t, loaded)
}

func TestFileStore_DeleteNotFound(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileStore(dir)
	require.NoError(t, err)

	err = store.Delete(context.Background(), "nonexistent")
	require.NoError(t, err)
}

func TestFileStore_List(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileStore(dir)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now()

	states := []*AgentWorkState{
		{AgentType: "developer", State: StateImplement, UpdatedAt: now, CreatedAt: now},
		{AgentType: "qa", State: StateIdle, UpdatedAt: now, CreatedAt: now},
	}

	for _, s := range states {
		require.NoError(t, store.Save(ctx, s))
	}

	listed, err := store.List(ctx)
	require.NoError(t, err)
	assert.Len(t, listed, 2)
}

func TestFileStore_ListEmpty(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileStore(dir)
	require.NoError(t, err)

	listed, err := store.List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, listed)
}

func TestFileStore_Overwrite(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileStore(dir)
	require.NoError(t, err)

	ctx := context.Background()
	now := time.Now()

	ws := &AgentWorkState{
		AgentType: "developer",
		State:     StateIdle,
		UpdatedAt: now,
		CreatedAt: now,
	}

	require.NoError(t, store.Save(ctx, ws))

	ws.State = StateImplement
	ws.IssueNumber = 10
	require.NoError(t, store.Save(ctx, ws))

	loaded, err := store.Load(ctx, "developer")
	require.NoError(t, err)
	assert.Equal(t, StateImplement, loaded.State)
	assert.Equal(t, 10, loaded.IssueNumber)
}
