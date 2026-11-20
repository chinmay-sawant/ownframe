import gzip
import importlib.util
from pathlib import Path
import tempfile
import unittest

SPEC = importlib.util.spec_from_file_location('generated_site', Path(__file__).with_name('check-generated-site.py'))
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class GeneratedSiteTests(unittest.TestCase):
    def setUp(self):
        self.site = Path(self.enterContext(tempfile.TemporaryDirectory()))
        (self.site / 'index.html').touch()
        for folder, binary in MODULE.DEMOS.items():
            directory = self.site / folder
            directory.mkdir(parents=True)
            (directory / 'index.html').touch()
            (directory / 'wasm_exec.js').touch()
            payload = MODULE.WASM_HEADER
            if binary.endswith('.gz'):
                payload = gzip.compress(payload)
            (directory / binary).write_bytes(payload)

    def test_complete_artifact(self):
        MODULE.check_site(self.site)

    def test_missing_wispr_flow(self):
        (self.site / 'demos/wispr-flow/ownframe.wasm').unlink()
        with self.assertRaisesRegex(ValueError, 'Missing WebAssembly.*wispr-flow'):
            MODULE.check_site(self.site)

    def test_uncompressed_spotify_copy(self):
        (self.site / 'demos/spotify-player/ownframe.wasm').touch()
        with self.assertRaisesRegex(ValueError, 'Unexpected WebAssembly.*spotify-player'):
            MODULE.check_site(self.site)
