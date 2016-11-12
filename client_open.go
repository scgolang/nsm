package nsm

import (
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/osc"
)

// openWait waits for an open message from Non Session Manager.
func (c *Client) openWait() error {
	timeout := time.After(c.Timeout)
	select {
	case <-timeout:
		return ErrTimeout
	case msg := <-c.openChan:
		if err := c.handleOpen(msg); err != nil {
			return errors.Wrap(err, "could not handle open message")
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
	displayName, err := msg.ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read display name")
	}
	clientID, err := msg.ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read client ID")
	}
	response, nsmerr := c.Session.Open(SessionInfo{
		Path:        projectPath,
		DisplayName: displayName,
		ClientID:    clientID,
	})
	if err := c.handle(AddressClientOpen, response, nsmerr); err != nil {
		return errors.Wrap(err, "could not respond to "+AddressClientOpen)
	}
	return nil
}
