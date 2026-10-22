#!/usr/bin/env python3
"""Sample an explicitly enabled debug endpoint; missing metrics fail the run."""
import argparse
import csv
from datetime import datetime, timezone
import json
from pathlib import Path
import platform
import subprocess
import sys
import time
from urllib.request import urlopen

ROOT = Path(__file__).resolve().parent.parent
STAT_FIELDS = ('Redraws', 'Parses', 'Cascades', 'Layouts', 'Repaints')


def number(mapping, key):
    value = mapping.get(key)
    if isinstance(value, bool) or not isinstance(value, (int, float)) or value < 0:
        raise ValueError(f'missing or invalid metric: {key}')
    return value


def metrics(body):
    stats, runtime = body.get('stats'), body.get('runtime')
    if not isinstance(stats, dict) or not isinstance(runtime, dict):
        raise ValueError('debug endpoint needs stats and runtime; enable the debug build and perf options')
    return [number(stats, key) for key in STAT_FIELDS] + [number(runtime, 'heap_alloc'), number(runtime, 'goroutines')]


def check_growth(first, current, heap_limit, goroutine_limit):
    if current[-2] - first[-2] > heap_limit:
        raise ValueError('heap growth exceeded the configured limit')
    if current[-1] - first[-1] > goroutine_limit:
        raise ValueError('goroutine growth exceeded the configured limit')
    if any(new < old for old, new in zip(first[:5], current[:5])):
        raise ValueError('page counters reset during the sample')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--minutes', type=float, default=15)
    parser.add_argument('--interval', type=float, default=30)
    parser.add_argument('--url', default='http://127.0.0.1:8128/debug/state')
    parser.add_argument('--out', type=Path, default=ROOT / 'temp/perf-sample.csv')
    parser.add_argument('--max-heap-growth-mib', type=float, default=64)
    parser.add_argument('--max-goroutine-growth', type=int, default=8)
    args = parser.parse_args()
    if not 0 < args.minutes <= 60 or args.interval <= 0 or args.max_heap_growth_mib < 0 or args.max_goroutine_growth < 0:
        parser.error('duration and interval must be positive; limits must be nonnegative; duration is at most 60 minutes')
    args.out.parent.mkdir(parents=True, exist_ok=True)
    commit = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip()
    dirty = bool(subprocess.check_output(['git', 'status', '--porcelain'], cwd=ROOT, text=True))
    start = time.monotonic()
    first = None
    count = 0
    with args.out.open('w', newline='') as stream:
        writer = csv.writer(stream)
        writer.writerow(['utc', 'elapsed_seconds', *STAT_FIELDS, 'heap_alloc', 'goroutines'])
        while True:
            with urlopen(args.url, timeout=5) as response:
                sample = metrics(json.load(response))
            if first is None:
                first = sample
            check_growth(first, sample, args.max_heap_growth_mib * 1024 * 1024, args.max_goroutine_growth)
            elapsed = time.monotonic() - start
            writer.writerow([datetime.now(timezone.utc).isoformat(), round(elapsed, 3), *sample])
            stream.flush()
            count += 1
            remaining = args.minutes * 60 - elapsed
            if remaining <= 0:
                break
            time.sleep(min(args.interval, remaining))
    args.out.with_suffix('.json').write_text(json.dumps({
        'source_commit': commit, 'source_dirty': dirty, 'platform': platform.platform(),
        'url': args.url, 'samples': count, 'seconds': time.monotonic() - start,
        'heap_growth_limit_mib': args.max_heap_growth_mib,
        'goroutine_growth_limit': args.max_goroutine_growth,
        'workload': 'external sampler; interactions must be driven separately',
    }, indent=2) + '\n')
    print(f'perf-sample: {count} valid samples; limits satisfied; {args.out}')


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        sys.exit(f'perf-sample: FAILED: {error}')
