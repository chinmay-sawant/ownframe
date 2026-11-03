#!/usr/bin/env python3
"""Measure an already-running native OwnFrame process without waking its window. Presents count OwnFrame draws, not GPU swaps."""
import argparse
import json
import os
from pathlib import Path
import time
import urllib.request


def snapshot(url):
    with urllib.request.urlopen(url, timeout=5) as response:
        state = json.load(response)
    return {key: state[key] for key in (
        "generation", "stats", "frame", "runtime", "sampled_at", "sample_age_ms"
    ) if key in state}


def process(pid):
    root = Path(f"/proc/{pid}")
    # The command name may contain spaces and parentheses.
    fields = (root / "stat").read_text().rsplit(")", 1)[1].split()
    ticks = int(fields[11]) + int(fields[12])
    rss = int((root / "statm").read_text().split()[1]) * os.sysconf("SC_PAGE_SIZE")
    return ticks, rss


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--pid", type=int, required=True, help="compiled application PID, not go run")
    parser.add_argument("--url", default="http://127.0.0.1:8129/debug/state")
    parser.add_argument("--seconds", type=float, default=30)
    parser.add_argument("--settle", type=float, default=10)
    parser.add_argument("--assert-static", action="store_true")
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    if args.seconds <= 0 or args.settle < 0:
        parser.error("seconds must be positive and settle must be nonnegative")
    time.sleep(args.settle)
    before = snapshot(args.url)
    ticks0, rss0 = process(args.pid)
    started = time.monotonic()
    time.sleep(args.seconds)
    ticks1, rss1 = process(args.pid)
    elapsed = time.monotonic() - started
    after = snapshot(args.url)
    presents0 = before.get("frame", {}).get("presents")
    presents1 = after.get("frame", {}).get("presents")
    presents = None if presents0 is None or presents1 is None else presents1 - presents0
    report = {
        "pid": args.pid, "seconds": elapsed,
        "cpu_percent_one_core": 100 * (ticks1 - ticks0) / os.sysconf("SC_CLK_TCK") / elapsed,
        "rss_before_bytes": rss0, "rss_after_bytes": rss1,
        "presents": presents,
        "presents_per_second": None if presents is None else presents / elapsed,
        "generation_changed": before["generation"] != after["generation"],
        "before": before, "after": after,
    }
    result = json.dumps(report, indent=2) + "\n"
    if args.output:
        args.output.write_text(result)
    print(result, end="")
    if args.assert_static and (presents != 0 or report["generation_changed"]):
        raise SystemExit("FAIL: static window drew OwnFrame content or changed, or present counter is unavailable")


if __name__ == "__main__":
    main()
