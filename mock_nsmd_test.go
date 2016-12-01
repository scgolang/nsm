package nsm

import (
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/osc"
	"golang.org/x/sync/errgroup"
)

type mockNsmdConfig struct {
	listenAddr    string
	announcePause time.Duration
}

// mockNsmd mocks an nsmd server.
type mockNsmd struct {
	mockNsmdConfig
	osc.Conn
	errgroup.Group

	t *testing.T

	announceChan chan osc.Message
}

// newMockNsmd creates a new mock nsmd server.
// This has the side effect of setting the NSM_URL to the listening
// address of the mock server.
func newMockNsmd(t *testing.T, config mockNsmdConfig) *mockNsmd {
	nsmd := &mockNsmd{
		mockNsmdConfig: config,
		t:              t,
		announceChan:   make(chan osc.Message),
	}
	nsmd.initialize()
	return nsmd
}

func (m *mockNsmd) initialize() {
	if m.announcePause == time.Duration(0) {
		m.announcePause = 10 * time.Millisecond
	}
	laddr, err := net.ResolveUDPAddr("udp", m.listenAddr)
	if err != nil {
		m.t.Fatalf("could not resolve udp address: %s", err)
	}
	conn, err := osc.ListenUDP("udp", laddr)
	if err != nil {
		m.t.Fatalf("could not create udp connection: %s", err)
	}
	if err := os.Setenv(NsmURL, conn.LocalAddr().String()); err != nil {
		m.t.Fatalf("could not set %s environment variable: %s", NsmURL, err)
	}
	m.Conn = conn

	m.Go(m.startOSC)
}

func (m *mockNsmd) startOSC() error {
	return m.Serve(m.dispatcher())
}

func (m *mockNsmd) dispatcher() osc.Dispatcher {
	return osc.Dispatcher{
		AddressServerAnnounce: func(msg osc.Message) error {
			time.Sleep(m.announcePause)

			// Send reply.
			return m.SendTo(msg.Sender, osc.Message{
				Address: AddressReply,
				Arguments: osc.Arguments{
					osc.String(AddressServerAnnounce),
					osc.String("session started"),
					osc.String("mock_nsmd"),
					osc.String(Capabilities{CapServerBroadcast, CapServerControl}.String()),
				},
			})
		},
	}
}

func (m *mockNsmd) Close() error {
	close(m.announceChan)
	errs := []string{}
	if err := os.Unsetenv(NsmURL); err != nil {
		errs = append(errs, err.Error())
	}
	if err := m.Conn.Close(); err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, " and "))
}
