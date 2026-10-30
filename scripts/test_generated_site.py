import importlib.util
from pathlib import Path
import tempfile
import types
import unittest
from unittest.mock import patch

SPEC = importlib.util.spec_from_file_location('generated_site', Path(__file__).with_name('check-generated-site.py'))
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


def _listing(argv, tracked='docs/index.html\0'):
    if argv[0] == 'npm':
        return types.SimpleNamespace(stdout='')
    if '--others' in argv:
        return types.SimpleNamespace(stdout='')
    return types.SimpleNamespace(stdout=tracked)


class GeneratedSiteTests(unittest.TestCase):
    def test_stale_and_unchanged_builds(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'docs').mkdir()
            page = root / 'docs/index.html'
            page.write_text('current')
            with patch.object(MODULE, 'ROOT', root), \
                    patch.object(MODULE.subprocess, 'run',
                                 side_effect=lambda *a, **k: _listing(a[0])):
                MODULE.main()
            def rebuilding(*args, **kwargs):
                if args[0][0] == 'npm':
                    page.write_text('rebuilt')
                    return types.SimpleNamespace(stdout='')
                return _listing(args[0])
            with patch.object(MODULE, 'ROOT', root), \
                    patch.object(MODULE.subprocess, 'run', side_effect=rebuilding):
                with self.assertRaisesRegex(ValueError, 'stale'):
                    MODULE.main()

    def test_ignored_files_are_not_compared(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'docs').mkdir()
            page = root / 'docs/index.html'
            page.write_text('current')
            ignored = root / 'docs/big.wasm'
            ignored.write_bytes(b'v1')
            def rebuilding(*args, **kwargs):
                if args[0][0] == 'npm':
                    ignored.write_bytes(b'v2')
                    return types.SimpleNamespace(stdout='')
                return _listing(args[0])
            with patch.object(MODULE, 'ROOT', root), \
                    patch.object(MODULE.subprocess, 'run', side_effect=rebuilding):
                MODULE.main()
