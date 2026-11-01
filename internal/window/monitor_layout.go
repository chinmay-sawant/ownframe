package window

import (
	"context"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// monitorRect is one display in global virtual-screen pixels. Origins come
// from xrandr on Linux, where Ebiten only reports sizes. The origin selects
// the monitor: a 2560 panel at +1920 is the right screen, a 1920 panel at +0
// is the left one, even when numbering is inverted.
type monitorRect struct {
	name string
	x, y int
	w, h int
}

const (
	// layoutExecTimeout bounds the xrandr probe so a hung X server never
	// stalls a double-click or a launch.
	layoutExecTimeout = 750 * time.Millisecond
	// layoutMinDimension rejects absurd monitor rects without clamping real ones.
	layoutMinDimension   = 1
	xrandrMinFields      = 3
	xrandrOriginParts    = 3
	monitorCenterDivisor = 2
)

// parseXrandrMonitors parses `xrandr --listmonitors` output. Each row looks
// like " 0: +XWAYLAND0 2560/677x1600/423+1920+0  XWAYLAND0".
func parseXrandrMonitors(out string) []monitorRect {
	var rects []monitorRect
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Monitors:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < xrandrMinFields {
			continue
		}
		// fields[1] is "+NAME" or "NAME", fields[2] is the geometry token.
		geom := ""
		name := ""
		for _, field := range fields[1:] {
			if strings.Contains(field, "x") && strings.Contains(field, "+") && strings.Contains(field, "/") {
				geom = field
			}
		}
		if geom == "" {
			continue
		}
		name = fields[len(fields)-1]
		rect, ok := parseXrandrGeometry(name, geom)
		if !ok {
			continue
		}
		rects = append(rects, rect)
	}
	return rects
}

// parseXrandrGeometry parses one "2560/677x1600/423+1920+0" token.
func parseXrandrGeometry(name, token string) (monitorRect, bool) {
	// Split off the +X+Y origin first; what remains is W/MMxH/MM.
	plus := strings.IndexByte(token, '+')
	if plus < 0 {
		return monitorRect{}, false
	}
	wh := token[:plus]
	origin := token[plus:]
	parts := strings.SplitN(origin, "+", xrandrOriginParts)
	if len(parts) != xrandrOriginParts {
		return monitorRect{}, false
	}
	ox, errX := strconv.Atoi(parts[1])
	oy, errY := strconv.Atoi(parts[2])
	if errX != nil || errY != nil {
		return monitorRect{}, false
	}
	xPos := strings.IndexByte(wh, 'x')
	if xPos < 0 {
		return monitorRect{}, false
	}
	wStr, hStr := wh[:xPos], wh[xPos+1:]
	if slash := strings.IndexByte(wStr, '/'); slash >= 0 {
		wStr = wStr[:slash]
	}
	if slash := strings.IndexByte(hStr, '/'); slash >= 0 {
		hStr = hStr[:slash]
	}
	w, errW := strconv.Atoi(wStr)
	h, errH := strconv.Atoi(hStr)
	if errW != nil || errH != nil || w < layoutMinDimension || h < layoutMinDimension {
		return monitorRect{}, false
	}
	if strings.TrimSpace(name) == "" {
		return monitorRect{}, false
	}
	return monitorRect{name: name, x: ox, y: oy, w: w, h: h}, true
}

// xrandrLayout probes monitor origins via xrandr. It reports false when not
// on Linux, when xrandr is missing, or when parsing yields nothing, so
// callers fall back to Ebiten sizes.
func xrandrLayout() ([]monitorRect, bool) {
	if runtime.GOOS != "linux" {
		return nil, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), layoutExecTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "xrandr", "--listmonitors")
	out, err := cmd.Output()
	if err != nil {
		return nil, false
	}
	rects := parseXrandrMonitors(string(out))
	if len(rects) == 0 {
		return nil, false
	}
	return rects, true
}

// ebitenLayoutFallback lays Ebiten monitors side by side from the origin.
// Ebiten reports sizes but not origins, so this only approximates the true
// union; xrandr above is preferred on Linux.
func ebitenLayoutFallback() []monitorRect {
	geoms := liveMonitorGeometries()
	rects := make([]monitorRect, 0, len(geoms))
	origin := 0
	for _, g := range geoms {
		if g.w < layoutMinDimension || g.h < layoutMinDimension {
			continue
		}
		rects = append(rects, monitorRect{name: g.name, x: origin, y: 0, w: g.w, h: g.h})
		origin += g.w
	}
	if len(rects) == 0 {
		return nil
	}
	return rects
}

// monitorLayout returns monitor rects with true origins when available.
func monitorLayout() []monitorRect {
	if rects, ok := xrandrLayout(); ok {
		if isWSLg() {
			return wslgMonitorLayout(rects)
		}
		return rects
	}
	return ebitenLayoutFallback()
}

// monitorForWindow chooses the largest visible area. The window center
// breaks a tie when it spans two monitors equally.
func monitorForWindow(wx, wy, ww, wh int, layout []monitorRect) (monitorRect, bool) {
	if len(layout) == 0 || ww <= 0 || wh <= 0 {
		return monitorRect{}, false
	}
	cx := wx + ww/monitorCenterDivisor
	cy := wy + wh/monitorCenterDivisor
	best := monitorRect{}
	bestArea := 0
	found := false
	for _, m := range layout {
		overlap := overlapArea(wx, wy, ww, wh, m)
		containsCenter := cx >= m.x && cx < m.x+m.w && cy >= m.y && cy < m.y+m.h
		if overlap > bestArea || (overlap == bestArea && overlap > 0 && containsCenter) {
			bestArea = overlap
			best = m
			found = true
		}
	}
	if !found || bestArea <= 0 {
		return monitorRect{}, false
	}
	return best, true
}

// overlapArea returns the pixel overlap of a window rect and a monitor.
func overlapArea(wx, wy, ww, wh int, m monitorRect) int {
	x1 := max(wx, m.x)
	y1 := max(wy, m.y)
	x2 := min(wx+ww, m.x+m.w)
	y2 := min(wy+wh, m.y+m.h)
	if x2 <= x1 || y2 <= y1 {
		return 0
	}
	return (x2 - x1) * (y2 - y1)
}

// virtualUnion returns the union bounds of the layout.
func virtualUnion(layout []monitorRect) (int, int, int, int, bool) {
	if len(layout) == 0 {
		return 0, 0, 0, 0, false
	}
	minX, minY := layout[0].x, layout[0].y
	maxX, maxY := layout[0].x+layout[0].w, layout[0].y+layout[0].h
	for _, m := range layout[1:] {
		minX = min(minX, m.x)
		minY = min(minY, m.y)
		maxX = max(maxX, m.x+m.w)
		maxY = max(maxY, m.y+m.h)
	}
	return minX, minY, maxX, maxY, true
}

// clampDragTargetInUnion keeps a ww by wh window grabbable inside the true
// union, unlike clampDragTarget which assumes the union starts at the origin.
func clampDragTargetInUnion(wx, wy, ww, wh, minX, minY, maxX, maxY int) (int, int) {
	if ww > 0 && ww-dragMinVisibleX < maxX-minX {
		low, high := minX-(ww-dragMinVisibleX), maxX-dragMinVisibleX
		if low > high {
			low = high
		}
		wx = min(max(wx, low), high)
	}
	if wh > 0 && wh-dragMinVisibleY < maxY-minY {
		low, high := minY-(wh-dragMinVisibleY), maxY-dragMinVisibleY
		if low > high {
			low = high
		}
		wy = min(max(wy, low), high)
	}
	return wx, wy
}
