package tv

import (
	"errors"
	"time"
)

func (c *Client) pair(old string) (string, error) {
	ch := make(chan Message, pairChanSize)
	if err := c.arm("register_0", ch); err != nil {
		return "", err
	}

	defer c.drop("register_0")

	if err := c.send(registerMsg(old)); err != nil {
		return "", err
	}

	timer := time.NewTimer(pairTimeout)
	defer timer.Stop()

	prompted := false

	for {
		select {
		case <-timer.C:
			if prompted {
				return "", errors.New("accept the prompt on the TV")
			}

			return "", errors.New("the TV did not pair")
		case m := <-ch:
			if m.Type == "registered" {
				if k := clientKey(m.Payload); k != "" {
					return k, nil
				}

				if old != "" {
					return old, nil
				}
			}

			if m.Type == "error" {
				if m.Error != "" {
					return "", errors.New(m.Error)
				}

				return "", errors.New("the TV refused pairing")
			}

			prompted = true
		}
	}
}

// Button sends one remote key on the pointer socket.
func (c *Client) Button(name string) error {
	return c.writePointer("type:button\nname:" + name + "\n\n")
}
