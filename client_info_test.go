package nsm

import (
	"testing"
)

func TestClientDirty(t *testing.T) {
	t.SkipNow()

	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	if _, err := NewClient(testConfig()); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: announce app: handle announce reply: read session manager capabilities: invalid type tag`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}
