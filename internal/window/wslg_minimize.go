package window

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/safego"
)

type hostWindowTarget struct {
	Title, HostTitle string
	Client, Frame    desktopRect
}

const windowsMinimizeScript = `
Add-Type -TypeDefinition @'
using System;
using System.Collections.Generic;
using System.Runtime.InteropServices;
using System.Text;
public static class OwnFrameMinimize {
 public delegate bool Callback(IntPtr h, IntPtr p);
 [StructLayout(LayoutKind.Sequential)] public struct Rect {public int X,Y,Right,Bottom;}
 [DllImport("user32.dll")] public static extern IntPtr SetThreadDpiAwarenessContext(IntPtr value);
 [DllImport("user32.dll")] static extern bool EnumWindows(Callback cb,IntPtr p);
 [DllImport("user32.dll")] static extern bool GetWindowRect(IntPtr h,out Rect r);
 [DllImport("user32.dll",CharSet=CharSet.Unicode)] static extern int GetWindowText(IntPtr h,StringBuilder s,int n);
 [DllImport("user32.dll")] public static extern bool ShowWindowAsync(IntPtr h,int action);
 public static List<IntPtr> Find(string title,string hostTitle,int x,int y,int w,int height,int fx,int fy,int fw,int fh) {
  var result=new List<IntPtr>();
  EnumWindows(delegate(IntPtr h,IntPtr p) {
   var text=new StringBuilder(1024);GetWindowText(h,text,text.Capacity);
   Rect r;
   if((text.ToString()==title || text.ToString()==hostTitle) && GetWindowRect(h,out r) &&
      ((r.X==x && r.Y==y && r.Right-r.X==w && r.Bottom-r.Y==height) ||
       (r.X==fx && r.Y==fy && r.Right-r.X==fw && r.Bottom-r.Y==fh))) result.Add(h);
   return true;
  },IntPtr.Zero);
  return result;
 }
}
'@
[void][OwnFrameMinimize]::SetThreadDpiAwarenessContext([IntPtr](-4))
[Console]::InputEncoding = New-Object System.Text.UTF8Encoding
$t=[Console]::In.ReadToEnd() | ConvertFrom-Json
$windows=[OwnFrameMinimize]::Find($t.Title,$t.HostTitle,$t.Client.X,$t.Client.Y,$t.Client.W,$t.Client.H,$t.Frame.X,$t.Frame.Y,$t.Frame.W,$t.Frame.H)
if($windows.Count -eq 1 -and [OwnFrameMinimize]::ShowWindowAsync($windows[0],6)) { [Console]::WriteLine('ok') }
`

func minimizeWSLg() bool {
	target, ok := nativeWindowTarget()
	if !ok {
		return false
	}
	target.HostTitle = target.Title + " (" + os.Getenv("WSL_DISTRO_NAME") + ")"
	data, err := json.Marshal(target)
	if err != nil {
		return false
	}
	// Resolve the X11 target on the game loop, then keep process startup and
	// the Windows call off it so focus and mouse releases keep flowing.
	safego.Go("WSLg minimize", func() {
		if !minimizeWindowsTarget(data) {
			log.Print("ownframe: Windows host did not accept the WSLg minimize request")
		}
	})
	return true
}

func minimizeWindowsTarget(data []byte) bool {
	ctx, cancel := context.WithTimeout(context.Background(), manualMaxLearnTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", windowsMinimizeScript)
	if _, err := exec.LookPath("powershell.exe"); err != nil {
		cmd = exec.CommandContext(ctx, "/mnt/c/Windows/System32/WindowsPowerShell/v1.0/powershell.exe", "-NoProfile", "-NonInteractive", "-Command", windowsMinimizeScript)
	}
	cmd.Stdin = bytes.NewReader(data)
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) == "ok"
}
