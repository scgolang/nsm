// Package nsm implements the non session manager OSC protocol.
package nsm

import (
	"io"
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
	ClientOpen     = "/nsm/client/open"
	ClientSave     = "/nsm/client/save"
	Reply          = "/reply"
	ServerAnnounce = "/nsm/server/announce"
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

	log io.Writer

	// TODO: use types specific to the address being handled
	openChan  chan *osc.Message
	replyChan chan *osc.Message
}

// NewClient creates a new nsm-enabled application.
// If config.Session is nil then ErrNilSession will be returned.
func NewClient(config ClientConfig) (*Client, error) {
	return NewClientG(config, nil)
}

// NewClientG creates a new nsm-enabled application whose goroutines
// are part of the provided errgroup.Group.
// If config.Session is nil then ErrNilSession will be returned.
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

// serveOSC listens for incoming messages from Non Session Manager.
func (c *Client) serveOSC() error {
	return c.Serve(c.dispatcher())
}

// dispatcher returns the osc Dispatcher for the nsm client.
func (c *Client) dispatcher() osc.Dispatcher {
	return osc.Dispatcher{
		Reply: func(msg *osc.Message) error {
			c.replyChan <- msg
			return nil
		},
		ClientOpen: func(msg *osc.Message) error {
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
