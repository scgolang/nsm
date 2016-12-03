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
	AddressServerAbort           = "/nsm/server/abort"
	AddressServerAdd             = "/nsm/server/add"
	AddressServerAnnounce        = "/nsm/server/announce"
	AddressServerClose           = "/nsm/server/close"
	AddressServerDuplicate       = "/nsm/server/duplicate"
	AddressServerList            = "/nsm/server/list"
	AddressServerNew             = "/nsm/server/new"
	AddressServerOpen            = "/nsm/server/open"
	AddressServerQuit            = "/nsm/server/quit"
	AddressServerSave            = "/nsm/server/save"
)

// DoneString is a string that is returned for replies (e.g. /nsm/server/list)
// that return more than one reply, to signal that the replies have ended.
const DoneString = "NSM_DONE"

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

	// Timeout is an amount of time we should wait for a response from the nsm server.
	Timeout time.Duration

	Session              Session
	ListenAddr           string
	DialNetwork          string
	NsmURL               string
	WaitForAnnounceReply bool
}

// Client represents an nsm client.
type Client struct {
	ClientConfig
	osc.Conn
	errgroup.Group

	ReplyChan  chan osc.Message
	closedChan chan struct{}
}

// NewClient creates a new nsm-enabled application.
// If config.Session is nil then ErrNilSession will be returned.
// If NSM_URL is not defined in the environment then ErrNoNsmURL will be returned.
// TODO: validate config?
func NewClient(config ClientConfig) (*Client, error) {
	if config.Session == nil {
		return nil, ErrNilSession
	}
	if config.Timeout == time.Duration(0) {
		config.Timeout = DefaultTimeout
	}
	// Create the client.
	c := &Client{
		ClientConfig: config,
		ReplyChan:    make(chan osc.Message),
		closedChan:   make(chan struct{}),
	}
	c.Defaults()

	if err := c.Initialize(); err != nil {
		return nil, errors.Wrap(err, "initialize client")
	}
	return c, nil
}

// Initialize initializes the client.
func (c *Client) Initialize() error {
	if c.closedChan == nil {
		c.closedChan = make(chan struct{})
	}
	if c.ReplyChan == nil {
		c.ReplyChan = make(chan osc.Message)
	}
	// Get connection.
	if err := c.DialUDP(c.ListenAddr); err != nil {
		return errors.Wrap(err, "dial udp")
	}

	// Start the OSC server.
	c.StartOSC()

	// Announce client.
	if err := c.Announce(); err != nil {
		_ = c.Close() // Best effort.
		return errors.Wrap(err, "announce app")
	}
	return nil
}

// Defaults sets default config values for the client.
func (c *Client) Defaults() {
	if c.ListenAddr == "" {
		c.ListenAddr = "0.0.0.0:0"
	}
	if c.DialNetwork == "" {
		c.DialNetwork = "udp"
	}
}

// DialUDP initializes the connection to non session manager.
func (c *Client) DialUDP(localAddr string) error {
	// Tests will sometimes initialize a conn without this method that does naughty things.
	if c.Conn != nil {
		return nil
	}
	// Allow client configuration to override the env var.
	var nsmURL string
	if c.NsmURL == "" {
		var ok bool
		// Look up NSM_URL environment variable.
		nsmURL, ok = os.LookupEnv(NsmURL)
		if !ok {
			return ErrNoNsmURL
		}
	} else {
		nsmURL = c.NsmURL
	}

	// Why?
	nsmURL = strings.TrimPrefix(nsmURL, "osc.udp://")
	nsmURL = strings.TrimSuffix(nsmURL, "/")

	// Get OSC connection.
	raddr, err := net.ResolveUDPAddr(c.DialNetwork, nsmURL)
	if err != nil {
		return errors.Wrap(err, "resolve udp remote address")
	}
	laddr, err := net.ResolveUDPAddr(c.DialNetwork, localAddr)
	if err != nil {
		return errors.Wrap(err, "resolve udp listening address")
	}
	conn, _ := osc.DialUDP(c.DialNetwork, laddr, raddr) // Never fails
	c.Conn = conn

	return nil
}

// StartOSC starts the osc server.
func (c *Client) StartOSC() {
	c.Go(c.serveOSC)
	c.Go(c.handleClientInfo)
}

// Close closes the nsm client.
func (c *Client) Close() error {
	close(c.ReplyChan)
	close(c.closedChan)
	return c.Conn.Close()
}

// serveOSC listens for incoming messages from Non Session Manager.
func (c *Client) serveOSC() error {
	return c.Serve(c.dispatcher())
}

// dispatcher returns the osc Dispatcher for the nsm client.
func (c *Client) dispatcher() osc.Dispatcher {
	d := osc.Dispatcher{
		AddressReply: func(msg osc.Message) error {
			c.ReplyChan <- msg
			return nil
		},
		AddressClientOpen: func(msg osc.Message) error {
			return c.handleOpen(msg)
		},
		AddressClientSave: func(msg osc.Message) error {
			response, nsmerr := c.Session.Save()
			return c.handle(AddressClientSave, response, nsmerr)
		},
		AddressClientSessionIsLoaded: func(msg osc.Message) error {
			return c.Session.IsLoaded()
		},
		AddressClientShowOptionalGUI: func(msg osc.Message) error {
			return c.Session.ShowGUI(true)
		},
		AddressClientHideOptionalGUI: func(msg osc.Message) error {
			return c.Session.ShowGUI(false)
		},
	}
	for address, method := range c.Session.Methods() {
		d[address] = method
	}
	return d
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
	msg := osc.Message{
		Address: AddressError,
		Arguments: osc.Arguments{
			osc.String(address),
			osc.Int(int32(err.Code())),
			osc.String(err.Error()),
		},
	}
	return errors.Wrap(c.Send(msg), "send error")
}

// handleReply handles the reply for a successful client operation.
func (c *Client) handleReply(address, message string) error {
	msg := osc.Message{
		Address: AddressReply,
		Arguments: osc.Arguments{
			osc.String(address),
			osc.String(message),
		},
	}
	return errors.Wrap(c.Send(msg), "send reply")
}
