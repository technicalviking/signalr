package signalr

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

//this test is not a unit test... at all....

func TestCallHub(t *testing.T) {
	//Assemble

	//config
	cfg := Config{
		Client: &http.Client{},
		ConnectionURL: &url.URL{
			Scheme: "https",
			Host:   "socket.bittrex.com",
		},
		NegotiatePath: "signalr/negotiate",
		ConnectPath:   "signalr/connect",
		ReconnectPath: "signalr/reconnect",
	}

	//client
	c := New(cfg).(*client)

	c.heartbeatChan = make(chan Heartbeat, 5)

	//connection
	c.Connect([]string{"c2"})

	var (
		//callHub payload with empty string as argument
		payload CallHubPayload = CallHubPayload{
			Hub:    "c2",
			Method: "GetAuthContext",
			Arguments: []interface{}{
				"",
			},
		}

		result string
	)

	timeout := time.NewTicker(time.Second * 10)

	//Act
	c.CallHub(payload, &result)

	//Assert
	select {
	case e := <-c.errChan:
		switch cast := e.(type) {
		case CallHubError:
			t.Logf("Callhub failed in an expected way: %s", cast.Error())
		default:
			t.Errorf("Error received %+s", cast.Error())
		}
	case <-timeout.C:
		if result == "" {
			t.Errorf("timeout hit without result being populated")
		}

	}

}
