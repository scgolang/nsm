package nsm

import (
	"testing"

	"github.com/scgolang/osc"
)

func TestClientOpenReplyMissingArguments(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	config := testConfig()
	config.Session = &mockSession{
		open: mockReply{Message: "session started"},
	}
	c, err := NewClient(config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }() // Best effort.

	reply := nsmd.OpenSession(osc.Message{
		Address: AddressClientOpen,
		Arguments: osc.Arguments{
			osc.String("./test-projects"),
			osc.String("display_name"),
			osc.String("client_id"),
		},
	})
	if expected, got := 2, len(reply.Arguments); expected != got {
		t.Fatalf("expected %d arguments, got %d", expected, got)
	}
	replyAddr, err := reply.Arguments[0].ReadString()
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := AddressClientOpen, replyAddr; expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
	replyMessage, err := reply.Arguments[1].ReadString()
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := `session started`, replyMessage; expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestClientOpenReplyError(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	config := testConfig()
	config.Session = &mockSession{
		open: mockReply{
			Err: NewError(ErrCreateFailed, "could not create new session"),
		},
	}
	c, err := NewClient(config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }() // Best effort.

	nsmErr := nsmd.OpenSessionError(osc.Message{
		Address: AddressClientOpen,
		Arguments: osc.Arguments{
			osc.String("./test-projects"),
			osc.String("display_name"),
			osc.String("client_id"),
		},
	})
	if nsmErr == nil {
		t.Fatal("expected error, got nil")
	}
	if expected, got := ErrCreateFailed, nsmErr.Code(); expected != got {
		t.Fatalf("expected %d, got %d", expected, got)
	}
	if expected, got := `could not create new session`, nsmErr.Error(); expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}
