#!/usr/bin/env python3
"""Require an explicit maintenance decision for every example root."""
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parent.parent
manifest = json.loads((ROOT / 'examples/maintenance.json').read_text())
files = subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard', 'examples'], cwd=ROOT, text=True).splitlines()
roots = {p.split('/')[1] for p in files if p.endswith('.go')}
maintained = manifest['maintained']
experimental = manifest['experimental']
assert len(experimental) == len(set(experimental)), 'examples-check: duplicate experimental root'
assert not maintained.keys() & set(experimental), 'examples-check: conflicting classifications'
classified = maintained.keys() | set(experimental)
assert roots == classified, f'examples-check: unclassified {roots - classified}, stale {classified - roots}'
assert all(maintained.values()), 'examples-check: maintained examples need a teaching purpose'
print(f'examples-check: maintained teaching set: {", ".join(maintained)}; {len(experimental)} experimental/support roots')
