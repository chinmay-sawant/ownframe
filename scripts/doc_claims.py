"""Checks for operational claims that Go snippet compilation cannot cover."""
from pathlib import Path
import re
import shlex


def check_versions(modules):
    versions = []
    for text in modules:
        go = re.search(r'^go (\S+)$', text, re.M)
        engine = re.search(r'github\.com/chinmay-sawant/blinkless\s+(\S+)', text)
        if not go or not engine:
            raise ValueError('docs-check: missing Go or Blinkless requirement')
        versions.append((go[1], engine[1]))
    if len(set(versions)) != 1:
        raise ValueError('docs-check: library and examples Go/Blinkless requirements disagree')
    return versions[0]


def check_perf_commands(text, options):
    for line in text.splitlines():
        if not line.startswith('sh scripts/perf-'):
            continue
        words = shlex.split(line)
        flags = options.get(words[1])
        if flags is None:
            raise ValueError(f'docs-check: unregistered performance command: {line}')
        args = iter(words[2:])
        for flag in args:
            if flag not in flags:
                raise ValueError(f'docs-check: invalid performance argument: {flag}')
            if flag != '--web':
                value = next(args, None)
                if value is None or value.startswith('--'):
                    raise ValueError(f'docs-check: missing value for {flag}')


def validate(root):
    _, engine = check_versions([(root / name).read_text() for name in ('go.mod', 'examples/go.mod')])
    android = (root / 'documentation/android-runtime.md').read_text()
    for pin in re.findall(r'blinkless (v[\w.\-]+)', android):
        if pin != engine:
            raise ValueError('docs-check: Android guide has a stale engine pin; link go.mod')
    performance = (root / 'documentation/performance.md').read_text()
    if 'ticking pages that do not call `UseFrameDirty`' in performance:
        raise ValueError('docs-check: performance guide describes the retired tick opt-in')
    sample = (root / 'scripts/perf-sample.py').read_text()
    sample_flags = set(re.findall(r"add_argument\('(--[\w-]+)'", sample))
    check_perf_commands(performance, {
        'scripts/perf-idle.sh': {'--example', '--web', '--seconds', '--url'},
        'scripts/perf-longrun.sh': sample_flags,
        'scripts/perf-baseline.sh': {'--out'},
    })
