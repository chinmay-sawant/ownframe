#!/usr/bin/env python3
"""Time invoice row clicks through the real -web server, click to picture.

Starts the app from the repo root, then drives it over HTTP exactly like the
browser shell does:

    go run ./examples/testing/ownframe -web -addr 127.0.0.1:8149

Per iteration the script reads the row center out of the shell imagemap,
POSTs /click with an Origin header, polls /state until the generation moves,
then fetches /frame.png. It reports click-post latency, click-to-visible
latency, frame-fetch latency, end-to-end latency, and frame bytes.

Targets are the first two data rows, which sit inside the default 1000x650
viewport with no scrolling. The cost matches any other row: the click commits
the same two class edits plus a text edit through the production patch path.

Run:  python3 scripts/invoice-webbench.py [--iters 50] [--addr 127.0.0.1:8149]
"""

import argparse
import http.client
import json
import re
import socket
import statistics
import subprocess
import sys
import time
from pathlib import Path

AREA_RE = re.compile(
    r'<area[^>]*alt="([^"]+)"[^>]*data-x="([^"]+)"[^>]*data-y="([^"]+)"[^>]*>'
    r'|<area[^>]*data-x="([^"]+)"[^>]*data-y="([^"]+)"[^>]*alt="([^"]+)"[^>]*>'
)
TARGETS = ("invoice-row-0", "invoice-row-1")


def wait_ready(host, port, timeout=30.0):
    deadline = time.time() + timeout
    while time.time() < deadline:
        try:
            with socket.create_connection((host, port), timeout=1):
                return
        except OSError:
            time.sleep(0.2)
    sys.exit("server did not listen on %s:%d" % (host, port))


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
        statistics.mean(values)))


class Driver:
    def __init__(self, host, port):
        self.host = host
        self.port = port
        self.origin = "http://%s:%d" % (host, port)
        self.conn = None

    def request(self, method, path, body=None, headers=None):
        if self.conn is None:
            self.conn = http.client.HTTPConnection(self.host, self.port, timeout=10)
        hdrs = dict(headers or {})
        try:
            self.conn.request(method, path, body=body, headers=hdrs)
            resp = self.conn.getresponse()
            data = resp.read()
            return resp.status, data
        except (ConnectionError, OSError):
            try:
                self.conn.close()
            finally:
                self.conn = None
            raise

    def get(self, path):
        return self.request("GET", path)

    def post_click(self, x, y):
        return self.request(
            "POST", "/click?x=%s&y=%s" % (x, y), headers={"Origin": self.origin})

    def generation(self):
        _, data = self.get("/state")
        return json.loads(data)["generation"]

    def areas(self):
        _, data = self.get("/")
        html = data.decode("utf-8", "replace")
        out = {}
        for m in AREA_RE.finditer(html):
            if m.group(1) is not None:
                alt, x, y = m.group(1), m.group(2), m.group(3)
            else:
                x, y, alt = m.group(4), m.group(5), m.group(6)
            out[alt] = (x, y)
        return out, html


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--addr", default="127.0.0.1:8149")
    parser.add_argument("--iters", type=int, default=50)
    parser.add_argument("--warmup", type=int, default=5)
    args = parser.parse_args()
    host, port = args.addr.rsplit(":", 1)
    port = int(port)

    root = Path(__file__).resolve().parent.parent
    server = subprocess.Popen(
        ["go", "run", "./examples/testing/ownframe",
         "-web", "-addr", args.addr],
        cwd=root, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    try:
        wait_ready(host, port)
        driver = Driver(host, port)

        areas, _ = driver.areas()
        missing = [t for t in TARGETS if t not in areas]
        if missing:
            sys.exit("targets missing from shell imagemap: %s" % missing)

        gen = driver.generation()

        def one_click(target):
            nonlocal gen
            x, y = areas[target]
            t0 = time.time()
            status, _ = driver.post_click(x, y)
            t_post = time.time()
            if status not in (303, 302, 200):
                sys.exit("POST /click returned %d" % status)
            deadline = t0 + 5.0
            while time.time() < deadline:
                if driver.generation() != gen:
                    break
                time.sleep(0.002)
            else:
                sys.exit("generation did not move after clicking %s" % target)
            gen = driver.generation()
            t_visible = time.time()
            status, png = driver.get("/frame.png")
            t_frame = time.time()
            if status != 200 or not png:
                sys.exit("GET /frame.png returned %d, %d bytes" % (status, len(png)))
            return ((t_post - t0) * 1e3, (t_visible - t0) * 1e3,
                    (t_frame - t_visible) * 1e3, (t_frame - t0) * 1e3, len(png))

        for i in range(args.warmup):
            one_click(TARGETS[i % 2])

        post_ms, vis_ms, frame_ms, total_ms, sizes = [], [], [], [], []
        for i in range(args.iters):
            target = TARGETS[i % 2]
            post, vis, frame, total, size = one_click(target)
            post_ms.append(post)
            vis_ms.append(vis)
            frame_ms.append(frame)
            total_ms.append(total)
            sizes.append(size)
            if i % 10 == 0:
                # The shell carries no readout text, so selection is proven by
                # the generation moving on every click: a miss hits no action,
                # bumps no version, and the poll above would time out. Here
                # just confirm the layout did not shift under the clicks.
                again, _ = driver.areas()
                for t in TARGETS:
                    if again.get(t) != areas[t]:
                        sys.exit("row coords moved mid-run for %s" % t)

        summary("click-post     ", post_ms)
        summary("click-to-visible", vis_ms)
        summary("frame-fetch    ", frame_ms)
        summary("end-to-end     ", total_ms)
        print("frame bytes avg %d min %d max %d over %d iters" % (
            sum(sizes) // len(sizes), min(sizes), max(sizes), args.iters))
    finally:
        server.terminate()
        server.wait(timeout=15)


if __name__ == "__main__":
    main()
