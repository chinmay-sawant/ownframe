//go:build windows && (amd64 || arm64)

package uia

import (
	"errors"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
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
	// orig is the window procedure this host replaced. wndProc reads it from
	// the window thread while AttachWindow and Close write it, so it is atomic.
	orig atomic.Uintptr

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
	// Read the current procedure before the host becomes visible to wndProc,
	// so a message that arrives between SetWindowLongPtrW and the final
	// store still forwards to a real procedure instead of null.
	cur, _, err := procGetWindowLong.Call(hwnd, uintptr(gwlpWndProcValue))
	if cur == 0 {
		h.rootObj.release()
		return nil, errors.New("uia: read window procedure: " + err.Error())
	}
	h.orig.Store(cur)
	hosts.Lock()
	if _, dup := hosts.m[hwnd]; dup {
		hosts.Unlock()
		h.rootObj.release()
		return nil, errors.New("uia: window already has a provider")
	}
	hosts.m[hwnd] = h
	hosts.Unlock()

	old, _, err := procSetWindowLong.Call(hwnd, uintptr(gwlpWndProcValue), wndProcCallback)
	if old == 0 {
		hosts.Lock()
		delete(hosts.m, hwnd)
		hosts.Unlock()
		h.rootObj.release()
		return nil, errors.New("uia: subclass window: " + err.Error())
	}
	h.orig.Store(old)
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

// objFor returns the object for id with a reference owned by the caller,
// who must release it (after give, which adds the client's own reference).
// The reference is taken under h.mu, so Publish and Close cannot drop the
// object's last reference and unregister it between lookup and use. It
// returns nil for an unknown id, a released object, or a closed host.
func (h *Host) objFor(id host.AccessibilityNodeID) *comObj {
	if id == "" {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return nil
	}
	if id == h.prov.Root() {
		return h.rootLocked()
	}
	if o, ok := h.elems[id]; ok {
		if o.tryAddRef() {
			return o
		}
		delete(h.elems, id)
	}
	if _, err := h.prov.Element(id); err != nil {
		return nil
	}
	o := newObj(h, id, false)
	h.elems[id] = o
	o.addRef() // the caller's reference; newObj's belongs to the cache
	return o
}

// acquireRoot returns the root object with a reference owned by the caller,
// or nil once the host is closed.
func (h *Host) acquireRoot() *comObj {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return nil
	}
	return h.rootLocked()
}

func (h *Host) rootLocked() *comObj {
	if h.rootObj.tryAddRef() {
		return h.rootObj
	}
	return nil
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
	if ch.Structure || len(ch.Props) > maxPropEvents {
		// The owning reference and KeepAlive hold the object across the
		// syscall: the slot address is a bare uintptr the GC cannot see.
		if root := h.acquireRoot(); root != nil {
			procRaiseStructure.Call(root.slotAddr(slotSimple), structChildInv, 0, 0)
			runtime.KeepAlive(root)
			root.release()
		}
	} else {
		for _, p := range ch.Props {
			h.raiseProperty(p)
		}
	}
	if ch.Focus != "" {
		if o := h.objFor(ch.Focus); o != nil {
			procRaiseEvent.Call(o.slotAddr(slotSimple), eventFocusChanged)
			runtime.KeepAlive(o)
			o.release()
			// Clients that follow the focus through WinEvents, such as
			// System.Windows.Automation and NVDA's MSAA path, then ask the
			// window for it.
			procNotifyWinEvent.Call(eventObjectFocus, h.hwnd, uintptr(uint32(objidClient&0xffffffff)), 0)
		}
	}
}

func (h *Host) raiseProperty(p PropertyChange) {
	e, err := h.prov.Element(p.ID)
	o := h.objFor(p.ID)
	if o == nil {
		return
	}
	defer o.release()
	if err != nil {
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
	runtime.KeepAlive(o)
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
	runtime.KeepAlive(h.rootObj)
	for _, o := range elems {
		procDisconnect.Call(o.slotAddr(slotSimple))
		runtime.KeepAlive(o)
		o.release()
	}
	h.rootObj.release()

	cur, _, _ := procGetWindowLong.Call(h.hwnd, uintptr(gwlpWndProcValue))
	if cur == wndProcCallback {
		procSetWindowLong.Call(h.hwnd, uintptr(gwlpWndProcValue), h.orig.Load())
		hosts.Lock()
		delete(hosts.m, h.hwnd)
		hosts.Unlock()
	}
}

// answersObject reports whether a WM_GETOBJECT request with this object ID is
// UI Automation's. Besides the root (UiaRootObjectId), UI Automation sends IDs
// from zero up when it delivers an event: the client's side asks the window
// for the element behind the event, and UiaReturnRawElementProvider resolves
// that ID to the right provider. A window that answers only the root leaves
// those requests to the default window procedure, which cannot, so UI
// Automation drops every event the window raises. OBJID_CLIENT gets the
// MSAA view of the same tree, which clients such as NVDA read. Other negative
// IDs (the title bar, the scroll bars) stay with the window.
func answersObject(id int32) bool {
	return id >= 0 || id == uiaRootObject || id == objidClient
}

func wndProc(hwnd, msg, wparam, lparam uintptr) (ret uintptr) {
	hosts.RLock()
	h := hosts.m[hwnd]
	hosts.RUnlock()
	if h == nil {
		return 0
	}
	if msg == wmGetObject && answersObject(int32(lparam)) && !h.isClosed() {
		defer func() {
			if recover() != nil {
				ret = 0
			}
		}()
		if root := h.acquireRoot(); root != nil {
			r, _, _ := procReturnRaw.Call(hwnd, wparam, lparam, root.slotAddr(slotSimple))
			runtime.KeepAlive(root)
			root.release()
			return r
		}
	}
	ret, _, _ = procCallWindowProc.Call(h.orig.Load(), hwnd, msg, wparam, lparam)
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
