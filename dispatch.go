package signalr

import (
	"encoding/json"
	"fmt"
)

//MessageDataPayload contains information from signalR peer based on subscription
type MessageDataPayload struct {
	HubName   string            `json:"H"`
	Method    string            `json:"M"`
	Arguments []json.RawMessage `json:"A"`
}

func (c *client) handleSocketData(message *serverMessage) {

	var payload MessageDataPayload

	for msg := range c.responseChan("default") {
		for _, curData := range msg.Data {
			if err := json.Unmarshal(curData, &payload); err != nil {
				c.errChan <- HubMessageError(fmt.Sprintf("Unable to unmarshal message data: %s", err.Error()))
				continue
			}

			c.messageChan <- payload
		}
	}
}
