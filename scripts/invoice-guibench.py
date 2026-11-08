#!/usr/bin/env python3
"""Click the invoice table in a real window and time it, press to painted.

Runs the app under Xvfb with the debug build, exactly like this:

    go run -tags ownframe_debug ./examples/testing/ownframe -perf \\
        -debug-addr 127.0.0.1:8129

Then drives the OS pointer with XTest: the click goes through the real Ebiten
window path, not pagetest. Coordinates come from GET /debug/state boxes, and
the window origin is calibrated from the state's cursor readout plus
XQueryPointer, so no app endpoint is needed for aiming.

Per click the script records press-to-visible latency from the state's
generation counter, plus the page counters (parses, cascades, layouts,
repaints, AllocFrame, LastRedraw) to prove what each click cost. At the end
it prints the perf frame ring (avg/p95/p99, presents, present fps) and saves
a heap profile of the live GUI process.

Targets are rows 0 and 1, the top two data rows, which are inside the default
1000x650 viewport with no scrolling. The cost matches any other row: the same
two class edits plus a text edit through the production patch path.

Run:  python3 scripts/invoice-guibench.py [--iters 30] [--display :99]
"""

import argparse
import ctypes
import ctypes.util
import http.client
import json
import socket
import statistics
import subprocess
import sys
import time
from pathlib import Path

SETTLE = 0.08
HOLD_NS = 50_000_000
STATE_TIMEOUT = 5.0


def summary(label, values):
    if not values:
        print("%s: no samples" % label)
        return
    ordered = sorted(values)
    print("%s p50 %.1fms p95 %.1fms max %.1fms avg %.1fms" % (
        label,
        ordered[len(ordered) // 2],
        ordered[len(ordered) * 95 // 100],
        ordered[-1],
        sum(values) / len(values)))


class X11:
    def __init__(self, display):
        self.x11 = ctypes.cdll.LoadLibrary(ctypes.util.find_library("X11"))
        self.xtst = ctypes.cdll.LoadLibrary(ctypes.util.find_library("Xtst"))
        self.x11.XOpenDisplay.restype = ctypes.c_void_p
        self.x11.XOpenDisplay.argtypes = [ctypes.c_char_p]
        self.x11.XDefaultRootWindow.restype = ctypes.c_ulong
        self.x11.XDisplayWidth.restype = ctypes.c_int
        self.x11.XDisplayHeight.restype = ctypes.c_int
        self.x11.XQueryPointer.restype = ctypes.c_int
        self.display = self.x11.XOpenDisplay(display.encode())
        if not self.display:
            sys.exit("cannot open the X display %s" % display)
        self.screen_w = self.x11.XDisplayWidth(self.display, 0)
        self.screen_h = self.x11.XDisplayHeight(self.display, 0)
        self.root = self.x11.XDefaultRootWindow(self.display)

    def pointer(self):
        root_r = ctypes.c_ulong()
        child_r = ctypes.c_ulong()
        rx, ry = ctypes.c_int(), ctypes.c_int()
        wx, wy = ctypes.c_int(), ctypes.c_int()
        mask = ctypes.c_uint()
        if not self.x11.XQueryPointer(self.display, self.root,
                                      ctypes.byref(root_r), ctypes.byref(child_r),
                                      ctypes.byref(rx), ctypes.byref(ry),
                                      ctypes.byref(wx), ctypes.byref(wy),
                                      ctypes.byref(mask)):
            sys.exit("XQueryPointer failed")
        return rx.value, ry.value

    def move(self, x, y):
        self.xtst.XTestFakeMotionEvent(self.display, -1, int(x), int(y), 0)
        self.x11.XFlush(self.display)

    def button(self, down):
        self.xtst.XTestFakeButtonEvent(self.display, 1, 1 if down else 0, 0)
        self.x11.XFlush(self.display)

    def close(self):
        self.x11.XCloseDisplay(self.display)


class Debug:
    def __init__(self, host, port):
        self.host = host
        self.port = port
        self.conn = None

    def get(self, path, timeout=10):
        if self.conn is None:
            self.conn = http.client.HTTPConnection(self.host, self.port, timeout=timeout)
        try:
            self.conn.request("GET", path)
            resp = self.conn.getresponse()
            return resp.status, resp.read()
        except (ConnectionError, OSError):
            try:
                self.conn.close()
            finally:
                self.conn = None
            raise

    def state(self):
        for _ in range(120):
            try:
                status, data = self.get("/debug/state")
            except (ConnectionError, OSError):
                time.sleep(0.5)
                continue
            if status == 503:
                time.sleep(0.5)
                continue
            if status != 200:
                sys.exit("GET /debug/state returned %d" % status)
            return json.loads(data)
        sys.exit("no debug snapshot published")


def free_display():
    for n in range(99, 120):
        if not Path("/tmp/.X11-unix/X%d" % n).exists():
            return ":%d" % n
    sys.exit("no free X display under /tmp/.X11-unix")


def wait_port(host, port, timeout=120.0):
    deadline = time.time() + timeout
    while time.time() < deadline:
        try:
            with socket.create_connection((host, port), timeout=1):
                return
        except OSError:
            time.sleep(0.5)
    sys.exit("debug server did not listen on %s:%d" % (host, port))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--iters", type=int, default=30)
    parser.add_argument("--warmup", type=int, default=5)
    parser.add_argument("--display", default=":99")
    parser.add_argument("--reuse-display", action="store_true",
                        help="use --display as-is (your real screen, so you can watch) "
                             "instead of starting Xvfb")
    parser.add_argument("--attach", action="store_true",
                        help="drive an already-running app on --display and --debug-addr; "
                             "do not start or stop anything")
    parser.add_argument("--debug-addr", default="127.0.0.1:8129")
    parser.add_argument("--heap-out", default="/tmp/opencode/invoice-gui-heap.pb.gz")
    args = parser.parse_args()
    dbg_host, dbg_port = args.debug_addr.rsplit(":", 1)
    dbg_port = int(dbg_port)

    root = Path(__file__).resolve().parent.parent
    log_path = Path("/tmp/opencode/invoice-gui.log")
    log_path.parent.mkdir(parents=True, exist_ok=True)

    if args.attach:
        display = args.display
        xvfb, app, log_f = None, None, None
        print("attached, keeping the running window")
    elif args.reuse_display:
        display = args.display
        print("reusing display %s, the window opens on your screen" % display)
        xvfb = None
    else:
        display = args.display
        if Path("/tmp/.X11-unix/X%s" % display.lstrip(":")).exists():
            display = free_display()
            print("display busy, using %s" % display)
        xvfb = subprocess.Popen(["Xvfb", display, "-screen", "0", "1280x1024x24"])
        time.sleep(1.0)
        if xvfb.poll() is not None:
            sys.exit("Xvfb %s did not start (busy display?)" % display)
    env = dict(__import__("os").environ)
    env["DISPLAY"] = display
    env["LIBGL_ALWAYS_SOFTWARE"] = "1"
    log_f = open(log_path, "wb")
    app = subprocess.Popen(
        ["go", "run", "-tags", "ownframe_debug",
         "./examples/testing/ownframe", "-perf",
         "-debug-addr", args.debug_addr],
        cwd=root, env=env, stdout=log_f, stderr=subprocess.STDOUT)
    x11 = None
    try:
        wait_port(dbg_host, dbg_port)
        dbg = Debug(dbg_host, dbg_port)
        state = dbg.state()
        x11 = X11(display)

        # Calibrate: the cursor readout publishes at most twice a second and
        # goes stale while the loop is parked, so after each move wait for
        # the readout to settle (two identical reads) before trusting it.
        # Three settled points must agree on the origin.
        def settled_cursor():
            deadline = time.time() + 8.0
            last = (None, time.time())
            while time.time() < deadline:
                cur = dbg.state()["cursor"]
                now = (cur["X"], cur["Y"])
                if now == last[0] and time.time() - last[1] >= 0.6:
                    return now
                if now != last[0]:
                    last = (now, time.time())
                time.sleep(0.15)
            sys.exit("cursor readout never settled")
        origins = []
        sw, sh = x11.screen_w, x11.screen_h
        for fx, fy in ((0.16, 0.20), (0.47, 0.49), (0.70, 0.29), (0.30, 0.70), (0.55, 0.15)):
            px, py = int(sw * fx), int(sh * fy)
            x11.move(px, py)
            time.sleep(0.3)
            cx, cy = settled_cursor()
            origins.append((px - cx, py - cy))
        # Majority vote: the window may not cover every probe point, so a
        # probe outside the window can report a stale cursor. At least three
        # of five must agree; the median of the agreeing cluster wins.
        origins.sort()
        best, cluster = origins[0], [origins[0]]
        for o in origins[1:]:
            if abs(o[0] - best[0]) <= 2 and abs(o[1] - best[1]) <= 2:
                cluster.append(o)
            elif len(cluster) == 1:
                best, cluster = o, [o]
        if len(cluster) < 3:
            sys.exit("calibration disagrees: %s" % (origins,))
        ox = sorted(o[0] for o in cluster)[len(cluster) // 2]
        oy = sorted(o[1] for o in cluster)[len(cluster) // 2]
        print("window origin %d,%d" % (ox, oy))

        boxes = dbg.state().get("boxes") or []
        want = {"row:0": None, "row:1": None}
        for b in boxes:
            if b.get("Action") in want and want[b["Action"]] is None:
                want[b["Action"]] = (b["X"] + b["W"] / 2, b["Y"] + b["H"] / 2)
        if any(v is None for v in want.values()):
            sys.exit("target rows missing from debug boxes")
        targets = [want["row:0"], want["row:1"]]

        gen = dbg.state()["generation"]
        base_stats = dbg.state().get("stats") or {}

        def one_click(target):
            nonlocal gen
            cx, cy = target
            x11.move(ox + cx, oy + cy)
            time.sleep(SETTLE)
            before = dbg.state().get("stats") or {}
            press_ns = time.time_ns()
            x11.button(True)
            deadline = time.time() + STATE_TIMEOUT
            while time.time() < deadline:
                cur = dbg.state()
                if cur["generation"] != gen:
                    observed = cur
                    break
                time.sleep(0.002)
            else:
                sys.exit("generation did not move after click")
            visible_ms = (time.time_ns() - press_ns) / 1e6
            gen = observed["generation"]
            left = HOLD_NS - (time.time_ns() - press_ns)
            if left > 0:
                time.sleep(left / 1e9)
            x11.button(False)
            after = observed.get("stats") or {}
            return visible_ms, before, after

        for i in range(args.warmup):
            one_click(targets[i % 2])

        vis_ms, deltas = [], []
        for i in range(args.iters):
            vis, before, after = one_click(targets[i % 2])
            vis_ms.append(vis)
            deltas.append({k: after.get(k, 0) - before.get(k, 0)
                           for k in ("Parses", "Cascades", "Layouts", "Repaints")})
            alloc = after.get("AllocFrame", 0)
            last = after.get("LastRedraw", 0)
            if i % 10 == 0:
                print("click %d visible %5.1fms alloc %7d last_redraw %8dns" % (
                    i, vis, alloc, last))

        summary("press->visible  ", vis_ms)
        for k in ("Parses", "Cascades", "Layouts", "Repaints"):
            vals = [d[k] for d in deltas]
            print("per-click %-8s min %d p50 %d max %d" % (
                k, min(vals), sorted(vals)[len(vals) // 2], max(vals)))

        final = dbg.state()
        frame = final.get("frame") or {}
        if frame:
            print("frame avg %.2fms p95 %.2fms p99 %.2fms presents %d fps %s repaints/s %s" % (
                frame.get("avg_ns", 0) / 1e6, frame.get("p95_ns", 0) / 1e6,
                frame.get("p99_ns", 0) / 1e6, frame.get("presents", 0),
                frame.get("present_fps"), frame.get("repaints_per_sec")))
        else:
            print("no frame ring (was -perf passed?)")
        print("idle parked: %s" % frame.get("idle"))

        status, heap = dbg.get("/debug/pprof/heap", timeout=30)
        if status == 200:
            Path(args.heap_out).write_bytes(heap)
            print("heap profile %d bytes -> %s" % (len(heap), args.heap_out))
        else:
            print("heap profile returned %d" % status)
    finally:
        if x11 is not None:
            x11.close()
        if not args.attach:
            app.terminate()
            try:
                app.wait(timeout=15)
            except subprocess.TimeoutExpired:
                app.kill()
            log_f.close()
        if xvfb is not None:
            xvfb.terminate()
            xvfb.wait(timeout=15)


if __name__ == "__main__":
    main()
