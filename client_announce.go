package nsm

import (
	"os"
	"time"

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
func (c *Client) newAnnounceMsg() (*osc.Message, error) {
	msg, err := osc.NewMessage(AddressServerAnnounce)
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
	timeout := time.After(c.Timeout)
	select {
	case <-timeout:
		return ErrTimeout
	case msg := <-c.replyChan:
		c.handleAnnounceReply(msg)
	}
	return nil
}

// handleAnnounceReply handles a reply to the announce message.
func (c *Client) handleAnnounceReply(msg *osc.Message) error {
	addr, err := msg.ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read reply address")
	}
	if addr != AddressServerAnnounce {
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
