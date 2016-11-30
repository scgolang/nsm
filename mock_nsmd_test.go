package nsm

import (
	"net"
	"os"
	"testing"

	"github.com/scgolang/osc"
	"golang.org/x/sync/errgroup"
)

// mockNsmd mocks an nsmd server.
type mockNsmd struct {
	osc.Conn
	errgroup.Group

	t            *testing.T
	announceChan chan osc.Message
}

// newMockNsmd creates a new mock nsmd server.
// This has the side effect of setting the NSM_URL to the listening
// address of the mock server.
func newMockNsmd(t *testing.T, listen string) *mockNsmd {
	nsmd := &mockNsmd{
		t:            t,
		announceChan: make(chan osc.Message),
	}
	nsmd.initialize(listen)
	return nsmd
}

func (m *mockNsmd) initialize(listen string) {
	laddr, err := net.ResolveUDPAddr("udp", listen)
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
			// Send reply.
			return m.SendTo(msg.Sender, osc.Message{
				Address: AddressReply,
				Sender:  m.LocalAddr(),
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
	return m.Conn.Close()
}
