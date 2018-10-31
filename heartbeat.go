package signalr

//Heartbeat struct used to inform consuming app that a signal has come in with no data, probably as a keepalive
type Heartbeat struct{}

//String implement Stringer interface
func (hb Heartbeat) String() string {
	return "Thump thump!"
}
