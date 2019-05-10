package signalr

import "fmt"

// ConnectError used when consuming app tries to connect when app is in broken state.
type ConnectError string

func (ce ConnectError) Error() string {
	return fmt.Sprintf("ConnectError: %s", string(ce))
}

type baseError struct {
	error
	prefix string
	source string
	err    error
}

func (b baseError) Error() string {
	return fmt.Sprintf(
		"%s \n Source: %s \n %Error: s",
		b.prefix,
		b.source,
		b.err.Error(),
	)
}

func newBaseError(prefix string, source string, err error) baseError {
	return baseError{
		prefix: fmt.Sprintf("SignalR %s", prefix),
		source: source,
		err:    err,
	}
}

//NegotiationError error created when negotiation step of connection fails.
type NegotiationError baseError

func NewNegotiationError(source string, err error) NegotiationError {
	return NegotiationError(
		newBaseError(
			"NegotiationError",
			source,
			err,
		),
	)
}

//SocketConnectionError error created when connectWebSocket step of connection fails.
type SocketConnectionError string

// Error implement Error interface
func (sce SocketConnectionError) Error() string {
	return fmt.Sprintf("SocketConnectionError: %s", string(sce))
}

//SocketError error created when websocket.ReadMessage or websocket.WriteMessage returns an error..
type SocketError baseError

func newSocketError(source string, err error) SocketError {
	return SocketError(
		newBaseError(
			"SocketError",
			source,
			err,
		),
	)
}

//TimeoutError error created when dispatch loop times out
type TimeoutError string

// Error implement Error interface
func (te TimeoutError) Error() string {
	return fmt.Sprintf("TimeoutError: %s", string(te))
}

//HubMessageError error created when unable to parse the messagedata from the serverMessage
type HubMessageError string

//Error implement the error interface
func (hme HubMessageError) Error() string {
	return fmt.Sprintf("HubMessageError: %s", string(hme))
}

// CallHubError error generated during an attempt to send a message to the Signalr hub
type CallHubError baseError

func newCallHubError(source string, err error) CallHubError {
	return CallHubError(
		newBaseError(
			"CallHubError",
			source,
			err,
		),
	)
}

type MDPParseError baseError

func newMDPParseError(source string, err error) MDPParseError {
	return MDPParseError(newBaseError(
		"MessageDataPayloadParseError",
		source,
		err,
	),
	)
}



// BrokenWebSocketError describes a broken websocket error
type BrokenWebSocketError struct {
	error
	prefix string
	source string
	err error
}

func (b BrokenWebSocketError) Error() string {
	return fmt.Sprintf("%s (%s): %s", b.prefix, b.source, b.err)
}

func NewBrokenWebSocketError(source string, err error) *BrokenWebSocketError{
	return &BrokenWebSocketError{
		prefix: "Broken Web Socket Error",
		source: source,
		err: err,
	}
}
