package nsm

import (
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	if _, err := NewClient(ClientConfig{Session: nil}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClientAnnounce(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	_ = newMockNsmd(t, "127.0.0.1:0")

	c, err := NewClient(ClientConfig{
		Name:         "test_client",
		Capabilities: Capabilities{"switch", "progress"},
		Major:        1,
		Minor:        2,
		PID:          os.Getpid(),
		Session:      &mockSession{},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
}
