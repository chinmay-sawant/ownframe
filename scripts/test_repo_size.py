import runpy
from pathlib import Path
import unittest

MOD = runpy.run_path(str(Path(__file__).with_name('check-repo-size.py')))
check_sizes = MOD['check_sizes']
MAX_BYTES = MOD['MAX_BYTES']


class RepoSizeCheckTests(unittest.TestCase):
    def test_accepts_allowlisted_large_files_and_small_files(self):
        sizes = {
            'small.go': 1024,
            'docs/demo.wasm': MAX_BYTES + 1,
        }
        self.assertEqual(check_sizes(sizes, {'docs/demo.wasm'}), 1)

    def test_rejects_new_oversized_file(self):
        sizes = {'new-blob.bin': MAX_BYTES + 1024}
        with self.assertRaisesRegex(ValueError, 'new-blob.bin'):
            check_sizes(sizes, set())

    def test_rejects_stale_allowlist_entry(self):
        sizes = {'docs/demo.wasm': 1024}
        with self.assertRaisesRegex(ValueError, 'stale allowlist entry'):
            check_sizes(sizes, {'docs/demo.wasm'})


if __name__ == '__main__':
    unittest.main()
