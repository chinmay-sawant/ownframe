//go:build linux && !android

package clipboard

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// fakeComp is a tiny in-process compositor that speaks just enough Wayland
// (registry, seat, data-control manager/device/source/offer) over a unix
// socket. It needs no real compositor and never touches the desktop clipboard.
type fakeComp struct {
	t       *testing.T
	ln      net.Listener
	manager string // interface advertised; empty means no data-control

	mu      sync.Mutex
	seeded  *string
	owner   *fakeClient
	ownerID uint32
	mimes   []string
	clients []*fakeClient
}

type fakeClient struct {
	comp *fakeComp
	conn *wlConn
	ifc  map[uint32]string
	sMim map[uint32][]string
	// serverID allocates ids for server-created objects (offers).
	serverID uint32
}

const fakeOfferBase = 0xff000000

func newFakeComp(t *testing.T, manager string) *fakeComp {
	t.Helper()
	dir := t.TempDir()
	ln, err := net.Listen("unix", filepath.Join(dir, "wayland-fake"))
	if err != nil {
		t.Skipf("unix sockets unavailable: %v", err)
	}
	f := &fakeComp{t: t, ln: ln, manager: manager}
	t.Setenv("WAYLAND_DISPLAY", "wayland-fake")
	t.Setenv("XDG_RUNTIME_DIR", dir)
	t.Setenv("DISPLAY", "")
	t.Setenv("PATH", t.TempDir())
	t.Cleanup(func() {
		_ = ln.Close()
		wlHeldMu.Lock()
		if wlHeld != nil {
			wlHeld.conn.close()
			wlHeld = nil
		}
		wlHeldMu.Unlock()
	})
	go f.accept()

	return f
}

func (f *fakeComp) seed(text string) {
	f.mu.Lock()
	f.seeded = &text
	f.owner = nil
	f.mu.Unlock()
}

func (f *fakeComp) accept() {
	for {
		c, err := f.ln.Accept()
		if err != nil {
			return
		}
		uc := c.(*net.UnixConn)
		cl := &fakeClient{
			comp: f, conn: &wlConn{c: uc}, ifc: map[uint32]string{1: "wl_display"},
			sMim: map[uint32][]string{}, serverID: fakeOfferBase,
		}
		f.mu.Lock()
		f.clients = append(f.clients, cl)
		f.mu.Unlock()
		go cl.run()
	}
}

func (c *fakeClient) run() {
	defer c.conn.close()
	for {
		obj, op, body, err := c.conn.recv()
		if err != nil {
			c.comp.dropOwner(c)
			return
		}
		c.handle(obj, op, &wlReader{b: body})
	}
}

func (f *fakeComp) dropOwner(c *fakeClient) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.owner == c {
		f.owner = nil
	}
}

func (c *fakeClient) handle(obj, op uint32, r *wlReader) {
	switch c.ifc[obj] {
	case "wl_display":
		if op == 0 { // sync
			cb := r.u32()
			_ = c.conn.send(cb, wlEvDone, wlArgs(nil).u32(0))
		} else { // get_registry
			reg := r.u32()
			c.ifc[reg] = "wl_registry"
			_ = c.conn.send(reg, wlEvGlobal, wlArgs(nil).u32(1).str("wl_seat").u32(7))
			if c.comp.manager != "" {
				_ = c.conn.send(reg, wlEvGlobal, wlArgs(nil).u32(2).str(c.comp.manager).u32(1))
			}
		}
	case "wl_registry":
		_ = r.u32()
		iface := r.str()
		_ = r.u32()
		c.ifc[r.u32()] = iface
	case c.comp.manager:
		c.manager(obj, op, r)
	case "device":
		if op == 0 {
			c.setSelection(r.u32())
		}
	case "source":
		if op == 0 {
			c.sMim[obj] = append(c.sMim[obj], r.str())
		}
	case "offer":
		if op == 0 {
			c.receive(r)
		}
	}
}

func (c *fakeClient) manager(_, op uint32, r *wlReader) {
	switch op {
	case wlReqCreateSource:
		c.ifc[r.u32()] = "source"
	case wlReqGetDevice:
		dev := r.u32()
		c.ifc[dev] = "device"
		c.comp.mu.Lock()
		seeded, owner := c.comp.seeded, c.comp.owner
		mimes := c.comp.mimes
		c.comp.mu.Unlock()
		switch {
		case seeded != nil:
			c.sendOffer(dev, []string{"text/plain;charset=utf-8"})
		case owner != nil:
			c.sendOffer(dev, mimes)
		default:
			_ = c.conn.send(dev, wlEvSelect, wlArgs(nil).u32(0))
		}
	}
}

func (c *fakeClient) sendOffer(dev uint32, mimes []string) {
	c.serverID++
	id := c.serverID
	c.ifc[id] = "offer"
	_ = c.conn.send(dev, wlEvDataOffr, wlArgs(nil).u32(id))
	for _, m := range mimes {
		_ = c.conn.send(id, wlEvOffer, wlArgs(nil).str(m))
	}
	_ = c.conn.send(dev, wlEvSelect, wlArgs(nil).u32(id))
}

func (c *fakeClient) setSelection(src uint32) {
	f := c.comp
	f.mu.Lock()
	prev, prevID := f.owner, f.ownerID
	f.seeded, f.owner, f.ownerID = nil, nil, 0
	if src != 0 {
		f.owner, f.ownerID, f.mimes = c, src, c.sMim[src]
	}
	f.mu.Unlock()
	if prev != nil {
		_ = prev.conn.send(prevID, wlEvCancel, nil)
	}
}

func (c *fakeClient) receive(r *wlReader) {
	mime := r.str()
	fd := c.conn.popFD()
	if fd == nil {
		return
	}
	f := c.comp
	f.mu.Lock()
	seeded, owner, ownerID := f.seeded, f.owner, f.ownerID
	f.mu.Unlock()
	switch {
	case seeded != nil:
		_, _ = io.WriteString(fd, *seeded)
		_ = fd.Close()
	case owner != nil:
		_ = owner.conn.send(ownerID, wlEvSend, wlArgs(nil).str(mime), int(fd.Fd())) //nolint:gosec // reason: live pipe fd.
		_ = fd.Close()
	default:
		_ = fd.Close()
	}
}

func TestWLWireEncoding(t *testing.T) {
	a := wlArgs(nil).u32(2)
	if !bytes.Equal(a, []byte{2, 0, 0, 0}) {
		t.Fatalf("u32 = %v", a)
	}
	s := wlArgs(nil).str("wl_seat")
	want := []byte{8, 0, 0, 0, 'w', 'l', '_', 's', 'e', 'a', 't', 0}
	if !bytes.Equal(s, want) {
		t.Fatalf("str = %v", s)
	}
	odd := wlArgs(nil).str("abc")
	if !bytes.Equal(odd, []byte{4, 0, 0, 0, 'a', 'b', 'c', 0}) {
		t.Fatalf("str pad = %v", odd)
	}
	r := &wlReader{b: append(append([]byte{}, s...), 9, 0, 0, 0)}
	if got := r.str(); got != "wl_seat" || r.u32() != 9 || r.err != nil {
		t.Fatalf("reader got %q err %v", got, r.err)
	}
	bad := &wlReader{b: []byte{200, 0, 0, 0, 1}}
	_ = bad.str()
	if bad.err == nil {
		t.Fatal("short string accepted")
	}
}

func TestWLNativeRoundTrip(t *testing.T) {
	for _, mgr := range wlManagers {
		t.Run(mgr, func(t *testing.T) {
			newFakeComp(t, mgr)
			ctx := context.Background()
			for _, text := range []string{"hello", "multi\nline ✓", "", "--flag"} {
				if !wlNativeWrite(ctx, text) {
					t.Fatalf("write %q not handled", text)
				}
				got, ok, handled := wlNativeRead(ctx)
				if !handled || !ok || got != text {
					t.Fatalf("read %q = %q ok=%v handled=%v", text, got, ok, handled)
				}
			}
			if !wlNativeAvailable() {
				t.Fatal("expected available")
			}
		})
	}
}

func TestWLNativeLargeText(t *testing.T) {
	newFakeComp(t, wlManagers[0])
	text := string(bytes.Repeat([]byte("abcdefgh"), 1<<17)) // 1 MiB, exceeds the pipe buffer
	if !wlNativeWrite(context.Background(), text) {
		t.Fatal("write not handled")
	}
	got, ok, handled := wlNativeRead(context.Background())
	if !handled || !ok || got != text {
		t.Fatalf("large read ok=%v handled=%v len=%d", ok, handled, len(got))
	}
}

func TestWLNativeReadsForeignSelection(t *testing.T) {
	f := newFakeComp(t, wlManagers[1])
	f.seed("from another app")
	got, ok, handled := wlNativeRead(context.Background())
	if !handled || !ok || got != "from another app" {
		t.Fatalf("got %q ok=%v handled=%v", got, ok, handled)
	}
}

func TestWLNativeClearAndEmpty(t *testing.T) {
	newFakeComp(t, wlManagers[0])
	ctx := context.Background()
	if _, ok, handled := wlNativeRead(ctx); !handled || ok {
		t.Fatalf("empty selection ok=%v handled=%v", ok, handled)
	}
	wlNativeWrite(ctx, "secret")
	if !wlNativeClear() {
		t.Fatal("clear not handled")
	}
	if got, ok, handled := wlNativeRead(ctx); !handled || ok || got != "" {
		t.Fatalf("after clear %q ok=%v handled=%v", got, ok, handled)
	}
}

func TestWLNativeReplacedSourceIsCancelled(t *testing.T) {
	newFakeComp(t, wlManagers[0])
	ctx := context.Background()
	wlNativeWrite(ctx, "one")
	wlHeldMu.Lock()
	first := wlHeld
	wlHeldMu.Unlock()
	wlNativeWrite(ctx, "two")

	got, ok, _ := wlNativeRead(ctx)
	if !ok || got != "two" {
		t.Fatalf("got %q ok=%v", got, ok)
	}
	deadline := time.Now().Add(2 * time.Second)
	for !first.cancelled && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	// first.conn was closed by the replacement; serve exits either way.
}

func TestWLNativeUnsupportedFallsBack(t *testing.T) {
	newFakeComp(t, "")
	if wlNativeWrite(context.Background(), "x") {
		t.Fatal("write handled without data-control")
	}
	if _, _, handled := wlNativeRead(context.Background()); handled {
		t.Fatal("read handled without data-control")
	}
	if wlNativeAvailable() {
		t.Fatal("available without data-control")
	}
	if st := diagnoseOS(); st.Backend == backendWaylandNative {
		t.Fatalf("Diagnose = %+v", st)
	}
}

func TestWLNativeNoSocket(t *testing.T) {
	t.Setenv("WAYLAND_DISPLAY", "wayland-missing")
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	if wlNativeWrite(context.Background(), "x") || wlNativeAvailable() {
		t.Fatal("handled with no socket")
	}
	t.Setenv("XDG_RUNTIME_DIR", "")
	if _, ok := wlSocketPath("wayland-0", ""); ok {
		t.Fatal("relative display without runtime dir accepted")
	}
	if p, ok := wlSocketPath("/run/x/wl", ""); !ok || p != "/run/x/wl" {
		t.Fatalf("absolute path = %q %v", p, ok)
	}
}

func TestWLNativeThroughRouting(t *testing.T) {
	newFakeComp(t, wlManagers[0])
	ctx := context.Background()
	if !writeOS(ctx, "routed") {
		t.Fatal("writeOS failed")
	}
	if got, ok := readOS(ctx); !ok || got != "routed" {
		t.Fatalf("readOS = %q %v", got, ok)
	}
	if st := diagnoseOS(); st.Backend != backendWaylandNative || !st.Available {
		t.Fatalf("Diagnose = %+v", st)
	}
	clearOS()
	if got, ok := readOS(ctx); ok || got != "" {
		t.Fatalf("after clear %q %v", got, ok)
	}
}

func TestWLNativeCompositorErrorFallsBack(t *testing.T) {
	dir := t.TempDir()
	ln, err := net.Listen("unix", filepath.Join(dir, "wayland-bad"))
	if err != nil {
		t.Skip(err)
	}
	defer ln.Close()
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		conn := &wlConn{c: c.(*net.UnixConn)}
		_, _, _, _ = conn.recv()
		_ = conn.send(1, wlEvError, wlArgs(nil).u32(1).u32(0).str("boom"))
	}()
	t.Setenv("WAYLAND_DISPLAY", "wayland-bad")
	t.Setenv("XDG_RUNTIME_DIR", dir)
	if wlNativeWrite(context.Background(), "x") {
		t.Fatal("handled despite protocol error")
	}
	_ = os.Remove(filepath.Join(dir, "wayland-bad"))
}
