"""Run repeated benchmarks, retain raw output, and compare compatible medians."""
import argparse
from datetime import datetime, timezone
import json
import math
import os
from pathlib import Path
import re
import signal
import statistics
import subprocess
import sys

ROOT = Path(__file__).resolve().parent.parent
GROUPS = (
    ('./internal/page', 'BenchmarkRedrawStages|BenchmarkRedrawWarm|BenchmarkRedrawCold|BenchmarkRedrawHeavyWarm'),
    ('./internal/replay', 'BenchmarkDrawRectScan'),
    ('./examples/platform/...', 'BenchmarkClickCount'),
    ('./examples/perf-benchmarks', 'BenchmarkNormalRedraw|BenchmarkLargeRedraw|BenchmarkFlappyTick'),
)


def provenance(root):
    def git(*args):
        result = subprocess.run(['git', *args], cwd=root, capture_output=True, text=True, check=False)
        return result.stdout.strip() if result.returncode == 0 else None
    env = json.loads(subprocess.check_output(['go', 'env', '-json', 'GOVERSION', 'GOOS', 'GOARCH', 'GOAMD64', 'CGO_ENABLED'], cwd=root, text=True))
    return {'source_commit': git('rev-parse', 'HEAD') or 'unknown',
            'source_dirty': None if git('rev-parse', '--git-dir') is None else bool(git('status', '--porcelain')),
            'go': env, 'gomaxprocs': os.environ.get('GOMAXPROCS', 'runtime default'),
            'cpu': os.cpu_count(), 'machine': os.uname().nodename if hasattr(os, 'uname') else os.environ.get('COMPUTERNAME', 'unknown')}


def summarize(output):
    rows = {}
    package = None
    for line in output.splitlines():
        if line.startswith('pkg: '):
            package = line[5:]
        fields = line.split()
        if not fields or not fields[0].startswith('Benchmark') or len(fields) < 4:
            continue
        if not fields[1].isdigit() or not package:
            continue
        name = re.sub(r'-\d+$', '', fields[0])
        metrics = rows.setdefault(f'{package}/{name}', {})
        for i in range(2, len(fields) - 1, 2):
            metrics.setdefault(fields[i + 1], []).append(float(fields[i]))
    if not rows:
        raise ValueError('no benchmark rows found')
    return {name: {unit: {'count': len(values), 'min': min(values),
                         'median': statistics.median(values),
                         'sample_p95': sorted(values)[math.ceil(len(values) * .95) - 1]}
                   for unit, values in metrics.items()} for name, metrics in rows.items()}


def compare(report, baseline, ratio):
    for key in ('go', 'gomaxprocs', 'machine', 'cpu', 'benchtime', 'count', 'excluded_benchmarks', 'gomaxprocs_effective'):
        if report.get(key) != baseline.get(key):
            raise ValueError(f'incompatible baseline: {key} differs')
    old = baseline.get('benchmarks', {})
    current = report['benchmarks']
    if current.keys() != old.keys():
        raise ValueError('baseline benchmark inventory differs')
    failures = []
    for name, metrics in current.items():
        before = old[name]['ns/op']['median']
        after = metrics['ns/op']['median']
        if after > before * ratio:
            failures.append(f'{name}: {after:g} ns/op > {ratio:g} x {before:g}')
    if failures:
        raise ValueError('median regression: ' + '; '.join(failures))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--root', type=Path, default=ROOT)
    parser.add_argument('--exclude', default='', help='exclude benchmark names matching this regex')
    parser.add_argument('--out', type=Path, default=ROOT / 'temp/perf-baseline.txt')
    parser.add_argument('--baseline', type=Path)
    parser.add_argument('--count', type=int, default=5)
    parser.add_argument('--benchtime', default='1s')
    parser.add_argument('--max-ratio', type=float, default=1.3)
    args = parser.parse_args()
    root = args.root.resolve()
    exclude = re.compile(args.exclude) if args.exclude else None
    if args.count < 2 or not math.isfinite(args.max_ratio) or args.max_ratio <= 1:
        parser.error('count must be at least 2 and max-ratio must be finite and greater than 1')
    args.out.parent.mkdir(parents=True, exist_ok=True)
    metadata = args.out.with_suffix('.json')
    if metadata.resolve() == args.out.resolve():
        parser.error('--out names raw benchmark output; use a suffix other than .json')
    report = {'status': 'running', 'started_at': datetime.now(timezone.utc).isoformat(),
              'count': args.count, 'benchtime': args.benchtime, 'max_ratio': args.max_ratio,
              'excluded_benchmarks': args.exclude, 'raw_output': str(args.out), 'benchmarks': {}}
    if args.baseline and args.baseline.resolve() == metadata.resolve():
        parser.error('baseline and output report must be different files')
    try:
        report.update(provenance(root))
        metadata.write_text(json.dumps(report, indent=2) + '\n')
        with args.out.open('w') as output:
            for package, pattern in GROUPS:
                command = ['go', 'test', '-p', '1', package, '-run', '^$', '-bench', pattern,
                           '-benchmem', '-count', str(args.count), '-benchtime', args.benchtime]
                result = subprocess.run(command, cwd=root, stdout=output, stderr=subprocess.STDOUT, check=False)
                output.flush()
                if result.returncode:
                    raise ValueError(f'{package}: benchmark command exited {result.returncode}; see {args.out}')
        report['gomaxprocs_effective'] = sorted({int(n) for n in re.findall(r'^Benchmark\S+-(\d+)\s+\d+\s', args.out.read_text(), re.M)})
        report['benchmarks'] = {name: metrics for name, metrics in summarize(args.out.read_text()).items()
                                if exclude is None or not exclude.search(name)}
        if not report['benchmarks']:
            raise ValueError('no benchmarks remain after exclusions')
        if any(metric['ns/op']['count'] != args.count for metric in report['benchmarks'].values()):
            raise ValueError('benchmark sample count does not match requested count')
        if args.baseline:
            baseline = json.loads(args.baseline.read_text())
            if baseline.get('status') != 'passed':
                raise ValueError('baseline did not pass')
            compare(report, baseline, args.max_ratio)
        report['status'] = 'passed'
        print(args.out.read_text(), end='')
        print(f'Wrote {metadata}')
    except BaseException as error:
        report['status'], report['error'] = 'failed', str(error)
        raise
    finally:
        report['finished_at'] = datetime.now(timezone.utc).isoformat()
        metadata.write_text(json.dumps(report, indent=2) + '\n')


def interrupted(signum, _frame):
    raise KeyboardInterrupt(f'signal {signum}')


if __name__ == '__main__':
    signal.signal(signal.SIGTERM, interrupted)
    try:
        main()
    except (Exception, KeyboardInterrupt) as error:
        sys.exit(f'perf-baseline: FAILED: {error}')
