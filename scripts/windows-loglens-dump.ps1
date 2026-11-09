# Capture a fresh Loglens process without sending mouse or keyboard input.
# The operator performs each named phase during its sampling window.
param(
    [Parameter(Mandatory=$true)][string]$Exe,
    [string]$OutputDirectory = (Join-Path $env:USERPROFILE 'Downloads\loglens-dumps'),
    [ValidateRange(1,20)][int]$Repeats = 1,
    [ValidateRange(2,120)][int]$SecondsPerPhase = 10,
    [ValidateRange(1,60)][int]$WarmupSeconds = 3,
    [ValidateRange(1260,1920)][int]$ContentWidth = 1440,
    [ValidateRange(720,1200)][int]$Height = 900,
    [ValidateSet('single','dual','unknown')][string]$TopologyLabel = 'unknown',
    [switch]$DevTools
)
$ErrorActionPreference = 'Stop'
$Exe = (Resolve-Path $Exe).Path
Add-Type -AssemblyName System.Net.Http
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
function Get-GpuMemory {
    $paths = @(
        '\GPU Process Memory(*)\Dedicated Usage',
        '\GPU Process Memory(*)\Shared Usage',
        '\GPU Process Memory(*)\Local Usage',
        '\GPU Process Memory(*)\Non Local Usage',
        '\GPU Process Memory(*)\Total Committed',
        '\GPU Adapter Memory(*)\Dedicated Usage',
        '\GPU Adapter Memory(*)\Shared Usage',
        '\GPU Adapter Memory(*)\Total Committed',
        '\GPU Local Adapter Memory(*)\Local Usage',
        '\GPU Non Local Adapter Memory(*)\Non Local Usage'
    )
    try {
        $counter = Get-Counter -Counter $paths -ErrorAction Stop
        $values = @($counter.CounterSamples | Where-Object {
            $_.InstanceName -match "^pid_$($proc.Id)_" -or $_.Path -match '\\GPU (Adapter|Local Adapter|Non Local Adapter) Memory\\'
        } | ForEach-Object {
            [ordered]@{ path = $_.Path; instance = $_.InstanceName; value_bytes = [int64]$_.CookedValue }
        })
        return @{ available = $true; counters = $values }
    } catch {
        return @{ available = $false; error = $_.Exception.Message; counters = @() }
    }
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
        gpu_memory = Get-GpuMemory
        state = $state
    }
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

$phases = @('idle','hover','scroll','resize','navigation','idle-after')
$phaseInstructions = @{
    idle = 'Leave the pointer and page still.'
    hover = 'Move the pointer across several sidebar hover targets.'
    scroll = 'Scroll down and back up inside the page.'
    resize = 'Change the window width, then leave it at the new width for sampling.'
    navigation = 'Switch among Logs, Traces, Incident room, and Overview.'
    'idle-after' = 'Restore the starting window size and Overview, then leave the app still.'
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
            topology_label = $TopologyLabel; content_width = $ContentWidth; height = $Height
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
            Start-Sleep -Seconds $WarmupSeconds
            $state = Fresh-State ([DateTime]::UtcNow)
            if ($state.width -ne $ContentWidth -or $state.height -ne $Height) {
                throw "Unexpected content viewport: $($state.width)x$($state.height); requested ${ContentWidth}x${Height}"
            }
            $report['pid'] = $proc.Id
            $report['window_width'] = $state.window_width
            $report['window_height'] = $state.window_height

            foreach ($phase in $phases) {
                $phaseDir = (New-Item -ItemType Directory -Path (Join-Path $directory $phase)).FullName
                Write-Host "Phase: $phase. $($phaseInstructions[$phase])"
                Write-Host "Prepare the app. Sampling starts when you press Enter and lasts $SecondsPerPhase seconds."
                [void](Read-Host)
                Save-Profile (Join-Path $phaseDir 'heap-before.pb.gz') 'heap?gc=1'
                Save-Profile (Join-Path $phaseDir 'allocs-before.pb.gz') 'allocs'
                [void](Fresh-State ([DateTime]::UtcNow))
                $samples = @((Get-Sample))
                $cpu = $http.GetByteArrayAsync("$base/debug/pprof/profile?seconds=$SecondsPerPhase")
                $started = $clock.Elapsed.TotalSeconds
                while ($clock.Elapsed.TotalSeconds-$started -lt $SecondsPerPhase) {
                    Start-Sleep -Milliseconds 500
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
                [void]$proc.CloseMainWindow()
                if (-not $proc.WaitForExit(5000)) { $proc.Kill(); $proc.WaitForExit() }
            }
        }
    }
} finally { $http.Dispose() }
Write-Output "Capture complete: $($root.FullName). Analyze with scripts/loglens-report.py."
