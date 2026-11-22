//go:build linux && !android

package clipboard

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// Minimal Wayland wire client: just enough framing, string/uint arguments and
// SCM_RIGHTS fd passing for the data-control clipboard protocols. No cgo, no
// libwayland, no external module.

var errWLProtocol = errors.New("wayland: malformed message")

const (
	wlMaxMsg   = 1 << 16
	wlOOBSpace = 256
	// wlHeaderSize is the object id plus the size and opcode word.
	wlHeaderSize = 8
	// wlWordSize is the 32-bit unit every argument and message pads to.
	wlWordSize   = 4
	wlPadMask    = wlWordSize - 1
	wlSizeShift  = 16
	wlOpcodeMask = 0xffff
	wlReadChunk  = 4096
	// wlFirstClientID is the first object id a client allocates; id 1 is
	// the wl_display.
	wlFirstClientID = 2
)

// wlConn frames Wayland messages over a unix stream socket.
type wlConn struct {
	c      *net.UnixConn
	rbuf   []byte
	fds    []*os.File
	nextID uint32
}

func wlSocketPath(display, runtimeDir string) (string, bool) {
	if display == "" {
		return "", false
	}
	if filepath.IsAbs(display) {
		return display, true
	}
	if runtimeDir == "" {
		return "", false
	}

	return filepath.Join(runtimeDir, display), true
}

func wlDial(ctx context.Context, path string, timeout time.Duration) (*wlConn, error) {
	c, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "unix", path)
	if err != nil {
		return nil, err
	}
	uc, ok := c.(*net.UnixConn)
	if !ok {
		_ = c.Close()
		return nil, errors.New("wayland: not a unix connection")
	}

	return &wlConn{c: uc, nextID: wlFirstClientID}, nil
}

func (w *wlConn) close() {
	_ = w.c.Close()
}

func (w *wlConn) newID() uint32 {
	id := w.nextID
	w.nextID++

	return id
}

// wlArgs builds a request body.
type wlArgs []byte

func (a wlArgs) u32(v uint32) wlArgs { return binary.LittleEndian.AppendUint32(a, v) }

func (a wlArgs) str(s string) wlArgs {
	n := len(s) + 1
	a = a.u32(uint32(n)) //nolint:gosec // reason: protocol strings are short constants.
	a = append(a, s...)
	a = append(a, 0)
	for n%wlWordSize != 0 {
		a = append(a, 0)
		n++
	}

	return a
}

// send writes one request, optionally passing fds as ancillary data.
func (w *wlConn) send(obj, opcode uint32, args wlArgs, fds ...int) error {
	size := wlHeaderSize + len(args)
	if size > wlMaxMsg {
		return errWLProtocol
	}
	msg := make([]byte, 0, size)
	msg = binary.LittleEndian.AppendUint32(msg, obj)
	msg = binary.LittleEndian.AppendUint32(msg, uint32(size)<<wlSizeShift|opcode)
	msg = append(msg, args...)

	var oob []byte
	if len(fds) > 0 {
		oob = syscall.UnixRights(fds...)
	}
	_, _, err := w.c.WriteMsgUnix(msg, oob, nil)

	return err
}

// recv returns the next event. Fds received with it queue on w.fds.
func (w *wlConn) recv() (obj, opcode uint32, body []byte, err error) {
	for {
		if len(w.rbuf) >= wlHeaderSize {
			obj = binary.LittleEndian.Uint32(w.rbuf)
			word := binary.LittleEndian.Uint32(w.rbuf[wlWordSize:])
			size := int(word >> wlSizeShift)
			if size < wlHeaderSize || size%wlWordSize != 0 {
				return 0, 0, nil, errWLProtocol
			}
			if len(w.rbuf) >= size {
				body = append([]byte(nil), w.rbuf[wlHeaderSize:size]...)
				w.rbuf = w.rbuf[size:]

				return obj, word & wlOpcodeMask, body, nil
			}
		}

		buf := make([]byte, wlReadChunk)
		oob := make([]byte, wlOOBSpace)
		n, oobn, _, _, rerr := w.c.ReadMsgUnix(buf, oob)
		if n < 0 {
			n, oobn = 0, 0
		}
		w.takeFDs(oob[:oobn])
		w.rbuf = append(w.rbuf, buf[:n]...)
		if n == 0 && rerr != nil {
			return 0, 0, nil, rerr
		}
		if n == 0 {
			return 0, 0, nil, io.ErrUnexpectedEOF
		}
	}
}

func (w *wlConn) takeFDs(oob []byte) {
	if len(oob) == 0 {
		return
	}
	msgs, err := syscall.ParseSocketControlMessage(oob)
	if err != nil {
		return
	}
	for i := range msgs {
		fds, perr := syscall.ParseUnixRights(&msgs[i])
		if perr != nil {
			continue
		}
		for _, fd := range fds {
			w.fds = append(w.fds, os.NewFile(uintptr(fd), "wl-fd"))
		}
	}
}

func (w *wlConn) popFD() *os.File {
	if len(w.fds) == 0 {
		return nil
	}
	f := w.fds[0]
	w.fds = w.fds[1:]

	return f
}

// wlReader decodes event bodies.
type wlReader struct {
	b   []byte
	err error
}

func (r *wlReader) u32() uint32 {
	if len(r.b) < wlWordSize {
		r.err = errWLProtocol

		return 0
	}
	v := binary.LittleEndian.Uint32(r.b)
	r.b = r.b[wlWordSize:]

	return v
}

func (r *wlReader) str() string {
	n := int(r.u32())
	pad := (n + wlPadMask) &^ wlPadMask
	if r.err != nil || pad > len(r.b) {
		r.err = errWLProtocol

		return ""
	}
	s := r.b[:n]
	r.b = r.b[pad:]
	if n > 0 && s[n-1] == 0 {
		s = s[:n-1]
	}

	return string(s)
}

func wlErrorf(format string, a ...any) error { return fmt.Errorf("wayland: "+format, a...) }
