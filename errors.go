package signalr

//NegotiationError error created when negotiation step of connection fails.
type NegotiationError string

// Error implement Error interface
func (ne NegotiationError) Error() string {
	return string(ne)
}

//SocketConnectionError error created when connectWebSocket step of connection fails.
type SocketConnectionError string

// Error implement Error interface
func (sce SocketConnectionError) Error() string {
	return string(sce)
}

//SocketError error created when websocket.ReadMessage returns an error..
type SocketError string

// Error implement Error interface
func (se SocketError) Error() string {
	return string(se)
}

//TimeoutError error created when dispatch loop times out
type TimeoutError string

// Error implement Error interface
func (te TimeoutError) Error() string {
	return string(te)
}
