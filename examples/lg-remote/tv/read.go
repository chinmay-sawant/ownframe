package tv

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/chinmay-sawant/ownframe/examples/lg-remote/ws"
)

func (c *Client) read(conn *ws.Conn) {
	for {
		b, err := conn.ReadText()
		if err != nil {
			c.fail(err)
			return
		}

		var m Message
		if json.Unmarshal(b, &m) != nil {
			continue
		}

		c.mu.Lock()
		ch := c.wait[string(m.ID)]
		c.mu.Unlock()
		if ch == nil {
			continue
		}

		select {
		case ch <- m:
		default:
		}
	}
}

func (c *Client) round(ctx context.Context, id string, v any) (Message, error) {
	ch := make(chan Message, 1)
	if err := c.arm(id, ch); err != nil {
		return Message{}, err
	}

	defer c.drop(id)

	if err := c.send(v); err != nil {
		return Message{}, err
	}

	select {
	case <-ctx.Done():
		return Message{}, errors.New("the TV did not answer")
	case m := <-ch:
		if m.Type == "error" && m.Error == "the TV closed the connection" {
			return Message{}, errors.New(m.Error)
		}

		return m, nil
	}
}

func (c *Client) drop(id string) {
	c.mu.Lock()
	delete(c.wait, id)
	c.mu.Unlock()
}

func (c *Client) Request(uri string, payload map[string]any) (map[string]any, error) {
	if payload == nil {
		payload = map[string]any{}
	}

	msg := map[string]any{
		"id":      c.next(),
		"type":    "request",
		"uri":     "ssap://" + uri,
		"payload": payload,
	}
	ctx, cancel := context.WithTimeout(context.Background(), roundTimeout)
	defer cancel()

	id, ok := msg["id"].(string)
	if !ok {
		return nil, errors.New("the request id is not a string")
	}
	m, err := c.round(ctx, id, msg)
	if err != nil {
		return nil, err
	}

	if m.Type == "error" {
		if m.Error != "" {
			return nil, errors.New(m.Error)
		}

		return nil, errors.New("the TV rejected the control")
	}

	p := mapOf(m.Payload)
	if v, ok := p["returnValue"].(bool); ok && !v {
		return nil, errors.New("the TV rejected the control")
	}

	return p, nil
}
