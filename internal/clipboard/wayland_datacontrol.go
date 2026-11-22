//go:build linux && !android

// Pure-Go Wayland clipboard over ext_data_control_v1 or zwlr_data_control_v1.
// Core wl_data_device needs a focused surface and an input serial owned by the
// window layer, so this client only uses the data-control protocols, which a
// background client may use. Compositors without them (GNOME Mutter, Weston)
// make the native path report "not handled" and the caller falls back to
// wl-copy / wl-paste.
package clipboard

import (
	"context"
	"errors"
	"io"
	"os"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

const (
	wlRegistryID = 2
	wlReadLimit  = 32 << 20
)

// Opcodes shared by ext_data_control_v1 and zwlr_data_control_v1 (the two
// protocols are wire-identical for the requests and events used here).
const (
	wlReqSync        = 0
	wlReqGetRegistry = 1
	wlReqBind        = 0

	wlReqCreateSource = 0
	wlReqGetDevice    = 1
	wlReqSetSelection = 0
	wlReqOfferReceive = 0
	wlReqSourceOffer  = 0

	wlEvError    = 0
	wlEvGlobal   = 0
	wlEvDone     = 0
	wlEvDataOffr = 0
	wlEvSelect   = 1
	wlEvOffer    = 0
	wlEvSend     = 0
	wlEvCancel   = 1
)

// wlManagerNames lists the data-control manager globals this client binds,
// in preference order.
func wlManagerNames() []string {
	return []string{"ext_data_control_manager_v1", "zwlr_data_control_manager_v1"}
}

// wlTextMimeTypes lists the text MIME types this client offers and accepts,
// in preference order.
func wlTextMimeTypes() []string {
	return []string{"text/plain;charset=utf-8", "text/plain", "UTF8_STRING", "STRING", "TEXT"}
}

// wlWriteTimeout bounds how long a paste waits for the receiving client.
const wlWriteTimeout = 5 * time.Second

var errWLUnsupported = errors.New("wayland: no data-control protocol")

type wlGlobal struct{ name, version uint32 }

// wlSession is one connection with the manager, seat and device bound.
type wlSession struct {
	conn    *wlConn
	globals map[string]wlGlobal
	manager uint32
	device  uint32
	source  uint32
	text    string

	offers    map[uint32][]string
	selection uint32
	gotSel    bool
	cancelled atomic.Bool
	callbacks map[uint32]bool
	err       error
}

func openWLSession(ctx context.Context, display, runtimeDir string, deadline time.Time) (*wlSession, error) {
	path, ok := wlSocketPath(display, runtimeDir)
	if !ok {
		return nil, errWLUnsupported
	}
	conn, err := wlDial(ctx, path, time.Until(deadline))
	if err != nil {
		return nil, err
	}
	_ = conn.c.SetDeadline(deadline)

	s := &wlSession{
		conn:      conn,
		globals:   map[string]wlGlobal{},
		offers:    map[uint32][]string{},
		callbacks: map[uint32]bool{},
	}
	if err := s.setup(); err != nil {
		conn.close()
		return nil, err
	}

	return s, nil
}

func (s *wlSession) setup() error {
	reg := s.conn.newID()
	if reg != wlRegistryID {
		return errWLProtocol
	}
	if err := s.conn.send(1, wlReqGetRegistry, wlArgs(nil).u32(reg)); err != nil {
		return err
	}
	if err := s.roundtrip(); err != nil {
		return err
	}

	var iface string
	for _, name := range wlManagerNames() {
		if _, ok := s.globals[name]; ok {
			iface = name
			break
		}
	}
	seat, hasSeat := s.globals["wl_seat"]
	if iface == "" || !hasSeat {
		return errWLUnsupported
	}

	s.manager = s.bind(s.globals[iface], iface)
	seatID := s.bind(seat, "wl_seat")
	s.device = s.conn.newID()
	if err := s.conn.send(s.manager, wlReqGetDevice, wlArgs(nil).u32(s.device).u32(seatID)); err != nil {
		return err
	}

	return s.err
}

func (s *wlSession) bind(g wlGlobal, iface string) uint32 {
	id := s.conn.newID()
	args := wlArgs(nil).u32(g.name).str(iface).u32(1).u32(id)
	if err := s.conn.send(wlRegistryID, wlReqBind, args); err != nil {
		s.err = err
	}

	return id
}

// roundtrip dispatches events until the compositor answers a sync.
func (s *wlSession) roundtrip() error {
	cb := s.conn.newID()
	if err := s.conn.send(1, wlReqSync, wlArgs(nil).u32(cb)); err != nil {
		return err
	}
	for !s.callbacks[cb] {
		if err := s.step(); err != nil {
			return err
		}
	}
	delete(s.callbacks, cb)

	return nil
}

// step reads and dispatches one event.
func (s *wlSession) step() error {
	if s.err != nil {
		return s.err
	}
	obj, op, body, err := s.conn.recv()
	if err != nil {
		return err
	}
	s.dispatch(obj, op, body)

	return s.err
}

func (s *wlSession) dispatch(obj, op uint32, body []byte) {
	r := &wlReader{b: body}
	switch {
	case obj == 1 && op == wlEvError:
		_ = r.u32()
		code := r.u32()
		s.err = wlErrorf("compositor error %d: %s", code, r.str())
	case obj == wlRegistryID && op == wlEvGlobal:
		name := r.u32()
		iface := r.str()
		ver := r.u32()
		if r.err == nil {
			if _, dup := s.globals[iface]; !dup {
				s.globals[iface] = wlGlobal{name, ver}
			}
		}
	case obj == s.device && s.device != 0:
		s.deviceEvent(op, r)
	case obj == s.source && s.source != 0:
		s.sourceEvent(op, r)
	default:
		if _, isOffer := s.offers[obj]; isOffer {
			if op == wlEvOffer {
				s.offers[obj] = append(s.offers[obj], r.str())
			}
		} else if op == wlEvDone {
			s.callbacks[obj] = true
		}
	}
}

func (s *wlSession) deviceEvent(op uint32, r *wlReader) {
	switch op {
	case wlEvDataOffr:
		s.offers[r.u32()] = nil
	case wlEvSelect:
		id := r.u32()
		if r.err == nil {
			s.selection = id
			s.gotSel = true
		}
	}
}

func (s *wlSession) sourceEvent(op uint32, r *wlReader) {
	switch op {
	case wlEvSend:
		_ = r.str()
		fd := s.conn.popFD()
		if fd != nil {
			go func(text string) {
				_ = fd.SetWriteDeadline(time.Now().Add(wlWriteTimeout))
				_, _ = io.WriteString(fd, text)
				_ = fd.Close()
			}(s.text)
		}
	case wlEvCancel:
		s.cancelled.Store(true)
	}
}

func pickWLMime(mimes []string) (string, bool) {
	for _, want := range wlTextMimeTypes() {
		if slices.Contains(mimes, want) {
			return want, true
		}
	}

	return "", false
}

// readText returns the selection text. ok is false when the selection is
// empty or has no text type.
func (s *wlSession) readText() (string, bool, error) {
	for !s.gotSel {
		if err := s.step(); err != nil {
			return "", false, err
		}
	}
	if s.selection == 0 {
		return "", false, nil
	}
	// Offer events precede the selection event; one more roundtrip makes
	// sure none are still in flight.
	if err := s.roundtrip(); err != nil {
		return "", false, err
	}
	mime, ok := pickWLMime(s.offers[s.selection])
	if !ok {
		return "", false, nil
	}

	pr, pw, err := os.Pipe()
	if err != nil {
		return "", false, err
	}
	defer func() { _ = pr.Close() }()

	args := wlArgs(nil).str(mime)
	sendErr := s.conn.send(s.selection, wlReqOfferReceive, args, int(pw.Fd()))
	_ = pw.Close()
	if sendErr != nil {
		return "", false, sendErr
	}

	_ = pr.SetReadDeadline(time.Now().Add(wlWait))
	data, err := io.ReadAll(io.LimitReader(pr, wlReadLimit+1))
	if err != nil {
		return "", false, err
	}
	if len(data) > wlReadLimit {
		return "", false, wlErrorf("selection larger than %d bytes", wlReadLimit)
	}

	return string(data), true, nil
}

// offerText makes this session the selection owner.
func (s *wlSession) offerText(text string) error {
	s.text = text
	s.source = s.conn.newID()
	if err := s.conn.send(s.manager, wlReqCreateSource, wlArgs(nil).u32(s.source)); err != nil {
		return err
	}
	for _, m := range wlTextMimeTypes() {
		if err := s.conn.send(s.source, wlReqSourceOffer, wlArgs(nil).str(m)); err != nil {
			return err
		}
	}
	if err := s.conn.send(s.device, wlReqSetSelection, wlArgs(nil).u32(s.source)); err != nil {
		return err
	}

	return s.roundtrip()
}

func (s *wlSession) clearSelection() error {
	if err := s.conn.send(s.device, wlReqSetSelection, wlArgs(nil).u32(0)); err != nil {
		return err
	}

	return s.roundtrip()
}

// serve answers send requests until the source is cancelled or the
// connection closes.
func (s *wlSession) serve() {
	_ = s.conn.c.SetDeadline(time.Time{})
	for !s.cancelled.Load() {
		if s.step() != nil {
			break
		}
	}
	s.conn.close()
}

// The offered selection must outlive the call that created it: the source
// answers other clients' paste requests until another client replaces it, so
// one process-wide holder is the design.
var (
	wlHeldMu sync.Mutex //nolint:gochecknoglobals // reason: the held source outlives the call that created it.
	wlHeld   *wlSession //nolint:gochecknoglobals // reason: the held source outlives the call that created it.
)

func wlNativeSession(ctx context.Context) (*wlSession, error) {
	deadline := time.Now().Add(wlWait)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}

	return openWLSession(ctx, os.Getenv("WAYLAND_DISPLAY"), os.Getenv("XDG_RUNTIME_DIR"), deadline)
}

// wlNativeWrite reports false when the compositor lacks data-control or the
// socket is unreachable, so the caller can try wl-copy. The selection is
// served from this process until another client replaces it.
func wlNativeWrite(ctx context.Context, text string) bool {
	s, err := wlNativeSession(ctx)
	if err != nil {
		return false
	}
	if err := s.offerText(text); err != nil {
		s.conn.close()
		return false
	}

	wlHeldMu.Lock()
	old := wlHeld
	wlHeld = s
	wlHeldMu.Unlock()
	if old != nil {
		old.conn.close()
	}
	go s.serve()

	return true
}

// wlNativeRead returns handled=true when the native client answered, even
// with an empty selection (ok false).
func wlNativeRead(ctx context.Context) (text string, ok, handled bool) {
	s, err := wlNativeSession(ctx)
	if err != nil {
		return "", false, false
	}
	defer s.conn.close()

	text, ok, err = s.readText()
	if err != nil {
		return "", false, false
	}

	return text, ok, true
}

func wlNativeClear() bool {
	wlHeldMu.Lock()
	old := wlHeld
	wlHeld = nil
	wlHeldMu.Unlock()
	if old != nil {
		old.conn.close()
	}

	s, err := wlNativeSession(context.Background())
	if err != nil {
		return false
	}
	defer s.conn.close()

	return s.clearSelection() == nil
}

// wlNativeAvailable probes the registry without touching the clipboard.
func wlNativeAvailable() bool {
	s, err := wlNativeSession(context.Background())
	if err != nil {
		return false
	}
	s.conn.close()

	return true
}
