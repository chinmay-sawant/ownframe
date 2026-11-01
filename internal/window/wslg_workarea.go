package window

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"
)

const hostWorkareaCacheTime = 5 * time.Second

// Query in physical desktop pixels, matching the normalized XRandR layout.
const windowsWorkareaScript = `
Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;
public static class OwnFrameDpi {
    [DllImport("user32.dll")]
    public static extern bool SetProcessDpiAwarenessContext(IntPtr value);
}
'@
[void][OwnFrameDpi]::SetProcessDpiAwarenessContext([IntPtr](-4))
Add-Type -AssemblyName System.Windows.Forms
$screens = @(foreach ($s in [System.Windows.Forms.Screen]::AllScreens) {
    @{
        bounds = @{x=$s.Bounds.X; y=$s.Bounds.Y; w=$s.Bounds.Width; h=$s.Bounds.Height}
        work = @{x=$s.WorkingArea.X; y=$s.WorkingArea.Y; w=$s.WorkingArea.Width; h=$s.WorkingArea.Height}
    }
})
ConvertTo-Json -Compress -Depth 3 -InputObject @($screens)
`

type desktopRect struct{ X, Y, W, H int }
type hostWorkarea struct{ Bounds, Work desktopRect }

func matchHostWorkarea(target monitorRect, screens []hostWorkarea) (workarea, bool) {
	for _, s := range screens {
		b, w := s.Bounds, s.Work
		if b.X != target.x || b.Y != target.y || b.W != target.w || b.H != target.h {
			continue
		}
		if w.W <= 0 || w.H <= 0 || w.X < b.X || w.Y < b.Y || w.X+w.W > b.X+b.W || w.Y+w.H > b.Y+b.H {
			continue
		}
		return workarea{x: w.X, y: w.Y, w: w.W, h: w.H}, true
	}
	return workarea{}, false
}

func wslgHostWorkareaLocked(target monitorRect) (workarea, bool) {
	if !isWSLg() {
		return workarea{}, false
	}
	if time.Since(manualMax.hostAreasAt) < hostWorkareaCacheTime {
		if area, ok := matchHostWorkarea(target, manualMax.hostAreas); ok {
			return area, true
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), manualMaxLearnTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", windowsWorkareaScript)
	if _, err := exec.LookPath("powershell.exe"); err != nil {
		cmd = exec.CommandContext(ctx, "/mnt/c/Windows/System32/WindowsPowerShell/v1.0/powershell.exe", "-NoProfile", "-NonInteractive", "-Command", windowsWorkareaScript)
	}
	out, err := cmd.Output()
	if err != nil {
		return workarea{}, false
	}
	var screens []hostWorkarea
	if json.Unmarshal([]byte(strings.TrimPrefix(strings.TrimSpace(string(out)), "\ufeff")), &screens) != nil {
		return workarea{}, false
	}
	manualMax.hostAreas, manualMax.hostAreasAt = screens, time.Now()
	return matchHostWorkarea(target, screens)
}
