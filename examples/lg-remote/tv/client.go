package tv

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/lg-remote/ws"
	"github.com/chinmay-sawant/ownframe/internal/secret"
)

// Client is one paired SSAP connection and its pointer socket.
// Key stays a secret.Secret so logs and %v dumps cannot leak it.
type Client struct {
	main        *ws.Conn
	in          *ws.Conn
	mu          sync.Mutex
	wait        map[string]chan Message
	dead        error
	n           int
	Key         secret.Secret
	Model       string
	hotstarID   string
	MACs        []string
	pointerAt   time.Time
	pointerDone <-chan struct{}
}

func (c *Client) next() string {
	c.mu.Lock()
	c.n++
	n := c.n
	c.mu.Unlock()

	return strconv.Itoa(n)
}

func (c *Client) Alive() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.dead == nil && c.main != nil
}

// Close drops both sockets.
func (c *Client) Close() {
	if c.main != nil {
		_ = c.main.Close()
	}

	if c.in != nil {
		_ = c.in.Close()
	}
}

func (c *Client) send(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return c.main.WriteText(b)
}

func (c *Client) fail(err error) {
	c.mu.Lock()
	if c.dead == nil {
		c.dead = err
	}

	wait := make([]chan Message, 0, len(c.wait))
	for _, ch := range c.wait {
		wait = append(wait, ch)
	}
	c.mu.Unlock()

	note := Message{Type: "error", Error: "the TV closed the connection"}
	for _, ch := range wait {
		select {
		case ch <- note:
		default:
		}
	}
}
