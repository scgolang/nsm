package nsm

import (
	"testing"
)

type ExampleClient struct {
	SessionInfo
}

func (c *ExampleClient) Open(info SessionInfo) (string, Error) {
	c.SessionInfo = info
	// Do what you need to do to open the session.
	return "Client has finished opening the session", nil
}

func (c *ExampleClient) Save() (string, Error) {
	// Do what you need to do to save the session.
	return "Client has finished saving the session", nil
}

func Example_client(t *testing.T) {
	c, err := NewClient(ClientConfig{
		Session: &ExampleClient{},
	})
	if err != nil {
		t.Fatal(err)
	}
	c.Wait()
}
