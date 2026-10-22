#!/usr/bin/env python3
"""Exercise rejected maintenance records, including optimized Python execution."""
from pathlib import Path
import runpy
import subprocess
import sys
import unittest

HERE = Path(__file__).resolve().parent
validate_ledger = runpy.run_path(str(HERE / 'check-plans.py'))['validate_ledger']
validate_inventory = runpy.run_path(str(HERE / 'check-examples.py'))['validate_inventory']


class MaintenanceChecks(unittest.TestCase):
    def test_valid_statuses(self):
        text = '\n'.join([
            '- [x] `DONE` implemented. Evidence: command exit 0.',
            '- [ ] `OPEN` awaiting work.',
            '- [-] `LATER` Reconsider when the host exists.',
            '- [s] `OLD` Replacement: current ledger.',
        ])
        self.assertEqual(validate_ledger(text), 4)

    def test_invalid_status_records_fail(self):
        cases = [
            '', '- [~] `A` ambiguous.', '- [x] `A` no proof.',
            '- [x] `A` Evidence: pending.', '- [-] `A` no condition.',
            '- [s] `A` no replacement.', '- [ ] no identifier.',
            '- [x] `A` [s] Evidence: conflicting.',
            '- [ ] `A` open.\n- [ ] `A` duplicate.',
        ]
        for text in cases:
            with self.subTest(text=text), self.assertRaises(ValueError):
                validate_ledger(text)

    def test_example_scope_must_match(self):
        valid = {'maintained': {'login': 'forms'}, 'experimental': ['demo']}
        validate_inventory(valid, {'login', 'demo'})
        for roots in [{'login'}, {'login', 'demo', 'surprise'}]:
            with self.subTest(roots=roots), self.assertRaises(ValueError):
                validate_inventory(valid, roots)
        invalid = [
            {'maintained': {'login': ''}, 'experimental': ['demo']},
            {'maintained': {'login': 'forms'}, 'experimental': ['demo', 'demo']},
            {'maintained': {'login': 'forms'}, 'experimental': ['login', 'demo']},
        ]
        for manifest in invalid:
            with self.subTest(manifest=manifest), self.assertRaises(ValueError):
                validate_inventory(manifest, {'login', 'demo'})

    def test_optimization_cannot_disable_validation(self):
        code = 'import runpy, sys; runpy.run_path(sys.argv[1])["validate_ledger"]("- [x] `A` no proof.")'
        result = subprocess.run([sys.executable, '-O', '-c', code, str(HERE / 'check-plans.py')], capture_output=True, text=True, check=False)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('implemented row needs evidence', result.stderr)
        code = 'import runpy, sys; runpy.run_path(sys.argv[1])["validate_inventory"]({"maintained":{},"experimental":[]},{"missing"})'
        result = subprocess.run([sys.executable, '-O', '-c', code, str(HERE / 'check-examples.py')], capture_output=True, text=True, check=False)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('unclassified', result.stderr)


if __name__ == '__main__':
    unittest.main()
