# SignalR

SignalR Connection Interface Written in Go

Forked from github.com/hweom/signalr.  Credit where credit is due.

## Installation

SignalR client can be included in any project:

`go get gitlab.com/techviking/signalr`

## Usage

### Import

This part is typical of most libraries.

    import (gitlab.com/techviking/signalr)

Each SignalR connection is represented by a seperate instance.  (THIS IS NOT A SINGLETON!)  To simplify some things, the parameter for the constructor is a `Config` struct.


### Instantiation
    peerURL, _ := url.Parse("https://fakesignalrpeer.com")
    
    //basic config
    cfg := signalr.Config{
      ConnectionURL: peerURL
    }

The config exposes fields allowing you to set a non-default `*http.Client` and `http.Header` if special work is required to connect to the original endpoint, and `...Path` fields to override the default endpoints for negotiation, connection, or reconnection if necessary.  I imagine most implementers won't need this, but it doesn't hurt.  The constructor takes a value, not a reference, so that if it needs defaults it doesn't mutate your input in a way that you might not want.

    client := signalr.New(cfg)


client is of type signalr.Connection, which is an `interface` type.  This should make mocking it out in your application tests easier.  (and also gives me an easier way to ensure the external interface is sufficient for user needs)

### Everything is done with channels!

Before you connect, you're going to want start your listeners:

    errChan := client.ListenToErrors() 
    dataChan := client.ListenToHubResponses()

The channels returned are buffered channels, so that rather than hanging, the code will intentionally be placed into a potential `panic` condition once they fill up.

### Okay, not QUITE everything...

The one exception to the above rule is scenarios where calls to `CallHub` result in an immediate response, detected when the response payload from the signalr peer has a matching identifier to a sent message.  In that scenario, the result message is sent back as a return fromm `CallHub`... unless it's an error.

### I hate `if err != nil`... so I don't ask you to use it.

That errChan above?  It gives ALL the errors from the lib, whether or not they're from the signalr peer.  You might be wondering if that makes the application more difficult to debug, but I've tried to cover that front too.  The error chan will NEVER output *just* an error.  It will be one of the error types defined in this lib's `errors.go` file.  If you're not sure which one you've got, you can do a type switch, or alternately parse the resulting string (each type's implementation of the Error interface indicates its own type).

### Miscellaneous

The websocket library used under the hood is github.com/gorilla/websocket.

Due to my own personal requirements using this lib, I've made efforts to have instances of the connection client be threadsafe, as far as being able to send messages from or read messages into any number of goroutines.


## TODOs

* MORE TESTING (especially related to connection resiliency).


## Contributions

Make a Pull Request (or "Merge Request" as Gitlab calls it).  I'm actively using this library, so any pull requests will receive prompt attention from me.
Fork the project and make it your own!  I only ask that if you do that, lemme know!  If a fork includes featuers I don't have time or desire to implement, I'll put a link to the forked repo with full credit!

Happy coding!
