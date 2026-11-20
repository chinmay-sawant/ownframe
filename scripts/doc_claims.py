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



def check_current_options(text):
    if re.search(r'ServeOptions\s*\{[^}]*\b(?:Perf|Debug)\s*:', text):
        raise ValueError('docs-check: use explicit InsecureAllow serve options')


def check_doc_pins(text, name, go_version, engine):
    for raw in re.findall(r'blinkless`?\s+`?(v\d+\.\d+\.\d+(?:-[\w.\-]+)?)', text):
        pin = raw.rstrip('.,;:)')
        if pin != engine:
            raise ValueError(f'docs-check: {name} has a stale blinkless pin {pin!r} (go.mod has {engine!r}); link go.mod')
    for pin in re.findall(r'\b(1\.\d+\.\d+)\b', text):
        if pin != go_version:
            raise ValueError(f'docs-check: {name} has a stale Go version {pin!r} (go.mod has {go_version!r})')


def check_relative_links(root, names):
    missing = []
    for name in names:
        path = root / name
        text = re.sub(r'^```.*?^```', '', path.read_text(), flags=re.M | re.S)
        for match in re.finditer(r'\[[^\]]+\]\(([^)]+)\)', text):
            target = match.group(1).strip().split('#')[0]
            if not target or target.startswith(('http://', 'https://', 'mailto:', 'file://')):
                continue
            resolved = (path.parent / target).resolve()
            if not resolved.exists():
                missing.append(f'{name} -> {target}')
    if missing:
        raise ValueError('docs-check: broken relative links:\n' + '\n'.join(f'  - {item}' for item in missing))


def validate(root):
    for path in [root / 'README.md', *sorted((root / 'documentation').glob('*.md')),
                 *sorted((root / 'examples').rglob('*.md'))]:
        check_current_options(path.read_text())
    go_version, engine = check_versions([(root / name).read_text() for name in ('go.mod', 'examples/go.mod')])
    for name in ('AGENTS.md', 'README.md', 'CONTRIBUTING.md', 'documentation/android-runtime.md'):
        check_doc_pins((root / name).read_text(), name, go_version, engine)
    check_relative_links(root, ('README.md', 'AGENTS.md', 'documentation/README.md', 'plans/README.md'))
    performance = (root / 'documentation/performance.md').read_text()
    if 'ticking pages that do not call `UseFrameDirty`' in performance:
        raise ValueError('docs-check: performance guide describes the retired tick opt-in')
    sample = (root / 'scripts/perf-sample.py').read_text()
    sample_flags = set(re.findall(r"add_argument\('(--[\w-]+)'", sample))
    check_perf_commands(performance, {
        'scripts/perf-idle.sh': {'--example', '--web', '--seconds', '--url'},
        'scripts/perf-longrun.sh': sample_flags,
        'scripts/perf-baseline.sh': {'--out', '--root', '--exclude', '--baseline', '--max-ratio', '--count', '--benchtime'},
    })
