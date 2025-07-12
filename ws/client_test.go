package ws

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	go newClient(t)
	go newClient(t)

	forever := make(chan bool)
	<-forever
}

func newClient(t *testing.T) {
	client := NewClient("ws://127.0.0.1:8000/ws", nil, Option{Name: "testclient", ReconnectWaitSecond: 0})
	client.SetHandler(func(c *Connect) {
		t.Logf("connect success\n")
	}, func(c *Connect, text string) (interface{}, error) {
		//t.Logf("text:%s\n", text)
		return text, nil
	}, func(c *Connect, binary []byte) (interface{}, error) {
		//t.Logf("binary:%s\n", string(binary))
		return binary, nil
	}, func(c *Connect, response interface{}) {
		//t.Logf("response:%s\n", response)
	})

	client.SetPingHandler(func(c *Connect) {
		t.Logf("%s", "reciver ping")
	})

	client.ConnectAwaitSuccess()
	t.Log("complete")
}
