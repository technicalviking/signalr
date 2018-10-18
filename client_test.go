package signalr

import (
	"net/http"
	"net/url"
	"testing"
	"time"
	//"fmt"
)

// TestNewDefault test the constructor with a default config
func TestNewDefault(t *testing.T) {
	//Assemble
	cfg := Config{}

	//Act
	conn := New(cfg)

	//Assert
	if conn == nil {
		t.Fatal("constructor returning nil")
	}

	cast, ok := conn.(*client)

	if !ok {
		t.Fatalf("constructor returning wierd type: %T", conn)
	}

	if Ready != cast.state {
		t.Errorf("default state expected to be %+v, received %+v", Ready, cast.state)
	}

	if cast.errChan == nil {
		t.Errorf("default err chan expected, <nil> found")
	}

	if cast.responseChannels == nil {
		t.Errorf("default response channels map expected.  <nil> found")
	}

	if defaultChan, ok := cast.responseChannels["default"]; !ok || defaultChan == nil {
		t.Errorf("Invalid default response channel detected. default key val exists: %t, channel val: %+v", ok, defaultChan)
	}

	sanitizedCfg := cast.config

	if sanitizedCfg.Client == nil {
		t.Errorf("default http client expected.  <nil> found.")
	}

	if sanitizedCfg.ConnectionURL.Host != "localhost:1337" {
		t.Errorf("unknown default host: %s", sanitizedCfg.ConnectionURL.Host)
	}

	if sanitizedCfg.NegotiatePath != negotiatePath {
		t.Errorf("default negotiation path - expected %s, found %s", negotiatePath, sanitizedCfg.NegotiatePath)
	}

	if sanitizedCfg.ConnectPath != connectPath {
		t.Errorf("default connection path - expected %s, found %s", connectPath, sanitizedCfg.ConnectPath)
	}

	if sanitizedCfg.ReconnectPath != reconnectPath {
		t.Errorf("default reconnection path - expected %s, found %s", reconnectPath, sanitizedCfg.ReconnectPath)
	}

}

// TestNewCustom test the constructor with a default config
func TestNewCustom(t *testing.T) {
	//Assemble

	connURL, err := url.Parse("https://argle.bargle.horse")

	if err != nil {
		t.Fatalf("unable to parse test url: %s", err.Error())
	}

	cfg := Config{
		Client: &http.Client{
			Timeout: time.Hour,
		},
		ConnectionURL: connURL,
		NegotiatePath: "/gimme",
		ConnectPath:   "plugin",
		ReconnectPath: "comeback",
	}

	//Act
	conn := New(cfg)

	//Assert
	if conn == nil {
		t.Fatal("constructor returning nil")
	}

	cast, ok := conn.(*client)

	if !ok {
		t.Fatalf("constructor returning wierd type: %T", conn)
	}

	if Ready != cast.state {
		t.Errorf("default state expected to be %+v, received %+v", Ready, cast.state)
	}

	if cast.errChan == nil {
		t.Errorf("default err chan expected, <nil> found")
	}

	if cast.responseChannels == nil {
		t.Errorf("default response channels map expected.  <nil> found")
	}

	if defaultChan, ok := cast.responseChannels["default"]; !ok || defaultChan == nil {
		t.Errorf("Invalid default response channel detected. default key val exists: %t, channel val: %+v", ok, defaultChan)
	}

	sanitizedCfg := cast.config

	if sanitizedCfg.Client == nil {
		t.Errorf("default http client expected.  <nil> found.")
	}

	if sanitizedCfg.ConnectionURL.Host != "argle.bargle.horse" {
		t.Errorf("unknown default host: %s", sanitizedCfg.ConnectionURL.Host)
	}

	if sanitizedCfg.NegotiatePath != cfg.NegotiatePath {
		t.Errorf("default negotiation path - expected %s, found %s", cfg.NegotiatePath, sanitizedCfg.NegotiatePath)
	}

	if sanitizedCfg.ConnectPath != cfg.ConnectPath {
		t.Errorf("default connection path - expected %s, found %s", cfg.ConnectPath, sanitizedCfg.ConnectPath)
	}

	if sanitizedCfg.ReconnectPath != cfg.ReconnectPath {
		t.Errorf("default reconnection path - expected %s, found %s", cfg.ReconnectPath, sanitizedCfg.ReconnectPath)
	}

}

func TestsetState(t *testing.T) {
	//Assemble

	cfg := Config{}

	conn := New(cfg) //set's intial state.
	testClient := conn.(*client)

	//Act
	testClient.setState(Connecting)

	//Assert

	if testClient.state != Connecting {
		t.Errorf("setState not acting like a proper setter.  expected %+v, got %+v", Connecting, testClient.state)
	}
}
