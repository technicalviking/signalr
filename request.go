package signalr

import (
	"encoding/json"
	"fmt"
)

//CallHubPayload parameters for sending message to signalr hub.  identifier is set internally.  Arguments must be json marshallable.
type CallHubPayload struct {
	Hub       string        `json:"H"`
	Method    string        `json:"M"`
	Arguments []interface{} `json:"A"`

	identifier string `json:"I"`
}

func (c *client) CallHub(payload CallHubPayload) json.RawMessage {
	//increment the message identifier.
	c.callHubIDMutex.Lock()
	payload.identifier = fmt.Sprintf("%d", c.nextID)
	c.nextID++
	c.callHubIDMutex.Unlock()

	//attempt to marshal the payload
	if data, err := json.Marshal(payload); err != nil {
		c.sendErr(CallHubError(err.Error()))
		return nil
	}

	//set the response future channel
	/*
		responseKey := request.Identifier
		responseChannel := sc.createResponseFuture(responseKey)
	*/

	//send the message payload to the signalr peer (port this method?)
	/* 	if err := sc.sendHubMessage(data); err != nil {
		return nil, err
	} */

	//get message from response.

	//if response has an error, pipe that to the error chan.

	//return the result

	/*

	   OLD CODEOLD CODEOLD CODEOLD CODEOLD CODE
	   OLD CODEOLD CODEOLD CODEOLD CODEOLD CODE
	   OLD CODEOLD CODEOLD CODEOLD CODEOLD CODE


	   	responseKey := request.Identifier
	   	responseChannel := sc.createResponseFuture(responseKey)

	   	var (
	   		response *serverMessage
	   		ok       bool
	   	)

	   	if response, ok = <-responseChannel; !ok {
	   		return nil, fmt.Errorf("Call to server returned no result")
	   	}

	   	if len(response.Error) > 0 {
	   		return nil, fmt.Errorf("%s", response.Error)
	   	}

			 return response.Result, nil */

	return nil
}
