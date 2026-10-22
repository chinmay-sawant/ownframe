package snip

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// psQuote wraps path in a PowerShell single-quoted literal. An embedded
// quote is doubled, so the path cannot close the string early.
func psQuote(path string) string {
	return "'" + strings.ReplaceAll(path, "'", "''") + "'"
}

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
		"$g.Dispose();$m.Save(" + psQuote(path) + ");$m.Dispose()"
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", //nolint:gosec // reason: fixed powershell with static flags, path is quoted.
		"-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("powershell: %v: %s", err, out)
	}
	raw, err := os.ReadFile(path) //nolint:gosec // reason: temp PNG from CreateTemp, local file.
	if err != nil {
		return nil, fmt.Errorf("powershell: %w", err)
	}
	return raw, nil
}
