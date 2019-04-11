package signalr

import (
	"encoding/json"
	"errors"
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
		c.sendErr(
			newCallHubError(
				fmt.Sprintf(
					"Unable to marshal the callhub payload: %+v",
					payload,
				),
				err,
			),
		)

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
		c.sendErr(
			newCallHubError(
				fmt.Sprintf("Call to method %s returned no result.", payload.Method),
				nil,
			),
		)
		return
	}
	if response.Error != "" {
		c.sendErr(
			newCallHubError(
				"Error detected within responseChan payload.",
				errors.New(response.Error),
			),
		)
		return
	}

	if err = json.Unmarshal(response.Result, resultPayload); err != nil {
		//e := fmt.Sprintf("Unable to parse response into type provided for call to %s: %s", payload.Method,
		//string(response.Result))
		c.sendErr(
			newCallHubError(
				fmt.Sprintf("Unable to parse response: \n Method: %s \n response.Result: %s \n",
					payload.Method,
					string(response.Result),
				),
				err,
			),
		)
	}

	return
}

func (c *client) sendHubMessage(data []byte) {
	c.socketWriteMutex.Lock()
	defer c.socketWriteMutex.Unlock()

	if err := c.socket.WriteMessage(websocket.TextMessage, data); err != nil {
		c.sendErr(
			newSocketError(
				"Unable to write message to socket hub",
				err,
			),
		)
	}

	return
}
