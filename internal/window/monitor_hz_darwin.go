//go:build darwin && !ios

package window

import (
	"sync"

	"github.com/ebitengine/purego"
)

const coreGraphicsPath = "/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics"

// cgPoint mirrors CoreGraphics CGPoint on 64-bit macOS (CGFloat == double).
type cgPoint struct {
	x float64
	y float64
}

type darwinCGProcs struct {
	once               sync.Once
	ready              bool
	displaysWithPoint  func(point cgPoint, maxDisplays uint32, displays *uint32, count *uint32) int32
	mainDisplayID      func() uint32
	copyDisplayMode    func(display uint32) uintptr
	modeGetRefreshRate func(mode uintptr) float64
	releaseDisplayMode func(mode uintptr)
}

//nolint:gochecknoglobals // reason: lazy one-time CoreGraphics dlopen/dlsym cache for monitor refresh queries.
var darwinCG darwinCGProcs

func (d *darwinCGProcs) init() {
	d.once.Do(func() {
		lib, err := purego.Dlopen(coreGraphicsPath, purego.RTLD_LAZY|purego.RTLD_LOCAL)
		if err != nil || lib == 0 {
			return
		}
		if !registerCGSym(&d.displaysWithPoint, lib, "CGGetDisplaysWithPoint") ||
			!registerCGSym(&d.mainDisplayID, lib, "CGMainDisplayID") ||
			!registerCGSym(&d.copyDisplayMode, lib, "CGDisplayCopyDisplayMode") ||
			!registerCGSym(&d.modeGetRefreshRate, lib, "CGDisplayModeGetRefreshRate") ||
			!registerCGSym(&d.releaseDisplayMode, lib, "CGDisplayModeRelease") {
			return
		}
		d.ready = true
	})
}

func registerCGSym(fptr any, lib uintptr, name string) bool {
	sym, err := purego.Dlsym(lib, name)
	if err != nil || sym == 0 {
		return false
	}
	purego.RegisterFunc(fptr, sym)
	return true
}

// detectMonitorHz queries the refresh rate (Hz) of the macOS display hosting
// the window center via CoreGraphics and purego (zero cgo). Falls back to
// CGMainDisplayID and then defaultMonitorHz when CoreGraphics is unavailable
// or reports 0 Hz.
func detectMonitorHz(posX, posY, winW, winH int) int {
	darwinCG.init()
	if !darwinCG.ready {
		return defaultMonitorHz
	}

	const half = 2
	pt := cgPoint{
		x: float64(posX) + float64(winW)/half,
		y: float64(posY) + float64(winH)/half,
	}

	var displayID uint32
	var count uint32
	if errCode := darwinCG.displaysWithPoint(pt, 1, &displayID, &count); errCode != 0 || count == 0 || displayID == 0 {
		displayID = darwinCG.mainDisplayID()
	}
	if displayID == 0 {
		return defaultMonitorHz
	}

	mode := darwinCG.copyDisplayMode(displayID)
	if mode == 0 {
		return defaultMonitorHz
	}
	defer darwinCG.releaseDisplayMode(mode)

	return sanitizeMonitorHz(roundMonitorHz(darwinCG.modeGetRefreshRate(mode)))
}
