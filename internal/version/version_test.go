package version

import "testing"

func TestDefaults(t *testing.T) {
	if Version != "dev" {
		t.Errorf("expected default Version to be 'dev', got %q", Version)
	}
	if Commit != "unknown" {
		t.Errorf("expected default Commit to be 'unknown', got %q", Commit)
	}
	if BuildDate != "unknown" {
		t.Errorf("expected default BuildDate to be 'unknown', got %q", BuildDate)
	}
}
