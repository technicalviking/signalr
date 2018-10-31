package signalr

import (
	"net/http"
	"net/url"
	"testing"
	"time"
	//"fmt"
)

func TestConnect(t *testing.T) {
	//Assemble
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

	c := New(cfg).(*client)

	timeout := time.NewTicker(time.Second * 30)
	c.heartbeatChan = make(chan Heartbeat, 5)

	//Act

	c.Connect([]string{"c2"})
	//assert
	select {
	case err := <-c.errChan:
		t.Fatalf("Error found!  %s\n", err.Error())
	case <-timeout.C:
		t.Fatalf("Timeout error")
	case r := <-c.responseChan("default"):
		t.Fatalf("what is r %+v", len(r.Data))
	case <-c.heartbeatChan:
	}

}

func TestNegotiate(t *testing.T) {
	//Assemble
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

	c := New(cfg).(*client)

	//act
	nresp := c.negotiate()

	//assert
	if nresp == nil {
		t.Errorf("unable to connect to sample server: %+v \n\n\n", nresp)
	}
}

/*func TestConnectWebSocket(t *testing.T) {

	//Assemble
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

	c := New(cfg).(*client)
	nresp := c.negotiate()

	c.heartbeatChan = make(chan Heartbeat)

	c.connectWebSocket(nresp, []string{"c2"})

	select {
	case e := <-c.errChan:
		t.Fatalf("error found! %s", e.Error())
	case hb := <-c.heartbeatChan:
		t.Fatalf("HEARTBEAT %s", hb)
	case <-c.responseChan("default"):
	}
}*/
