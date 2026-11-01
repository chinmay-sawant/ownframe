# Windows desktop regression fixture: a 1920 panel left of a 2560 panel,
# sharing the horizontal row y=620. Moves the real cursor for this test.
param([Parameter(Mandatory=$true)][string]$Executable, [switch]$Reverse)
$ErrorActionPreference = 'Stop'
Add-Type @'
using System;
using System.Runtime.InteropServices;
public static class DragCheck {
 [StructLayout(LayoutKind.Sequential)] public struct Rect { public int Left,Top,Right,Bottom; }
 [DllImport("user32.dll")] public static extern bool SetProcessDpiAwarenessContext(IntPtr context);
 [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr h, out Rect r);
 [DllImport("user32.dll")] public static extern bool SetWindowPos(IntPtr h, IntPtr after, int x,int y,int w,int hgt,uint flags);
 [DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr h);
 [DllImport("user32.dll")] public static extern bool SetCursorPos(int x,int y);
 [DllImport("user32.dll")] public static extern void mouse_event(uint flags,uint x,uint y,uint data,UIntPtr extra);
}
'@
[void][DragCheck]::SetProcessDpiAwarenessContext([IntPtr](-4))
$p = Start-Process $Executable -PassThru -RedirectStandardOutput "$Executable.stdout.log" -RedirectStandardError "$Executable.stderr.log"
try {
 Start-Sleep -Seconds 4
 $p.Refresh()
 if ($p.HasExited) { throw 'Application exited at startup' }
 $h = $p.MainWindowHandle
 [void][DragCheck]::SetForegroundWindow($h)
 [void][DragCheck]::SetWindowPos($h,[IntPtr]::Zero,$(if ($Reverse) { -1200 } else { 300 }),600,1120,760,0x14)
 [void][DragCheck]::SetWindowPos($h,[IntPtr](-1),0,0,0,0,3)
 Start-Sleep -Seconds 1
 $startX = $(if ($Reverse) { -1000 } else { 500 })
 [void][DragCheck]::SetCursorPos($startX,620)
 Start-Sleep -Milliseconds 200
 [DragCheck]::mouse_event(2,0,0,0,[UIntPtr]::Zero)
 Start-Sleep -Milliseconds 200
 $maxError = 0
 for ($i=1; $i -le 100; $i++) {
  $x = $startX + $(if ($Reverse) { 15*$i } else { -15*$i })
  [void][DragCheck]::SetCursorPos($x,620)
  Start-Sleep -Milliseconds 40
  $r = New-Object DragCheck+Rect
  if (![DragCheck]::GetWindowRect($h,[ref]$r)) { $p.Refresh(); throw "Window vanished. exited=$($p.HasExited) exitCode=$($p.ExitCode)" }
  $errorPx = [Math]::Abs(($x - $r.Left) - 200)
  $maxError = [Math]::Max($maxError,$errorPx)
 }
 Start-Sleep -Milliseconds 300
 $r = New-Object DragCheck+Rect
 if (![DragCheck]::GetWindowRect($h,[ref]$r)) { $p.Refresh(); throw "Window vanished. exited=$($p.HasExited) exitCode=$($p.ExitCode)" }
 $finalError = [Math]::Abs(($x - $r.Left) - 200)
 "maxGrabError=$maxError finalGrabError=$finalError window=$($r.Left),$($r.Top)"
 if ($maxError -gt 100 -or $finalError -gt 30) { throw 'Drag lost its grab point across monitors' }
} finally {
 [DragCheck]::mouse_event(4,0,0,0,[UIntPtr]::Zero)
 if (!$p.HasExited) { [void]$p.CloseMainWindow() }
}
