#!/usr/bin/env python3
"""Exercise real window controls on the recorded 1920/2560 WSLg fixture.

Use --executable PATH --launches 3 to cover close/relaunch input state.
The script moves the desktop pointer and leaves a failing process open.
"""
import argparse
import ctypes
import json
import pathlib
import re
import subprocess
import time


TITLE = "Postmark · JMAP Mail Preview (Ubuntu)"
ROOT = pathlib.Path(__file__).resolve().parent


def client_rect():
    tree = subprocess.check_output(["xwininfo", "-root", "-tree"], text=True)
    line = next(line for line in tree.splitlines() if "JMAP Mail Preview" in line)
    data = subprocess.check_output(["xwininfo", "-id", line.split()[0]], text=True)
    keys = ["Absolute upper-left X", "Absolute upper-left Y", "Width", "Height"]
    rect = [int(re.search(rf"{key}:\s*(-?\d+)", data).group(1)) for key in keys]
    rect[0] -= 1920  # Recorded fixture: XWAYLAND0 is at X11 x=1920.
    return rect


def x11_left_down():
    """Read the stale server state that used to survive closing on press."""
    x11 = ctypes.CDLL("libX11.so.6")
    x11.XOpenDisplay.restype = ctypes.c_void_p
    x11.XDefaultRootWindow.argtypes = [ctypes.c_void_p]
    x11.XDefaultRootWindow.restype = ctypes.c_ulong
    x11.XQueryPointer.argtypes = [ctypes.c_void_p, ctypes.c_ulong] + [ctypes.c_void_p] * 7
    x11.XCloseDisplay.argtypes = [ctypes.c_void_p]
    display = x11.XOpenDisplay(None)
    if not display:
        raise RuntimeError("X11 display unavailable")
    try:
        root = x11.XDefaultRootWindow(display)
        root_return, child = ctypes.c_ulong(), ctypes.c_ulong()
        rx, ry, wx, wy = (ctypes.c_int() for _ in range(4))
        mask = ctypes.c_uint()
        ok = x11.XQueryPointer(
            display, root, ctypes.byref(root_return), ctypes.byref(child),
            ctypes.byref(rx), ctypes.byref(ry), ctypes.byref(wx), ctypes.byref(wy),
            ctypes.byref(mask),
        )
        if not ok:
            raise RuntimeError("X11 pointer query failed")
        return bool(mask.value & 256)
    finally:
        x11.XCloseDisplay(display)


class WindowCheck:
    def __init__(self):
        script = subprocess.check_output(
            ["wslpath", "-w", str(ROOT / "windows-window-test-server.ps1")], text=True,
        ).strip()
        self.log = open("/tmp/jmap-window-server.log", "w", encoding="utf-8")
        self.server = subprocess.Popen(
            ["/mnt/c/Windows/System32/WindowsPowerShell/v1.0/powershell.exe",
             "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", script],
            stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=self.log, text=True,
        )
        self.handle = None

    def rpc(self, op, **data):
        self.server.stdin.write(json.dumps(dict(op=op, **data)) + "\n")
        self.server.stdin.flush()
        line = self.server.stdout.readline()
        while line and not line.strip():
            line = self.server.stdout.readline()
        if not line:
            raise RuntimeError("Window test server exited")
        answer = json.loads(line)
        if "error" in answer:
            raise RuntimeError(answer["error"])
        return answer

    def find(self):
        deadline = time.monotonic() + 10
        while time.monotonic() < deadline:
            try:
                self.handle = self.rpc("find", title=TITLE)["handle"]
                client_rect()
                return
            except (RuntimeError, StopIteration):
                time.sleep(.05)
        raise RuntimeError("Mail window did not appear")

    def click(self, offset):
        self.rpc("focus", handle=self.handle)
        previous = client_rect()
        for _ in range(15):
            time.sleep(.1)
            current = client_rect()
            if current == previous:
                break
            previous = current
        else:
            raise RuntimeError("Window geometry did not settle before click")
        x, y, w, _ = current
        routing = self.rpc("click", handle=self.handle, x=x+w-offset, y=y+36)
        print("click", offset, "routing", routing, flush=True)

    def wait_minimized(self, expected):
        deadline = time.monotonic() + 5
        while time.monotonic() < deadline:
            state = self.rpc("status", handle=self.handle)
            if not state["exists"]:
                raise RuntimeError("Mail window exited")
            if state["topmost"]:
                raise RuntimeError("Mail window became always on top")
            if state["minimized"] == expected:
                return
            time.sleep(.1)
        raise RuntimeError(f"Minimized state did not become {expected}")

    def controls(self, cycles):
        for cycle in range(1, cycles+1):
            self.click(130)
            self.wait_minimized(True)
            self.rpc("restore", handle=self.handle)
            self.wait_minimized(False)
            time.sleep(.4)
            print("cycle", cycle, "minimized and restored", flush=True)
        before = client_rect()
        self.click(85)
        time.sleep(1)
        restored = client_rect()
        if restored == before:
            raise RuntimeError("Maximize button did not change window geometry")
        self.click(85)
        time.sleep(1)
        expanded = client_rect()
        if expanded != before:
            raise RuntimeError(f"Wrong workarea after maximize: {expanded}, expected {before}")
        print("restore", restored, "maximize", expanded, flush=True)
        self.click(202)
        time.sleep(.4)
        tree = subprocess.check_output(["xwininfo", "-root", "-tree"], text=True)
        window = next(line.split()[0] for line in tree.splitlines() if "JMAP Mail Preview" in line)
        subprocess.run(["import", "-window", window, "/tmp/jmap-controls-settings.png"], check=True)
        print("Settings clicked after minimize/maximize; screenshot saved", flush=True)

    def close(self):
        try:
            if self.server.poll() is None:
                self.rpc("quit")
        except (RuntimeError, BrokenPipeError):
            pass
        finally:
            self.server.stdin.close()
            self.server.wait(timeout=5)
            self.log.close()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--executable", type=pathlib.Path)
    parser.add_argument("--launches", type=int, default=1)
    parser.add_argument("--cycles", type=int, default=3)
    args = parser.parse_args()
    if args.launches < 1 or args.cycles < 1:
        parser.error("launches and cycles must be positive")
    if args.launches > 1 and not args.executable:
        parser.error("multiple launches require --executable")
    check = WindowCheck()
    try:
        for launch in range(1, args.launches+1):
            process = None
            if args.executable:
                with open(f"/tmp/jmap-launch-{launch}.log", "w", encoding="utf-8") as log:
                    process = subprocess.Popen([str(args.executable.resolve())], stdout=log, stderr=log)
            check.find()
            print("launch", launch, flush=True)
            check.controls(args.cycles)
            if process:
                check.click(40)
                process.wait(timeout=5)
                if process.returncode:
                    raise RuntimeError(f"Mail exited with status {process.returncode}")
                if x11_left_down():
                    raise RuntimeError("Close left the X11 mouse button stuck down")
                print("closed cleanly; X11 left button released", flush=True)
    finally:
        check.close()


if __name__ == "__main__":
    main()
