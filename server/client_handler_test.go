package server

import (
	"reflect"
	"testing"
)

var c1 *fakeClient
var c2 *fakeClient
var clh ClientHandler

func TestBasicClientHandler(t *testing.T) {
	c1 = &fakeClient{host: "foo", port: 1234}
	c2 = &fakeClient{host: "bar", port: 1234}

	clh = DefaultClientHandler()
	err := clh.Accept(c1)
	if err != nil {
		t.Error("Unexpected Error Accepting PeerClient")
	}

	err = clh.Accept(c2)
	if err != nil {
		t.Error("Unexpected Error Accepting PeerClient")
	}

	clients := clh.Clients()
	if len(clients) != 2 {
		t.Error("Unexpected client size ", clh.Clients())
	}
}
func TestErrorOnAddTwiceSameClient(t *testing.T) {
	err := clh.Accept(c2)
	if err == nil {
		t.Error("Unexpected Error Accepting PeerClient", err)
	}
	if err.Error() != "Client: bar:1234 Already registered" {
		t.Error("Unexpected error message", err.Error())
	}
}

func TestToRemoveCLientFromCLientHandler(t *testing.T) {
	err := clh.Remove(c2)
	if err != nil {
		t.Error("Unexpected Error Removing PeerClient", err)
	}

	clients := clh.Clients()
	if len(clients) != 1 {
		t.Error("Unexpected client size ", clh.Clients())
	}
}

func TestErrorOnRemoveInexistentCLient(t *testing.T) {
	err := clh.Remove(c2)
	if err == nil {
		t.Error("Unexpected Error Remove PeerClient", err)
	}
	if err.Error() != "Client Not found" {
		t.Error("Unexpected error message", err.Error())
	}
}

type fakeClient struct {
	host    string
	port    int
	msgChan chan Message
}

func (f *fakeClient) Node() Node {
	return Node{Host: f.host, Port: f.port}
}
func (f *fakeClient) Id() ID {
	return ID(0)
}
func (f *fakeClient) Run() {
}
func (f *fakeClient) Send(m Message) error {
	f.msgChan <- m
	return nil
}
func (f *fakeClient) ReceiveChan() (v chan Message) {
	return f.msgChan
}

func (f *fakeClient) Exit() {
}
func (f *fakeClient) SayHello() {
}
func (f *fakeClient) Identify(n Node) {
}

func (f *fakeClient) Mode() string {
	return ""
}

func (f *fakeClient) From() Node {
	return Node{Host: f.host, Port: f.port}
}

func TestBasicFakeClientTest(t *testing.T) {
	ch := make(chan Message, 10)
	fkc := &fakeClient{"localhost", 9000, ch}

	msg := Hello{
		Id:      999,
		From:    Node{"localhost", 9000},
		Details: map[string]interface{}{"foo": "bar"},
	}

	fkc.Send(msg)

	rcvMsg := <-fkc.ReceiveChan()
	if !reflect.DeepEqual(msg, rcvMsg) {
		t.Errorf("Expected %s, got %s", msg, rcvMsg)
	}
}
