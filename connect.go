package signalr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type negotiationResponse struct {
	ConnectionToken         string
	URL                     string
	ConnectionID            string
	KeepAliveTimeout        float32
	DisconnectTimeout       float32
	ConnectionTimeout       float32
	TryWebSockets           bool
	ProtocolVersion         string
	TransportConnectTimeout float32
	LogPollDelay            float32
}

func (c *client) Connect(hubs []string) {
	if c.State() == Broken {
		c.errChan <- ConnectError("Client in broken state.  Check config or create new client instance.")
		return
	}

	c.state = Connecting

	nResp := c.negotiate()
	c.connectWebSocket(nResp, hubs)

	go c.listenToWebSocketData(time.Second * time.Duration(nResp.KeepAliveTimeout))
	//go c.beginDispatchLoop(nResp.KeepAliveTimeout)
}

func (c *client) negotiate() *negotiationResponse {
	var (
		request  *http.Request
		response *http.Response
		result   negotiationResponse
		err      error
		body     []byte
	)

	negotiationURL := url.URL{
		Scheme: c.config.ConnectionURL.Scheme,
		Host:   c.config.ConnectionURL.Host,
		Path:   c.config.NegotiatePath,
		RawQuery: url.Values{
			"clientProtocol": []string{"1.5"},
			"_":              []string{fmt.Sprintf("%d", time.Now().Unix()*1000)},
		}.Encode(),
	}

	if request, err = http.NewRequest("GET", negotiationURL.String(), nil); err != nil {
		c.errChan <- NegotiationError(err.Error())
		c.stateChan <- Broken
		return nil
	}

	for k, values := range c.config.RequestHeaders {
		for _, val := range values {
			request.Header.Add(k, val)
		}
	}

	if response, err = c.config.Client.Do(request); err != nil {
		c.errChan <- NegotiationError(err.Error())
		return nil
	}

	defer response.Body.Close()

	if body, err = ioutil.ReadAll(response.Body); err != nil {
		c.errChan <- NegotiationError(err.Error())
		return nil
	}

	if err = json.Unmarshal(body, &result); err != nil {
		err = fmt.Errorf("Failed to parse response '%s': %s", string(body), err.Error())
		c.errChan <- NegotiationError(err.Error())
		return nil
	}

	return &result
}

func (c *client) connectWebSocket(params *negotiationResponse, hubs []string) {
	if c.State() == Broken {
		return
	}

	connectionURL := url.URL{
		Scheme: c.config.ConnectionURL.Scheme,
		Host:   c.config.ConnectionURL.Host,
		Path:   c.config.ConnectPath,
		RawQuery: url.Values{
			"transport":       []string{"webSockets"},
			"clientProtocol":  []string{params.ProtocolVersion},
			"connectionToken": []string{params.ConnectionToken},
			"connectionData":  []string{string(castHubNamesToString(hubs))},
			"_":               []string{fmt.Sprintf("%d", time.Now().Unix()*1000)},
		}.Encode(),
	}

	socketDialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              c.config.Client.Jar,
	}

	var err error

	for i := 0; i <= 5; i++ {
		if i == 5 {
			c.stateChan <- Broken
			c.errChan <- SocketConnectionError("MAX RETRIES REACHED.  ABORTING CONNECTION.")
			break
		}

		backoff := math.Pow(2.0, float64(i))
		time.Sleep(time.Second * time.Duration(backoff))

		if c.socket, _, err = socketDialer.Dial(connectionURL.String(), c.config.RequestHeaders); err != nil {
			c.errChan <- SocketConnectionError(err.Error())
		}
	}
}

func castHubNamesToString(hubs []string) []byte {
	var connectionData = make([]struct {
		Name string `json:"Name"`
	}, len(hubs))
	for i, h := range hubs {
		connectionData[i].Name = h
	}
	connectionDataBytes, _ := json.Marshal(connectionData)

	return connectionDataBytes
}
