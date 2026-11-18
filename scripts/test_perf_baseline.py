import copy
import unittest

from perf_baseline import compare, summarize, validate_list_inventory


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

    def test_list_inventory_requires_both_modes_and_all_workloads(self):
        prefix = 'github.com/chinmay-sawant/ownframe/examples/perf-benchmarks/BenchmarkListWorkloads/'
        rows = {prefix + mode + '/' + name: {} for mode in ('PerfOff', 'PerfOn')
                for name in ('Unchanged', 'TextUpdate', 'ScrollOffset', 'Resize', 'WindowedScroll')}
        validate_list_inventory(rows)
        del rows[prefix + 'PerfOff/WindowedScroll']
        with self.assertRaisesRegex(ValueError, 'inventory'):
            validate_list_inventory(rows)

    def test_different_workload_conditions_rejected(self):
        report = {'benchmarks': {}, 'suite': 'list', 'conditions': 'idle'}
        baseline = dict(report, conditions='concurrent tests')
        with self.assertRaisesRegex(ValueError, 'conditions'):
            compare(report, baseline, 1.3)
