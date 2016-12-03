package nsm

import (
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

	if _, err := NewClient(config); err != nil {
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
	t.SkipNow()

	var (
		config    = testConfig()
		dirtyChan = make(chan bool)
	)
	config.Session = &mockSession{dirtyChan: dirtyChan}
	config.WaitForAnnounceReply = false

	c := clientFailSend(t, config, 1)

	if err := c.Initialize(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case dirtyChan <- false:
	}
	println("waiting...")
	if err := c.Wait(); err == nil {
		t.Fatal("expected error, got nil")
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

	if _, err := NewClient(config); err != nil {
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
