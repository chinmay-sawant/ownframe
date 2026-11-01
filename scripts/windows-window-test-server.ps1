$ErrorActionPreference='Stop'
[Console]::InputEncoding=New-Object System.Text.UTF8Encoding
$common=Get-Content (Join-Path $PSScriptRoot 'windows-drag-check.ps1') -Raw
Add-Type ([regex]::Match($common,"Add-Type @'([\s\S]*?)'@").Groups[1].Value)
Add-Type @'
using System;using System.Text;using System.Collections.Generic;using System.Runtime.InteropServices;
public static class LiveHost {
 [StructLayout(LayoutKind.Sequential)] public struct Point {public int X,Y;}
 [DllImport("user32.dll")] public static extern bool GetPhysicalCursorPos(out Point p);
 [DllImport("user32.dll")] public static extern IntPtr WindowFromPoint(Point p);
 [DllImport("user32.dll")] public static extern IntPtr GetAncestor(IntPtr h,uint n);
 [DllImport("user32.dll")] public static extern IntPtr GetForegroundWindow();

 public delegate bool Callback(IntPtr h,IntPtr p);
 [DllImport("user32.dll")] static extern bool EnumWindows(Callback c,IntPtr p);
 [DllImport("user32.dll",CharSet=CharSet.Unicode)] static extern int GetWindowText(IntPtr h,StringBuilder s,int n);
 [DllImport("user32.dll")] public static extern bool IsWindowVisible(IntPtr h);
 [DllImport("user32.dll")] public static extern bool IsIconic(IntPtr h);
 [DllImport("user32.dll")] public static extern bool IsWindow(IntPtr h);
 [DllImport("user32.dll")] public static extern bool ShowWindowAsync(IntPtr h,int n);
 [DllImport("user32.dll",EntryPoint="GetWindowLongPtrW")] public static extern IntPtr GetWindowLongPtr(IntPtr h,int index);
 [DllImport("user32.dll")] public static extern bool SetPhysicalCursorPos(int x,int y);
 [DllImport("user32.dll")] public static extern IntPtr SetThreadDpiAwarenessContext(IntPtr c);
 public static List<IntPtr> Find(string title) {
  var found=new List<IntPtr>();
  EnumWindows(delegate(IntPtr h,IntPtr p){var s=new StringBuilder(1024);GetWindowText(h,s,s.Capacity);if(s.ToString()==title)found.Add(h);return true;},IntPtr.Zero);
  return found;
 }
}
'@
[void][LiveHost]::SetThreadDpiAwarenessContext([IntPtr](-4))
while($line=[Console]::In.ReadLine()) {
 try {
  $r=$line|ConvertFrom-Json
  $h=[IntPtr]([long]$r.handle)
  switch($r.op) {
   'foreground' {$answer=@{handle=[LiveHost]::GetForegroundWindow().ToInt64()}}
   'find' {$v=[LiveHost]::Find($r.title);if($v.Count -ne 1){throw "Window matches: $($v.Count)"};$answer=@{handle=$v[0].ToInt64()}}
   'focus' {
    [void][DragCheck]::SetWindowPos($h,[IntPtr]::Zero,0,0,0,0,3)
    [void][DragCheck]::SetForegroundWindow($h)
    Start-Sleep -Milliseconds 400
    $answer=@{focused=([LiveHost]::GetForegroundWindow() -eq $h)}
   }
   'click' {
    $ready=$false
    for($attempt=0;$attempt -lt 10;$attempt++) {
     [void][DragCheck]::SetWindowPos($h,[IntPtr]::Zero,0,0,0,0,3)
     [void][DragCheck]::SetForegroundWindow($h)
     Start-Sleep -Milliseconds 300
     $start=New-Object LiveHost+Point
     [void][LiveHost]::GetPhysicalCursorPos([ref]$start)
     for($step=1;$step -le 5;$step++){
      [void][LiveHost]::SetPhysicalCursorPos([int]($start.X+($r.x-$start.X)*$step/5),[int]($start.Y+($r.y-$start.Y)*$step/5))
      Start-Sleep -Milliseconds 30
     }
     Start-Sleep -Milliseconds 150
     $point=New-Object LiveHost+Point
     [void][LiveHost]::GetPhysicalCursorPos([ref]$point)
     if([LiveHost]::GetAncestor([LiveHost]::WindowFromPoint($point),2) -eq $h) {$ready=$true;break}
    }
    if(!$ready){throw ('Target window did not receive pointer focus; no click sent. hit='+[LiveHost]::GetAncestor([LiveHost]::WindowFromPoint($point),2).ToInt64()+' visible='+[LiveHost]::IsWindowVisible($h))}
    [DragCheck]::mouse_event(2,0,0,0,[UIntPtr]::Zero)
    Start-Sleep -Milliseconds 100
    [DragCheck]::mouse_event(4,0,0,0,[UIntPtr]::Zero)
    $point=New-Object LiveHost+Point
    [void][LiveHost]::GetPhysicalCursorPos([ref]$point)
    $hit=[LiveHost]::GetAncestor([LiveHost]::WindowFromPoint($point),2)
    $answer=@{clicked=$true;x=$point.X;y=$point.Y;hit=$hit.ToInt64();target=$h.ToInt64();foreground=[LiveHost]::GetForegroundWindow().ToInt64()}
   }
   'restore' {[void][LiveHost]::ShowWindowAsync($h,9);[void][DragCheck]::SetForegroundWindow($h);$answer=@{requested=$true}}
   'status' {$rect=New-Object DragCheck+Rect;[void][DragCheck]::GetWindowRect($h,[ref]$rect);$answer=@{exists=[LiveHost]::IsWindow($h);minimized=[LiveHost]::IsIconic($h);topmost=(([LiveHost]::GetWindowLongPtr($h,-20).ToInt64() -band 8) -ne 0);rect=@{x=$rect.Left;y=$rect.Top;w=$rect.Right-$rect.Left;h=$rect.Bottom-$rect.Top}}}
   'quit' {break}
   default {throw 'Unknown operation'}
  }
  [Console]::WriteLine(($answer|ConvertTo-Json -Compress))
 } catch { [Console]::WriteLine((@{error=$_.Exception.Message}|ConvertTo-Json -Compress)) }
 if($r.op -eq 'quit'){break}
}
