#!/usr/bin/env python3
"""Drive the live JMAP window using its X11 client coordinates on this dual-monitor fixture."""
import pathlib
import re
import subprocess
import time

ROOT = pathlib.Path(__file__).resolve().parent
PS = '/mnt/c/Windows/System32/WindowsPowerShell/v1.0/powershell.exe'
SCRIPT = subprocess.check_output(['wslpath', '-w', str(ROOT / 'windows-pointer-gesture.ps1')], text=True).strip()

def rect():
    tree = subprocess.check_output(['xwininfo', '-root', '-tree'], text=True)
    line = next(line for line in tree.splitlines() if 'JMAP Mail Preview' in line)
    wid = line.split()[0]
    data = subprocess.check_output(['xwininfo', '-id', wid], text=True)
    values = [int(re.search(rf'{key}:\s*(-?\d+)', data).group(1)) for key in ['Absolute upper-left X', 'Absolute upper-left Y', 'Width', 'Height']]
    values[0] -= 1920  # XRandR primary origin for the captured 2560/1920 fixture.
    return values

def gesture(start, end=None):
    args = [PS, '-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', SCRIPT, '-StartX', str(start[0]), '-StartY', str(start[1])]
    args += ['-DoubleClick'] if end is None else ['-EndX', str(end[0]), '-EndY', str(end[1])]
    subprocess.run(args, check=True, capture_output=True, text=True)

for right_percent in [10, 90, 50, 60, 80, 20]:
    x, y, w, h = rect()
    # The saved normal size in this fixture is the left panel's 1920x1032 workarea.
    target_x = round(-1920 * (100 - right_percent) / 100)
    gesture((x + 400, y + 20), (target_x + 400, 380))
    before = rect()
    print(f'{right_percent}% right: dragged={before}', flush=True)
    x, y, w, h = before
    gesture((x + 400, y + 20))
    expected = [0, 0, 2560, 1528] if right_percent >= 50 else [-1920, 320, 1920, 1032]
    deadline = time.monotonic() + 3
    while time.monotonic() < deadline:
        actual = rect()
        if actual == expected:
            break
        time.sleep(.1)
    print(f'maximized={actual} expected={expected}', flush=True)
    if actual != expected:
        raise SystemExit('Maximize missed the selected monitor workarea')
