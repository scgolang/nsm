package nsm

import (
	"context"
	"testing"
	"time"

	"github.com/scgolang/osc"
)

func TestClientOpen(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	config := testConfig()
	config.Session = &mockSession{
		open: mockReply{Message: "session started"},
	}
	c := newClient(t, config)
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

func TestClientOpenNoArguments(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	config := testConfig()
	config.Session = &mockSession{
		open: mockReply{Message: "session started"},
	}
	c := newClient(t, config)
	defer func() { _ = c.Close() }() // Best effort.

	if err := nsmd.SendTo(c.LocalAddr(), osc.Message{Address: AddressClientOpen}); err != nil {
		t.Fatal(err)
	}
	errChan := make(chan error)
	go func() {
		errChan <- c.Wait()
	}()

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case err := <-errChan:
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	}
}

func TestClientOpenFirstArgumentWrongType(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	config := testConfig()
	config.Session = &mockSession{
		open: mockReply{Message: "session started"},
	}
	c := newClient(t, config)
	defer func() { _ = c.Close() }() // Best effort.

	if err := nsmd.SendTo(c.LocalAddr(), osc.Message{
		Address: AddressClientOpen,
		Arguments: osc.Arguments{
			osc.Int(10),
			osc.String("foo"),
			osc.String("bar"),
		},
	}); err != nil {
		t.Fatal(err)
	}
	errChan := make(chan error)
	go func() {
		errChan <- c.Wait()
	}()

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case err := <-errChan:
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	}
}

func TestClientOpenSecondArgumentWrongType(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	config := testConfig()
	config.Session = &mockSession{
		open: mockReply{Message: "session started"},
	}
	c := newClient(t, config)
	defer func() { _ = c.Close() }() // Best effort.

	if err := nsmd.SendTo(c.LocalAddr(), osc.Message{
		Address: AddressClientOpen,
		Arguments: osc.Arguments{
			osc.String("foo"),
			osc.Int(10),
			osc.String("bar"),
		},
	}); err != nil {
		t.Fatal(err)
	}
	errChan := make(chan error)
	go func() {
		errChan <- c.Wait()
	}()

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case err := <-errChan:
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	}
}

func TestClientOpenThirdArgumentWrongType(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	config := testConfig()
	config.Session = &mockSession{
		open: mockReply{Message: "session started"},
	}
	c := newClient(t, config)
	defer func() { _ = c.Close() }() // Best effort.

	if err := nsmd.SendTo(c.LocalAddr(), osc.Message{
		Address: AddressClientOpen,
		Arguments: osc.Arguments{
			osc.String("foo"),
			osc.String("bar"),
			osc.Int(10),
		},
	}); err != nil {
		t.Fatal(err)
	}
	errChan := make(chan error)
	go func() {
		errChan <- c.Wait()
	}()

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case err := <-errChan:
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
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
	c := newClient(t, config)
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

func TestClientOpenSendReplyError(t *testing.T) {
	t.SkipNow()

	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	config := testConfig()
	config.Session = &mockSession{
		open: mockReply{
			Err: NewError(ErrCreateFailed, "could not create new session"),
		},
	}
	config.WaitForAnnounceReply = false

	c := clientFailSend(context.Background(), t, config, 1)
	if err := c.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }() // Best effort.

	if err := nsmd.SendTo(c.LocalAddr(), osc.Message{
		Address: AddressClientOpen,
		Arguments: osc.Arguments{
			osc.String("./test-projects"),
			osc.String("display_name"),
			osc.String("client_id"),
		},
	}); err != nil {
		t.Fatal(err)
	}
	errChan := make(chan error)
	go func() {
		errChan <- c.Wait()
	}()

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case err := <-errChan:
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	}
}
