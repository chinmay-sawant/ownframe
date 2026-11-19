#!/usr/bin/env python3
"""Reject newly tracked files larger than 5 MiB unless explicitly allowlisted."""
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parent.parent
MAX_BYTES = 5 * 1024 * 1024
ALLOWLIST_PATH = Path(__file__).resolve().with_name('repo-size-allowlist.txt')


def load_allowlist(path=ALLOWLIST_PATH):
    entries = set()
    for raw in path.read_text().splitlines():
        line = raw.strip()
        if not line or line.startswith('#'):
            continue
        entries.add(line)
    return entries


def check_sizes(file_sizes, allowlist, max_bytes=MAX_BYTES):
    oversized = {path: size for path, size in file_sizes.items() if size > max_bytes}
    unexpected = sorted(path for path in oversized if path not in allowlist)
    stale = sorted(path for path in allowlist if path not in oversized)
    errors = []
    for path in unexpected:
        mb = oversized[path] / (1024 * 1024)
        errors.append(f'new tracked file exceeds 5 MiB ({mb:.1f} MiB): {path}')
    for path in stale:
        errors.append(f'stale allowlist entry (removed or <= 5 MiB): {path}')
    if errors:
        raise ValueError('repo-size-check:\n' + '\n'.join(f'  - {err}' for err in errors))
    return len(oversized)


def tracked_file_sizes(root=ROOT):
    raw = subprocess.check_output(['git', 'ls-files', '-z'], cwd=root)
    sizes = {}
    for entry in raw.split(b'\0'):
        if not entry:
            continue
        rel = entry.decode('utf-8')
        path = root / rel
        if path.is_file():
            sizes[rel] = path.stat().st_size
    return sizes


def main():
    count = check_sizes(tracked_file_sizes(ROOT), load_allowlist(ALLOWLIST_PATH))
    print(f'repo-size-check: {count} allowlisted >5 MiB files; 0 new violations')


if __name__ == '__main__':
    try:
        main()
    except ValueError as error:
        sys.exit(str(error))
