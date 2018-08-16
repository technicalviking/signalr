package signalr

import (
	"net/http"
	"net/url"
	"sync"

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
	Disconnected ConnectionState = iota
	Connecting
	Reconnecting
	Connected
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
	//channel used to read errors.
	errChan chan error

	socket *websocket.Conn
}

//meant to be run as a goroutine
func (c *client) trackState() {
	for newState := range c.stateChan {
		c.stateMutex.Lock()
		c.state = newState
		c.stateMutex.Unlock()
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
		state:     Disconnected,
		stateChan: make(chan ConnectionState, 1),
		errChan:   make(chan error, 2),
	}

	go new.trackState()

	return new
}
