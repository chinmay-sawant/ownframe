#!/usr/bin/env python3
"""Fail when rebuilding the website changes the checked-in output."""
from pathlib import Path
import hashlib
import subprocess
import sys

ROOT = Path(__file__).resolve().parent.parent


def snapshot(directory):
    """Hash the files git would commit: tracked plus untracked non-ignored.

    Build intermediates that git ignores (such as the unpacked Spotify
    wasm, which stays under no circumstances in the repo past GitHub's
    100 MB cap) must not fail the check; they differ per machine and can
    never be checked in.
    """
    rel = str(directory.relative_to(ROOT))
    tracked = subprocess.run(['git', 'ls-files', '-z', '--', rel],
                             cwd=ROOT, check=True, capture_output=True,
                             text=True).stdout.split('\0')
    others = subprocess.run(['git', 'ls-files', '--others', '--exclude-standard',
                             '-z', '--', rel],
                            cwd=ROOT, check=True, capture_output=True,
                            text=True).stdout.split('\0')
    # A rebuild can delete a tracked hashed asset and add its replacement.
    # The deleted path is still listed by ls-files but no longer on disk;
    # record it as missing so the comparison below reports it as changed
    # instead of crashing on the read.
    snapshot = {}
    for name in tracked + others:
        if not name:
            continue
        path = ROOT / name
        snapshot[name] = hashlib.sha256(path.read_bytes()).hexdigest() if path.is_file() else None
    return snapshot


def main():
    before = snapshot(ROOT / 'docs')
    subprocess.run(['npm', 'run', 'build'], cwd=ROOT / 'frontend', check=True)
    after = snapshot(ROOT / 'docs')
    changed = sorted(name for name in before.keys() | after.keys() if before.get(name) != after.get(name))
    if changed:
        raise ValueError('site-check: generated output was stale; review and include the rebuilt files: ' + ', '.join(changed))
    print('site-check: generated output matches the source')


if __name__ == '__main__':
    try:
        main()
    except (ValueError, subprocess.CalledProcessError) as error:
        sys.exit(str(error))
