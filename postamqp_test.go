package main

import (
	"bytes"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	//start server at localhost 8100 before running this
	//pumps 200 messages into rabitmq
	buf := bytes.NewBufferString("{\"msg\":\"Hello world\"}")
	f := func() {
		for i := 0; i < 200; i++ {
			resp, err := http.Post("http://127.0.0.1:8100/", "application/json", buf)
			if err != nil {
				t.Errorf("Problem: %s", err)
				break
			}
			if resp.Status != "200 OK" {
				t.Errorf("Unexpected response: %s", resp.Status)
				break
			}
		}
	}
	f()
}
