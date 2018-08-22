package signalr

import "fmt"

// ConnectError used when consuming app tries to connect when app is in broken state.
type ConnectError string

func (ce ConnectError) Error() string {
	return fmt.Sprintf("ConnectError: %s", string(ce))
}

//NegotiationError error created when negotiation step of connection fails.
type NegotiationError string

// Error implement Error interface
func (ne NegotiationError) Error() string {
	return fmt.Sprintf("NegotiationError: %s", string(ne))
}

//SocketConnectionError error created when connectWebSocket step of connection fails.
type SocketConnectionError string

// Error implement Error interface
func (sce SocketConnectionError) Error() string {
	return fmt.Sprintf("SocketConnectionError: %s", string(sce))
}

//SocketError error created when websocket.ReadMessage or websocket.WriteMessage returns an error..
type SocketError string

// Error implement Error interface
func (se SocketError) Error() string {
	return fmt.Sprintf("SocketError: %s", string(se))
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
type CallHubError string

func (che CallHubError) Error() string {
	return fmt.Sprintf("CallHubError: %s", string(che))
}
