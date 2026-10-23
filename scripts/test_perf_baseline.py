import copy
import unittest

from perf_baseline import compare, summarize


class PerfBaselineTests(unittest.TestCase):
    def test_summary_and_regression(self):
        metrics = summarize('pkg: example\nBenchmarkRedraw-8 100 100 ns/op 20 B/op 2 allocs/op\nBenchmarkRedraw-8 100 120 ns/op 20 B/op 2 allocs/op\n')
        self.assertEqual(metrics['example/BenchmarkRedraw']['ns/op']['median'], 110)
        report = {'benchmarks': metrics}
        baseline = copy.deepcopy(report)
        compare(report, baseline, 1.3)
        report['benchmarks']['example/BenchmarkRedraw']['ns/op']['median'] = 165
        with self.assertRaisesRegex(ValueError, 'median regression'):
            compare(report, baseline, 1.3)

    def test_incompatible_baseline_rejected(self):
        report = {'machine': 'first', 'benchmarks': {}}
        with self.assertRaisesRegex(ValueError, 'machine'):
            compare(report, {'machine': 'second'}, 1.3)

    def test_missing_rows_rejected(self):
        with self.assertRaisesRegex(ValueError, 'no benchmark'):
            summarize('PASS\n')
        with self.assertRaisesRegex(ValueError, 'inventory'):
            compare({'benchmarks': {'new': {}}}, {'benchmarks': {}}, 1.3)
