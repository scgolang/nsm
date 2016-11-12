package nsm

import (
	"strings"
)

// Capability is a capability of an nsm client.
type Capability string

// CapSep is the separator in the capabilities string
const CapSep = ":"

// Capabilities.
const (
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
