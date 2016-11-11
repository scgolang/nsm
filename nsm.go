package nsm

import (
	"net"
	"os"

	"github.com/pkg/errors"
	"github.com/scgolang/osc"
)

// OSC addresses.
const (
	AddressAnnounce = "/nsm/server/announce"
)

// NsmURL is the name of the NSM url environment variable.
var NsmURL = "NSM_URL"

// Common errors.
var (
	ErrNoNsmURL = errors.New("No " + NsmURL + " environment variable")
)

// NSM holds all the state for an nsm-enabled application.
type NSM struct {
	osc.Conn
}

// Announce announces a new nsm application.
func (n *NSM) Announce() error {
	msg, err := osc.NewMessage(AddressAnnounce)
	if err != nil {
		return errors.Wrap(err, "could not create osc message")
	}
	n.Send(msg)
	return nil
}

// New creates a new nsm-enabled application.
func New(appName string) (*NSM, error) {
	// Get connection.
	nsmURL, ok := os.LookupEnv(NsmURL)
	if !ok {
		return nil, ErrNoNsmURL
	}
	conn, err := dialUDP("0.0.0.0:0", nsmURL)
	if err != nil {
		return nil, errors.Wrap(err, "could not dial udp")
	}
	n := &NSM{Conn: conn}

	// Announce app.
	if err := n.Announce(); err != nil {
		return nil, errors.Wrap(err, "could not announce app")
	}

	return n, nil
}

func dialUDP(localAddr, nsmURL string) (osc.Conn, error) {
	laddr, err := net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		return nil, errors.Wrap(err, "could not resolve udp listening address")
	}
	raddr, err := net.ResolveUDPAddr("udp", nsmURL)
	if err != nil {
		return nil, errors.Wrap(err, "could not resolve udp remote address")
	}
	return osc.DialUDP("udp", laddr, raddr)
}
