import runpy
from pathlib import Path
import tempfile
import unittest

MOD = runpy.run_path(str(Path(__file__).with_name('check-ui-generated.py')))
compare_generated = MOD['compare_generated']
check_ui_generated = MOD['check_ui_generated']


class UIGeneratedCheckTests(unittest.TestCase):
    def test_accepts_matching_generated_content(self):
        content = b'// Code generated\npackage ui\n'
        self.assertEqual(compare_generated(content, content), len(content))

    def test_rejects_drifted_generated_content(self):
        committed = b'// Code generated\npackage ui\nvar old = 1\n'
        generated = b'// Code generated\npackage ui\nvar updated = 2\n'
        with self.assertRaisesRegex(ValueError, r'go generate \./ui'):
            compare_generated(committed, generated)

    def test_check_ui_generated_without_modifying_target(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            root = Path(tmpdir)
            target = root / 'ui' / 'properties_generated.go'
            target.parent.mkdir(parents=True)
            committed = b'package ui\n// committed\n'
            target.write_bytes(committed)

            def fake_match(out_path, root=None):
                out_path.write_bytes(committed)

            self.assertEqual(check_ui_generated(root=root, generator=fake_match), len(committed))

            def fake_drift(out_path, root=None):
                out_path.write_bytes(b'package ui\n// drifted\n')

            with self.assertRaisesRegex(ValueError, r'out of date'):
                check_ui_generated(root=root, generator=fake_drift)
            self.assertEqual(target.read_bytes(), committed)


if __name__ == '__main__':
    unittest.main()
