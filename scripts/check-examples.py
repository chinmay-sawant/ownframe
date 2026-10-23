#!/usr/bin/env python3
"""Require an explicit maintenance decision for every example root."""
import json
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parent.parent


def validate_inventory(manifest, roots):
    maintained = manifest['maintained']
    experimental = manifest['experimental']
    if len(experimental) != len(set(experimental)):
        raise ValueError('examples-check: duplicate experimental root')
    if maintained.keys() & set(experimental):
        raise ValueError('examples-check: conflicting classifications')
    classified = maintained.keys() | set(experimental)
    if roots != classified:
        raise ValueError(f'examples-check: unclassified {roots - classified}, stale {classified - roots}')
    if not all(maintained.values()):
        raise ValueError('examples-check: maintained examples need a teaching purpose')
    return maintained, experimental


def main():
    manifest = json.loads((ROOT / 'examples/maintenance.json').read_text())
    files = subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard', 'examples'], cwd=ROOT, text=True).splitlines()
    roots = {p.split('/')[1] for p in files if p.endswith('.go')}
    maintained, experimental = validate_inventory(manifest, roots)
    print(f'examples-check: maintained teaching set: {", ".join(maintained)}; {len(experimental)} experimental/support roots')


if __name__ == '__main__':
    try:
        main()
    except (ValueError, KeyError) as error:
        sys.exit(str(error))
