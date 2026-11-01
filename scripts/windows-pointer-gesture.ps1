param([int]$StartX,[int]$StartY,[int]$EndX,[int]$EndY,[switch]$DoubleClick)
$ErrorActionPreference='Stop'
$source=Get-Content (Join-Path $PSScriptRoot 'windows-drag-check.ps1') -Raw
Add-Type ([regex]::Match($source,"Add-Type @'([\s\S]*?)'@").Groups[1].Value)
[void][DragCheck]::SetProcessDpiAwarenessContext([IntPtr](-4))
$p=Get-Process msrdc | Where-Object {$_.MainWindowTitle -like '*JMAP Mail Preview*'} | Select-Object -First 1
if(!$p){throw 'JMAP window not found'}
# Raise within the normal window stack; never mark the application topmost.
[void][DragCheck]::SetWindowPos($p.MainWindowHandle,[IntPtr]::Zero,0,0,0,0,3)
[void][DragCheck]::SetForegroundWindow($p.MainWindowHandle)
[void][DragCheck]::SetCursorPos($StartX,$StartY)
Start-Sleep -Milliseconds 250
if($DoubleClick) {
 for($i=0;$i -lt 2;$i++) {
  [DragCheck]::mouse_event(2,0,0,0,[UIntPtr]::Zero)
  Start-Sleep -Milliseconds 70
  [DragCheck]::mouse_event(4,0,0,0,[UIntPtr]::Zero)
  Start-Sleep -Milliseconds 100
 }
} else {
 try {
  [DragCheck]::mouse_event(2,0,0,0,[UIntPtr]::Zero)
  Start-Sleep -Milliseconds 150
  for($i=1;$i -le 100;$i++) {
   [void][DragCheck]::SetCursorPos([int]($StartX+($EndX-$StartX)*$i/100),[int]($StartY+($EndY-$StartY)*$i/100))
   Start-Sleep -Milliseconds 15
  }
 } finally { [DragCheck]::mouse_event(4,0,0,0,[UIntPtr]::Zero) }
}
Start-Sleep -Milliseconds 500
