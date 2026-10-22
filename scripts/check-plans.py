#!/usr/bin/env python3
"""Reject ambiguous states in the authoritative maintenance ledger."""
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parent.parent
text = (ROOT / 'plans/README.md').read_text()
rows = re.findall(r'^- \[([^]]+)\] (.+)$', text, re.M)
assert rows, 'plans-check: no status rows'
seen = set()
for state, row in rows:
    assert state in ('x', ' ', '-', 's'), f'plans-check: unknown state {state!r}'
    assert not re.search(r'\[[x~s-]\]', row), f'plans-check: conflicting states: {row}'
    identifier = re.match(r'`([^`]+)` ', row)
    assert identifier, f'plans-check: row needs a stable identifier: {row}'
    assert identifier[1] not in seen, f'plans-check: duplicate identifier {identifier[1]}'
    seen.add(identifier[1])
    if state == 'x':
        assert 'Evidence:' in row and 'pending' not in row.lower(), f'plans-check: implemented row needs evidence: {row}'
    if state == 's':
        assert 'Replacement:' in row, f'plans-check: superseded row needs replacement: {row}'
    if state == '-':
        assert 'Reconsider' in row, f'plans-check: deferred row needs a condition: {row}'
for path in (ROOT / 'plans/v0.0.3/prod-ready-task').glob('*.md'):
    assert 'superseded' in path.read_text().lower(), f'plans-check: missing archive notice: {path}'
    assert not re.search(r'^- \[x\]', path.read_text(), re.M), f'plans-check: archive claims current completion: {path}'
print(f'plans-check: {len(rows)} rows with distinct states')
