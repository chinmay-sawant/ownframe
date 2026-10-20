#!/usr/bin/env python3
"""Click every Start button in the CSS preview through the real X11 pointer.

The app must run with perf sampling and the walk hook, from the repo root:

    CSS_PREVIEW_WALK=1 go run ./examples/css-preview -perf

The walk hook closes the devtools dock so both catalog columns stay
clickable; perf sampling stays on. It serves these local endpoints:

    /goto?i=N   scroll Start button N under a fixed row
    /aim        index, client x, client y, property name
    /cursor     client cursor position, used to calibrate the window origin
    /state      next index, started count, selected name, last redraw (us)

This script moves the X pointer and sends real button press/release events
with XTest, so every click goes through the window pointer path. It measures
the time from the release to the page reporting the item as started, and
retries a missed click.

Run:  python3 scripts/css-preview-click-test.py
"""

import ctypes
import ctypes.util
import sys
import time
import urllib.request

BASE = "http://127.0.0.1:6971/"
CALIBRATE_AT = (700, 700)


def open_display():
    x11 = ctypes.cdll.LoadLibrary(ctypes.util.find_library("X11"))
    xtst = ctypes.cdll.LoadLibrary(ctypes.util.find_library("Xtst"))
    x11.XOpenDisplay.restype = ctypes.c_void_p
    display = x11.XOpenDisplay(None)
    if not display:
        sys.exit("cannot open the X display")
    return x11, xtst, display


def get(path):
    with urllib.request.urlopen(BASE + path, timeout=3) as response:
        return response.read().decode().strip()


def move(xtst, x11, display, x, y):
    xtst.XTestFakeMotionEvent(display, -1, x, y, 0)
    x11.XFlush(display)


def click(x11, xtst, display, x, y):
    move(xtst, x11, display, x, y)
    time.sleep(0.08)
    xtst.XTestFakeButtonEvent(display, 1, 1, 0)
    x11.XFlush(display)
    time.sleep(0.05)
    xtst.XTestFakeButtonEvent(display, 1, 0, 0)
    x11.XFlush(display)


def aim_for(index):
    for _ in range(400):
        fields = get("aim").split()
        if len(fields) >= 4 and int(fields[0]) == index:
            return int(fields[1]), int(fields[2]), fields[3]
        time.sleep(0.01)
    return None


def main():
    x11, xtst, display = open_display()

    # Calibrate the window origin: the app reports the client position of
    # the real pointer, so subtract it from a known screen point.
    move(xtst, x11, display, *CALIBRATE_AT)
    time.sleep(0.4)
    cx, cy = map(int, get("cursor").split())
    ox, oy = CALIBRATE_AT[0] - cx, CALIBRATE_AT[1] - cy
    print("window origin", ox, oy)

    latencies = []
    failures = []
    total = 43

    for i in range(total):
        get("goto?i=%d" % i)
        aim = aim_for(i)
        if aim is None:
            failures.append((i, "no aim"))
            print("%02d NO AIM" % i)
            continue

        before = int(get("state").split()[1])
        ok = False
        for _ in range(3):
            started = time.time()
            click(x11, xtst, display, ox + aim[0], oy + aim[1])
            while time.time() - started < 1.2:
                state = get("state").split()
                if len(state) >= 4 and int(state[1]) > before and state[2] == aim[2]:
                    latencies.append(time.time() - started)
                    ok = True
                    break
                time.sleep(0.01)
            if ok:
                break
            aim = aim_for(i) or aim
            time.sleep(0.2)

        if not ok:
            failures.append((i, aim[2]))
        print("%02d %-40s %s %4.0fms" % (
            i, aim[2], "ok" if ok else "FAIL",
            (latencies[-1] * 1000) if ok and latencies else 0.0))

    latencies.sort()
    print("clicked %d/%d failures %s" % (len(latencies), total, failures))
    if latencies:
        print("p50 %.0fms p95 %.0fms max %.0fms avg %.0fms" % (
            latencies[len(latencies) // 2] * 1000,
            latencies[len(latencies) * 95 // 100] * 1000,
            latencies[-1] * 1000,
            sum(latencies) / len(latencies) * 1000))
    print("final state", get("state"))


if __name__ == "__main__":
    main()
