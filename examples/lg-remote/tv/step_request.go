package tv

import (
	"errors"
	"strings"
	"time"
)

func (c *Client) startStep(spec string) (func() (string, bool, error), error) {
	kind, name, _ := strings.Cut(spec, ":")
	uri, label := "", ""
	if name != "Up" && name != "Down" {
		return nil, errors.New("unknown step")
	}
	switch kind {
	case "vol":
		uri, label = "audio/volume"+name, "Volume "+strings.ToLower(name)
	case "ch":
		uri, label = "tv/channel"+name, "Channel"
	default:
		return nil, errors.New("unknown step")
	}
	id := c.next()
	ch := make(chan Message, 1)
	if err := c.arm(id, ch); err != nil {
		return nil, err
	}
	if err := c.send(map[string]any{"id": id, "type": "request",
		"uri": "ssap://" + uri, "payload": map[string]any{}}); err != nil {
		c.drop(id)
		return nil, err
	}
	timer := time.NewTimer(stepTimeout)
	return func() (string, bool, error) {
		defer timer.Stop()
		defer c.drop(id)
		select {
		case <-timer.C:
			return "", false, errors.New("the TV did not answer")
		case m := <-ch:
			if m.Type == "error" {
				return "", m.Error != "the TV closed the connection", errors.New(m.Error)
			}
			if v, ok := mapOf(m.Payload)["returnValue"].(bool); ok && !v {
				return "", true, errors.New("the TV rejected the control")
			}
			return label, false, nil
		}
	}, nil
}
