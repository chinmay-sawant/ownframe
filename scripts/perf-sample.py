#!/usr/bin/env python3
"""Sample debug metrics, retain failure metadata, and check post-warmup RSS growth."""
import argparse
import csv
from datetime import datetime, timezone
import json
import math
from pathlib import Path
import platform
import signal
import subprocess
import sys
import time
from urllib.request import urlopen

ROOT = Path(__file__).resolve().parent.parent
STAT_FIELDS = ('Redraws', 'Parses', 'Cascades', 'Layouts', 'Repaints')


def number(mapping, key):
    value = mapping.get(key)
    if isinstance(value, bool) or not isinstance(value, (int, float)) or not math.isfinite(value) or value < 0:
        raise ValueError(f'missing or invalid metric: {key}')
    return value


def metrics(body):
    stats, runtime = body.get('stats'), body.get('runtime')
    if not isinstance(stats, dict) or not isinstance(runtime, dict):
        raise ValueError('debug endpoint needs stats and runtime; enable the debug build and perf options')
    rss = number(runtime, 'rss_bytes') if 'rss_bytes' in runtime else None
    return [number(stats, key) for key in STAT_FIELDS] + [number(runtime, 'heap_alloc'), number(runtime, 'goroutines'), rss]


def check_growth(first, current, heap_limit, goroutine_limit):
    if current[5] - first[5] > heap_limit:
        raise ValueError('heap growth exceeded the configured limit')
    if current[6] - first[6] > goroutine_limit:
        raise ValueError('goroutine growth exceeded the configured limit')
    if any(new < old for old, new in zip(first[:5], current[:5])):
        raise ValueError('page counters reset during the sample')


def rss_slope(samples):
    """Least-squares bytes/second after warmup; samples are (elapsed, rss)."""
    if len(samples) < 3 or samples[-1][0] - samples[0][0] < 60:
        return None
    mean_t = sum(t for t, _ in samples) / len(samples)
    mean_rss = sum(rss for _, rss in samples) / len(samples)
    numerator = sum((t - mean_t) * (rss - mean_rss) for t, rss in samples)
    denominator = sum((t - mean_t) ** 2 for t, _ in samples)
    return numerator / denominator


def check_rss(samples, limit_mib_hour):
    slope = rss_slope(samples)
    if slope is not None and limit_mib_hour > 0 and slope * 3600 > limit_mib_hour * 1024 * 1024:
        raise ValueError('post-warmup RSS slope exceeded the configured limit')
    return slope


def source_info():
    try:
        return git_source_info()
    except OSError:
        return {'source_commit': 'unknown', 'source_dirty': None}


def git_source_info():
    result = subprocess.run(['git', 'rev-parse', 'HEAD'], cwd=ROOT, capture_output=True, text=True, check=False)
    if result.returncode:
        return {'source_commit': 'unknown', 'source_dirty': None}
    dirty = subprocess.run(['git', 'status', '--porcelain'], cwd=ROOT, capture_output=True, text=True, check=False)
    return {'source_commit': result.stdout.strip(), 'source_dirty': bool(dirty.stdout) if dirty.returncode == 0 else None}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--minutes', type=float, default=15)
    parser.add_argument('--interval', type=float, default=30)
    parser.add_argument('--url', default='http://127.0.0.1:8128/debug/state')
    parser.add_argument('--out', type=Path, default=ROOT / 'temp/perf-sample.csv')
    parser.add_argument('--max-heap-growth-mib', type=float, default=64)
    parser.add_argument('--max-goroutine-growth', type=int, default=8)
    parser.add_argument('--warmup-seconds', type=float, default=30)
    parser.add_argument('--max-rss-growth-mib-hour', type=float, default=20)
    args = parser.parse_args()
    numbers = (args.minutes, args.interval, args.max_heap_growth_mib, args.warmup_seconds, args.max_rss_growth_mib_hour)
    if not all(math.isfinite(n) for n in numbers) or not 0 < args.minutes <= 60 or args.interval <= 0 or min(numbers[2:]) < 0 or args.max_goroutine_growth < 0:
        parser.error('duration and interval must be positive; limits and warmup must be nonnegative; duration is at most 60 minutes')
    args.out.parent.mkdir(parents=True, exist_ok=True)
    start = time.monotonic()
    first, rss_samples, count = None, [], 0
    report = {**source_info(), 'status': 'running', 'started_at': datetime.now(timezone.utc).isoformat(),
              'platform': platform.platform(), 'url': args.url, 'samples': 0,
              'heap_growth_limit_mib': args.max_heap_growth_mib,
              'goroutine_growth_limit': args.max_goroutine_growth,
              'rss_growth_limit_mib_hour': args.max_rss_growth_mib_hour,
              'warmup_seconds': args.warmup_seconds,
              'workload': 'external sampler; interactions must be driven separately'}
    metadata = args.out.with_suffix('.json')
    metadata.write_text(json.dumps(report, indent=2) + '\n')
    try:
        with args.out.open('w', newline='') as stream:
            writer = csv.writer(stream)
            writer.writerow(['utc', 'elapsed_seconds', *STAT_FIELDS, 'heap_alloc', 'goroutines', 'rss_bytes'])
            while True:
                with urlopen(args.url, timeout=5) as response:
                    sample = metrics(json.load(response))
                if first is None:
                    first = sample
                elapsed = time.monotonic() - start
                writer.writerow([datetime.now(timezone.utc).isoformat(), round(elapsed, 3), *sample])
                stream.flush()
                count += 1
                check_growth(first, sample, args.max_heap_growth_mib * 1024 * 1024, args.max_goroutine_growth)
                if sample[7] is None and args.max_rss_growth_mib_hour > 0:
                    raise ValueError('RSS unavailable; use --max-rss-growth-mib-hour 0 to disable its limit explicitly')
                if elapsed >= args.warmup_seconds and sample[7] is not None:
                    rss_samples.append((elapsed, sample[7]))
                    check_rss(rss_samples, args.max_rss_growth_mib_hour)
                remaining = args.minutes * 60 - elapsed
                if remaining <= 0:
                    break
                time.sleep(min(args.interval, remaining))
        if args.max_rss_growth_mib_hour > 0 and rss_slope(rss_samples) is None:
            raise ValueError('RSS budget needs at least three post-warmup samples spanning 60 seconds')
        report['status'] = 'passed'
    except BaseException as error:
        report['status'], report['error'] = 'failed', str(error) or type(error).__name__
        raise
    finally:
        slope = rss_slope(rss_samples)
        report.update(samples=count, seconds=time.monotonic() - start,
                      finished_at=datetime.now(timezone.utc).isoformat(),
                      rss_slope_mib_hour=None if slope is None else slope * 3600 / 1024 / 1024)
        metadata.write_text(json.dumps(report, indent=2) + '\n')
    print(f'perf-sample: {count} valid samples; limits satisfied; {args.out}')


def interrupted(signum, _frame):
    raise KeyboardInterrupt(f'signal {signum}')


if __name__ == '__main__':
    signal.signal(signal.SIGTERM, interrupted)
    try:
        main()
    except (Exception, KeyboardInterrupt) as error:
        sys.exit(f'perf-sample: FAILED: {error}')
