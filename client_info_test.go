package nsm

import (
	"context"
	"testing"
	"time"
)

func TestClientDirty(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	var (
		config    = testConfig()
		dirtyChan = make(chan bool)
	)
	config.Session = &mockSession{dirtyChan: dirtyChan}

	if _, err := NewClient(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case dirtyChan <- false:
	}
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case isDirty := <-nsmd.dirtyChan:
		if expected, got := isDirty, false; expected != got {
			t.Fatalf("expected %t, got %t", expected, got)
		}
	}
}

func TestClientDirtyFailSend(t *testing.T) {
	var (
		config    = testConfig()
		dirtyChan = make(chan bool)
	)
	config.NsmURL = "osc.udp://127.0.0.1:55555"
	config.Session = &mockSession{dirtyChan: dirtyChan}
	config.WaitForAnnounceReply = false

	c := clientFailSend(context.Background(), t, config, 1)

	if err := c.Initialize(); err != nil {
		t.Fatal(err)
	}

	// Setup a channel for the error from c.Wait
	errChan := make(chan error)
	go func() {
		errChan <- c.Wait()
	}()

	// Signal that the client has unsaved changes.
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case dirtyChan <- true:
	}

	// Receive an error on the c.Wait channel.
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case err := <-errChan:
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if expected, got := `send dirty message: oops`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientGuiShowing(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	var (
		config         = testConfig()
		showingGuiChan = make(chan bool)
	)
	config.Session = &mockSession{showingGuiChan: showingGuiChan}

	if _, err := NewClient(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case showingGuiChan <- true:
	}
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case guiIsShowing := <-nsmd.guiShowingChan:
		if expected, got := guiIsShowing, true; expected != got {
			t.Fatalf("expected %t, got %t", expected, got)
		}
	}
}

func TestClientGuiShowingFailSend(t *testing.T) {
	var (
		config         = testConfig()
		showingGuiChan = make(chan bool)
	)
	config.NsmURL = "osc.udp://127.0.0.1:55555"
	config.Session = &mockSession{showingGuiChan: showingGuiChan}
	config.WaitForAnnounceReply = false

	c := clientFailSend(context.Background(), t, config, 1)

	if err := c.Initialize(); err != nil {
		t.Fatal(err)
	}

	// Setup a channel for the error from c.Wait
	errChan := make(chan error)
	go func() {
		errChan <- c.Wait()
	}()

	// Signal that the client has unsaved changes.
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case showingGuiChan <- false:
	}

	// Receive an error on the c.Wait channel.
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case err := <-errChan:
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if expected, got := `send gui-showing message: oops`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}
