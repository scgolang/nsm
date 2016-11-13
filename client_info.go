package nsm

import (
	"github.com/pkg/errors"
	"github.com/scgolang/osc"
)

// handleClientInfo runs a goroutine that handles the
// client to server informational messages.
func (c *Client) handleClientInfo() error {
	// TODO: use a context.Context to receive cancellation
	for {
		select {
		case isDirty := <-c.Session.Dirty():
			if err := c.sendDirty(isDirty); err != nil {
				return errors.Wrap(err, "could not send dirty message")
			}
		case isGUIShowing := <-c.Session.GUIShowing():
			if err := c.sendGUIShowing(isGUIShowing); err != nil {
				return errors.Wrap(err, "could not send gui-showing message")
			}
		case x := <-c.Session.Progress():
			if err := c.sendProgress(x); err != nil {
				return errors.Wrap(err, "could not send progress message")
			}
		case clientStatus := <-c.Session.ClientStatus():
			if err := c.sendClientStatus(clientStatus); err != nil {
				return errors.Wrap(err, "could not send client status message")
			}
		}
	}
	return nil
}

// sendDirty sends an OSC message telling Non Session Manager
// if the client has unsaved changes.
func (c *Client) sendDirty(isDirty bool) error {
	var address string
	if isDirty {
		address = AddressClientIsDirty
	} else {
		address = AddressClientIsClean
	}
	msg, err := osc.NewMessage(address)
	if err != nil {
		return errors.Wrap(err, "could not create osc message")
	}
	return c.Send(msg) // Will get wrapped by caller
}

// sendGUIShowing sends an OSC message telling Non Session Manager
// if the client has unsaved changes.
func (c *Client) sendGUIShowing(isGUIShowing bool) error {
	var address string
	if isGUIShowing {
		address = AddressClientGUIShowing
	} else {
		address = AddressClientGUIHidden
	}
	msg, err := osc.NewMessage(address)
	if err != nil {
		return errors.Wrap(err, "could not create osc message")
	}
	return c.Send(msg) // Will get wrapped by caller
}

// sendProgress sends a progress measurement to Non Session Manager.
func (c *Client) sendProgress(x float32) error {
	msg, err := osc.NewMessage(AddressClientProgress)
	if err != nil {
		return errors.Wrap(err, "could not create osc message")
	}
	if err := msg.WriteFloat32(x); err != nil {
		return errors.Wrap(err, "could not add float32 to progress message")
	}
	return c.Send(msg) // Will get wrapped by caller
}

// sendClientStatus sends a progress measurement to Non Session Manager.
func (c *Client) sendClientStatus(clientStatus ClientStatus) error {
	msg, err := osc.NewMessage(AddressClientStatus)
	if err != nil {
		return errors.Wrap(err, "could not create osc message")
	}
	if err := msg.WriteInt32(int32(clientStatus.Priority)); err != nil {
		return errors.Wrap(err, "could not add int32 to client status message")
	}
	if err := msg.WriteString(clientStatus.Message); err != nil {
		return errors.Wrap(err, "could not add string to client status message")
	}
	return c.Send(msg) // Will get wrapped by caller
}
