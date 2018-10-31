package signalr

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

//CallHubPayload parameters for sending message to signalr hub.  identifier is set internally.  Arguments must be json marshallable.
type CallHubPayload struct {
	Hub        string        `json:"H"`
	Method     string        `json:"M"`
	Arguments  []interface{} `json:"A"`
	Identifier string        `json:"I"`
}

// CallHub send a message to the signalr peer.  Sets unique identifier in threadsafe way.
func (c *client) CallHub(payload CallHubPayload, resultPayload interface{}) {
	//increment the message identifier.
	c.callHubIDMutex.Lock()
	payload.Identifier = fmt.Sprintf("%d", c.nextID)
	c.nextID++
	c.callHubIDMutex.Unlock()

	var (
		data []byte
		err  error
	)

	//attempt to marshal the payload
	if data, err = json.Marshal(payload); err != nil {
		c.sendErr(CallHubError(err.Error()))
		return
	}

	//set the response future channel
	c.setResponseChan(payload.Identifier)
	//send the message payload to the signalr peer
	c.sendHubMessage(data)

	var (
		response *serverMessage
	)

	response = <-(c.responseChan(payload.Identifier))

	if response == nil {
		c.sendErr(CallHubError(fmt.Sprintf("Call to method %s returned no result.", payload.Method)))
		return
	}
	if response.Error != "" {
		c.sendErr(CallHubError(response.Error))
		return
	}

	if err = json.Unmarshal(response.Result, resultPayload); err != nil {
		e := fmt.Sprintf("Unable to parse response into type provided for call to %s: %s", payload.Method, string(response.Result))
		c.sendErr(CallHubError(e))
	}

	return
}

func (c *client) sendHubMessage(data []byte) {
	c.socketWriteMutex.Lock()
	defer c.socketWriteMutex.Unlock()

	if err := c.socket.WriteMessage(websocket.TextMessage, data); err != nil {
		c.sendErr(SocketError(err.Error()))
	}

	return
}
