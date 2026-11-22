//go:build windows && (amd64 || arm64)

package uia

import (
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// The COM layer needs a 64-bit calling convention: every syscall callback
// takes the same four register-sized arguments, which is only sound when
// Windows passes the first four arguments in registers (amd64 and arm64).

var (
	uiaCore = windows.NewLazySystemDLL("uiautomationcore.dll")
	oleAut  = windows.NewLazySystemDLL("oleaut32.dll")
	user32  = windows.NewLazySystemDLL("user32.dll")

	procReturnRaw       = uiaCore.NewProc("UiaReturnRawElementProvider")
	procHostFromHwnd    = uiaCore.NewProc("UiaHostProviderFromHwnd")
	procRaiseEvent      = uiaCore.NewProc("UiaRaiseAutomationEvent")
	procRaiseProperty   = uiaCore.NewProc("UiaRaiseAutomationPropertyChangedEvent")
	procRaiseStructure  = uiaCore.NewProc("UiaRaiseStructureChangedEvent")
	procClientsListen   = uiaCore.NewProc("UiaClientsAreListening")
	procDisconnect      = uiaCore.NewProc("UiaDisconnectProvider")
	procSysAllocStrLen  = oleAut.NewProc("SysAllocStringLen")
	procSysFreeString   = oleAut.NewProc("SysFreeString")
	procSafeArrayCreate = oleAut.NewProc("SafeArrayCreateVector")
	procSafeArrayPut    = oleAut.NewProc("SafeArrayPutElement")
	procSafeArrayFree   = oleAut.NewProc("SafeArrayDestroy")

	procEnumWindows    = user32.NewProc("EnumWindows")
	procWindowPID      = user32.NewProc("GetWindowThreadProcessId")
	procClassName      = user32.NewProc("GetClassNameW")
	procIsVisible      = user32.NewProc("IsWindowVisible")
	procSetWindowLong  = user32.NewProc("SetWindowLongPtrW")
	procGetWindowLong  = user32.NewProc("GetWindowLongPtrW")
	procCallWindowProc = user32.NewProc("CallWindowProcW")
	procClientToScreen = user32.NewProc("ClientToScreen")
	procGetClientRect  = user32.NewProc("GetClientRect")
	procGetDpiForWnd   = user32.NewProc("GetDpiForWindow")
	procNotifyWinEvent = user32.NewProc("NotifyWinEvent")
)

// HRESULT values the provider returns.
const (
	sOK           = 0
	eNoInterface  = 0x80004002
	eFail         = 0x80004005
	ePointer      = 0x80004003
	eNotImpl      = 0x80004001
	eElementNoAvl = 0x80040201
	eElementNoEn  = 0x80040200
	eNotSupported = 0x80040204
	hrTimeout     = 0x800705B4
	vtEmpty       = 0
	vtI4          = 3
	vtBstr        = 8
	vtBool        = 11
	vtUnknown     = 13
	variantTrue   = 0xFFFF
	uiaRootObject = -25
	objidClient   = -4
	// eventObjectFocus is EVENT_OBJECT_FOCUS.
	eventObjectFocus = 0x8005
	// providerOptionsServerSide is ProviderOptions_ServerSideProvider. The
	// enum is ClientSideProvider = 1, ServerSideProvider = 2; reporting 1
	// from a provider hosted in the window it serves makes UI Automation
	// treat it as a client-side one and deliver none of its events.
	providerOptionsServerSide = 0x2
	wmGetObject               = 0x003D
	wmNCDestroy               = 0x0082
	gwlpWndProc               = -4
	glfwClassName             = "GLFW30"
	maxPropEvents             = 64
	uiaAppendRtID             = 3
	structChildInv            = 2
)

// ptr turns an address handed to a callback into a pointer. The pointee is
// COM-owned memory, so go vet's unsafeptr check is satisfied by routing the
// conversion through an unsafe.Pointer-typed load.
func ptr(a uintptr) unsafe.Pointer { return *(*unsafe.Pointer)(unsafe.Pointer(&a)) }

func putU32(a uintptr, v uint32) {
	if a != 0 {
		*(*uint32)(ptr(a)) = v
	}
}

func putPtr(a, v uintptr) {
	if a != 0 {
		*(*uintptr)(ptr(a)) = v
	}
}

// variant is a COM VARIANT. It is 24 bytes on 64-bit Windows.
type variant struct {
	vt       uint16
	_        [3]uint16
	val      uintptr
	reserved uintptr
}

func putVariant(a uintptr, vt uint16, val uintptr) {
	if a == 0 {
		return
	}
	*(*variant)(ptr(a)) = variant{vt: vt, val: val}
}

func allocBSTR(s string) uintptr {
	s = strings.ReplaceAll(s, "\x00", "")
	u, err := windows.UTF16FromString(s)
	if err != nil {
		return 0
	}
	u = u[:len(u)-1]
	var first *uint16
	if len(u) > 0 {
		first = &u[0]
	}
	r, _, _ := procSysAllocStrLen.Call(uintptr(unsafe.Pointer(first)), uintptr(len(u)))
	return r
}

func freeBSTR(b uintptr) {
	if b != 0 {
		procSysFreeString.Call(b)
	}
}

// runtimeIDArray builds the SAFEARRAY of int32 that GetRuntimeId returns.
func runtimeIDArray(rid int32) uintptr {
	psa, _, _ := procSafeArrayCreate.Call(vtI4, 0, 2)
	if psa == 0 {
		return 0
	}
	vals := [2]int32{uiaAppendRtID, rid}
	for i := range vals {
		idx := int32(i)
		hr, _, _ := procSafeArrayPut.Call(psa, uintptr(unsafe.Pointer(&idx)), uintptr(unsafe.Pointer(&vals[i])))
		if int32(hr) < 0 {
			procSafeArrayFree.Call(psa)
			return 0
		}
	}
	return psa
}
