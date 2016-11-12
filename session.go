package nsm

// Code is the type of an error code.
type Code int

// Error codes.
const (
	ErrGeneral         Code = -1
	ErrIncompatibleAPI Code = -2
	ErrBlacklisted     Code = -3
	ErrLaunchFailed    Code = -4
	ErrNoSuchFile      Code = -5
	ErrSoSessionOpen   Code = -6
	ErrUnsavedChanges  Code = -7
	ErrNotNow          Code = -8
	ErrBadProject      Code = -9
	ErrCreateFailed    Code = -10
)

// Error extends the builtin error interface to add an nsm error code.
type Error interface {
	error
	Code() Code
}

// NewError creates an error with an nsm-specific error code.
func NewError(code Code, msg string) Error {
	return nsmError{msg: msg, code: code}
}

// nsmError represents an nsm-specific error.
type nsmError struct {
	msg  string
	code Code
}

// Error returns an error message.
func (e nsmError) Error() string {
	return e.msg
}

// Code returns an nsm-specific error code.
func (e nsmError) Code() Code {
	return e.code
}

// Session represents the behavior of a client
// with respect to the control messages that are
// sent by Non Session Manager.
// The only methods in this interface that are actually
// REQUIRED by all implementations are Open and Save.
// The other methods are all optional and mostly depend on
// the capabilities that a given client wishes to implement
// (e.g. progress reports for long-running open and save operations).
type Session interface {
	// Announce will be called when the server has replied
	// to the client's announce message.
	// The server sends back information about itself in the reply.
	Announce(ServerInfo)

	// Open tells the client to open a session.
	// If a client has not specified CapSwitch in their
	// capabilities then this method will only be called once.
	Open(SessionInfo) (string, Error)

	// This method will only be called after Open has been
	// called, and may be called any number of times during
	// a session. If the user aborts the session this method
	// will not be called.
	Save() (string, Error)

	// This method will be invoked when Non Session Manager
	// has started all the clients for a given session.
	SessionIsLoaded()

	// ShowGUI is an optional method that may be invoked in response
	// to server commands. If your nsm client has the CapGUI capability
	// then it should show/hide the GUI based on the bool parameter.
	ShowGUI(bool)

	// Dirty is an optional client to server informational message
	// that tells Non Session Manager that a client has unsaved changes.
	// If your client does not support CapDirty then return nil from
	// this method.
	Dirty() chan bool

	// If your client has the CapGUI capability then it must
	// return a channel that can be used to notify Non Session Manager
	// that the client's GUI is hidden (false) or showing (true).
	GUIShowing() chan bool

	// Progress can be used by clients with the CapProgress capability
	// to indicate an ongoing save or open operation.
	// Note that clients with CapProgress are still required to call
	// Open or Save when the operation has completed.
	// The float32 that is sent on this channel must be between 0 and 1,
	// with 1 indicating completion.
	Progress() chan float32

	// Message can be used by clients with the CapMessage capability
	// to send status updates to Non Session Manager.
	Message() chan ClientStatus
}

// SessionInfo contains the data a client receives
// in an Open client control message.
// Note that the optional methods from the Session interface
// are implemented on SessionInfo.
// This means that if your Session implementation embeds
// SessionInfo (which is might a good idea anyways because
// your app should probably cache info about the current session)
// you can easily avoid writing boilerplate no-op methods
// for capabilities you do not wish to implement.
type SessionInfo struct {
	// Path is the path where a client can store
	// their project-specific data.
	// Path can be a directory tree or a file.
	// This is up to the client.
	// If a project exists at the provided path
	// the project must be opened immediately.
	// Otherwise a new project must immediately
	// be created at Path.
	Path string

	// DisplayName is the name of the client as
	// displayed in Non Session Manager.
	DisplayName string

	// ClientID should be used by clients that expect
	// more than one instance to be part of a single session.
	// ClientID should be prepended to any names it registers
	// with subsystems that could be used by other instances.
	// For example, clients that create JACK connections
	// should prepend ClientID to the JACK client name.
	ClientID string
}

// SessionIsLoaded is a no-op.
func (s SessionInfo) SessionIsLoaded() {
}

// ShowGUI is a no-op.
func (s SessionInfo) ShowGUI(show bool) {
}

// Dirty returns nil.
func (s SessionInfo) Dirty() chan bool {
	return nil
}

// GUIShowing returns nil.
func (s SessionInfo) GUIShowing() chan bool {
	return nil
}

// Progress returns nil.
func (s SessionInfo) Progress() chan float32 {
	return nil
}

// Message returns nil.
func (s SessionInfo) Message() chan ClientStatus {
	return nil
}

// ClientStatus represents a status sent to Non Session Manager.
// Priority should be used to indicate how important it is for
// the user to see the status message, with 0 being lowest priority
// and 3 being highest.
type ClientStatus struct {
	Priority int
	Message  string
}

// ServerInfo contains info about Non Session Manager itself.
// Message is a message that was received in reply to the client's
// announce message. Name is the name of the session manager.
// Capabilities describes the capabilities of the session manager.
type ServerInfo struct {
	Message      string
	Name         string
	Capabilities Capabilities
}
