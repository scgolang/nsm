package nsm

import (
	"io"
	"os"
	"path"
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
	msg, err := osc.NewMessage(Reply)
	if err != nil {
		return errors.Wrap(err, "could not create open reply message")
	}
	if err := msg.WriteString(ClientOpen); err != nil {
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
