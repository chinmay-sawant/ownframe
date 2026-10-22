package ws

import (
	"io"
	"time"
)

// WriteText sends one masked text frame.
func (c *Conn) WriteText(p []byte) error {
	c.wmu <- struct{}{}
	defer func() { <-c.wmu }()
	_ = c.c.SetWriteDeadline(time.Now().Add(writeTimeout))

	return writeFrame(c.c, opcodeText, p, true)
}

// ReadText returns the next text payload. Ping frames get a pong.
func (c *Conn) ReadText() ([]byte, error) {
	var buf []byte

	for {
		fin, op, payload, err := readFrame(c.r)
		if err != nil {
			return nil, err
		}

		switch op {
		case opcodeText, opcodeCont:
			buf = append(buf, payload...)
			if fin {
				return buf, nil
			}
		case opcodeClose:
			return nil, io.EOF
		case opcodePing:
			if err := c.writeCtrl(opcodePong, payload); err != nil {
				return nil, err
			}
		default:
		}
	}
}

// Close sends a close frame and closes the socket.
func (c *Conn) Close() error {
	_ = c.writeCtrl(opcodeClose, nil)

	return c.c.Close()
}

func (c *Conn) writeCtrl(op byte, p []byte) error {
	c.wmu <- struct{}{}
	defer func() { <-c.wmu }()
	_ = c.c.SetWriteDeadline(time.Now().Add(writeTimeout))

	return writeFrame(c.c, op, p, true)
}

// Ping keeps the pointer connection active without sending a TV command.
func (c *Conn) Ping() error { return c.writeCtrl(opcodePing, nil) }
