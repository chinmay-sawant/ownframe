# Run on Windows against the separate Telegram debug package.
# Captures require the debuggable APK's explicit validation launch extra.
param(
    [Parameter(Mandatory=$true)][ValidatePattern('^[A-Za-z0-9._:-]+$')][string]$Serial,
    [Parameter(Mandatory=$true)][string]$Apk,
    [Parameter(Mandatory=$true)][ValidatePattern('^[0-9a-f]{40}$')][string]$SourceCommit,
    [string]$Adb = 'adb.exe',
    [ValidateRange(1,65535)][int]$Port = 5038,
    [string]$OutDir = "$env:TEMP\ownframe-android-session"
)
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
Add-Type -AssemblyName System.Drawing
$package = 'com.chinmaysawant.telegram.debug'
$component = "$package/com.chinmaysawant.telegram.MainActivity"
$adbArgs = @('-P', "$Port", '-s', $Serial)
if ((Test-Path $OutDir) -and (Get-ChildItem $OutDir | Select-Object -First 1)) {
    throw 'Choose an empty output directory; previous results must not be reused.'
}
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
function Invoke-Adb {
    param([string[]]$CommandArgs)
    $output = & $Adb @adbArgs @CommandArgs 2>&1
    if ($LASTEXITCODE -ne 0) { throw "ADB failed: $CommandArgs : $output" }
    return ($output -join "`n")
}
function Assert-Demo {
    $policy = Invoke-Adb -CommandArgs @('shell', 'dumpsys', 'window', 'policy')
    $keyguard = ($policy -split 'KeyguardServiceDelegate', 2)[-1]
    if ($keyguard -notmatch '(?m)^\s*showing=false\s*$') { throw 'Unlock the phone before validation.' }
    $activity = Invoke-Adb -CommandArgs @('shell', 'dumpsys', 'activity', 'activities')
    if ($activity -notmatch 'topResumedActivity=.*com\.chinmaysawant\.telegram\.debug/') {
        throw 'Demo is not foreground; input refused.'
    }
}
function Save-Screen {
    param([string]$Name)
    Assert-Demo
    $info = New-Object System.Diagnostics.ProcessStartInfo
    $info.FileName = $Adb
    $info.Arguments = "-P $Port -s $Serial exec-out screencap -p"
    $info.UseShellExecute = $false
    $info.RedirectStandardOutput = $true
    $process = [System.Diagnostics.Process]::Start($info)
    $file = [IO.File]::Create((Join-Path $OutDir "$Name.png"))
    try { $process.StandardOutput.BaseStream.CopyTo($file) } finally { $file.Dispose() }
    $process.WaitForExit()
    if ($process.ExitCode -ne 0) { throw 'Screenshot capture failed.' }
    $process.Dispose()
    $bytes = [IO.File]::ReadAllBytes((Join-Path $OutDir "$Name.png"))
    if ($bytes.Length -lt 8 -or [BitConverter]::ToString($bytes,0,8) -ne '89-50-4E-47-0D-0A-1A-0A') { throw 'Invalid PNG capture.' }
    $bitmap = [Drawing.Bitmap]::FromFile((Join-Path $OutDir "$Name.png"))
    $visible = $false
    try {
        foreach ($x in @(0.25,0.5,0.75)) {
            foreach ($y in @(0.25,0.5,0.75)) {
                $pixel = $bitmap.GetPixel([int]($bitmap.Width*$x), [int]($bitmap.Height*$y))
                if ($pixel.R -ne 0 -or $pixel.G -ne 0 -or $pixel.B -ne 0) { $visible = $true }
            }
        }
    } finally { $bitmap.Dispose() }
    if (-not $visible) { throw 'Black capture: secure or blank surface; no visual validation is possible.' }

}
function Tap-Demo {
    param([int]$X, [int]$Y)
    Assert-Demo
    Invoke-Adb -CommandArgs @('shell', 'input', 'swipe', "$X", "$Y", "$X", "$Y", '150') | Out-Null
    Start-Sleep -Milliseconds 500
}
# Coordinates below are specific to the 1080x2400 Pixel 7 portrait fixture.
# Reject other device geometry instead of tapping an unrelated target.
$size = Invoke-Adb -CommandArgs @('shell', 'wm', 'size')
if ($size -notmatch 'Physical size: 1080x2400' -or $size -match 'Override size:') {
    throw "This session driver requires the Pixel 7 1080x2400 geometry: $size"
}
$oldAuto = Invoke-Adb -CommandArgs @('shell', 'settings', 'get', 'system', 'accelerometer_rotation')
$oldRotation = Invoke-Adb -CommandArgs @('shell', 'settings', 'get', 'system', 'user_rotation')
$record = [ordered]@{
    source_commit = $SourceCommit
    apk_sha256 = (Get-FileHash -Algorithm SHA256 $Apk).Hash.ToLowerInvariant()
    device = (Invoke-Adb -CommandArgs @('shell', 'getprop', 'ro.product.model'))
    android = (Invoke-Adb -CommandArgs @('shell', 'getprop', 'ro.build.version.release'))
    state = 'running'
    limits = 'Pixel 7 coordinates; synthetic committed characters, not complex IME composition; screenshot review required'
}
try {
    Invoke-Adb -CommandArgs @('install', '-r', $Apk) | Out-Null
    Start-Sleep -Seconds 1
    Invoke-Adb -CommandArgs @('shell', 'am', 'force-stop', '--user', '0', $package) | Out-Null
    Invoke-Adb -CommandArgs @('shell', 'am', 'start', '-W', '-f', '0x10008000', '-n', $component, '--ez', 'ownframe.validation', 'true') | Out-Null
    Start-Sleep -Seconds 2
    # Android can restore the pre-update task with its old intent first.
    Invoke-Adb -CommandArgs @('shell', 'am', 'start', '-W', '-f', '0x20000000', '-n', $component, '--ez', 'ownframe.validation', 'true') | Out-Null
    Save-Screen '01-collapsed'
    Tap-Demo 200 2255
    Save-Screen '02-expanded'
    Tap-Demo 1018 1420
    Save-Screen '03-closed'
    Tap-Demo 300 480
    Tap-Demo 400 2250
    foreach ($letter in @('q','a')) {
        Assert-Demo
        Invoke-Adb -CommandArgs @('shell', 'input', 'text', $letter) | Out-Null
        Start-Sleep -Milliseconds 500
    }
    $ime = Invoke-Adb -CommandArgs @('shell', 'dumpsys', 'input_method')
    if ($ime -notmatch 'mInputShown=true') { throw 'Composer tap did not open the soft keyboard.' }
    Save-Screen '04-keyboard-qa'
    Invoke-Adb -CommandArgs @('shell', 'settings', 'put', 'system', 'accelerometer_rotation', '0') | Out-Null
    foreach ($rotation in @('1','3','0')) {
        Assert-Demo
        Invoke-Adb -CommandArgs @('shell', 'settings', 'put', 'system', 'user_rotation', $rotation) | Out-Null
        Start-Sleep -Seconds 1
        Save-Screen "05-rotation-$rotation"
    }
    Assert-Demo
    $before = Invoke-Adb -CommandArgs @('shell', 'pidof', $package)
    Invoke-Adb -CommandArgs @('shell', 'input', 'keyevent', '3') | Out-Null
    Start-Sleep -Seconds 1
    $paused = Invoke-Adb -CommandArgs @('shell', 'pidof', $package)
    Invoke-Adb -CommandArgs @('shell', 'am', 'start', '-W', '-f', '0x20000000', '-n', $component, '--ez', 'ownframe.validation', 'true') | Out-Null
    Start-Sleep -Seconds 1
    Assert-Demo
    $resumed = Invoke-Adb -CommandArgs @('shell', 'pidof', $package)
    if ($before -ne $paused -or $before -ne $resumed) { throw 'Process restarted during background/resume.' }
    $record.same_process_resume = $true
    $ime = Invoke-Adb -CommandArgs @('shell', 'dumpsys', 'input_method')
    if ($ime -match 'mInputShown=true') {
        Invoke-Adb -CommandArgs @('shell', 'input', 'keyevent', '4') | Out-Null
        Start-Sleep -Seconds 1
    }
    $ime = Invoke-Adb -CommandArgs @('shell', 'dumpsys', 'input_method')
    if ($ime -match 'mInputShown=true') { throw 'Keyboard remained visible after dismissal.' }
    Save-Screen '06-resumed'
    $bitmap = [System.Drawing.Bitmap]::FromFile((Join-Path $OutDir '06-resumed.png'))
    try {
        foreach ($fraction in @(0.75, 0.84, 0.88)) {
            $pixel = $bitmap.GetPixel(10, [int]($bitmap.Height * $fraction))
            if ($pixel.R -gt 248 -and $pixel.G -gt 248 -and $pixel.B -gt 248) {
                throw 'Blank keyboard gap remains below the thread after resume.'
            }
        }
    } finally { $bitmap.Dispose() }
    $record.keyboard_gap_check = 'passed: lower thread edge is not blank white'
    $record.state = 'captured-awaiting-visual-review'
} catch {
    $record.state = 'failed'
    $record.error = $_.Exception.Message
    throw
} finally {
    try {
        Invoke-Adb -CommandArgs @('shell', 'settings', 'put', 'system', 'user_rotation', $oldRotation) | Out-Null
        Invoke-Adb -CommandArgs @('shell', 'settings', 'put', 'system', 'accelerometer_rotation', $oldAuto) | Out-Null
        $record.rotation_restored = $true
    } finally {
        $record | ConvertTo-Json -Depth 4 | Set-Content -Encoding UTF8 (Join-Path $OutDir 'session.json')
    }
}
Write-Output "Session captured in $OutDir. Review every screenshot before declaring success."
