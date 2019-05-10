package signalr

//Connection specify interface methods that allow consumer to interact with a connection type.
type Connection interface {
	State() ConnectionState
	Connect([]string) error
	CallHub(CallHubPayload, interface{}) error

	ListenToErrors() <-chan error
	ListenToHubResponses() <-chan MessageDataPayload
	ListenToHeartbeat() <-chan Heartbeat
	SubscribeToState() <-chan ConnectionState


}
