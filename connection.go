package signalr

//Connection specify interface methods that allow consumer to interact with a connection type.
type Connection interface {
	State() ConnectionState
	Connect([]string)
	negotiate() *negotiationResponse
	connectWebSocket(*negotiationResponse, []string)

	/* 	ErrorChannel()
	   	MessageChannel()


	   	SetMaxRetries()
	   	Close() */
}
