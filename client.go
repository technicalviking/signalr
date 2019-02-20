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
	negotiatePath string = "negotiate"
	connectPath   string = "connect"
	reconnectPath string = "reconnect"
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
	ConnectionURL *url.URL `json:"url"`

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

//MessageDataPayload contains information from signalR peer based on subscription
type MessageDataPayload struct {
	HubName   string            `json:"H"`
	Method    string            `json:"M"`
	Arguments []json.RawMessage `json:"A"`
}

//client implemntation of Connection interface.
type client struct {
	//persist sanitized config
	config Config

	//store current state of connection
	state ConnectionState
	//mutex to make changes to state threadsafe
	stateMutex sync.RWMutex

	//active websocket, assigned durinng connection process.
	socket *websocket.Conn
	//writing to the signalr websocket should be a threadsafe operation to conform to gorlla/websocket docs re: one go-routine for writing.
	socketWriteMutex sync.Mutex

	//internal pipe for server messages
	responseChannels     map[string]chan *serverMessage
	responseChannelMutex sync.RWMutex

	nextID         int
	callHubIDMutex sync.Mutex

	//pipes read by lib consumers:

	//channel used to read errors.
	errChan      chan error
	errChanMutex sync.Mutex

	//channel used to read message stream from peer
	messageChan      chan MessageDataPayload
	messageChanMutex sync.Mutex

	//channel used to listen to heartbeat updates.
	heartbeatChan      chan Heartbeat
	heartbeatChanMutex sync.Mutex

	//external pipe for server messages
	/*routedMessageChan      chan interface{}
	routedMessageChanMutex sync.RWMutex */
}

func (c *client) setState(newState ConnectionState) {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()

	//cannot change state once broken.
	if c.state < Broken {
		c.state = newState
	}
}

func (c *client) State() ConnectionState {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()

	return c.state
}

//listenToWebSocketData receives all signals from the current websocket.  uses timeout based on signalr negotiation response
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
				c.sendErr(SocketError(fmt.Sprintf("Unable to parse message from payload %s: %s\n", string(data), err.Error())))
				continue
			case net.Error:
				c.sendErr(TimeoutError(fmt.Sprintf("Keepalive timeout reached: %s", err.Error())))
			default:
				c.sendErr(SocketError(err.Error()))
			}

			c.setState(Disconnected)
			return
		}

		//no error
		go c.dispatchMessage(message)
	}

}

func (c *client) dispatchMessage(msg serverMessage) {

	if msg.Error != "" {
		c.sendErr(HubMessageError(fmt.Sprintf("Error from signalr hub: %s", msg.Error)))
		return
	}

	if len(msg.Identifier) > 0 {

		if rc := c.responseChan(msg.Identifier); rc != nil {
			rc <- &msg
			c.delResponseChan(msg.Identifier)
		} else if len(msg.Data) > 0 { //if "Data" is not empty, presume it's a subscription response.
			for dataIndex := range msg.Data {
				var dataPayload MessageDataPayload
				if parseErr := json.Unmarshal(msg.Data[dataIndex], &dataPayload); parseErr != nil {
					c.errChan <- parseErr
				} else {
					c.messageChan <- dataPayload
					c.heartbeatChan <- NormalHeartbeat("Heartbeat refreshed by subscription signal.")
				}
			}
		} else {
			c.heartbeatChan <- (AwkwardHeartbeat(fmt.Sprintf("No listener found for message with ID %s: %+v", msg.Identifier, msg)))
		}
	} else {
		c.heartbeatChan <- NormalHeartbeat("Default Heartbeat.")
	}
}

func (c *client) responseChan(key string) chan *serverMessage {
	c.responseChannelMutex.RLock()
	defer c.responseChannelMutex.RUnlock()

	return c.responseChannels[key]
}

func (c *client) setResponseChan(key string) {
	c.responseChannelMutex.Lock()
	defer c.responseChannelMutex.Unlock()

	c.responseChannels[key] = make(chan *serverMessage)
}

func (c *client) delResponseChan(key string) {
	c.responseChannelMutex.Lock()
	defer c.responseChannelMutex.Unlock()

	if rc, ok := c.responseChannels[key]; ok {
		close(rc)
		delete(c.responseChannels, key)
	}
}

func (c *client) sendErr(err error) {
	c.errChanMutex.Lock()
	defer c.errChanMutex.Unlock()

	c.errChan <- err
}

// ListenToErrors get reference to errors generated by signalr peer
func (c *client) ListenToErrors() <-chan error {
	return c.errChan
}

// ListenToHubResponses get reference to messages generated by signalr peer
func (c *client) ListenToHubResponses() <-chan MessageDataPayload {
	return c.messageChan
}

// ListenToHeartbeat get reference to heartbeat messages generated by signalr peer
func (c *client) ListenToHeartbeat() <-chan Heartbeat {
	if c.heartbeatChan == nil {
		c.heartbeatChan = make(chan Heartbeat)
	}

	return c.heartbeatChan
}

//New generates a new client based on user data.  Specifying an invalid url will not fail until the connection steps.
func New(c Config) Connection {

	if c.ConnectionURL == nil {
		c.ConnectionURL = &url.URL{}
	}

	// Don't care what the prior scheme was.  Force HTTPS.
	c.ConnectionURL.Scheme = "https"

	if c.ConnectionURL.Host == "" {
		c.ConnectionURL.Host = "localhost:1337"
	}

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
		config:           c,
		state:            Ready,
		nextID:           1,
		errChan:          make(chan error, 5),
		messageChan:      make(chan MessageDataPayload),
		responseChannels: map[string]chan *serverMessage{},
	}

	return new
}
