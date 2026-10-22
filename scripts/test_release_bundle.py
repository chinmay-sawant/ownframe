import hashlib
import importlib.util
import io
from pathlib import Path
import tarfile
import tempfile
import unittest
import zipfile

SPEC = importlib.util.spec_from_file_location('release_bundle', Path(__file__).with_name('check-release-bundle.py'))
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class ReleaseBundleTests(unittest.TestCase):
    def make_bundle(self, root):
        with tarfile.open(root / 'app-linux.tar.gz', 'w:gz') as archive:
            for name, content, mode in [('app/app', b'executable', 0o755), ('app/app.desktop', b'desktop', 0o644), ('app/README.md', b'docs', 0o644)]:
                info = tarfile.TarInfo(name)
                info.mode, info.size = mode, len(content)
                archive.addfile(info, io.BytesIO(content))
        with zipfile.ZipFile(root / 'app-wasm.zip', 'w') as archive:
            for name, content in {'ownframe.wasm': b'\x00asmdata', 'wasm_exec.js': b'loader', 'index.html': b'page'}.items():
                archive.writestr(name, content)
        self.checksums(root)

    def checksums(self, root):
        (root / 'SHA256SUMS').write_text(''.join(hashlib.sha256(path.read_bytes()).hexdigest() + '  ' + path.name + '\n' for path in sorted(root.iterdir()) if path.name.endswith(('.zip', '.tar.gz'))))

    def test_valid_bundle_and_corrupt_digest(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.make_bundle(root)
            self.assertEqual(MODULE.verify(root), 2)
            (root / 'app-wasm.zip').write_bytes(b'changed')
            with self.assertRaisesRegex(ValueError, 'checksum mismatch'):
                MODULE.verify(root)

    def test_unsafe_paths_and_missing_members(self):
        for members in ({'../escape': b'x'}, {'ownframe.wasm': b'not wasm', 'wasm_exec.js': b'x', 'index.html': b'x'}, {'ownframe.wasm': b'\x00asm'}):
            with self.subTest(members=members), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                self.make_bundle(root)
                with zipfile.ZipFile(root / 'app-wasm.zip', 'w') as archive:
                    for name, data in members.items():
                        archive.writestr(name, data)
                self.checksums(root)
                with self.assertRaises(ValueError):
                    MODULE.verify(root)

    def test_extra_archive_and_duplicate_checksum_fail(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.make_bundle(root)
            checksums = root / 'SHA256SUMS'
            original = checksums.read_text()
            checksums.write_text(original + original.splitlines()[0] + '\n')
            with self.assertRaisesRegex(ValueError, 'duplicate checksum'):
                MODULE.verify(root)
            checksums.write_text(original)
            (root / 'unlisted.zip').write_bytes(b'x')
            with self.assertRaisesRegex(ValueError, 'disagree'):
                MODULE.verify(root)
