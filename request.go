package signalr

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
)

//CallHubPayload parameters for sending message to signalr hub.  identifier is set internally.  Arguments must be json marshallable.
type CallHubPayload struct {
	Hub        string        `json:"H,omitempty"`
	Method     string        `json:"M,omitempty"`
	Arguments  []interface{} `json:"A,omitempty"`
	Identifier string        `json:"I,omitempty"`
}

var (
  pingMessage = CallHubPayload{}
)

// SendPing Sends a ping message to the signalr hub.
func (c client) SendPing() error {
  var result string
  return c.CallHub(pingMessage, &result)
}

// CallHub send a message to the signalr peer.  Sets unique identifier in threadsafe way.
// Result of the callhub is set into resultPayload
func (c *client) CallHub(payload CallHubPayload, resultPayload interface{}) error {
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
		err = newCallHubError(
			fmt.Sprintf(
				"Unable to marshal the callhub payload: %+v",
				payload,
			),
			err,
		)
		c.sendErr(err)

		return err
	}

	//set the response future channel
	c.setResponseChan(payload.Identifier)
	//send the message payload to the signalr peer
	if err = c.sendHubMessage(data); err != nil {
		return err
	}

	var (
		response *serverMessage
	)

	response = <-(c.responseChan(payload.Identifier))

	if response == nil {
		err = newCallHubError(
			fmt.Sprintf("Call to method %s returned no result.", payload.Method),
			nil,
		)
		c.sendErr(err)
		return err
	}
	if response.Error != "" {
		err = newCallHubError(
			"Error detected within responseChan payload.",
			errors.New(response.Error),
		)
		c.sendErr(err)
		return err
	}

	if err = json.Unmarshal(response.Result, resultPayload); err != nil {
		err = newCallHubError(
			fmt.Sprintf("Unable to parse response: \n Method: %s \n response.Result: %s \n",
				payload.Method,
				string(response.Result),
			),
			err,
		)
		c.sendErr(err)
		return err
	}

	return nil
}

func (c *client) sendHubMessage(data []byte) error {
	c.socketWriteMutex.Lock()
	defer c.socketWriteMutex.Unlock()

	if err := c.socket.WriteMessage(websocket.TextMessage, data); err != nil {
		err = newSocketError(
			"Unable to write message to socket hub",
			err,
		)
		c.sendErr(err)
		return err
	}

	return nil
}
