#!/usr/bin/env python3
"""Build the Pages artifact and check its WebAssembly payloads."""
import gzip
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parent.parent
WASM_HEADER = b'\x00asm\x01\x00\x00\x00'
DEMOS = {
    'demo': 'demo.wasm',
    'demos/builder': 'demo.wasm',
    'demos/expense-tracker': 'demo.wasm',
    'demos/hello': 'demo.wasm',
    'demos/spotify-player': 'ownframe.wasm.gz',
    'demos/telegram': 'ownframe.wasm',
    'demos/ui-builder': 'demo.wasm',
    'demos/wispr-flow': 'ownframe.wasm',
    'demos/workspace-inspector': 'demo.wasm',
}


def check_site(site):
    if not (site / 'index.html').is_file():
        raise ValueError('Pages index.html is missing')
    expected = {site / folder / binary for folder, binary in DEMOS.items()}
    actual = set(site.rglob('*.wasm')) | set(site.rglob('*.wasm.gz'))
    if missing := expected - actual:
        raise ValueError(f'Missing WebAssembly: {", ".join(str(p.relative_to(site)) for p in sorted(missing))}')
    if extra := actual - expected:
        raise ValueError(f'Unexpected WebAssembly: {", ".join(str(p.relative_to(site)) for p in sorted(extra))}')
    for folder, binary in DEMOS.items():
        directory = site / folder
        for name in ('index.html', 'wasm_exec.js'):
            if not (directory / name).is_file():
                raise ValueError(f'Missing demo file: {folder}/{name}')
        wasm = directory / binary
        open_wasm = gzip.open if wasm.suffix == '.gz' else open
        with open_wasm(wasm, 'rb') as stream:
            if stream.read(len(WASM_HEADER)) != WASM_HEADER:
                raise ValueError(f'Invalid WebAssembly header: {folder}/{binary}')


def main():
    subprocess.run(['npm', 'run', 'build'], cwd=ROOT / 'frontend', check=True)
    check_site(ROOT / 'frontend/dist')
    print(f'site-check: Pages artifact contains {len(DEMOS)} WebAssembly demos')


if __name__ == '__main__':
    try:
        main()
    except (ValueError, OSError, EOFError, subprocess.CalledProcessError) as error:
        sys.exit(str(error))
