//go:build windows && (amd64 || arm64)

package uia

import (
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"golang.org/x/sys/windows"
)

// TestLiveEventDelivery hosts the provider in a plain Win32 window and keeps
// changing focus and names, so an external UIA client can check that the
// events arrive. It needs a desktop and a listening client, so it only runs
// when OWNFRAME_UIA_LIVE=1. It isolates the provider from Ebiten: if a client
// hears these events but not the ones from a real ownframe window, the
// difference is in the window, not in the COM layer.
func TestLiveEventDelivery(t *testing.T) {
	if os.Getenv("OWNFRAME_UIA_LIVE") != "1" {
		t.Skip("set OWNFRAME_UIA_LIVE=1 on a Windows desktop to run the live probe")
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	u32 := windows.NewLazySystemDLL("user32.dll")
	var (
		registerClass = u32.NewProc("RegisterClassExW")
		createWindow  = u32.NewProc("CreateWindowExW")
		showWindow    = u32.NewProc("ShowWindow")
		getMessage    = u32.NewProc("GetMessageW")
		translate     = u32.NewProc("TranslateMessage")
		dispatch      = u32.NewProc("DispatchMessageW")
		defProc       = u32.NewProc("DefWindowProcW")
		setTimer      = u32.NewProc("SetTimer")
		postQuit      = u32.NewProc("PostQuitMessage")
	)

	const (
		wmTimer   = 0x0113
		wmDestroy = 0x0002
	)

	var hostRef *Host
	ticks := 0
	names := []string{"Login", "Login one", "Login two", "Login three"}
	focuses := []host.AccessibilityNodeID{"go", "email", "tog", "email"}

	proc := syscall.NewCallback(func(hwnd, msg, wparam, lparam uintptr) uintptr {
		switch msg {
		case wmTimer:
			ticks++
			tree := sampleTree()
			tree.Revision = uint64(ticks) + 1
			tree.Nodes[0].Name = names[ticks%len(names)]
			tree.Focused = focuses[ticks%len(focuses)]
			for i := range tree.Nodes {
				tree.Nodes[i].Focused = tree.Nodes[i].ID == tree.Focused
			}
			if hostRef != nil {
				hostRef.Publish(tree)
			}
			if ticks > 14 {
				postQuit.Call(0)
			}
			return 0
		case wmDestroy:
			postQuit.Call(0)
			return 0
		}
		r, _, _ := defProc.Call(hwnd, msg, wparam, lparam)
		return r
	})

	className, _ := windows.UTF16PtrFromString("OwnframeUIAProbe")
	title, _ := windows.UTF16PtrFromString("UIAProbe")

	type wndClassEx struct {
		Size       uint32
		Style      uint32
		WndProc    uintptr
		ClsExtra   int32
		WndExtra   int32
		Instance   uintptr
		Icon       uintptr
		Cursor     uintptr
		Background uintptr
		MenuName   *uint16
		ClassName  *uint16
		IconSm     uintptr
	}

	wc := wndClassEx{WndProc: proc, ClassName: className, Background: 6}
	wc.Size = uint32(unsafe.Sizeof(wc))

	if r, _, err := registerClass.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		t.Fatalf("RegisterClassEx: %v", err)
	}

	const wsOverlappedVisible = 0x00CF0000 | 0x10000000

	hwnd, _, err := createWindow.Call(0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(title)),
		wsOverlappedVisible, uintptr(0xFFFFF8F8), 340, 400, 300, 0, 0, 0, 0) // x = -1800
	if hwnd == 0 {
		t.Fatalf("CreateWindowEx: %v", err)
	}

	showWindow.Call(hwnd, 5)

	prov := NewProvider(nil)
	prov.Update(sampleTree())

	hostRef, err = AttachWindow(hwnd, prov)
	if err != nil {
		t.Fatalf("AttachWindow: %v", err)
	}
	defer hostRef.Close()

	setTimer.Call(hwnd, 1, uintptr(1500), 0)

	deadline := time.Now().Add(40 * time.Second)

	var msg [48]byte

	for time.Now().Before(deadline) {
		r, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&msg[0])), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}

		translate.Call(uintptr(unsafe.Pointer(&msg[0])))
		dispatch.Call(uintptr(unsafe.Pointer(&msg[0])))
	}
}
