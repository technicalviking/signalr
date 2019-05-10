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

func (c *client) Connect(hubs []string) error {
	if c.State() == Broken {
		return ConnectError("Client in broken state.  Check config or create new client instance.")
	}

	c.state = Connecting

	nResp, negotiationErr := c.negotiate()
	if negotiationErr != nil {
		return negotiationErr
	}

	if err := c.connectWebSocket(nResp, hubs); err != nil {
		return err
	}

	return c.handleSocketCommunication(nResp, hubs)

}

func (c *client) handleSocketCommunication(nResp *negotiationResponse, hubs []string) error {
	for {
		c.listenToWebSocketData(time.Second * time.Duration(nResp.KeepAliveTimeout)) //20 seconds, as of 2019.04.16 --DM

		//if the code gets here, that means the socket disconnected.
		c.setState(Reconnecting)
		if err := c.reconnectWebSocket(nResp, hubs); err != nil {
			return err
		}
	}
}

func (c *client) negotiate() (*negotiationResponse, error) {
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
		err = NewNegotiationError("Unable to create new request", err)
		c.sendErr(err)
		c.setState(Broken)
		return nil, err
	}

	for k, values := range c.config.RequestHeaders {
		for _, val := range values {
			request.Header.Add(k, val)
		}
	}

	if response, err = c.config.Client.Do(request); err != nil {
		err = NewNegotiationError("Unable to execute negotiation request", err)
		c.sendErr(err)
		c.setState(Broken)
		return nil, err
	}

	defer response.Body.Close()

	if body, err = ioutil.ReadAll(response.Body); err != nil {
		err = NewNegotiationError("Unable to read negotiation response body", err)
		c.sendErr(err)
		c.setState(Broken)
		return nil, err
	}

	if err = json.Unmarshal(body, &result); err != nil {
		err = NewNegotiationError(
			fmt.Sprintf("Unable to parse negotiation response: %s", string(body)),
			err,
		)
		c.sendErr(err)
		c.setState(Broken)
		return nil, err
	}

	return &result, nil
}

func (c *client) connectWebSocket(params *negotiationResponse, hubs []string) error {
	if c.State() == Broken {
		return NewBrokenWebSocketError(
			"connectWebSocket",
			fmt.Errorf("unable to connect, client object in broken state"),
		)
	}

	connectionURL := url.URL{
		Scheme: "wss",
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

	var (
		err  error
		resp *http.Response
	)

	for i := 0; i <= 5; i++ {
		if i == 5 {
			c.setState(Broken)
			err = SocketConnectionError("MAX RETRIES REACHED.  ABORTING CONNECTION.")
			c.sendErr(err)
			return err
		}

		backoff := math.Pow(2.0, float64(i))
		time.Sleep(time.Second * time.Duration(backoff))
		//@todo incorporate the currently ignored http response parameter into socketConnectionError
		if c.socket, resp, err = socketDialer.Dial(connectionURL.String(), c.config.RequestHeaders); err != nil {
			c.sendErr(
				SocketConnectionError(
					fmt.Sprintf(
						"\n Unable to dial successfully: %s \n HTTP Response: %+v\n",
						err.Error(),
						resp,
					),
				),
			)
		} else {
			c.setState(Connected)
			break
		}
	}

	return nil
}

func (c *client) reconnectWebSocket(params *negotiationResponse, hubs []string) error {
	if c.State() == Broken {
		return NewBrokenWebSocketError(
			"reconnectWebSocket",
			fmt.Errorf("unable to reconnect, client object in broken state"),
		)
	}

	// if we get here without having recieved a single message, try to connect instead.
	if c.messageID == "" {
		return c.connectWebSocket(params, hubs)

	}

	connectionURL := url.URL{
		Scheme: "wss",
		Host:   c.config.ConnectionURL.Host,
		Path:   c.config.ReconnectPath,
		RawQuery: url.Values{
			"transport":       []string{"webSockets"},
			"clientProtocol":  []string{params.ProtocolVersion},
			"connectionToken": []string{params.ConnectionToken},
			"connectionData":  []string{string(castHubNamesToString(hubs))},
			"messageId":       []string{c.messageID},
			"_":               []string{fmt.Sprintf("%d", time.Now().Unix()*1000)},
		}.Encode(),
	}

	socketDialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 30 * time.Second,
		Jar:              c.config.Client.Jar,
	}

	var (
		err  error
		resp *http.Response
	)

	for i := 0; i <= 5; i++ {
		if i == 5 {
			c.setState(Broken)
			err = SocketConnectionError("MAX RETRIES REACHED.  ABORTING CONNECTION.")
			c.sendErr(err)
			return err
		}

		backoff := math.Pow(2.0, float64(i))
		time.Sleep(time.Second * time.Duration(backoff))
		//@todo incorporate the currently ignored http response parameter into socketConnectionError
		if c.socket, resp, err = socketDialer.Dial(connectionURL.String(), c.config.RequestHeaders); err != nil {
			c.sendErr(
				SocketConnectionError(
					fmt.Sprintf(
						"\n Unable to dial successfully: %s \n HTTP Response: %+v\n",
						err.Error(),
						resp,
					),
				),
			)
		} else {
			c.setState(Connected)
			break
		}
	}

	return nil
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
