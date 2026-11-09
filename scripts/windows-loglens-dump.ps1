# Run from Windows PowerShell. Each repetition launches its own process.
param(
    [Parameter(Mandatory=$true)][string]$Exe,
    [string]$OutputDirectory = (Join-Path $env:USERPROFILE 'Downloads\loglens-dumps'),
    [ValidateRange(1,20)][int]$Repeats = 3,
    [ValidateRange(2,120)][int]$SecondsPerPhase = 10,
    [ValidateRange(1,60)][int]$WarmupSeconds = 3,
    [ValidateRange(1260,1920)][int]$ContentWidth = 1440,
    [ValidateRange(720,1200)][int]$Height = 900,
    [switch]$DevTools
)
$ErrorActionPreference = 'Stop'
$Exe = (Resolve-Path $Exe).Path
Add-Type -AssemblyName System.Net.Http
Add-Type @'
using System;
using System.Runtime.InteropServices;
public static class LoglensInput {
    [StructLayout(LayoutKind.Sequential)] public struct Rect { public int Left,Top,Right,Bottom; }
    [StructLayout(LayoutKind.Sequential)] public struct Point { public int X,Y; }
    [DllImport("user32.dll")] public static extern bool SetProcessDpiAwarenessContext(IntPtr context);
    [DllImport("user32.dll")] public static extern bool GetClientRect(IntPtr h, out Rect r);
    [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr h, out Rect r);
    [DllImport("user32.dll")] public static extern bool ClientToScreen(IntPtr h, ref Point p);
    [DllImport("user32.dll")] public static extern bool SetWindowPos(IntPtr h,IntPtr after,int x,int y,int w,int z,uint flags);
    [DllImport("user32.dll")] public static extern bool ShowWindow(IntPtr h,int command);
    [DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr h);
    [DllImport("user32.dll")] public static extern IntPtr GetForegroundWindow();
    [DllImport("user32.dll")] public static extern bool SetCursorPos(int x,int y);
    [DllImport("user32.dll")] public static extern void mouse_event(uint flags,uint x,uint y,int data,UIntPtr extra);
}
'@
[void][LoglensInput]::SetProcessDpiAwarenessContext([IntPtr](-4))
$stamp = (Get-Date).ToString('yyyyMMdd-HHmmss-fff')
$root = New-Item -ItemType Directory -Path (Join-Path $OutputDirectory $stamp)
$encoding = New-Object System.Text.UTF8Encoding($false)
$http = New-Object System.Net.Http.HttpClient
$http.Timeout = [TimeSpan]::FromSeconds($SecondsPerPhase + 30)

function Write-JSON($path, $value) {
    [IO.File]::WriteAllText($path, ($value | ConvertTo-Json -Depth 30), $encoding)
}
function Read-State {
    return ($http.GetStringAsync("$base/debug/state").GetAwaiter().GetResult() | ConvertFrom-Json)
}
function Save-Profile($path, $kind) {
    $bytes = $http.GetByteArrayAsync("$base/debug/pprof/$kind").GetAwaiter().GetResult()
    [IO.File]::WriteAllBytes($path, $bytes)
}
function Get-Sample {
    $state = Read-State
    $proc.Refresh()
    return [ordered]@{
        elapsed_seconds = $clock.Elapsed.TotalSeconds
        utc = (Get-Date).ToUniversalTime().ToString('o')
        cpu_seconds = $proc.TotalProcessorTime.TotalSeconds
        working_set_bytes = $proc.WorkingSet64
        private_bytes = $proc.PrivateMemorySize64
        handles = $proc.HandleCount
        threads = $proc.Threads.Count
        state = $state
    }
}
function Move-Pointer($x, $y, $state) {
    if ([LoglensInput]::GetForegroundWindow() -ne $handle) {
        throw 'Loglens lost foreground focus. Capture stopped before sending input.'
    }
    $rect = New-Object LoglensInput+Rect
    [void][LoglensInput]::GetClientRect($handle, [ref]$rect)
    $point = New-Object LoglensInput+Point
    $point.X = [int]($x * ($rect.Right-$rect.Left) / $state.window_width)
    $point.Y = [int]($y * ($rect.Bottom-$rect.Top) / $state.window_height)
    [void][LoglensInput]::ClientToScreen($handle, [ref]$point)
    if (-not [LoglensInput]::SetCursorPos($point.X, $point.Y)) { throw 'Cannot move pointer' }
}
function Point-Control($state, $action, $click) {
    $box = $state.boxes | Where-Object {
        $_.Action -eq $action -and $_.W -gt 0 -and $_.H -gt 0 -and $_.X -lt 180
    } | Select-Object -First 1
    if (-not $box) { throw "Visible sidebar target missing: $action" }
    Move-Pointer ($box.X+$box.W/2) ($box.Y+$box.H/2) $state
    if ($click) {
        Start-Sleep -Milliseconds 60
        [LoglensInput]::mouse_event(2,0,0,0,[UIntPtr]::Zero)
        try { Start-Sleep -Milliseconds 60 }
        finally { [LoglensInput]::mouse_event(4,0,0,0,[UIntPtr]::Zero) }
    }
}
function Set-ClientSize($width, $height, $state) {
    $client = New-Object LoglensInput+Rect
    $window = New-Object LoglensInput+Rect
    [void][LoglensInput]::GetClientRect($handle,[ref]$client)
    [void][LoglensInput]::GetWindowRect($handle,[ref]$window)
    $scaleX = ($client.Right-$client.Left) / $state.window_width
    $scaleY = ($client.Bottom-$client.Top) / $state.window_height
    $w = [int]($width*$scaleX + $window.Right-$window.Left-$client.Right+$client.Left)
    $h = [int]($height*$scaleY + $window.Bottom-$window.Top-$client.Bottom+$client.Top)
    if (-not [LoglensInput]::SetWindowPos($handle,[IntPtr]::Zero,0,0,$w,$h,4)) { throw 'Resize failed' }
}
function Fresh-State($after) {
    $deadline = [DateTime]::UtcNow.AddSeconds(10)
    do {
        $state = Read-State
        if ([DateTimeOffset]::Parse($state.sampled_at).UtcDateTime -gt $after) { return $state }
        Start-Sleep -Milliseconds 50
    } while ([DateTime]::UtcNow -lt $deadline)
    throw 'No fresh state snapshot within 10 seconds'
}

try {
    for ($repeat=1; $repeat -le $Repeats; $repeat++) {
        $directory = (New-Item -ItemType Directory -Path (Join-Path $root.FullName "run-$repeat")).FullName
        $listener = New-Object Net.Sockets.TcpListener([Net.IPAddress]::Loopback,0)
        $listener.Start()
        $port = $listener.LocalEndpoint.Port
        $listener.Stop()
        $base = "http://127.0.0.1:$port"
        $arguments = "-perf -debug-addr 127.0.0.1:$port"
        if ($DevTools) { $arguments += ' -dev' }
        $report = [ordered]@{
            status = 'running'; exe = $Exe; sha256 = (Get-FileHash $Exe -Algorithm SHA256).Hash
            arguments = $arguments; devtools = [bool]$DevTools; repeat = $repeat
            content_width = $ContentWidth; height = $Height
            seconds_per_phase = $SecondsPerPhase; warmup_seconds = $WarmupSeconds
            phases = @(); error = $null
        }
        Write-JSON (Join-Path $directory 'capture.json') $report
        $proc = $null
        try {
            $proc = Start-Process -FilePath $Exe -ArgumentList $arguments -PassThru `
                -RedirectStandardOutput (Join-Path $directory 'stdout.txt') `
                -RedirectStandardError (Join-Path $directory 'stderr.txt')
            $clock = [Diagnostics.Stopwatch]::StartNew()
            $ready = $false
            while ($clock.Elapsed.TotalSeconds -lt 30) {
                if ($proc.HasExited) { throw "Loglens exited with code $($proc.ExitCode)" }
                try { $state = Read-State; $ready = $true; break } catch { Start-Sleep -Milliseconds 100 }
            }
            if (-not $ready) { throw 'Debug endpoint did not start; build with -tags ownframe_debug' }
            $proc.Refresh()
            $handle = $proc.MainWindowHandle
            if ($handle -eq [IntPtr]::Zero) { throw 'Native window missing' }
            [void][LoglensInput]::ShowWindow($handle,9)
            [void][LoglensInput]::SetForegroundWindow($handle)
            Start-Sleep -Milliseconds 300
            $width = $ContentWidth
            if ($DevTools) { $width += 340 }
            Set-ClientSize $width $Height $state
            Start-Sleep -Seconds $WarmupSeconds
            $state = Fresh-State ([DateTime]::UtcNow)
            if ($state.window_width -ne $width -or $state.window_height -ne $Height -or $state.width -ne $ContentWidth) {
                throw "Unexpected viewport: window=$($state.window_width)x$($state.window_height) content=$($state.width); requested content=${ContentWidth}x${Height}"
            }
            $report['pid'] = $proc.Id
            foreach ($phase in @('idle','hover','scroll','resize','navigation','idle-after')) {
                Set-ClientSize $width $Height (Read-State)
                Start-Sleep -Milliseconds 300
                Point-Control (Read-State) 'nav:overview' $true
                Start-Sleep -Milliseconds 300
                $state = Read-State
                Move-Pointer ($ContentWidth-30) 20 $state
                Start-Sleep -Seconds $WarmupSeconds
                $phaseDir = (New-Item -ItemType Directory -Path (Join-Path $directory $phase)).FullName
                Save-Profile (Join-Path $phaseDir 'heap-before.pb.gz') 'heap?gc=1'
                Save-Profile (Join-Path $phaseDir 'allocs-before.pb.gz') 'allocs'
                [void](Fresh-State ([DateTime]::UtcNow))
                $samples = @((Get-Sample))
                $cpu = $http.GetByteArrayAsync("$base/debug/pprof/profile?seconds=$SecondsPerPhase")
                $started = $clock.Elapsed.TotalSeconds
                $step = 0
                while ($clock.Elapsed.TotalSeconds-$started -lt $SecondsPerPhase) {
                    $state = Read-State
                    switch ($phase) {
                        'hover' {
                            $targets = @('nav:overview','nav:logs','nav:traces','nav:incident')
                            Point-Control $state $targets[$step % $targets.Count] $false
                        }
                        'scroll' {
                            Move-Pointer ($ContentWidth-80) ($Height/2) $state
                            $delta = -360
                            if ($step % 8 -ge 4) { $delta = 360 }
                            [LoglensInput]::mouse_event(0x800,0,0,$delta,[UIntPtr]::Zero)
                        }
                        'resize' {
                            $targetWidth = $width
                            if ($step % 2 -eq 0) { $targetWidth -= 160 }
                            Set-ClientSize $targetWidth $Height $state
                        }
                        'navigation' {
                            $targets = @('nav:logs','nav:traces','nav:incident','nav:overview')
                            Point-Control $state $targets[$step % $targets.Count] $true
                        }
                    }
                    $step++
                    $next = $started + $step*0.5
                    do {
                        Start-Sleep -Milliseconds 20
                    } while ($clock.Elapsed.TotalSeconds -lt $next)
                    $samples += ,(Get-Sample)
                }
                [IO.File]::WriteAllBytes((Join-Path $phaseDir 'cpu.pb.gz'), $cpu.GetAwaiter().GetResult())
                [void](Fresh-State ([DateTime]::UtcNow))
                $samples += ,(Get-Sample)
                Write-JSON (Join-Path $phaseDir 'samples.json') @($samples)
                Save-Profile (Join-Path $phaseDir 'allocs-after.pb.gz') 'allocs'
                Save-Profile (Join-Path $phaseDir 'heap-after.pb.gz') 'heap?gc=1'
                $report.phases += $phase
                Write-JSON (Join-Path $directory 'capture.json') $report
                Write-Output "run=$repeat phase=$phase samples=$($samples.Count) output=$phaseDir"
            }
            $report.status = 'passed'
        } catch {
            $report.status = 'failed'
            $report.error = $_.Exception.Message
            throw
        } finally {
            Write-JSON (Join-Path $directory 'capture.json') $report
            if ($proc -and -not $proc.HasExited) {
                # Only close the fresh process launched by this repetition.
                [void]$proc.CloseMainWindow()
                if (-not $proc.WaitForExit(5000)) { $proc.Kill(); $proc.WaitForExit() }
            }
        }
    }
} finally { $http.Dispose() }
Write-Output "Capture complete: $($root.FullName). Analyze with scripts/loglens-report.py."
