package signalr

import "encoding/json"

//Connection specify interface methods that allow consumer to interact with a connection type.
type Connection interface {
	State() ConnectionState
	Connect([]string)

	ListenToErrors() <-chan error
	ListenToHubResponses() <-chan MessageDataPayload

	CallHub(CallHubPayload) json.RawMessage
}
