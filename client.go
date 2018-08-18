package signalr

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//default values for configuartion
const (
	defaultScheme string = "https"
	socketScheme  string = "wss"
	signalRPath   string = "signalr"
	negotiatePath string = signalRPath + "/negotiate"
	connectPath   string = signalRPath + "/connect"
	reconnectPath string = signalRPath + "/reconnect"
)

//ConnectionState int representing current state of the SignalR Client
type ConnectionState int

//SignalR Client State Values
const (
	Ready ConnectionState = iota
	Connecting
	Reconnecting
	Connected
	Disconnected
	Broken
)

//Config define options required for connecting to a signalr endpoint.
type Config struct {
	//Client allows the consumer to override the default http client as needed. (Cloudflare issues anyone?)
	Client *http.Client

	//URL for the signalr endpoint.  uses url.URL package to ensure valid url is used.
	ConnectionURL url.URL `json:"url"`

	//URI path for negotiation portion of the connection.  Defaults to "/signalr/negotiate"
	NegotiatePath string `json:"negotiate_path,omitempty"`

	//URI path for websocket connection.  Defaults to "/signalr/connect"
	ConnectPath string `json:"connct_path,omitempty"`

	//URI path for websocket reconnect.  Defaults to "/signalr/reconnect"
	ReconnectPath string `json:"reconnect_path,omitempty"`

	// RequestHeaders additional header parameters to add to the negotiation HTTP request.
	RequestHeaders http.Header `json:"request_headers,omitempty"`
}

type serverMessage struct {
	Cursor     string            `json:"C"`
	Data       []json.RawMessage `json:"M"`
	Result     json.RawMessage   `json:"R"`
	Identifier string            `json:"I"`
	Error      string            `json:"E"`
}

//client implemntation of Connection interface.
type client struct {
	//persist sanitized config
	config Config
	//store current state of connection
	state ConnectionState
	//mutex to make changes to state threadsafe
	stateMutex sync.RWMutex
	//internal channel used to trigger state changes
	stateChan chan ConnectionState

	//active websocket, assigned durinng connection process.
	socket *websocket.Conn

	//internal pipe for server messages
	responseChannels     map[string]chan *serverMessage
	responseChannelMutex sync.RWMutex

	//pipes read by lib consumers:

	//channel used to read errors.
	errChan chan error

	//channel used to read message stream from peer
	messageChan chan MessageDataPayload

	//external pipe for server messages
	/*routedMessageChan      chan interface{}
	routedMessageChanMutex sync.RWMutex */
}

//meant to be run as a goroutine
func (c *client) trackState() {
	for newState := range c.stateChan {
		c.stateMutex.Lock()
		c.state = newState
		c.stateMutex.Unlock()
	}
}

//listenToWebSocketData receives all signals from the current websocket.  does not handle socket signal timeout logic AFAIK
func (c *client) listenToWebSocketData(timeout time.Duration) {
	var (
		data    []byte
		message serverMessage
	)

	for {
		c.socket.SetReadDeadline(time.Now().Add(timeout))
		if err := c.socket.ReadJSON(&message); err != nil {
			switch err.(type) {
			case *json.UnmarshalTypeError:
				c.errChan <- SocketError(fmt.Sprintf("Unable to parse message from payload %s: %s\n", string(data), err.Error()))
				continue
			case net.Error:
				c.errChan <- TimeoutError(fmt.Sprintf("Keepalive timeout reached: %s", err.Error()))
			default:
				c.errChan <- SocketError(err.Error())
			}

			c.stateChan <- Disconnected
			return
		}

		//no error
		c.dispatchMessage(&message)
	}

}

func (c *client) dispatchMessage(msg *serverMessage) {
	if len(msg.Identifier) > 0 {
		c.responseChan(msg.Identifier) <- msg
		c.delResponseChan(msg.Identifier)
	} else {
		c.responseChan("default") <- msg
	}
}

func (c *client) responseChan(key string) chan *serverMessage {
	c.responseChannelMutex.RLock()
	defer c.responseChannelMutex.RUnlock()

	return c.responseChannels[key]
}

func (c *client) delResponseChan(key string) {
	c.responseChannelMutex.Lock()
	defer c.responseChannelMutex.Unlock()

	if rc, ok := c.responseChannels[key]; ok {
		close(rc)
		delete(c.responseChannels, key)
	}
}

func (c *client) State() ConnectionState {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()

	return c.state
}

//New generates a new client based on user data.  Specifying an invalid url will not fail until the connection steps.
func New(c Config) Connection {
	// Don't care what the prior scheme was.  Force HTTPS.
	c.ConnectionURL.Scheme = "https"

	if c.NegotiatePath == "" {
		c.NegotiatePath = negotiatePath
	}

	if c.ConnectPath == "" {
		c.ConnectPath = connectPath
	}

	if c.ReconnectPath == "" {
		c.ReconnectPath = reconnectPath
	}

	if c.Client == nil {
		c.Client = &http.Client{}
	}

	new := &client{
		config:    c,
		state:     Ready,
		stateChan: make(chan ConnectionState, 1),
		errChan:   make(chan error, 2),
		responseChannels: map[string]chan *serverMessage{
			"default": make(chan *serverMessage),
		},
	}

	go new.trackState()

	return new
}
