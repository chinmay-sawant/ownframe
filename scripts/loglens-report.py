#!/usr/bin/env python3
"""Summarize native Loglens captures using elapsed timestamps and counter deltas."""
import argparse
from datetime import datetime
import json
import math
from pathlib import Path
import statistics
import sys

COUNTERS = ('Redraws', 'Parses', 'Cascades', 'FullLayouts', 'HoverPatches',
            'HoverFallbacks', 'HoverPatchErrors')


def number(values, key):
    value = values[key]
    if isinstance(value, bool) or not isinstance(value, (int, float)) or not math.isfinite(value) or value < 0:
        raise ValueError(f'invalid {key}: {value!r}')
    return value


def delta(first, last, key):
    result = number(last, key) - number(first, key)
    if result < 0:
        raise ValueError(f'counter reset: {key}')
    return result


def seconds(first, last):
    elapsed = (datetime.fromisoformat(last.replace('Z', '+00:00')) -
               datetime.fromisoformat(first.replace('Z', '+00:00'))).total_seconds()
    if elapsed <= 0:
        raise ValueError('timestamps must advance')
    return elapsed


def summarize(samples):
    if len(samples) < 2:
        raise ValueError('need at least two samples')
    for first, last in zip(samples, samples[1:]):
        if last['elapsed_seconds'] <= first['elapsed_seconds']:
            raise ValueError('sample clock did not advance')
        for name in COUNTERS:
            delta(first['state']['stats'], last['state']['stats'], name)
        delta(first['state']['frame'], last['state']['frame'], 'presents')
        delta(first['state']['runtime'], last['state']['runtime'], 'total_alloc')
    first, last = samples[0], samples[-1]
    a, b = first['state'], last['state']
    elapsed = number(last, 'elapsed_seconds') - number(first, 'elapsed_seconds')
    frame_elapsed = seconds(a['sampled_at'], b['sampled_at'])
    runtime_elapsed = seconds(a['runtime']['sampled_at'], b['runtime']['sampled_at'])
    alloc = delta(a['runtime'], b['runtime'], 'total_alloc')
    presents = delta(a['frame'], b['frame'], 'presents')
    first_failures = failure_counts(a['stats'])
    last_failures = failure_counts(b['stats'])
    failure_deltas = {}
    for key in first_failures.keys() | last_failures.keys():
        change = last_failures.get(key, 0) - first_failures.get(key, 0)
        if change < 0:
            raise ValueError(f'counter reset: hover failure {key}')
        if change:
            failure_deltas[key] = change
    gpu_memory = gpu_memory_summary(samples)
    result = {
        'elapsed_seconds': elapsed,
        'frame_sample_seconds': frame_elapsed,
        'runtime_sample_seconds': runtime_elapsed,
        'presents': presents,
        'presents_per_second': presents / frame_elapsed,
        'allocated_bytes': alloc,
        'allocated_bytes_per_second': alloc / runtime_elapsed,
        'allocated_bytes_per_present': alloc / presents if presents else None,
        'cpu_seconds': delta(first, last, 'cpu_seconds'),
        'working_set_first_bytes': number(first, 'working_set_bytes'),
        'working_set_final_bytes': number(last, 'working_set_bytes'),
        'working_set_peak_bytes': max(number(s, 'working_set_bytes') for s in samples),
        'private_final_bytes': number(last, 'private_bytes'),
        'private_peak_bytes': max(number(s, 'private_bytes') for s in samples),
        'pipeline_delta': {key: delta(a['stats'], b['stats'], key) for key in COUNTERS},
        'hover_patch_failure_delta': failure_deltas,
        'gpu_memory': gpu_memory,
        'gc_cycles': delta(a['runtime'], b['runtime'], 'num_gc'),
        'gc_pause_ns': delta(a['runtime'], b['runtime'], 'gc_pause_total_ns'),
        'runtime_managed_final_bytes': number(b['runtime'], 'sys') - number(b['runtime'], 'heap_released'),
        'last_sampled_redraw_alloc_bytes': number(b['stats'], 'AllocFrame'),
        'max_observed_rolling_p99_ns': max(number(s['state']['frame'], 'p99_ns') for s in samples),
        'max_observed_last_redraw_ns': max(number(s['state']['stats'], 'LastRedraw') for s in samples),
        'window_widths': sorted({s['state']['window_width'] for s in samples}),
        'scroll_y_range': [min(s['state']['scroll_y'] for s in samples), max(s['state']['scroll_y'] for s in samples)],
        'generations': delta(a, b, 'generation'),
    }
    if result['runtime_managed_final_bytes'] < 0:
        raise ValueError('released heap exceeds runtime memory')
    return result


def failure_counts(stats):
    counts = {}
    for entry in stats.get('HoverPatchFailures', []):
        key = '/'.join((entry['stage'], entry['reason'], entry['target']))
        value = entry['count']
        if isinstance(value, bool) or not isinstance(value, int) or value < 0:
            raise ValueError(f'invalid hover failure count: {value!r}')
        counts[key] = value
    return counts


def gpu_memory_summary(samples):
    totals = {}
    available = False
    for sample in samples:
        gpu = sample.get('gpu_memory')
        if not gpu or not gpu.get('available'):
            continue
        available = True
        for counter in gpu.get('counters', []):
            key = counter['path']
            value = counter['value_bytes']
            entry = totals.setdefault(key, {'peak_bytes': 0, 'final_bytes': 0})
            entry['peak_bytes'] = max(entry['peak_bytes'], value)
            entry['final_bytes'] = value
    return {'available': available, 'counters': totals}


def report(directory):
    runs = []
    for manifest in sorted(directory.glob('run-*/capture.json')):
        metadata = json.loads(manifest.read_text(encoding='utf-8-sig'))
        if metadata['status'] != 'passed':
            raise ValueError(f'{manifest}: capture did not pass: {metadata.get("error")}')
        phases = {}
        for phase in metadata['phases']:
            samples = json.loads((manifest.parent / phase / 'samples.json').read_text(encoding='utf-8-sig'))
            phases[phase] = summarize(samples)
        required = {'idle', 'hover', 'scroll', 'resize', 'navigation', 'idle-after'}
        if set(phases) != required:
            raise ValueError(f'{manifest}: missing workload phases')
        if phases['hover']['pipeline_delta']['HoverPatches'] == 0:
            raise ValueError(f'{manifest}: hover phase did not exercise scoped patches')
        if phases['scroll']['scroll_y_range'][0] == phases['scroll']['scroll_y_range'][1]:
            raise ValueError(f'{manifest}: scroll phase did not move the page')
        if len(phases['resize']['window_widths']) < 2:
            raise ValueError(f'{manifest}: resize phase did not change window width')
        if phases['navigation']['pipeline_delta']['Parses'] == 0:
            raise ValueError(f'{manifest}: navigation phase did not rebuild a screen')
        runs.append({'metadata': metadata, 'phases': phases})
    if not runs:
        raise ValueError(f'no run-*/capture.json in {directory}')
    configurations = {(r['metadata']['sha256'], r['metadata']['devtools'],
                       r['metadata']['content_width'], r['metadata']['height'],
                       r['metadata']['seconds_per_phase'], r['metadata']['warmup_seconds'],
                       r['metadata'].get('window_width'), r['metadata'].get('window_height'),
                       r['metadata'].get('topology_label', 'unknown')) for r in runs}
    if len(configurations) != 1:
        raise ValueError('capture mixes binaries or configurations')
    return {
        'runs': runs,
        'repetitions': len(runs),
        'medians': {
            phase: {key: statistics.median(r['phases'][phase][key] for r in runs)
                    for key in ('allocated_bytes_per_second', 'presents_per_second',
                                'cpu_seconds', 'working_set_final_bytes', 'private_final_bytes')}
            for phase in runs[0]['phases']
        },
        'limits': [
            'Profiles include instrumentation overhead. Heap profiles use gc=1 outside measured phases.',
            'Frame percentiles describe rolling 120-frame host work, not input latency or GPU completion.',
            'Allocation/frame means a sampled redraw only. Allocation/present uses runtime and present deltas with independently sampled boundaries.',
            'Working set, private bytes, Go-managed memory, and heap-profile live bytes are different measurements.',
        ],
    }


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('capture', type=Path)
    parser.add_argument('--out', type=Path)
    args = parser.parse_args()
    try:
        result = report(args.capture)
        encoded = json.dumps(result, indent=2) + '\n'
        if args.out:
            args.out.write_text(encoded)
        else:
            print(encoded, end='')
    except (OSError, KeyError, ValueError, TypeError) as error:
        print(f'FAILED: {error}', file=sys.stderr)
        return 1
    return 0


if __name__ == '__main__':
    sys.exit(main())
