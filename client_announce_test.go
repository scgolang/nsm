package nsm

import (
	"net"
	"testing"

	"github.com/scgolang/osc"
)

func TestClientAnnounceReplyMissingArguments(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	// This mock also sends an invalid announce reply (missing arguments)
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr: "127.0.0.1:0",
		announceReply: osc.Message{
			Address: AddressReply,
			Arguments: osc.Arguments{
				osc.String(AddressServerAnnounce),
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

func TestClientAnnounceReplyFirstArgumentWrongType(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	// This mock also sends an invalid announce reply (missing arguments)
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr: "127.0.0.1:0",
		announceReply: osc.Message{
			Address: AddressReply,
			Arguments: osc.Arguments{
				osc.Int(3),
				osc.Int(3),
				osc.Int(3),
				osc.Int(3),
			},
		},
	})
	defer func() { _ = nsmd.Close() }() // Best effort.

	if _, err := NewClient(testConfig()); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: announce app: waiting for announce reply: first argument of reply message should be a string: invalid type tag`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientAnnounceSendError(t *testing.T) {
	raddr, err := net.ResolveUDPAddr("udp", "example.com:8000")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := osc.DialUDP("udp", nil, raddr)
	if err != nil {
		t.Fatal(err)
	}
	c := &Client{
		Conn:         mockSendErr{Conn: conn},
		ClientConfig: testConfig(),
	}
	if err := c.Announce(); err == nil {
		t.Fatal("expected error, got nil")
	}
}
