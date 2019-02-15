package signalr

import (
	"fmt"
)

//Heartbeat interface used to inform consuming app that a signal has come in with no data, probably as a keepalive
type Heartbeat interface {
	fmt.Stringer
}

type NormalHeartbeat string

//String implement Stringer interface
func (hb NormalHeartbeat) String() string {
	return fmt.Sprintf("Thump thump! - %s", hb)
}

//AwkwardHeartbeat test signalr implementation sends last hubmessage as aa heartbeat...?  Thanks bittrex.
type AwkwardHeartbeat string

//String implement Stringer interface
func (hb AwkwardHeartbeat) String() string {
	return "Thud thud!"
}

//GetError is used instead of implementing error because of how fmt.Sprintf and Printf work to use the Error interface before trying stringer
func (hb AwkwardHeartbeat) GetError() string {
	return string(hb)
}
