package nsm

import (
	"github.com/pkg/errors"
	"github.com/scgolang/osc"
)

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
		ProjectPath: projectPath,
		DisplayName: displayName,
		ClientID:    clientID,
		LocalAddr:   c.LocalAddr(),
	})
	if err := c.handle(AddressClientOpen, response, nsmerr); err != nil {
		return errors.Wrap(err, "could not respond to "+AddressClientOpen)
	}
	return nil
}
