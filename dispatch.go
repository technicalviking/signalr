package signalr

import (
	"encoding/json"
	"fmt"
	"time"
)

type serverMessage struct {
	Cursor     string            `json:"C"`
	Data       []json.RawMessage `json:"M"`
	Result     json.RawMessage   `json:"R"`
	Identifier string            `json:"I"`
	Error      string            `json:"E"`
}

func (c *client) beginDispatchLoop(timeout float32) {

	keepAliveTime := time.Now()

	t := time.NewTicker(time.Second)
	dataChan := c.listenToWebSocket()

	for {
		select {
		case data, ok := <-dataChan:
			if !ok {
				t.Stop()
				c.stateChan <- Disconnected
				return
			}

			//sc.handleSocketData(data)
		case <-t.C:

			if time.Since(keepAliveTime) > time.Duration(timeout)*time.Second {
				c.errChan <- TimeoutError("keepalive timeout reached.  RECONNECTING.")
				t.Stop()
				c.socket.Close()

				return
			}
		}

	}

}

func (c *client) listenToWebSocket() chan serverMessage {
	socketDataChan := make(chan serverMessage)

	go func() {
		defer close(socketDataChan)
		for {
			var (
				data []byte
				err  error
			)

			if _, data, err = c.socket.ReadMessage(); err != nil {
				c.errChan <- SocketError(err.Error())
				return
			}

			var message serverMessage
			if err = json.Unmarshal(data, &message); err != nil {
				c.errChan <- SocketError(fmt.Sprintf("Unable to parse message: %s\n", err.Error()))
				continue
			}
			socketDataChan <- message
		}
	}()

	return socketDataChan
}
