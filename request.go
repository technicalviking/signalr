package signalr

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

//CallHubPayload parameters for sending message to signalr hub.  identifier is set internally.  Arguments must be json marshallable.
type CallHubPayload struct {
	Hub       string        `json:"H"`
	Method    string        `json:"M"`
	Arguments []interface{} `json:"A"`

	identifier string `json:"I"`
}

// CallHub send a message to the signalr peer.  Sets unique identifier in threadsafe way.
func (c *client) CallHub(payload CallHubPayload) json.RawMessage {
	//increment the message identifier.
	c.callHubIDMutex.Lock()
	payload.identifier = fmt.Sprintf("%d", c.nextID)
	c.nextID++
	c.callHubIDMutex.Unlock()

	var (
		data []byte
		err  error
	)

	//attempt to marshal the payload
	if data, err = json.Marshal(payload); err != nil {
		c.sendErr(CallHubError(err.Error()))
		return nil
	}

	//set the response future channel
	c.setResponseChan(payload.identifier)

	//send the message payload to the signalr peer
	c.sendHubMessage(data)

	var (
		response *serverMessage
		ok       bool
	)

	if response, ok = <-(c.responseChan(payload.identifier)); !ok {
		c.sendErr(CallHubError(fmt.Sprintf("Call to method %s returned no result.", payload.Method)))
		return nil
	}

	return response.Result
}

func (c *client) sendHubMessage(data []byte) {
	c.socketWriteMutex.Lock()
	defer c.socketWriteMutex.Unlock()

	if err := c.socket.WriteMessage(websocket.TextMessage, data); err != nil {
		c.sendErr(SocketError(err.Error()))
	}

	return
}
