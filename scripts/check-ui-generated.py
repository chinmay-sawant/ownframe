#!/usr/bin/env python3
"""Verify ui/properties_generated.go matches scripts/gen-ui-properties output."""
import difflib
import os
from pathlib import Path
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parent.parent
TARGET_REL = 'ui/properties_generated.go'
MAX_DIFF_LINES = 40


def compare_generated(actual_bytes, expected_bytes, target_rel=TARGET_REL):
    if actual_bytes == expected_bytes:
        return len(actual_bytes)
    actual_lines = actual_bytes.decode('utf-8', errors='replace').splitlines(keepends=True)
    expected_lines = expected_bytes.decode('utf-8', errors='replace').splitlines(keepends=True)
    diff = list(difflib.unified_diff(
        actual_lines,
        expected_lines,
        fromfile=f'{target_rel} (committed)',
        tofile=f'{target_rel} (generated)',
        n=2,
    ))
    preview = ''.join(diff[:MAX_DIFF_LINES])
    if len(diff) > MAX_DIFF_LINES:
        preview += f'... ({len(diff) - MAX_DIFF_LINES} more diff lines)\n'
    raise ValueError(
        f'ui-gen-check: {target_rel} is out of date; run `go generate ./ui`\n{preview}'
    )


def generate_to(out_path, root=ROOT):
    env = os.environ.copy()
    env['GOWORK'] = 'off'
    result = subprocess.run(
        ['go', '-C', str(root / 'scripts'), 'run', './gen-ui-properties', '-out', str(out_path)],
        cwd=root,
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        raise ValueError(f'ui-gen-check: generator failed:\n{result.stdout}')


def check_ui_generated(root=ROOT, generator=generate_to):
    target = root / TARGET_REL
    if not target.is_file():
        raise ValueError(f'ui-gen-check: missing {TARGET_REL}')
    with tempfile.TemporaryDirectory() as tmpdir:
        tmp_out = Path(tmpdir) / 'properties_generated.go'
        generator(tmp_out, root=root)
        expected = tmp_out.read_bytes()
    actual = target.read_bytes()
    return compare_generated(actual, expected, target_rel=TARGET_REL)


def main():
    size = check_ui_generated(ROOT)
    print(f'ui-gen-check: {TARGET_REL} matches generator ({size} bytes)')


if __name__ == '__main__':
    try:
        main()
    except (ValueError, OSError) as error:
        sys.exit(str(error))
