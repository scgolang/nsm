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
	AddressClientOpen            = "/nsm/client/open"
	AddressClientSave            = "/nsm/client/save"
	AddressClientSessionIsLoaded = "/nsm/client/session_is_loaded"
	AddressClientShowOptionalGUI = "/nsm/client/show_optional_gui"
	AddressClientHideOptionalGUI = "/nsm/client/hide_optional_gui"
	AddressClientIsDirty         = "/nsm/client/is_dirty"
	AddressClientIsClean         = "/nsm/client/is_clean"
	AddressClientGUIHidden       = "/nsm/client/gui_is_hidden"
	AddressClientGUIShowing      = "/nsm/client/gui_is_showing"
	AddressClientProgress        = "/nsm/client/progress"
	AddressClientStatus          = "/nsm/client/message"
	AddressError                 = "/error"
	AddressReply                 = "/reply"
	AddressServerAnnounce        = "/nsm/server/announce"
)

// NsmURL is the name of the NSM url environment variable.
var NsmURL = "NSM_URL"

// DefaultTimeout is the default timeout for waiting for
// a reply from Non Session Manager.
var DefaultTimeout = 5 * time.Second

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

	// TODO: need the ability to requeue replies.
	ReplyChan chan *osc.Message
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
	if config.Timeout == time.Duration(0) {
		config.Timeout = DefaultTimeout
	}
	// Create the client.
	c := &Client{
		ClientConfig: config,
		ReplyChan:    make(chan *osc.Message),
	}
	if err := c.initialize(g); err != nil {
		return nil, errors.Wrap(err, "could not initialize client")
	}
	return c, nil
}

// initialize initializes the client.
func (c *Client) initialize(g *errgroup.Group) error {
	// Get connection.
	if err := c.dialUDP("0.0.0.0:0"); err != nil {
		return errors.Wrap(err, "could not dial udp")
	}

	// Start the OSC server.
	c.startOSC(g)

	// Announce client.
	if err := c.announce(); err != nil {
		return errors.Wrap(err, "could not announce app")
	}

	return nil
}

// dialUDP initializes the connection to non session manager.
func (c *Client) dialUDP(localAddr string) error {
	// Look up NSM_URL environment variable.
	nsmURL, ok := os.LookupEnv(NsmURL)
	if !ok {
		return ErrNoNsmURL
	}

	// Why?
	nsmURL = strings.TrimPrefix(nsmURL, "osc.udp://")
	nsmURL = strings.TrimSuffix(nsmURL, "/")

	// Get OSC connection.
	raddr, err := net.ResolveUDPAddr("udp", nsmURL)
	if err != nil {
		return errors.Wrap(err, "could not resolve udp remote address")
	}
	laddr, err := net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		return errors.Wrap(err, "could not resolve udp listening address")
	}
	conn, err := osc.DialUDP("udp", laddr, raddr)
	if err != nil {
		return err
	}
	c.Conn = conn

	return nil
}

// startOSC starts the osc server.
func (c *Client) startOSC(g *errgroup.Group) {
	if g != nil {
		c.Group = g
	} else {
		c.Group = &errgroup.Group{}
	}
	c.Go(c.serveOSC)
	c.Go(c.handleClientInfo)
}

// wait waits for a reply to a message that was sent to address.
func (c *Client) wait(address string) error {
	timeout := time.After(c.Timeout)
	select {
	case <-timeout:
		return ErrTimeout
	case msg := <-c.ReplyChan:
		if msg.Address() != address {
			// TODO: requeue message
		}
		switch address {
		case AddressClientOpen:
			if err := c.handleOpen(msg); err != nil {
				return errors.Wrap(err, "could not handle open reply")
			}
		case AddressServerAnnounce:
			if err := c.handleAnnounceReply(msg); err != nil {
				return errors.Wrap(err, "could not handle announce reply")
			}
		}
	}
	return nil
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
	d := osc.Dispatcher{
		AddressReply: func(msg *osc.Message) error {
			c.ReplyChan <- msg
			return nil
		},
		AddressClientOpen: func(msg *osc.Message) error {
			return c.handleOpen(msg)
		},
		AddressClientSave: func(msg *osc.Message) error {
			response, nsmerr := c.Session.Save()
			return c.handle(AddressClientSave, response, nsmerr)
		},
		AddressClientSessionIsLoaded: func(msg *osc.Message) error {
			return c.Session.IsLoaded()
		},
		AddressClientShowOptionalGUI: func(msg *osc.Message) error {
			return c.Session.ShowGUI(true)
		},
		AddressClientHideOptionalGUI: func(msg *osc.Message) error {
			return c.Session.ShowGUI(false)
		},
	}
	for address, method := range c.Session.Methods() {
		d[address] = method
	}
	return d
}

// Close closes the nsm client.
func (c *Client) Close() error {
	close(c.ReplyChan)
	return c.Conn.Close()
}
