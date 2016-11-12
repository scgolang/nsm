package nsm

import (
	"net"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/osc"
	"golang.org/x/sync/errgroup"
)

// OSC addresses.
const (
	AddressClientOpen     = "/nsm/client/open"
	AddressClientSave     = "/nsm/client/save"
	AddressError          = "/error"
	AddressReply          = "/reply"
	AddressServerAnnounce = "/nsm/server/announce"
)

// NsmURL is the name of the NSM url environment variable.
var NsmURL = "NSM_URL"

// Common errors.
var (
	ErrNilSession = errors.New("Session must be provided")
	ErrNoNsmURL   = errors.New("No " + NsmURL + " environment variable")
	ErrTimeout    = errors.New("timeout")
)

// ClientConfig represents the configuration of an nsm client.
type ClientConfig struct {
	Name         string
	Capabilities Capabilities
	Major        int32
	Minor        int32
	PID          int
	Timeout      time.Duration
	Session      Session
}

// Client represents an nsm client.
type Client struct {
	ClientConfig
	osc.Conn
	*errgroup.Group

	// TODO: use types specific to the address being handled
	openChan  chan *osc.Message
	replyChan chan *osc.Message
}

// NewClient creates a new nsm-enabled application.
// If config.Session is nil then ErrNilSession will be returned.
// If NSM_URL is not defined in the environment then ErrNoNsmURL will be returned.
func NewClient(config ClientConfig) (*Client, error) {
	return NewClientG(config, nil)
}

// NewClientG creates a new nsm-enabled application whose goroutines
// are part of the provided errgroup.Group.
// If config.Session is nil then ErrNilSession will be returned.
// If NSM_URL is not defined in the environment then ErrNoNsmURL will be returned.
func NewClientG(config ClientConfig, g *errgroup.Group) (*Client, error) {
	if config.Session == nil {
		return nil, ErrNilSession
	}
	// Create the client.
	c := &Client{
		ClientConfig: config,

		openChan:  make(chan *osc.Message),
		replyChan: make(chan *osc.Message),
	}

	// Get connection.
	conn, err := c.dialUDP("0.0.0.0:0")
	if err != nil {
		return nil, errors.Wrap(err, "could not dial udp")
	}
	c.Conn = conn

	// Start the OSC server.
	if g != nil {
		c.Group = g
	} else {
		var gg errgroup.Group
		c.Group = &gg
	}
	c.Go(c.serveOSC)

	// Announce app.
	if err := c.announce(); err != nil {
		return nil, errors.Wrap(err, "could not announce app")
	}

	return c, nil
}

// handle handles the return values from a Session's method.
// The method must be associated with the provided address,
// e.g. ClientOpen should be passed after calling a Session's
// Open method.
func (c *Client) handle(address, message string, err Error) error {
	if err != nil {
		return c.handleError(address, err)
	}
	return c.handleReply(address, message)
}

// handleError handles the reply for a successful client operation.
func (c *Client) handleError(address string, err Error) error {
	msg, errr := osc.NewMessage(AddressError)
	if errr != nil {
		return errors.Wrap(errr, "could not create error reply for "+address)
	}
	if errr := msg.WriteInt32(int32(err.Code())); err != nil {
		return errors.Wrap(errr, "could not add address to error reply")
	}
	if errr := msg.WriteString(err.Error()); err != nil {
		return errors.Wrap(errr, "could not add message to error reply")
	}
	return errors.Wrap(c.Send(msg), "could not send error")
}

// handleReply handles the reply for a successful client operation.
func (c *Client) handleReply(address, message string) error {
	msg, err := osc.NewMessage(AddressReply)
	if err != nil {
		return errors.Wrap(err, "could not create reply for "+address)
	}
	if err := msg.WriteString(address); err != nil {
		return errors.Wrap(err, "could not add address to reply")
	}
	if err := msg.WriteString(message); err != nil {
		return errors.Wrap(err, "could not add message to reply")
	}
	return errors.Wrap(c.Send(msg), "could not send reply")
}

// serveOSC listens for incoming messages from Non Session Manager.
func (c *Client) serveOSC() error {
	return c.Serve(c.dispatcher())
}

// dispatcher returns the osc Dispatcher for the nsm client.
func (c *Client) dispatcher() osc.Dispatcher {
	return osc.Dispatcher{
		AddressReply: func(msg *osc.Message) error {
			c.replyChan <- msg
			return nil
		},
		AddressClientOpen: func(msg *osc.Message) error {
			c.openChan <- msg
			return nil
		},
	}
}

// Close closes the nsm client.
func (c *Client) Close() error {
	close(c.replyChan)
	return c.Conn.Close()
}

// dialUDP initializes the connection to non session manager.
func (c *Client) dialUDP(localAddr string) (osc.Conn, error) {
	nsmURL, ok := os.LookupEnv(NsmURL)
	if !ok {
		return nil, ErrNoNsmURL
	}

	// Why?
	nsmURL = strings.TrimPrefix(nsmURL, "osc.udp://")
	nsmURL = strings.TrimSuffix(nsmURL, "/")

	raddr, err := net.ResolveUDPAddr("udp", nsmURL)
	if err != nil {
		return nil, errors.Wrap(err, "could not resolve udp remote address")
	}
	laddr, err := net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		return nil, errors.Wrap(err, "could not resolve udp listening address")
	}
	return osc.DialUDP("udp", laddr, raddr)
}
