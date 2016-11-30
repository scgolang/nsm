package nsm

import (
	"os"

	"github.com/pkg/errors"
	"github.com/scgolang/osc"
)

// Announce announces a new nsm application.
func (c *Client) announce() error {
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
func (c *Client) newAnnounceMsg() (osc.Message, error) {
	msg := osc.Message{
		Address: AddressServerAnnounce,
		Arguments: osc.Arguments{
			osc.String(os.Args[0]),
			osc.String(c.Capabilities.String()),
			osc.String(os.Args[0]),
			osc.Int(c.Major),
			osc.Int(c.Minor),
			osc.Int(int32(c.PID)),
		},
	}
	// Write name.
	if c.Name != "" {
		msg.Arguments[0] = osc.String(c.Name)
	}
	return msg, nil
}

// announceWait waits for a reply to the announce message.
func (c *Client) announceWait() error {
	return c.wait(AddressServerAnnounce)
}

// handleAnnounceReply handles a reply to the announce message.
func (c *Client) handleAnnounceReply(msg osc.Message) error {
	addr, err := msg.Arguments[0].ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read reply address")
	}
	if addr != AddressServerAnnounce {
		// TODO: put the message back in a queue and keep waiting
		return nil
	}

	serverMsg, err := msg.Arguments[1].ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read reply message")
	}
	smName, err := msg.Arguments[2].ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read session manager name")
	}
	capsRaw, err := msg.Arguments[3].ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read session manager capabilities")
	}
	return c.Session.Announce(ServerInfo{
		Message:      serverMsg,
		ServerName:   smName,
		Capabilities: parseCapabilities(capsRaw),
	})
}
