package ws

import (
	"io"
	"time"
)

// WriteText sends one masked text frame.
func (c *Conn) WriteText(p []byte) error {
	if len(p) > maxMsgLen {
		return ErrTooLarge
	}

	c.wmu <- struct{}{}
	defer func() { <-c.wmu }()
	_ = c.c.SetWriteDeadline(time.Now().Add(writeTimeout))

	return writeFrame(c.c, opcodeText, p, true)
}

// ReadText returns the next text payload. Ping frames get a pong.
// Every frame carries a read deadline so a stalled peer cannot park
// the reader, and the reassembled message is capped at maxMsgLen.
func (c *Conn) ReadText() ([]byte, error) {
	var buf []byte

	for {
		_ = c.c.SetReadDeadline(time.Now().Add(readTimeout))

		fin, op, payload, err := readFrame(c.r)
		if err != nil {
			return nil, err
		}

		switch op {
		case opcodeText, opcodeCont:
			if len(buf)+len(payload) > maxMsgLen {
				return nil, ErrTooLarge
			}

			buf = append(buf, payload...)
			if fin {
				_ = c.c.SetReadDeadline(time.Time{})

				return buf, nil
			}
		case opcodeClose:
			return nil, io.EOF
		case opcodePing:
			if len(payload) > maxCtrlLen {
				return nil, ErrTooLarge
			}

			if err := c.writeCtrl(opcodePong, payload); err != nil {
				return nil, err
			}
		case opcodePong:
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
	if len(p) > maxCtrlLen {
		return ErrTooLarge
	}

	c.wmu <- struct{}{}
	defer func() { <-c.wmu }()
	_ = c.c.SetWriteDeadline(time.Now().Add(writeTimeout))

	return writeFrame(c.c, op, p, true)
}

// Ping keeps the pointer connection active without sending a TV command.
func (c *Conn) Ping() error { return c.writeCtrl(opcodePing, nil) }
