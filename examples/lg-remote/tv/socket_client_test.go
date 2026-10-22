package tv

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/lg-remote/ws"
)

func testClient(t *testing.T, payload func(map[string]any) map[string]any) *Client {
	t.Helper()
	url := socketURL(t, func(b []byte) []byte {
		var req map[string]any
		if json.Unmarshal(b, &req) != nil {
			return nil
		}
		out, _ := json.Marshal(map[string]any{"id": req["id"], "type": "response", "payload": payload(req)})
		return out
	})
	main, err := ws.Dial(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	c := &Client{main: main, wait: map[string]chan Message{}}
	go c.read(main)
	t.Cleanup(c.Close)
	return c
}
