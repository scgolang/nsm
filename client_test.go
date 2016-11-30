package nsm

import (
	"os"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	if _, err := NewClient(ClientConfig{Session: nil}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClientAnnounce(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	_ = newMockNsmd(t, mockNsmdConfig{
		listenAddr: "127.0.0.1:0",
	})

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
	defer func() { _ = c.Close() }() // Best effort.
}

func TestClientAnnounceTimeout(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	_ = newMockNsmd(t, mockNsmdConfig{
		listenAddr:    "127.0.0.1:0",
		announcePause: 100 * time.Millisecond,
	})

	_, err := NewClient(ClientConfig{
		Name:         "test_client",
		Capabilities: Capabilities{"switch", "progress"},
		Major:        1,
		Minor:        2,
		PID:          os.Getpid(),
		Session:      &mockSession{},
		Timeout:      10 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if expected, got := `initialize client: announce app: waiting for announce reply: timeout`, err.Error(); expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestClientNoNsmUrl(t *testing.T) {
}
