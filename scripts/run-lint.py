#!/usr/bin/env python3
"""Run every lint gate, summarizing success and preserving failure diagnostics."""
import argparse
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parent.parent


def run_stage(name, command, verbose=False):
    try:
        result = subprocess.run(command, cwd=ROOT, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, check=False)
    except OSError as error:
        print(f'FAIL {name}: {error}', flush=True)
        return False
    passed = result.returncode == 0
    print(f'{"PASS" if passed else "FAIL"} {name}', flush=True)
    if verbose or not passed:
        print(result.stdout, end='' if result.stdout.endswith('\n') else '\n', flush=True)
    return passed


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('version', help='declared golangci-lint version')
    parser.add_argument('--verbose', action='store_true', help='print successful stage output too')
    args = parser.parse_args()
    stages = [
        ('maintenance tooling', ['make', '--no-print-directory', 'tools-check']),
        ('linter version regression', ['bash', 'scripts/lint_test.sh']),
        (f'Go library and examples ({args.version})', ['bash', 'scripts/lint.sh', args.version]),
        ('documentation snippets', ['python3', 'scripts/check-doc-snippets.py']),
        ('plan states', ['python3', 'scripts/check-plans.py']),
        ('example maintenance inventory', ['python3', 'scripts/check-examples.py']),
        ('example internal imports', ['bash', 'scripts/check-example-internal.sh']),
        ('website ESLint', ['npm', '--prefix', 'frontend', 'run', 'lint']),
    ]
    passed = True
    for name, command in stages:
        # Do not short circuit: a failed gate must not hide later diagnostics.
        passed = run_stage(name, command, args.verbose) and passed
    return 0 if passed else 1


if __name__ == '__main__':
    sys.exit(main())
