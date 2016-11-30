package nsm

import (
	"testing"
	// "github.com/scgolang/osc"
)

type mockSession struct {
	SessionInfo
}

func newMockSession(t *testing.T) Session {
	return &mockSession{
		SessionInfo: SessionInfo{},
	}
}

func (m *mockSession) Open(info SessionInfo) (string, Error) {
	m.SessionInfo = info
	return "", nil
}

func (m *mockSession) Save() (string, Error) {
	return "", nil
}

// func (m *mockSession) Methods() osc.Dispatcher {
// 	return osc.Dispatcher{}
// }
