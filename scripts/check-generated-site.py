#!/usr/bin/env python3
"""Fail when rebuilding the website changes the checked-in output."""
from pathlib import Path
import hashlib
import subprocess
import sys

ROOT = Path(__file__).resolve().parent.parent


def snapshot(directory):
    return {str(path.relative_to(directory)): hashlib.sha256(path.read_bytes()).hexdigest()
            for path in directory.rglob('*') if path.is_file()}


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
