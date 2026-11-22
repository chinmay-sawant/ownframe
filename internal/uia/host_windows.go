//go:build windows && (amd64 || arm64)

package uia

import (
	"errors"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"golang.org/x/sys/windows"
)

var (
	// ErrNoWindow means the process has no visible Ebiten window yet.
	ErrNoWindow = errors.New("uia: no window found")
	// ErrNoTree means the provider has no tree to serve yet.
	ErrNoTree = errors.New("uia: provider has no tree")
)

// Host attaches a Provider to one native window. It subclasses the window
// procedure to answer WM_GETOBJECT, serves the elements as COM providers, and
// raises UIA events when the tree changes.
type Host struct {
	hwnd uintptr
	prov *Provider
	orig uintptr

	rootObj *comObj

	mu     sync.Mutex
	elems  map[host.AccessibilityNodeID]*comObj
	closed bool

	evMu      sync.Mutex
	pending   Change
	hasEvents bool
	wake      chan struct{}
	done      chan struct{}
	wg        sync.WaitGroup
}

var hosts = struct {
	sync.RWMutex
	m map[uintptr]*Host
}{m: map[uintptr]*Host{}}

var wndProcCallback = syscall.NewCallback(wndProc)

// Attach finds the process's window and starts serving prov on it. prov must
// already hold a tree, so the first WM_GETOBJECT sees a root.
func Attach(prov *Provider) (*Host, error) {
	if prov.Root() == "" {
		return nil, ErrNoTree
	}
	hwnd := findProcessWindow()
	if hwnd == 0 {
		return nil, ErrNoWindow
	}
	return AttachWindow(hwnd, prov)
}

// AttachWindow serves prov on hwnd. Tests use it with their own window.
func AttachWindow(hwnd uintptr, prov *Provider) (*Host, error) {
	if prov.Root() == "" {
		return nil, ErrNoTree
	}
	h := &Host{
		hwnd:  hwnd,
		prov:  prov,
		elems: map[host.AccessibilityNodeID]*comObj{},
		wake:  make(chan struct{}, 1),
		done:  make(chan struct{}),
	}
	h.rootObj = newObj(h, "", true)
	hosts.Lock()
	if _, dup := hosts.m[hwnd]; dup {
		hosts.Unlock()
		return nil, errors.New("uia: window already has a provider")
	}
	hosts.m[hwnd] = h
	hosts.Unlock()

	old, _, err := procSetWindowLong.Call(hwnd, uintptr(gwlpWndProcValue), wndProcCallback)
	if old == 0 {
		hosts.Lock()
		delete(hosts.m, hwnd)
		hosts.Unlock()
		return nil, errors.New("uia: subclass window: " + err.Error())
	}
	h.orig = old
	h.wg.Add(1)
	go h.eventLoop()
	return h, nil
}

// gwlpWndProcValue is GWLP_WNDPROC (-4) as a register-sized argument.
var gwlpWndProcValue = int64(gwlpWndProc)

// Provider returns the provider the host serves.
func (h *Host) Provider() *Provider { return h.prov }

// Publish installs tree and raises the matching UIA events asynchronously.
// Call it on the page's owning loop, right after reading the tree.
func (h *Host) Publish(tree host.AccessibilityTree) {
	ch := h.prov.Update(tree)
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	if ch.Document {
		for id, o := range h.elems {
			o.release()
			delete(h.elems, id)
		}
	} else {
		for id, o := range h.elems {
			if _, err := h.prov.Element(id); err != nil {
				o.release()
				delete(h.elems, id)
			}
		}
	}
	h.mu.Unlock()
	h.queue(ch)
}

// objFor returns the object for id. The caller owns a reference only after
// give; objFor itself does not add one.
func (h *Host) objFor(id host.AccessibilityNodeID) *comObj {
	if id == "" {
		return nil
	}
	if id == h.prov.Root() {
		return h.rootObj
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return nil
	}
	if o, ok := h.elems[id]; ok {
		return o
	}
	if _, err := h.prov.Element(id); err != nil {
		return nil
	}
	o := newObj(h, id, false)
	h.elems[id] = o
	return o
}

func (h *Host) queue(ch Change) {
	h.evMu.Lock()
	if !h.hasEvents {
		h.pending = Change{}
		h.hasEvents = true
	}
	h.pending.Document = h.pending.Document || ch.Document
	h.pending.Structure = h.pending.Structure || ch.Structure
	h.pending.Props = append(h.pending.Props, ch.Props...)
	if ch.Focus != "" {
		h.pending.Focus = ch.Focus
	}
	h.evMu.Unlock()
	select {
	case h.wake <- struct{}{}:
	default:
	}
}

func (h *Host) eventLoop() {
	defer h.wg.Done()
	for {
		select {
		case <-h.done:
			return
		case <-h.wake:
		}
		h.evMu.Lock()
		ch := h.pending
		h.pending, h.hasEvents = Change{}, false
		h.evMu.Unlock()
		h.raise(ch)
	}
}

func (h *Host) raise(ch Change) {
	if h.isClosed() {
		return
	}
	if r, _, _ := procClientsListen.Call(); r == 0 {
		return
	}
	root := h.rootObj.slotAddr(slotSimple)
	if ch.Structure || len(ch.Props) > maxPropEvents {
		procRaiseStructure.Call(root, structChildInv, 0, 0)
	} else {
		for _, p := range ch.Props {
			h.raiseProperty(p)
		}
	}
	if ch.Focus != "" {
		if o := h.objFor(ch.Focus); o != nil {
			procRaiseEvent.Call(o.slotAddr(slotSimple), eventFocusChanged)
		}
	}
}

func (h *Host) raiseProperty(p PropertyChange) {
	e, err := h.prov.Element(p.ID)
	o := h.objFor(p.ID)
	if err != nil || o == nil {
		return
	}
	var id uintptr
	var next variant
	switch p.Property {
	case PropName:
		id, next = propName, variant{vt: vtBstr, val: allocBSTR(e.Name)}
	case PropValue:
		if !e.ValuePattern {
			return
		}
		id, next = propValueValue, variant{vt: vtBstr, val: allocBSTR(e.Value)}
	case PropToggle:
		id, next = propToggleState, variant{vt: vtI4, val: uintptr(e.Toggled)}
	case PropEnabled:
		id, next = propIsEnabled, variant{vt: vtBool, val: boolVariant(e.Enabled)}
	case PropSelected:
		id, next = propSelItemSelected, variant{vt: vtBool, val: boolVariant(e.Selected)}
	default:
		return
	}
	// VARIANT is larger than 16 bytes, so both supported ABIs pass the
	// two values by reference to a caller-made copy.
	old := variant{}
	procRaiseProperty.Call(o.slotAddr(slotSimple), id, uintptr(unsafe.Pointer(&old)), uintptr(unsafe.Pointer(&next)))
	if next.vt == vtBstr {
		freeBSTR(next.val)
	}
}

func boolVariant(b bool) uintptr {
	if b {
		return variantTrue
	}
	return 0
}

func (h *Host) isClosed() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.closed
}

// ClientGeometry reports where the window's client area sits on screen, in
// physical pixels, and the display scale (physical pixels per CSS pixel).
func (h *Host) ClientGeometry() (Geometry, bool) {
	var pt struct{ x, y int32 }
	if r, _, _ := procClientToScreen.Call(h.hwnd, uintptr(unsafe.Pointer(&pt))); r == 0 {
		return Geometry{}, false
	}
	var rc struct{ l, t, r, b int32 }
	if r, _, _ := procGetClientRect.Call(h.hwnd, uintptr(unsafe.Pointer(&rc))); r == 0 {
		return Geometry{}, false
	}
	scale := 1.0
	if procGetDpiForWnd.Find() == nil {
		if dpi, _, _ := procGetDpiForWnd.Call(h.hwnd); dpi != 0 {
			scale = float64(dpi) / 96
		}
	}
	return Geometry{
		OriginX: float64(pt.x), OriginY: float64(pt.y), Scale: scale,
		Width: float64(rc.r - rc.l), Height: float64(rc.b - rc.t),
	}, true
}

// Close stops serving. UIA clients lose their elements. If another
// subclass was installed after ours, the procedure stays in the chain and
// only forwards.
func (h *Host) Close() {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	h.closed = true
	elems := h.elems
	h.elems = map[host.AccessibilityNodeID]*comObj{}
	h.mu.Unlock()

	close(h.done)
	h.wg.Wait()
	h.prov.Stop()
	procDisconnect.Call(h.rootObj.slotAddr(slotSimple))
	for _, o := range elems {
		procDisconnect.Call(o.slotAddr(slotSimple))
		o.release()
	}
	h.rootObj.release()

	cur, _, _ := procGetWindowLong.Call(h.hwnd, uintptr(gwlpWndProcValue))
	if cur == wndProcCallback {
		procSetWindowLong.Call(h.hwnd, uintptr(gwlpWndProcValue), h.orig)
		hosts.Lock()
		delete(hosts.m, h.hwnd)
		hosts.Unlock()
	}
}

func wndProc(hwnd, msg, wparam, lparam uintptr) (ret uintptr) {
	hosts.RLock()
	h := hosts.m[hwnd]
	hosts.RUnlock()
	if h == nil {
		return 0
	}
	if msg == wmGetObject && int32(lparam) == uiaRootObject && !h.isClosed() {
		defer func() {
			if recover() != nil {
				ret = 0
			}
		}()
		r, _, _ := procReturnRaw.Call(hwnd, wparam, lparam, h.rootObj.slotAddr(slotSimple))
		return r
	}
	ret, _, _ = procCallWindowProc.Call(h.orig, hwnd, msg, wparam, lparam)
	if msg == wmNCDestroy {
		hosts.Lock()
		delete(hosts.m, hwnd)
		hosts.Unlock()
	}
	return ret
}

var (
	enumMu     sync.Mutex
	enumResult uintptr
	enumPID    uint32
	enumProc   = syscall.NewCallback(func(hwnd, _ uintptr) uintptr {
		var pid uint32
		procWindowPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
		if pid != enumPID {
			return 1
		}
		if vis, _, _ := procIsVisible.Call(hwnd); vis == 0 {
			return 1
		}
		var name [16]uint16
		n, _, _ := procClassName.Call(hwnd, uintptr(unsafe.Pointer(&name[0])), uintptr(len(name)))
		if windows.UTF16ToString(name[:n]) != glfwClassName {
			return 1
		}
		enumResult = hwnd
		return 0
	})
)

// findProcessWindow returns this process's visible Ebiten (GLFW) window.
func findProcessWindow() uintptr {
	enumMu.Lock()
	defer enumMu.Unlock()
	enumResult, enumPID = 0, uint32(os.Getpid())
	procEnumWindows.Call(enumProc, 0)
	return enumResult
}
