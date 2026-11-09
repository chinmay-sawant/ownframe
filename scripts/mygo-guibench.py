#!/usr/bin/env python3
"""Click the MyGo invoice table in a real window and time it, press to painted.

Runs the app under Xvfb (no window manager, so no decorations to offset the
math), finds the "Invoice Table" window by title, and drives the OS pointer
with XTest through the real window path. Row centers come from the headless
layout each run, so clicks track the code, not stale constants:

    GOWORK=off go test -run TestPrintRowCenters -v .

MyGo exposes no debug endpoint, so visibility is read from pixels: a small
strip over the clicked row is screenshotted before the press and polled until
it changes, then checked for the selection tint #dbeafe. That is true painted
output, stricter than a generation counter.

Targets are rows 0 and 1 (INV-1001, INV-1002), the top two data rows, inside
the default 1000x650 viewport with no scrolling.

Run:  python3 scripts/mygo-guibench.py [--iters 30]
"""

import argparse
import ctypes
import ctypes.util
import re
import socket
import statistics
import subprocess
import sys
import time
from pathlib import Path

SETTLE = 0.08
HOLD_NS = 50_000_000
STATE_TIMEOUT = 5.0
SELECT_RGB = b"\xdb\xea\xfa"


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


class XImage(ctypes.Structure):
    _fields_ = [
        ("width", ctypes.c_int),
        ("height", ctypes.c_int),
        ("xoffset", ctypes.c_int),
        ("format", ctypes.c_int),
        ("data", ctypes.c_char_p),
        ("byte_order", ctypes.c_int),
        ("bitmap_unit", ctypes.c_int),
        ("bitmap_bit_order", ctypes.c_int),
        ("bitmap_pad", ctypes.c_int),
        ("depth", ctypes.c_int),
        ("bytes_per_line", ctypes.c_int),
        ("bits_per_pixel", ctypes.c_int),
    ]


class X11:
    def __init__(self, display):
        self.x11 = ctypes.cdll.LoadLibrary(ctypes.util.find_library("X11"))
        self.xtst = ctypes.cdll.LoadLibrary(ctypes.util.find_library("Xtst"))
        x = self.x11
        x.XOpenDisplay.restype = ctypes.c_void_p
        x.XOpenDisplay.argtypes = [ctypes.c_char_p]
        x.XDefaultRootWindow.restype = ctypes.c_ulong
        x.XQueryTree.restype = ctypes.c_int
        x.XFetchName.restype = ctypes.c_int
        x.XGetWindowAttributes.restype = ctypes.c_int
        x.XGetImage.restype = ctypes.POINTER(XImage)
        x.XInternAtom.restype = ctypes.c_ulong
        x.XInternAtom.argtypes = [ctypes.c_void_p, ctypes.c_char_p, ctypes.c_int]
        self.display = x.XOpenDisplay(display.encode())
        if not self.display:
            sys.exit("cannot open the X display %s" % display)
        self.root = x.XDefaultRootWindow(self.display)
        self.display_name = display

    def find_window(self, title, timeout=60.0):
        deadline = time.time() + timeout
        while time.time() < deadline:
            root_r = ctypes.c_ulong()
            parent_r = ctypes.c_ulong()
            children = ctypes.c_void_p()
            n = ctypes.c_uint()
            if self.x11.XQueryTree(self.display, self.root,
                                   ctypes.byref(root_r), ctypes.byref(parent_r),
                                   ctypes.byref(children), ctypes.byref(n)):
                count = n.value
                if count:
                    arr = (ctypes.c_ulong * count).from_address(children.value)
                    ids = list(arr)
                    for wid in ids:
                        name = ctypes.c_char_p()
                        if self.x11.XFetchName(self.display, wid, ctypes.byref(name)) and name.value:
                            try:
                                text = name.value.decode("utf-8", "replace")
                            finally:
                                self.x11.XFree(name)
                            if title in text:
                                self.x11.XFree(children)
                                return wid
                    self.x11.XFree(children)
            time.sleep(0.5)
        sys.exit("window %r never appeared" % title)

    def window_pos(self, wid):
        ext = self.frame_extents(wid)
        x, y, w, h = self.client_pos(wid)
        return x + ext[0], y + ext[2], w, h

    def client_pos(self, wid):
        class Attr(ctypes.Structure):
            _fields_ = [("x", ctypes.c_int), ("y", ctypes.c_int),
                        ("width", ctypes.c_int), ("height", ctypes.c_int),
                        ("border_width", ctypes.c_int), ("depth", ctypes.c_int),
                        ("visual", ctypes.c_void_p), ("root", ctypes.c_ulong),
                        ("class", ctypes.c_int), ("bit_gravity", ctypes.c_int),
                        ("win_gravity", ctypes.c_int), ("backing_store", ctypes.c_int),
                        ("backing_planes", ctypes.c_ulong), ("backing_pixel", ctypes.c_ulong),
                        ("save_under", ctypes.c_int), ("colormap", ctypes.c_ulong),
                        ("map_installed", ctypes.c_int), ("map_state", ctypes.c_int),
                        ("all_event_masks", ctypes.c_long), ("your_event_mask", ctypes.c_long),
                        ("do_not_propagate_mask", ctypes.c_long),
                        ("override_redirect", ctypes.c_int),
                        ("screen", ctypes.c_void_p)]
        attr = Attr()
        if not self.x11.XGetWindowAttributes(self.display, wid, ctypes.byref(attr)):
            sys.exit("cannot read window geometry")
        return attr.x, attr.y, attr.width, attr.height

    def frame_extents(self, wid):
        atom = self.x11.XInternAtom(self.display, b"_NET_FRAME_EXTENTS", 1)
        if not atom:
            return (0, 0, 0, 0)
        self.x11.XGetWindowProperty.restype = ctypes.c_int
        atype = ctypes.c_ulong()
        aformat = ctypes.c_int()
        nitems = ctypes.c_ulong()
        after = ctypes.c_ulong()
        prop = ctypes.c_void_p()
        if self.x11.XGetWindowProperty(self.display, wid, atom, 0, 4, 0, 0,
                                       ctypes.byref(atype), ctypes.byref(aformat),
                                       ctypes.byref(nitems), ctypes.byref(after),
                                       ctypes.byref(prop)) != 0 or not prop.value:
            return (0, 0, 0, 0)
        try:
            if nitems.value < 4:
                return (0, 0, 0, 0)
            vals = (ctypes.c_ulong * 4).from_address(prop.value)
            return (vals[0], vals[1], vals[2], vals[3])
        finally:
            self.x11.XFree(prop)

    def shot(self, x, y, w, h):
        img = self.x11.XGetImage(self.display, self.root, int(x), int(y),
                                 int(w), int(h), 0xFFFFFFFF, 2)
        if not img:
            sys.exit("XGetImage failed")
        try:
            n = img.contents.bytes_per_line * img.contents.height
            return ctypes.string_at(img.contents.data, n)
        finally:
            self.x11.XDestroyImage(img)

    def move(self, x, y):
        self.xtst.XTestFakeMotionEvent(self.display, -1, int(x), int(y), 0)
        self.x11.XFlush(self.display)

    def button(self, down):
        self.xtst.XTestFakeButtonEvent(self.display, 1, 1 if down else 0, 0)
        self.x11.XFlush(self.display)

    def close(self):
        self.x11.XCloseDisplay(self.display)


def free_display():
    for n in range(99, 120):
        if not Path("/tmp/.X11-unix/X%d" % n).exists():
            return ":%d" % n
    sys.exit("no free X display under /tmp/.X11-unix")


def row_centers(mygo_dir):
    import os as _os
    env = dict(_os.environ, GOWORK="off")
    out = subprocess.run(["go", "test", "-run", "TestPrintRowCenters", "-v", "."],
                         cwd=mygo_dir, env=env,
                         capture_output=True, text=True, timeout=180)
    centers = {}
    for m in re.finditer(r"CENTER (\S+) ([\d.]+) ([\d.]+)", out.stdout):
        centers[m.group(1)] = (float(m.group(2)), float(m.group(3)))
    if "INV-1001" not in centers or "INV-1002" not in centers:
        sys.exit("probe failed:\n%s\n%s" % (out.stdout, out.stderr))
    return centers


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--iters", type=int, default=30)
    parser.add_argument("--warmup", type=int, default=5)
    parser.add_argument("--display", default=None,
                        help="X display; default starts Xvfb on a free one")
    parser.add_argument("--reuse-display", action="store_true",
                        help="use --display as-is instead of starting Xvfb")
    args = parser.parse_args()

    root = Path(__file__).resolve().parent.parent
    mygo_dir = root / "examples" / "testing" / "mygo"
    Path("/tmp/opencode").mkdir(parents=True, exist_ok=True)

    centers = row_centers(str(mygo_dir))
    print("targets %s" % (centers,))

    if args.reuse_display:
        if not args.display:
            sys.exit("--reuse-display needs --display")
        display = args.display
        print("reusing display %s, the window opens on your screen" % display)
        xvfb = None
    else:
        display = free_display()
        xvfb = subprocess.Popen(["Xvfb", display, "-screen", "0", "1280x1024x24"])
        time.sleep(1.0)
        if xvfb.poll() is not None:
            sys.exit("Xvfb %s did not start" % display)
    env = dict(__import__("os").environ)
    env["DISPLAY"] = display
    env["GOWORK"] = "off"
    env["LIBGL_ALWAYS_SOFTWARE"] = "1"
    # This box is Wayland-first, and GTK takes the Wayland backend by
    # default, which X11 cannot see or click. Force X11 so the window lands
    # in the X tree our pointer drives.
    env["GDK_BACKEND"] = "x11"
    log_f = open("/tmp/opencode/mygo-gui.log", "wb")
    app = subprocess.Popen(["go", "run", "."],
                           cwd=mygo_dir, env=env,
                           stdout=log_f, stderr=subprocess.STDOUT)
    x11 = None
    try:
        x11 = X11(display)
        wid = x11.find_window("Invoice Table")
        ox, oy, ww, wh = x11.window_pos(wid)
        print("window at %d,%d size %dx%d" % (ox, oy, ww, wh))
        if (ww, wh) != (1000, 650):
            print("warning: window is %dx%d, want 1000x650" % (ww, wh))
        time.sleep(1.0)

        targets = [centers["INV-1001"], centers["INV-1002"]]

        def strip(row_cx, row_cy):
            return (ox, oy + row_cy - 20, 1000, 40)

        def one_click(row):
            cx, cy = targets[row]
            before = x11.shot(*strip(cx, cy))
            t0 = time.time_ns()
            x11.move(ox + cx, oy + cy)
            time.sleep(SETTLE)
            x11.button(True)
            deadline = time.time() + STATE_TIMEOUT
            while time.time() < deadline:
                after = x11.shot(*strip(cx, cy))
                if after != before:
                    break
                time.sleep(0.005)
            else:
                sys.exit("pixels never changed for row %d" % row)
            painted_ms = (time.time_ns() - t0) / 1e6
            if SELECT_RGB not in after:
                sys.exit("row %d changed but has no selection tint" % row)
            left = HOLD_NS - (time.time_ns() - t0)
            if left > 0:
                time.sleep(left / 1e9)
            x11.button(False)
            return painted_ms

        for i in range(args.warmup):
            one_click(i % 2)

        vis_ms = []
        for i in range(args.iters):
            vis = one_click(i % 2)
            vis_ms.append(vis)
            if i % 10 == 0:
                print("click %d visible %5.1fms" % (i, vis))

        summary("press->painted   ", vis_ms)
    finally:
        if x11 is not None:
            x11.close()
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
