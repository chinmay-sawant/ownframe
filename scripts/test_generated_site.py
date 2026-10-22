import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

SPEC = importlib.util.spec_from_file_location('generated_site', Path(__file__).with_name('check-generated-site.py'))
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class GeneratedSiteTests(unittest.TestCase):
    def test_stale_and_unchanged_builds(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'docs').mkdir()
            page = root / 'docs/index.html'
            page.write_text('current')
            with patch.object(MODULE, 'ROOT', root), patch.object(MODULE.subprocess, 'run'):
                MODULE.main()
            with patch.object(MODULE, 'ROOT', root), patch.object(MODULE.subprocess, 'run', side_effect=lambda *a, **k: page.write_text('rebuilt')):
                with self.assertRaisesRegex(ValueError, 'stale'):
                    MODULE.main()
