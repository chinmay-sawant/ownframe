package tv

import "errors"

// arm registers a waiter for one request id. A closed connection fails
// the call instead of leaking the waiter.
func (c *Client) arm(id string, ch chan Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.dead != nil {
		return errors.New("the TV closed the connection")
	}

	c.wait[id] = ch

	return nil
}
