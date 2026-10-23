import os
from pathlib import Path
import tempfile
import subprocess
import unittest
import zipfile
from unittest.mock import patch

from release_archive import checksums, pack
from sign_release import sign


class ReleaseArchiveTests(unittest.TestCase):
    def test_metadata_changes_do_not_change_archive_bytes(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            stage = root / 'stage'
            stage.mkdir()
            binary = stage / 'app'
            binary.write_bytes(b'executable')
            binary.chmod(0o755)
            for suffix in ('.zip', '.tar.gz'):
                first, second = root / ('first' + suffix), root / ('second' + suffix)
                os.utime(binary, (1000000000, 1000000000))
                binary.chmod(0o755)
                pack(stage, first, ['app'], 1234567890)
                os.utime(binary, (1800000000, 1800000000))
                binary.chmod(0o775)
                pack(stage, second, ['app'], 1234567890)
                self.assertEqual(first.read_bytes(), second.read_bytes())

    def test_signing_repackages_windows_executable_and_refreshes_sum(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            stage = root / 'stage'
            stage.mkdir()
            (stage / 'app.exe').write_bytes(b'unsigned')
            output = root / 'ownframe-app-windows-amd64.zip'
            pack(stage, output, ['app.exe'], 0)
            checksums(root)
            before = (root / 'SHA256SUMS').read_text()
            (root / 'SHA256SUMS.sig').write_text('stale')
            calls = []

            def signer(*args):
                calls.append(args)
                Path(args[-1]).write_bytes(b'signed')

            env = dict.fromkeys(('AZURE_TENANT_ID', 'AZURE_CLIENT_ID', 'AZURE_CLIENT_SECRET', 'AZURE_VAULT_URL', 'AZURE_CERT_NAME'), 'fixture')
            with patch.dict(os.environ, env), patch('sign_release.run', signer):
                self.assertEqual(sign(root, 'windows'), 1)
            self.assertEqual(len(calls), 1)
            self.assertEqual(calls[0][:2], ('AzureSignTool', 'sign'))
            with zipfile.ZipFile(output) as archive:
                self.assertEqual(archive.read('app.exe'), b'signed')
            self.assertNotEqual(before, (root / 'SHA256SUMS').read_text())
            self.assertFalse((root / 'SHA256SUMS.sig').exists())

    def test_macos_staples_app_before_final_archive(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            stage = root / 'stage'
            app = stage / 'App.app'
            app.mkdir(parents=True)
            (app / 'binary').write_bytes(b'app')
            output = root / 'ownframe-app-macos-arm64.zip'
            pack(stage, output, ['App.app'], 0)
            calls = []

            def signer(*args):
                calls.append(args)
                if args[:3] == ('xcrun', 'stapler', 'staple'):
                    (Path(args[-1]) / 'ticket').write_text('stapled')

            env = dict.fromkeys(('APPLE_DEVELOPER_ID', 'NOTARY_KEY_ID', 'NOTARY_ISSUER_ID', 'NOTARY_KEY'), 'fixture')
            with patch.dict(os.environ, env), patch('sign_release.run', signer):
                sign(root, 'macos')
            with zipfile.ZipFile(output) as archive:
                self.assertEqual(archive.read('App.app/ticket'), b'stapled')
            self.assertTrue(any(call[:3] == ('xcrun', 'notarytool', 'submit') for call in calls))
            self.assertTrue(any(call[:3] == ('xcrun', 'stapler', 'validate') for call in calls))

    def test_shell_entry_invokes_real_stub_tool_for_archived_executable(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            stage, tools, dist = root / 'stage', root / 'tools', root / 'dist'
            for directory in (stage, tools, dist):
                directory.mkdir()
            (stage / 'app.exe').write_bytes(b'unsigned')
            archive = dist / 'ownframe-app-windows-amd64.zip'
            pack(stage, archive, ['app.exe'], 0)
            tool = tools / 'AzureSignTool'
            tool.write_text('#!/usr/bin/env python3\nimport os, pathlib, sys\npathlib.Path(os.environ["SIGN_LOG"]).write_text("called")\npathlib.Path(sys.argv[-1]).write_bytes(b"stub-signed")\n')
            tool.chmod(0o755)
            env = os.environ.copy()
            env.update(dict.fromkeys(('AZURE_TENANT_ID', 'AZURE_CLIENT_ID', 'AZURE_CLIENT_SECRET', 'AZURE_VAULT_URL', 'AZURE_CERT_NAME'), 'fixture'))
            env.update(PATH=str(tools) + os.pathsep + env['PATH'], OWNFRAME_DIST_DIR=str(dist), SIGN_LOG=str(root / 'calls'), COSIGN='0')
            script = Path(__file__).resolve().parent / 'sign-release.sh'
            subprocess.run(['sh', str(script), '--platform', 'windows'], env=env, check=True, capture_output=True, text=True)
            self.assertEqual((root / 'calls').read_text(), 'called')
            with zipfile.ZipFile(archive) as signed:
                self.assertEqual(signed.read('app.exe'), b'stub-signed')

    def test_no_signable_input_fails(self):
        with tempfile.TemporaryDirectory() as temp:
            with self.assertRaisesRegex(ValueError, 'no signable'):
                sign(Path(temp), 'all')


if __name__ == '__main__':
    unittest.main()
