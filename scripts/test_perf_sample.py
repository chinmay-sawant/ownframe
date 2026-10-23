import importlib.util
import json
import io
import tempfile
from pathlib import Path
import subprocess
import sys
import unittest
from unittest import mock

SPEC = importlib.util.spec_from_file_location('perf_sample', Path(__file__).with_name('perf-sample.py'))
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class PerfSampleTests(unittest.TestCase):
    def test_metrics_required(self):
        valid = {'stats': dict.fromkeys(MODULE.STAT_FIELDS, 1), 'runtime': {'heap_alloc': 10, 'goroutines': 2, 'rss_bytes': 100}}
        self.assertEqual(MODULE.metrics(valid), [1] * 5 + [10, 2, 100])
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
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / 'sample.csv'
            result = subprocess.run([sys.executable, str(Path(MODULE.__file__)), '--minutes', '0.001', '--url', 'http://127.0.0.1:1/debug/state', '--out', str(output)], capture_output=True, text=True, check=False)
            self.assertNotEqual(result.returncode, 0)
            self.assertIn('FAILED', result.stderr)
            report = json.loads(output.with_suffix('.json').read_text())
            self.assertEqual(report['status'], 'failed')
            self.assertTrue(report['source_commit'])
            self.assertIn('rss_bytes', output.read_text())

    def test_rss_slope_after_warmup(self):
        mib = 1024 * 1024
        self.assertIsNone(MODULE.rss_slope([(0, mib), (30, 2*mib)]))
        self.assertEqual(MODULE.check_rss([(0, mib), (30, mib), (60, mib)], 20), 0)
        climb = [(0, mib), (1800, 16*mib), (3600, 31*mib)]
        with self.assertRaisesRegex(ValueError, 'RSS slope'):
            MODULE.check_rss(climb, 20)


    def test_heap_failure_keeps_last_sample_and_metadata(self):
        body = {'stats': dict.fromkeys(MODULE.STAT_FIELDS, 1),
                'runtime': {'heap_alloc': 10, 'goroutines': 2, 'rss_bytes': 100}}
        before = io.StringIO(json.dumps(body))
        body['runtime']['heap_alloc'] = 20
        after = io.StringIO(json.dumps(body))
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / 'sample.csv'
            argv = ['perf-sample.py', '--out', str(output), '--minutes', '0.01',
                    '--max-heap-growth-mib', '0', '--max-rss-growth-mib-hour', '0']
            with mock.patch.object(sys, 'argv', argv), \
                    mock.patch.object(MODULE, 'urlopen', side_effect=[before, after]), \
                    mock.patch.object(MODULE.time, 'sleep'), \
                    mock.patch.object(MODULE.time, 'monotonic', side_effect=[0, .01, .02, .03]):
                with self.assertRaisesRegex(ValueError, 'heap growth'):
                    MODULE.main()
            report = json.loads(output.with_suffix('.json').read_text())
            self.assertEqual(report['status'], 'failed')
            self.assertEqual(report['samples'], 2)
            self.assertEqual(len(output.read_text().splitlines()), 3)

    def test_interruption_finalizes_metadata(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / 'sample.csv'
            with mock.patch.object(sys, 'argv', ['perf-sample.py', '--out', str(output)]), \
                    mock.patch.object(MODULE, 'urlopen', side_effect=KeyboardInterrupt('signal 15')):
                with self.assertRaises(KeyboardInterrupt):
                    MODULE.main()
            report = json.loads(output.with_suffix('.json').read_text())
            self.assertEqual(report['status'], 'failed')
            self.assertEqual(report['error'], 'signal 15')
