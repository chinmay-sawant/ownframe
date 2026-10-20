#!/usr/bin/env python3
"""Click every Start button in the CSS preview through the real X11 pointer.

The app must run with perf sampling and the walk hook, from the repo root:

    CSS_PREVIEW_WALK=1 go run ./examples/css-preview -perf

The walk hook closes the devtools dock so both catalog columns stay
clickable; perf sampling stays on. It serves these local endpoints:

    /goto?i=N   scroll Start button N under a fixed row
    /aim        index, client x, client y, property name
    /cursor     client cursor position, used to calibrate the window origin
    /state      next, started, selected, last redraw (us), pipeline counters,
                press timestamp (ns), started sequence

This script moves the X pointer and sends real button press/release events
with XTest, so every click goes through the window pointer path. The timed
interval starts at the press. It reports the app's press-to-started time from
the handler timestamp, the press-to-painted time from the next layout, and the
page's own redraw time. The 80 ms hover settle and 50 ms hold are kept for a
real pointer sequence but are outside the timed interval.

The walk hook stops the previous sample before each /goto, so each measurement
is one click instead of a soak of all 43 animations. Use /reset to clear the
started rows and run the whole catalog again in the same window.

Run:  python3 scripts/css-preview-click-test.py [--runs 3]

The script closes the application with /quit after the last pass. Use
--no-close to keep it open.
"""

import argparse
import ctypes
import ctypes.util
import http.client
import statistics
import sys
import time

HOST, PORT = "127.0.0.1", 6971
CALIBRATE_AT = (700, 700)
TOTAL = 43
SETTLE = 0.08
HOLD_NS = 50_000_000
STATE_TIMEOUT_NS = 1_200_000_000

conn = None


def get(path):
    global conn
    if conn is None:
        conn = http.client.HTTPConnection(HOST, PORT, timeout=3)
    try:
        conn.request("GET", "/" + path.lstrip("/"))
        return conn.getresponse().read().decode().strip()
    except (ConnectionError, OSError):
        conn.close()
        conn = None
        raise


def open_display():
    x11 = ctypes.cdll.LoadLibrary(ctypes.util.find_library("X11"))
    xtst = ctypes.cdll.LoadLibrary(ctypes.util.find_library("Xtst"))
    x11.XOpenDisplay.restype = ctypes.c_void_p
    display = x11.XOpenDisplay(None)
    if not display:
        sys.exit("cannot open the X display")
    return x11, xtst, display


def move(xtst, x11, display, x, y):
    xtst.XTestFakeMotionEvent(display, -1, x, y, 0)
    x11.XFlush(display)


def button(xtst, x11, display, down):
    xtst.XTestFakeButtonEvent(display, 1, 1 if down else 0, 0)
    x11.XFlush(display)


def state():
    return get("state").split()


def started(fields, before_seq, before_started, name):
    return (len(fields) >= 20 and int(fields[19]) > before_seq
            and int(fields[1]) > before_started and fields[2] == name)


def aim_for(index):
    for _ in range(400):
        fields = get("aim").split()
        if len(fields) >= 4 and int(fields[0]) == index:
            return int(fields[1]), int(fields[2]), fields[3]
        time.sleep(0.01)
    return None


def one_click(xtst, x11, display, ox, oy, aim):
    ax, ay, name = aim
    move(xtst, x11, display, ox + ax, oy + ay)
    time.sleep(SETTLE)

    before = state()
    before_started, before_seq = int(before[1]), int(before[19])
    before_layouts = int(before[7])

    press_ns = time.time_ns()
    button(xtst, x11, display, True)

    observed, app_started_ns = None, 0
    while time.time_ns() - press_ns < HOLD_NS + 300_000_000:
        fields = state()
        if started(fields, before_seq, before_started, name):
            observed = fields
            app_started_ns = int(fields[18])
            break
        time.sleep(0.001)

    left = HOLD_NS - (time.time_ns() - press_ns)
    if left > 0:
        time.sleep(left / 1e9)
    button(xtst, x11, display, False)

    if observed is None:
        deadline = time.time_ns() + STATE_TIMEOUT_NS
        while time.time_ns() < deadline:
            fields = state()
            if started(fields, before_seq, before_started, name):
                observed = fields
                app_started_ns = int(fields[18])
                break
            time.sleep(0.001)
    if observed is None:
        return None

    painted_ns = 0
    deadline = time.time_ns() + STATE_TIMEOUT_NS
    while time.time_ns() < deadline:
        fields = state()
        if len(fields) >= 8 and int(fields[7]) > before_layouts:
            painted_ns = time.time_ns()
            observed = fields
            break
        time.sleep(0.001)

    started_ms = (app_started_ns - press_ns) / 1e6 if app_started_ns else None
    painted_ms = (painted_ns - press_ns) / 1e6 if painted_ns else None
    redraw_ms = int(observed[3]) / 1e3
    return name, started_ms, painted_ms, redraw_ms, observed


def summary(label, values):
    if not values:
        print("%s: no samples" % label)
        return
    values = sorted(values)
    print("%s p50 %.1fms p95 %.1fms max %.1fms avg %.1fms total %.2fs" % (
        label,
        values[len(values) // 2],
        values[len(values) * 95 // 100],
        values[-1],
        statistics.mean(values),
        sum(values) / 1000.0))


def run_pass(x11, xtst, display, ox, oy):
    get("reset")
    for _ in range(400):
        fields = state()
        if len(fields) >= 2 and int(fields[1]) == 0:
            break
        time.sleep(0.01)

    started_ms, painted_ms, redraw_ms = [], [], []
    failures = []

    for i in range(TOTAL):
        get("goto?i=%d" % i)
        aim = aim_for(i)
        if aim is None:
            failures.append((i, "no aim"))
            print("%02d NO AIM" % i)
            continue

        result = None
        for _ in range(3):
            result = one_click(xtst, x11, display, ox, oy, aim)
            if result is not None:
                break
            aim = aim_for(i) or aim
            time.sleep(0.2)

        if result is None:
            failures.append((i, aim[2]))
            print("%02d %-40s FAIL" % (i, aim[2]))
            continue

        name, press_to_started, press_to_painted, redraw, _ = result
        if press_to_started is not None:
            started_ms.append(press_to_started)
        if press_to_painted is not None:
            painted_ms.append(press_to_painted)
        redraw_ms.append(redraw)
        print("%02d %-40s ok started %6.1fms painted %6.1fms redraw %6.1fms" % (
            i, name,
            press_to_started if press_to_started is not None else -1,
            press_to_painted if press_to_painted is not None else -1,
            redraw))

    print("clicked %d/%d failures %s" % (len(redraw_ms), TOTAL, failures))
    return started_ms, painted_ms, redraw_ms, failures


def close_app():
    try:
        get("quit")
    except (ConnectionError, OSError):
        pass


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--runs", type=int, default=1, help="full catalog passes")
    parser.add_argument("--no-close", action="store_true", help="leave the app running")
    args = parser.parse_args()

    x11, xtst, display = open_display()

    move(xtst, x11, display, *CALIBRATE_AT)
    time.sleep(0.4)
    cx, cy = map(int, get("cursor").split())
    ox, oy = CALIBRATE_AT[0] - cx, CALIBRATE_AT[1] - cy
    print("window origin", ox, oy)

    started_all, painted_all, redraw_all, failures_all = [], [], [], []
    try:
        for run in range(args.runs):
            print("pass %d/%d" % (run + 1, args.runs))
            started, painted, redraw, failures = run_pass(x11, xtst, display, ox, oy)
            started_all += started
            painted_all += painted
            redraw_all += redraw
            failures_all += failures
    finally:
        if not args.no_close:
            close_app()

    summary("press->started", started_all)
    summary("press->painted", painted_all)
    summary("redraw", redraw_all)
    print("passes %d, failures %s" % (args.runs, failures_all))


if __name__ == "__main__":
    main()
