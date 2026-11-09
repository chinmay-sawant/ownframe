import copy
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest

SPEC = importlib.util.spec_from_file_location('loglens_report', Path(__file__).with_name('loglens-report.py'))
REPORT = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(REPORT)


def sample(second, presents, allocated):
    timestamp = f'2026-11-09T12:00:{second:02d}+00:00'
    return {
        'elapsed_seconds': second, 'cpu_seconds': second / 2,
        'working_set_bytes': 1000 + second, 'private_bytes': 2000 + second,
        'state': {
            'sampled_at': timestamp, 'generation': second,
            'window_width': 1440, 'scroll_y': second,
            'stats': {**dict.fromkeys(REPORT.COUNTERS, second), 'AllocFrame': 42, 'LastRedraw': 100,
                      'HoverPatchFailures': []},
            'frame': {'presents': presents, 'p99_ns': 20},
            'runtime': {'sampled_at': timestamp, 'total_alloc': allocated,
                        'num_gc': second, 'gc_pause_total_ns': second * 100,
                        'sys': 10000, 'heap_released': 1000},
        },
    }


class ReportTests(unittest.TestCase):
    def test_elapsed_and_counter_deltas(self):
        result = REPORT.summarize([sample(1, 10000, 1000000), sample(11, 10020, 1004000)])
        self.assertEqual(result['elapsed_seconds'], 10)
        self.assertEqual(result['presents_per_second'], 2)
        self.assertEqual(result['allocated_bytes_per_second'], 400)
        self.assertEqual(result['allocated_bytes_per_present'], 200)
        self.assertEqual(result['last_sampled_redraw_alloc_bytes'], 42)
        self.assertEqual(result['runtime_managed_final_bytes'], 9000)

    def test_runtime_uses_its_own_sample_clock(self):
        a, b = sample(1, 10, 1000), sample(11, 20, 2000)
        b['state']['runtime']['sampled_at'] = '2026-11-09T12:00:06+00:00'
        self.assertEqual(REPORT.summarize([a, b])['allocated_bytes_per_second'], 200)

    def test_resets_and_nonfinite_values_fail(self):
        a, b = sample(1, 10, 1000), sample(11, 20, 2000)
        for group, field in [('stats', 'FullLayouts'), ('runtime', 'total_alloc'), ('frame', 'presents')]:
            bad = copy.deepcopy(b)
            bad['state'][group][field] = 0
            with self.subTest(group=group), self.assertRaises(ValueError):
                REPORT.summarize([a, bad])
        b['working_set_bytes'] = float('nan')
        with self.assertRaises(ValueError):
            REPORT.summarize([a, b])

    def test_idle_has_no_allocation_per_present(self):
        result = REPORT.summarize([sample(1, 10, 1000), sample(11, 10, 2000)])
        self.assertIsNone(result['allocated_bytes_per_present'])

    def test_complete_capture_and_missing_action(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            run = root / 'run-1'
            run.mkdir()
            phases = ['idle', 'hover', 'scroll', 'resize', 'navigation', 'idle-after']
            metadata = {'status': 'passed', 'sha256': 'binary-hash', 'devtools': False,
                        'content_width': 1440, 'height': 900, 'seconds_per_phase': 10,
                        'warmup_seconds': 3, 'phases': phases}
            (run / 'capture.json').write_text(json.dumps(metadata))
            for phase in phases:
                directory = run / phase
                directory.mkdir()
                first, last = sample(1, 100, 1000), sample(11, 200, 2000)
                if phase == 'resize':
                    last['state']['window_width'] = 1280
                (directory / 'samples.json').write_text(json.dumps([first, last]))
            result = REPORT.report(root)
            self.assertEqual(result['repetitions'], 1)
            self.assertEqual(result['medians']['hover']['allocated_bytes_per_second'], 100)
            filename = run / 'scroll' / 'samples.json'
            samples = json.loads(filename.read_text())
            samples[-1]['state']['scroll_y'] = samples[0]['state']['scroll_y']
            filename.write_text(json.dumps(samples))
            with self.assertRaisesRegex(ValueError, 'scroll phase did not move'):
                REPORT.report(root)

    def test_failed_capture_is_rejected(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            run = root / 'run-1'
            run.mkdir()
            (run / 'capture.json').write_text(json.dumps({'status': 'failed', 'error': 'lost focus'}))
            with self.assertRaisesRegex(ValueError, 'lost focus'):
                REPORT.report(root)

    def test_phase_failure_counts_and_gpu_memory(self):
        first, last = sample(1, 100, 1000), sample(11, 200, 2000)
        first['state']['stats']['HoverPatchFailures'] = [
            {'stage': 'selector', 'reason': 'target-missing', 'target': 'menu', 'count': 2}]
        last['state']['stats']['HoverPatchFailures'] = [
            {'stage': 'selector', 'reason': 'target-missing', 'target': 'menu', 'count': 5}]
        first['gpu_memory'] = {'available': True, 'counters': [
            {'path': 'gpu-process/dedicated', 'value_bytes': 10}]}
        last['gpu_memory'] = {'available': True, 'counters': [
            {'path': 'gpu-process/dedicated', 'value_bytes': 30}]}
        result = REPORT.summarize([first, last])
        self.assertEqual(result['hover_patch_failure_delta'], {
            'selector/target-missing/menu': 3})
        self.assertEqual(result['gpu_memory']['counters']['gpu-process/dedicated'], {
            'peak_bytes': 30, 'final_bytes': 30})


if __name__ == '__main__':
    unittest.main()
