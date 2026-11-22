//go:build windows && (amd64 || arm64)

package uia

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"golang.org/x/sys/windows"
)

// One comObj is one UIA element. It carries a vtable pointer for every
// interface it can implement, so a single object answers QueryInterface for
// all of them. The interface pointer handed to a client is the address of the
// matching slot, which resolve maps back to the object and the interface.
const (
	slotSimple = iota
	slotFragment
	slotRoot
	slotInvoke
	slotValue
	slotToggle
	slotSelection
	numSlots
)

// UIA property and pattern identifiers.
const (
	propProcessID       = 30002
	propControlType     = 30003
	propLocalizedType   = 30004
	propName            = 30005
	propHasFocus        = 30008
	propIsFocusable     = 30009
	propIsEnabled       = 30010
	propAutomationID    = 30011
	propHelpText        = 30013
	propIsControl       = 30016
	propIsContent       = 30017
	propIsPassword      = 30019
	propIsOffscreen     = 30022
	propFrameworkID     = 30024
	propValueValue      = 30045
	propValueReadOnly   = 30046
	propSelItemSelected = 30079
	propToggleState     = 30086
	propProviderDesc    = 30107

	patternInvoke        = 10000
	patternValue         = 10002
	patternSelectionItem = 10010
	patternToggle        = 10015

	eventFocusChanged = 20005
)

var (
	iidUnknown       = windows.GUID{Data4: [8]byte{0xC0, 0, 0, 0, 0, 0, 0, 0x46}}
	iidSimple        = windows.GUID{Data1: 0xd6dd68d1, Data2: 0x86fd, Data3: 0x4332, Data4: [8]byte{0x86, 0x66, 0x9a, 0xbe, 0xde, 0xa2, 0xd2, 0x4c}}
	iidFragment      = windows.GUID{Data1: 0xf7063da8, Data2: 0x8359, Data3: 0x439c, Data4: [8]byte{0x92, 0x97, 0xbb, 0xc5, 0x29, 0x9a, 0x7d, 0x87}}
	iidFragmentRoot  = windows.GUID{Data1: 0x620ce2a5, Data2: 0xab8f, Data3: 0x40a9, Data4: [8]byte{0x86, 0xcb, 0xde, 0x3c, 0x75, 0x59, 0x9b, 0x58}}
	iidInvoke        = windows.GUID{Data1: 0x54fcb24b, Data2: 0xe18e, Data3: 0x47a2, Data4: [8]byte{0xb4, 0xd3, 0xec, 0xcb, 0xe7, 0x75, 0x99, 0xa2}}
	iidValue         = windows.GUID{Data1: 0xc7935180, Data2: 0x6fb3, Data3: 0x4201, Data4: [8]byte{0xb1, 0x74, 0x7d, 0xf7, 0x3a, 0xdb, 0xf6, 0x4a}}
	iidToggle        = windows.GUID{Data1: 0x56d00bd0, Data2: 0xc4f4, Data3: 0x433c, Data4: [8]byte{0xa8, 0x36, 0x1a, 0x52, 0xa5, 0x7e, 0x08, 0x92}}
	iidSelectionItem = windows.GUID{Data1: 0x2acad808, Data2: 0xb2d4, Data3: 0x452d, Data4: [8]byte{0xa4, 0x07, 0x91, 0xff, 0x1a, 0xd1, 0x67, 0xb2}}
)

type comObj struct {
	slots  [numSlots]uintptr
	ref    atomic.Int32
	id     host.AccessibilityNodeID
	isRoot bool
	h      *Host
}

type ifaceRef struct {
	o    *comObj
	slot int
}

var registry = struct {
	sync.RWMutex
	m map[uintptr]ifaceRef
}{m: map[uintptr]ifaceRef{}}

func (o *comObj) slotAddr(slot int) uintptr { return uintptr(unsafe.Pointer(&o.slots[slot])) }

// newObj registers an object holding one reference, owned by the host cache.
func newObj(h *Host, id host.AccessibilityNodeID, isRoot bool) *comObj {
	buildVtables()
	o := &comObj{id: id, isRoot: isRoot, h: h}
	o.ref.Store(1)
	for i := range o.slots {
		o.slots[i] = vtables[i]
	}
	registry.Lock()
	for i := range o.slots {
		registry.m[o.slotAddr(i)] = ifaceRef{o, i}
	}
	registry.Unlock()
	return o
}

func (o *comObj) addRef() int32 { return o.ref.Add(1) }

func (o *comObj) release() int32 {
	n := o.ref.Add(-1)
	if n <= 0 {
		registry.Lock()
		for i := range o.slots {
			delete(registry.m, o.slotAddr(i))
		}
		registry.Unlock()
	}
	return n
}

func resolve(this uintptr) (*comObj, int, bool) {
	registry.RLock()
	r, ok := registry.m[this]
	registry.RUnlock()
	return r.o, r.slot, ok
}

// node returns the page node this object stands for. The root object is
// permanent, so it follows the current document's root.
func (o *comObj) node() host.AccessibilityNodeID {
	if o.isRoot {
		return o.h.prov.Root()
	}
	return o.id
}

func hr(v uint32) uintptr { return uintptr(v) }

func errHR(err error) uintptr {
	switch {
	case err == nil:
		return sOK
	case errors.Is(err, ErrNotAvailable), errors.Is(err, host.ErrAccessibilityStale),
		errors.Is(err, host.ErrAccessibilityNode), errors.Is(err, host.ErrAccessibilityUnavailable):
		return hr(eElementNoAvl)
	case errors.Is(err, ErrNotEnabled), errors.Is(err, ErrReadOnly), errors.Is(err, host.ErrAccessibilityDisabled):
		return hr(eElementNoEn)
	case errors.Is(err, ErrNotSupported), errors.Is(err, host.ErrAccessibilityAction):
		return hr(eNotSupported)
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return hr(hrTimeout)
	}
	return hr(eFail)
}

var (
	vtOnce  sync.Once
	vtables [numSlots]uintptr
	// vtStore keeps the vtable arrays reachable for the life of the process.
	vtStore [numSlots][]uintptr
)

type callback func(o *comObj, a1, a2, a3 uintptr) uintptr

func guard(fn callback) uintptr {
	return syscall.NewCallback(func(this, a1, a2, a3 uintptr) (ret uintptr) {
		defer func() {
			if recover() != nil {
				ret = hr(eFail)
			}
		}()
		o, _, ok := resolve(this)
		if !ok {
			return hr(eElementNoAvl)
		}
		return fn(o, a1, a2, a3)
	})
}

// guardSlot is guard for methods that also need the interface the call arrived on.
func guardSlot(fn func(o *comObj, slot int, a1, a2, a3 uintptr) uintptr) uintptr {
	return syscall.NewCallback(func(this, a1, a2, a3 uintptr) (ret uintptr) {
		defer func() {
			if recover() != nil {
				ret = hr(eFail)
			}
		}()
		o, slot, ok := resolve(this)
		if !ok {
			return hr(eElementNoAvl)
		}
		return fn(o, slot, a1, a2, a3)
	})
}

func buildVtables() {
	vtOnce.Do(func() {
		qi := guardSlot(func(o *comObj, _ int, riid, ppv, _ uintptr) uintptr { return o.queryInterface(riid, ppv) })
		addRef := syscall.NewCallback(func(this, _, _, _ uintptr) uintptr {
			if o, _, ok := resolve(this); ok {
				return uintptr(o.addRef())
			}
			return 0
		})
		release := syscall.NewCallback(func(this, _, _, _ uintptr) uintptr {
			if o, _, ok := resolve(this); ok {
				return uintptr(max(o.release(), 0))
			}
			return 0
		})
		unknown := []uintptr{qi, addRef, release}
		simple := append(slices3(unknown),
			guard(func(_ *comObj, p, _, _ uintptr) uintptr { putU32(p, 1); return sOK }), // ProviderOptions_ServerSideProvider
			guard((*comObj).getPattern),
			guard((*comObj).getProperty),
			guard((*comObj).getHostProvider),
		)
		fragment := append(slices3(unknown),
			guard((*comObj).navigate),
			guard((*comObj).getRuntimeID),
			guard((*comObj).getBounds),
			guard(func(_ *comObj, pp, _, _ uintptr) uintptr { putPtr(pp, 0); return sOK }), // GetEmbeddedFragmentRoots
			guard((*comObj).setFocus),
			guard((*comObj).getFragmentRoot),
		)
		root := append(slices3(unknown),
			guard((*comObj).fromPoint),
			guard((*comObj).getFocus),
		)
		invoke := append(slices3(unknown), guard(func(o *comObj, _, _, _ uintptr) uintptr {
			return errHR(o.h.prov.Invoke(o.node()))
		}))
		value := append(slices3(unknown),
			guard(func(o *comObj, s, _, _ uintptr) uintptr {
				if s == 0 {
					return hr(ePointer)
				}
				return errHR(o.h.prov.SetValue(o.node(), windows.UTF16PtrToString((*uint16)(ptr(s)))))
			}),
			guard((*comObj).getValue),
			guard(func(o *comObj, p, _, _ uintptr) uintptr {
				e, err := o.h.prov.Element(o.node())
				if err != nil || !e.ValuePattern {
					return errHR(errOr(err, ErrNotSupported))
				}
				putU32(p, boolU32(e.ReadOnly || !e.Enabled))
				return sOK
			}),
		)
		toggle := append(slices3(unknown),
			guard(func(o *comObj, _, _, _ uintptr) uintptr { return errHR(o.h.prov.Toggle(o.node())) }),
			guard(func(o *comObj, p, _, _ uintptr) uintptr {
				e, err := o.h.prov.Element(o.node())
				if err != nil || !e.Toggle {
					return errHR(errOr(err, ErrNotSupported))
				}
				putU32(p, uint32(e.Toggled))
				return sOK
			}),
		)
		selection := append(slices3(unknown),
			guard(func(o *comObj, _, _, _ uintptr) uintptr { return errHR(o.h.prov.Select(o.node())) }),
			guard(func(o *comObj, _, _, _ uintptr) uintptr { return errHR(o.h.prov.Select(o.node())) }),
			guard(func(*comObj, uintptr, uintptr, uintptr) uintptr { return hr(0x80131509) }), // InvalidOperationException
			guard(func(o *comObj, p, _, _ uintptr) uintptr {
				e, err := o.h.prov.Element(o.node())
				if err != nil || !e.SelectionItem {
					return errHR(errOr(err, ErrNotSupported))
				}
				putU32(p, boolU32(e.Selected))
				return sOK
			}),
			guard(func(_ *comObj, p, _, _ uintptr) uintptr { putPtr(p, 0); return sOK }), // no selection container
		)
		for i, tbl := range map[int][]uintptr{
			slotSimple: simple, slotFragment: fragment, slotRoot: root,
			slotInvoke: invoke, slotValue: value, slotToggle: toggle, slotSelection: selection,
		} {
			vtStore[i] = tbl
			vtables[i] = uintptr(unsafe.Pointer(&tbl[0]))
		}
	})
}

func slices3(u []uintptr) []uintptr { return append(make([]uintptr, 0, 12), u...) }

func errOr(err, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}

func boolU32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func (o *comObj) queryInterface(riid, ppv uintptr) uintptr {
	if ppv == 0 || riid == 0 {
		return hr(ePointer)
	}
	putPtr(ppv, 0)
	g := *(*windows.GUID)(ptr(riid))
	slot := -1
	switch g {
	case iidUnknown, iidSimple:
		slot = slotSimple
	case iidFragment:
		slot = slotFragment
	case iidFragmentRoot:
		if o.isRoot {
			slot = slotRoot
		}
	default:
		if e, err := o.h.prov.Element(o.node()); err == nil {
			switch {
			case g == iidInvoke && e.Invoke:
				slot = slotInvoke
			case g == iidValue && e.ValuePattern:
				slot = slotValue
			case g == iidToggle && e.Toggle:
				slot = slotToggle
			case g == iidSelectionItem && e.SelectionItem:
				slot = slotSelection
			}
		}
	}
	if slot < 0 {
		return hr(eNoInterface)
	}
	o.addRef()
	putPtr(ppv, o.slotAddr(slot))
	return sOK
}

// give hands out an interface pointer with a reference for the client.
func (o *comObj) give(out uintptr, slot int) {
	o.addRef()
	putPtr(out, o.slotAddr(slot))
}

func (o *comObj) getPattern(id, out, _ uintptr) uintptr {
	if out == 0 {
		return hr(ePointer)
	}
	putPtr(out, 0)
	e, err := o.h.prov.Element(o.node())
	if err != nil {
		return errHR(err)
	}
	switch id {
	case patternInvoke:
		if e.Invoke {
			o.give(out, slotInvoke)
		}
	case patternValue:
		if e.ValuePattern {
			o.give(out, slotValue)
		}
	case patternToggle:
		if e.Toggle {
			o.give(out, slotToggle)
		}
	case patternSelectionItem:
		if e.SelectionItem {
			o.give(out, slotSelection)
		}
	}
	return sOK
}

func (o *comObj) getProperty(id, out, _ uintptr) uintptr {
	if out == 0 {
		return hr(ePointer)
	}
	putVariant(out, vtEmpty, 0)
	e, err := o.h.prov.Element(o.node())
	if err != nil {
		return errHR(err)
	}
	str := func(s string) {
		if s != "" {
			putVariant(out, vtBstr, allocBSTR(s))
		}
	}
	b := func(v bool) {
		if v {
			putVariant(out, vtBool, variantTrue)
		} else {
			putVariant(out, vtBool, 0)
		}
	}
	switch id {
	case propProcessID:
		putVariant(out, vtI4, uintptr(windows.GetCurrentProcessId()))
	case propControlType:
		putVariant(out, vtI4, uintptr(e.ControlType))
	case propLocalizedType:
		str(e.LocalizedType)
	case propName:
		str(e.Name)
	case propHasFocus:
		b(e.Focused)
	case propIsFocusable:
		b(e.Focusable)
	case propIsEnabled:
		b(e.Enabled)
	case propAutomationID:
		str(e.AutomationID)
	case propHelpText:
		str(e.HelpText)
	case propIsControl, propIsContent:
		b(true)
	case propIsPassword:
		b(e.Password)
	case propIsOffscreen:
		_, ok, _ := o.h.prov.Bounds(e.ID)
		b(!ok)
	case propFrameworkID:
		str("Ownframe")
	case propProviderDesc:
		str("ownframe UIA provider")
	case propValueValue:
		if e.ValuePattern {
			putVariant(out, vtBstr, allocBSTR(e.Value))
		}
	case propValueReadOnly:
		if e.ValuePattern {
			b(e.ReadOnly || !e.Enabled)
		}
	case propSelItemSelected:
		if e.SelectionItem {
			b(e.Selected)
		}
	case propToggleState:
		if e.Toggle {
			putVariant(out, vtI4, uintptr(e.Toggled))
		}
	}
	return sOK
}

func (o *comObj) getHostProvider(out, _, _ uintptr) uintptr {
	if out == 0 {
		return hr(ePointer)
	}
	putPtr(out, 0)
	if !o.isRoot {
		return sOK
	}
	var p uintptr
	r, _, _ := procHostFromHwnd.Call(o.h.hwnd, uintptr(unsafe.Pointer(&p)))
	if int32(r) < 0 {
		return r
	}
	putPtr(out, p)
	return sOK
}

func (o *comObj) navigate(dir, out, _ uintptr) uintptr {
	if out == 0 {
		return hr(ePointer)
	}
	putPtr(out, 0)
	if dir > uintptr(LastChild) {
		return hr(eFail)
	}
	if o.isRoot && (Direction(dir) == NextSibling || Direction(dir) == PreviousSibling) {
		return sOK
	}
	id, ok, err := o.h.prov.Navigate(o.node(), Direction(dir))
	if err != nil {
		return errHR(err)
	}
	if !ok {
		return sOK
	}
	if t := o.h.objFor(id); t != nil {
		t.give(out, slotFragment)
	}
	return sOK
}

func (o *comObj) getRuntimeID(out, _, _ uintptr) uintptr {
	if out == 0 {
		return hr(ePointer)
	}
	putPtr(out, 0)
	if o.isRoot {
		return sOK // the window provider supplies the runtime ID
	}
	e, err := o.h.prov.Element(o.node())
	if err != nil {
		return errHR(err)
	}
	putPtr(out, runtimeIDArray(e.RuntimeID))
	return sOK
}

func (o *comObj) getBounds(out, _, _ uintptr) uintptr {
	if out == 0 {
		return hr(ePointer)
	}
	r, _, err := o.h.prov.Bounds(o.node())
	if err != nil {
		return errHR(err)
	}
	*(*[4]float64)(ptr(out)) = [4]float64{r.X, r.Y, r.W, r.H}
	return sOK
}

func (o *comObj) setFocus(_, _, _ uintptr) uintptr {
	return errHR(o.h.prov.SetFocus(o.node()))
}

func (o *comObj) getFragmentRoot(out, _, _ uintptr) uintptr {
	if out == 0 {
		return hr(ePointer)
	}
	o.h.rootObj.give(out, slotRoot)
	return sOK
}

// fromPoint cannot read its two double arguments: syscall callbacks only
// receive integer registers, and the x64 and arm64 conventions pass doubles
// in floating-point registers. It reports "no child here" so UIA falls back to
// the window element, and writes the result through the one pointer it can
// locate (the fourth integer-class argument on amd64, the first on arm64).
func (o *comObj) fromPoint(a1, _, a3 uintptr) uintptr {
	out := a3
	if runtimeArm64 {
		out = a1
	}
	putPtr(out, 0)
	return sOK
}

func (o *comObj) getFocus(out, _, _ uintptr) uintptr {
	if out == 0 {
		return hr(ePointer)
	}
	putPtr(out, 0)
	if id := o.h.prov.Focused(); id != "" {
		if t := o.h.objFor(id); t != nil {
			t.give(out, slotFragment)
		}
	}
	return sOK
}

func (o *comObj) getValue(out, _, _ uintptr) uintptr {
	if out == 0 {
		return hr(ePointer)
	}
	putPtr(out, 0)
	e, err := o.h.prov.Element(o.node())
	if err != nil || !e.ValuePattern || e.Password {
		return errHR(errOr(err, ErrNotSupported))
	}
	putPtr(out, allocBSTR(e.Value))
	return sOK
}

// runtimeArm64 picks the argument slot fromPoint writes to.
const runtimeArm64 = runtime.GOARCH == "arm64"
