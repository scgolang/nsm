package nsm

import (
	"os"
	"testing"
	"time"

	"github.com/scgolang/osc"
)

func testConfig() ClientConfig {
	return ClientConfig{
		Name:                 "test_client",
		Capabilities:         Capabilities{"switch", "progress"},
		Major:                1,
		Minor:                2,
		PID:                  os.Getpid(),
		Session:              &mockSession{},
		WaitForAnnounceReply: true,
	}
}

func TestNewClient(t *testing.T) {
	if _, err := NewClient(ClientConfig{Session: nil}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClientAnnounce(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	if err := os.Unsetenv(NsmURL); err != nil {
		t.Fatalf("unset NSM_URL %s", err)
	}

	c, err := NewClient(ClientConfig{
		Name:         "test_client",
		Capabilities: Capabilities{"switch", "progress"},
		Major:        1,
		Minor:        2,
		PID:          os.Getpid(),
		Session:      &mockSession{},
		NsmURL:       nsmd.LocalAddr().String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }() // Best effort.
}

func TestClientAnnounceTimeout(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr:    "127.0.0.1:0",
		announcePause: 100 * time.Millisecond,
	})
	defer func() { _ = nsmd.Close() }() // Best effort.

	_, err := NewClient(ClientConfig{
		Name:                 "test_client",
		Capabilities:         Capabilities{"switch", "progress"},
		Major:                1,
		Minor:                2,
		PID:                  os.Getpid(),
		Session:              &mockSession{},
		Timeout:              10 * time.Millisecond,
		WaitForAnnounceReply: true,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if expected, got := `initialize client: announce app: timeout`, err.Error(); expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestClientNoNsmUrl(t *testing.T) {
	if err := os.Unsetenv(NsmURL); err != nil {
		t.Fatal(err)
	}
	if _, err := NewClient(testConfig()); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: dial udp: No `+NsmURL+` environment variable`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientGarbageNsmUrl(t *testing.T) {
	if err := os.Setenv(NsmURL, "garbage"); err != nil {
		t.Fatal(err)
	}
	if _, err := NewClient(testConfig()); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: dial udp: resolve udp remote address: missing port in address garbage`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientGarbageListenAddr(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr:    "127.0.0.1:0",
		announcePause: 100 * time.Millisecond,
	})
	defer func() { _ = nsmd.Close() }() // Best effort.

	config := testConfig()
	config.ListenAddr = "garbage"
	if _, err := NewClient(config); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: dial udp: resolve udp listening address: missing port in address garbage`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientReplyNoArguments(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr:    "127.0.0.1:0",
		announceReply: osc.Message{Address: AddressReply},
	})
	defer func() { _ = nsmd.Close() }() // Best effort.

	if _, err := NewClient(testConfig()); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: announce app: handle announce reply: expected 4 arguments in announce reply, got 0`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientReplyFirstArgumentWrongAddress(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr: "127.0.0.1:0",
		announceReply: osc.Message{
			Address: AddressReply,
			Arguments: osc.Arguments{
				osc.String("/foo/bar"),
			},
		},
	})
	defer func() { _ = nsmd.Close() }() // Best effort.

	if _, err := NewClient(testConfig()); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: announce app: handle announce reply: expected 4 arguments in announce reply, got 1`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}
