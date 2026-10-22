package tv

import (
	"github.com/chinmay-sawant/ownframe/examples/lg-remote/ws"
	"time"
)

func (c *Client) pointerFailed() bool {
	select {
	case <-c.pointerDone:
		return true
	default:
		return false
	}
}

func watchPointer(conn *ws.Conn) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { _ = conn.Close() }()
		for {
			if _, err := conn.ReadText(); err != nil {
				return
			}
		}
	}()
	go func() {
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		for {
			select {
			case <-done:
				return
			case <-tick.C:
				if conn.Ping() != nil {
					_ = conn.Close()
					return
				}
			}
		}
	}()
	return done
}
