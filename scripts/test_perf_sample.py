import importlib.util
from pathlib import Path
import subprocess
import sys
import unittest

SPEC = importlib.util.spec_from_file_location('perf_sample', Path(__file__).with_name('perf-sample.py'))
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class PerfSampleTests(unittest.TestCase):
    def test_metrics_required(self):
        valid = {'stats': dict.fromkeys(MODULE.STAT_FIELDS, 1), 'runtime': {'heap_alloc': 10, 'goroutines': 2}}
        self.assertEqual(MODULE.metrics(valid), [1] * 5 + [10, 2])
        for body in ({}, {'stats': {}}, {'stats': valid['stats'], 'runtime': {}}, {'stats': valid['stats'], 'runtime': {'heap_alloc': True, 'goroutines': 2}}):
            with self.subTest(body=body), self.assertRaises(ValueError):
                MODULE.metrics(body)

    def test_growth_and_counter_reset_fail(self):
        before = [1] * 5 + [10, 2]
        MODULE.check_growth(before, before, 10, 1)
        for after in ([1] * 5 + [21, 2], [1] * 5 + [10, 4], [0] * 5 + [10, 2]):
            with self.subTest(after=after), self.assertRaises(ValueError):
                MODULE.check_growth(before, after, 10, 1)

    def test_missing_endpoint_fails_process(self):
        result = subprocess.run([sys.executable, str(Path(MODULE.__file__)), '--minutes', '0.001', '--url', 'http://127.0.0.1:1/debug/state'], capture_output=True, text=True, check=False)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('FAILED', result.stderr)
