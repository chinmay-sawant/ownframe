#!/usr/bin/env python3
"""Prove the lint summary hides success noise, never failure diagnostics."""
import contextlib
import io
from pathlib import Path
import runpy
import sys
import unittest

run_stage = runpy.run_path(str(Path(__file__).with_name('run-lint.py')))['run_stage']


class LintOutputTests(unittest.TestCase):
    def stage(self, code, verbose=False):
        output = io.StringIO()
        with contextlib.redirect_stdout(output):
            passed = run_stage('fixture', [sys.executable, '-c', code], verbose)
        return passed, output.getvalue()

    def test_success_is_one_line(self):
        passed, output = self.stage('print("noisy tool output")')
        self.assertTrue(passed)
        self.assertEqual(output, 'PASS fixture\n')

    def test_failure_preserves_stdout_stderr_and_exit_status(self):
        passed, output = self.stage('import sys; print("source.go:7: error"); print("details", file=sys.stderr); sys.exit(2)')
        self.assertFalse(passed)
        self.assertIn('FAIL fixture', output)
        self.assertIn('source.go:7: error', output)
        self.assertIn('details', output)

    def test_verbose_keeps_success_details(self):
        passed, output = self.stage('print("full details")', verbose=True)
        self.assertTrue(passed)
        self.assertIn('full details', output)

    def test_missing_tool_is_a_failure(self):
        output = io.StringIO()
        with contextlib.redirect_stdout(output):
            passed = run_stage('missing', [str(Path(__file__).with_name('nonexistent-lint-tool'))])
        self.assertFalse(passed)
        self.assertIn('FAIL missing', output.getvalue())


if __name__ == '__main__':
    unittest.main()
