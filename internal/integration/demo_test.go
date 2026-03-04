package integration

import (
	"testing"
)

// TestDemonstrateRepoSpecificPaths verifies that the demo functionality works
func TestDemonstrateRepoSpecificPaths(t *testing.T) {
	err := DemonstrateRepoSpecificPaths()
	if err != nil {
		t.Fatalf("Demo failed: %v", err)
	}
}