package snip

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// windowsFullscreen copies the primary screen via PowerShell System.Drawing.
func windowsFullscreen(ctx context.Context) ([]byte, error) {
	path, cleanup, err := tempPNG()
	if err != nil {
		return nil, err
	}
	defer cleanup()
	script := "Add-Type -AssemblyName System.Drawing;" +
		"Add-Type -AssemblyName System.Windows.Forms;" +
		"$b=[System.Windows.Forms.Screen]::PrimaryScreen.Bounds;" +
		"$m=New-Object System.Drawing.Bitmap($b.Width,$b.Height);" +
		"$g=[System.Drawing.Graphics]::FromImage($m);" +
		"$g.CopyFromScreen($b.Location,[System.Drawing.Point]::Empty,$b.Size);" +
		"$g.Dispose();$m.Save('" + path + "');$m.Dispose()"
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile",
		"-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("powershell: %v: %s", err, out)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("powershell: %w", err)
	}
	return raw, nil
}
