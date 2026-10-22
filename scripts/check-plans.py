#!/usr/bin/env python3
"""Reject ambiguous states in the authoritative maintenance ledger."""
from pathlib import Path
import re
import sys

ROOT = Path(__file__).resolve().parent.parent


def validate_ledger(text):
    rows = re.findall(r'^- \[([^]]+)\] (.+)$', text, re.M)
    if not rows:
        raise ValueError('plans-check: no status rows')
    seen = set()
    for state, row in rows:
        if state not in ('x', ' ', '-', 's'):
            raise ValueError(f'plans-check: unknown state {state!r}')
        if re.search(r'\[[x~s-]\]', row):
            raise ValueError(f'plans-check: conflicting states: {row}')
        identifier = re.match(r'`([^`]+)` ', row)
        if not identifier:
            raise ValueError(f'plans-check: row needs a stable identifier: {row}')
        if identifier[1] in seen:
            raise ValueError(f'plans-check: duplicate identifier {identifier[1]}')
        seen.add(identifier[1])
        if state == 'x' and ('Evidence:' not in row or 'pending' in row.lower()):
            raise ValueError(f'plans-check: implemented row needs evidence: {row}')
        if state == 's' and 'Replacement:' not in row:
            raise ValueError(f'plans-check: superseded row needs replacement: {row}')
        if state == '-' and 'Reconsider' not in row:
            raise ValueError(f'plans-check: deferred row needs a condition: {row}')
    return len(rows)


def main():
    count = validate_ledger((ROOT / 'plans/README.md').read_text())
    for path in (ROOT / 'plans/v0.0.3/prod-ready-task').glob('*.md'):
        text = path.read_text()
        if 'superseded' not in text.lower():
            raise ValueError(f'plans-check: missing archive notice: {path}')
        if re.search(r'^- \[x\]', text, re.M):
            raise ValueError(f'plans-check: archive claims current completion: {path}')
    print(f'plans-check: {count} rows with distinct states')


if __name__ == '__main__':
    try:
        main()
    except ValueError as error:
        sys.exit(str(error))
