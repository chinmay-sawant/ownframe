package window

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// defaultMonitorHz is the fallback refresh rate when the host cannot
	// report one or reports a sentinel/out-of-range value.
	defaultMonitorHz = DefaultMaxFPS
	// minValidMonitorHz rejects Windows' 0/1 "hardware default" sentinels
	// and non-positive values; maxValidMonitorHz rejects corrupt EDID values.
	minValidMonitorHz = 2
	maxValidMonitorHz = 480
	// monitorCheckInterval bounds how often the active loop polls window
	// placement for a monitor move.
	monitorCheckInterval = 250 * time.Millisecond
	// monitorReprobeInterval re-queries the active monitor's refresh rate
	// even while the window stays put, catching in-place display mode or
	// battery refresh-rate switches without per-frame syscalls.
	monitorReprobeInterval = 2 * time.Second
)

// monitorSample records the window's placement and active monitor pointer so
// monitor-change detection only calls the OS display API when the window moves
// or after monitorReprobeInterval.
type monitorSample struct {
	posX, posY int
	winW, winH int
	monitor    *ebiten.MonitorType
}

// monitorProbeFunc returns the refresh rate (Hz) of the monitor hosting the
// window at the given desktop rectangle.
type monitorProbeFunc func(posX, posY, winW, winH int) int

// sanitizeMonitorHz returns hz when it is a plausible display refresh rate
// in [minValidMonitorHz, maxValidMonitorHz], otherwise defaultMonitorHz (60).
func sanitizeMonitorHz(hz int) int {
	if hz < minValidMonitorHz || hz > maxValidMonitorHz {
		return defaultMonitorHz
	}
	return hz
}

// defaultMonitorSample reads the live window position, size, and Ebiten
// monitor handle on the game goroutine.
func defaultMonitorSample() monitorSample {
	x, y := ebiten.WindowPosition()
	w, h := ebiten.WindowSize()
	return monitorSample{
		posX:    x,
		posY:    y,
		winW:    w,
		winH:    h,
		monitor: ebiten.Monitor(),
	}
}
