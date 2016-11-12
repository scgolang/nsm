// Package nsm implements the non session manager OSC protocol.
package nsm

import (
	"io"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/osc"
	"golang.org/x/sync/errgroup"
)

// OSC addresses.
const (
	AddressAnnounce = "/nsm/server/announce"
)

// Capability is a capability of an nsm client.
type Capability string

// Capabilities.
const (
	CapSep                 = ":"
	CapSwitch   Capability = "switch"
	CapDirty    Capability = "dirty"
	CapProgress Capability = "progress"
	CapMessage  Capability = "message"
	CapGUI      Capability = "optional-gui"
)

// Capabilities represents a list of capabilities
type Capabilities []Capability

// String converts capabilities to a string.
func (caps Capabilities) String() string {
	if len(caps) == 0 {
		return ""
	}
	ss := make([]string, len(caps))
	for i, cap := range caps {
		ss[i] = string(cap)
	}
	return CapSep + strings.Join(ss, CapSep) + CapSep
}

// NsmURL is the name of the NSM url environment variable.
var NsmURL = "NSM_URL"

// Common errors.
var (
	ErrNoNsmURL = errors.New("No " + NsmURL + " environment variable")
	ErrTimeout  = errors.New("timeout")
)

// ClientConfig represents the configuration of an nsm client.
type ClientConfig struct {
	Name         string
	Capabilities Capabilities
	Major        int32
	Minor        int32
	PID          int
	Timeout      time.Duration
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

// Announce announces a new nsm application.
func (c *Client) Announce() error {
	// Send the announce message.
	msg, err := c.newAnnounceMsg()
	if err != nil {
		return errors.Wrap(err, "could not create announce message")
	}
	if err := c.Send(msg); err != nil {
		return errors.Wrap(err, "could not send announce message")
	}

	// Wait for the server's reply.
	if err := c.announceWait(); err != nil {
		return errors.Wrap(err, "waiting for announce reply")
	}

	return nil
}

// newAnnounceMsg creates a new announce message.
func (c *Client) newAnnounceMsg() (*osc.Message, error) {
	msg, err := osc.NewMessage(AddressAnnounce)
	if err != nil {
		return nil, errors.Wrap(err, "could not create osc message")
	}
	// Write name.
	if c.Name != "" {
		if err := msg.WriteString(c.Name); err != nil {
			return nil, errors.Wrap(err, "could not write name")
		}
	} else {
		if err := msg.WriteString(os.Args[0]); err != nil {
			return nil, errors.Wrap(err, "could not write name")
		}
	}
	// Write capabilities.
	if err := msg.WriteString(c.Capabilities.String()); err != nil {
		return nil, errors.Wrap(err, "could not write capabilities")
	}
	// Write executable.
	if err := msg.WriteString(os.Args[0]); err != nil {
		return nil, errors.Wrap(err, "could not write executable")
	}
	// Write version.
	if err := msg.WriteInt32(c.Major); err != nil {
		return nil, errors.Wrap(err, "could not write major")
	}
	if err := msg.WriteInt32(c.Minor); err != nil {
		return nil, errors.Wrap(err, "could not write minor")
	}
	// Write PID.
	if err := msg.WriteInt32(int32(c.PID)); err != nil {
		return nil, errors.Wrap(err, "could not write pid")
	}
	return msg, nil
}

// announceWait waits for a reply to the announce message.
func (c *Client) announceWait() error {
	var (
		timeout  = time.After(c.Timeout)
		received = 0
	)
	for {
		select {
		case <-timeout:
			return ErrTimeout
		case msg := <-c.openChan:
			io.WriteString(c.log, "received open\n")
			if err := c.handleOpen(msg); err != nil {
				return errors.Wrap(err, "could not handle open message")
			}
			received++
			if received == 2 {
				break
			}
		case msg := <-c.replyChan:
			io.WriteString(c.log, "received announce\n")
			c.handleAnnounceReply(msg)
			received++
			if received == 2 {
				break
			}
		}
	}
	return nil
}

// handleOpen handles the open message.
func (c *Client) handleOpen(msg *osc.Message) error {
	projectPath, err := msg.ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read project path")
	}
	if err := os.Mkdir(projectPath, 0755); err != nil {
		if !os.IsExist(err) {
			return errors.Wrap(err, "could not open project directory")
		}
	}
	w, err := os.Create(path.Join(projectPath, c.Name+".log"))
	if err != nil {
		return errors.Wrap(err, "could not open log file")
	}
	c.log = w

	if _, err := io.WriteString(w, "Started client "+c.Name+"\n"); err != nil {
		return errors.Wrap(err, "could not write to log")
	}
	if err := c.sendOpenReply(); err != nil {
		return errors.Wrap(err, "could not send open reply")
	}
	return nil
}

// sendOpenReply sends a reply to the open message.
func (c *Client) sendOpenReply() error {
	msg, err := osc.NewMessage("/reply")
	if err != nil {
		return errors.Wrap(err, "could not create open reply message")
	}
	if err := msg.WriteString("/nsm/client/open"); err != nil {
		return errors.Wrap(err, "could not write address to open reply message")
	}
	if err := msg.WriteString("Client " + c.Name + " started"); err != nil {
		return errors.Wrap(err, "could not write message to open reply message")
	}
	if err := c.Send(msg); err != nil {
		return errors.Wrap(err, "could not send open reply message")
	}
	_, err = io.WriteString(c.log, "Sent open reply message\n")
	return err
}

// handleAnnounceReply handles a reply to the announce message.
func (c *Client) handleAnnounceReply(msg *osc.Message) error {
	addr, err := msg.ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read reply address")
	}
	if addr != AddressAnnounce {
		// TODO: put the message back in a queue and keep waiting
		os.Stderr.Write([]byte("received reply for " + addr))
		return nil
	}

	serverMsg, err := msg.ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read reply message")
	}
	os.Stdout.Write([]byte("reply message: " + serverMsg))

	smName, err := msg.ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read session manager name")
	}
	os.Stdout.Write([]byte("session manager name: " + smName))
	return nil
}

// ServeOSC listens for incoming messages from Non Session Manager.
func (c *Client) ServeOSC() error {
	return c.Serve(c.Dispatcher())
}

// Dispatcher returns the osc Dispatcher for the nsm client.
func (c *Client) Dispatcher() osc.Dispatcher {
	return osc.Dispatcher{
		"/reply": func(msg *osc.Message) error {
			c.replyChan <- msg
			return nil
		},
		"/nsm/client/open": func(msg *osc.Message) error {
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

// NewClient creates a new nsm-enabled application.
func NewClient(config ClientConfig) (*Client, error) {
	return NewClientG(config, nil)
}

// NewClientG creates a new nsm-enabled application whose goroutines
// are part of the provided errgroup.Group.
func NewClientG(config ClientConfig, g *errgroup.Group) (*Client, error) {
	w, err := os.Create(path.Join(os.Getenv("HOME"), "."+config.Name))
	if err != nil {
		return nil, errors.Wrap(err, "could not open file")
	}
	// Create the client.
	c := &Client{
		ClientConfig: config,

		log:       w,
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
	c.Go(c.ServeOSC)

	// Announce app.
	if err := c.Announce(); err != nil {
		return nil, errors.Wrap(err, "could not announce app")
	}

	return c, nil
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
