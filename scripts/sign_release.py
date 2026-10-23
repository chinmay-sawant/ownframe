"""Sign executable members, notarize macOS apps, and replace release archives."""
import argparse
import os
from pathlib import Path, PurePosixPath
import shutil
import sys
import subprocess
import tempfile
import zipfile

from release_archive import checksums, pack


def require(*names):
    for name in names:
        if not os.environ.get(name):
            raise ValueError(f'sign-release: required credential unset: {name}')


def run(*args):
    result = subprocess.run(args, check=False)
    if result.returncode:
        # Azure arguments contain credentials. Never include argv in an error.
        raise ValueError(f'{args[0]} failed with exit status {result.returncode}')


def unpack(archive, directory):
    with zipfile.ZipFile(archive) as source:
        for item in source.infolist():
            name = PurePosixPath(item.filename)
            mode = item.external_attr >> 16
            if name.is_absolute() or '..' in name.parts or '\\' in item.filename or ':' in item.filename:
                raise ValueError(f'unsafe signing archive member: {item.filename}')
            if mode & 0o170000 == 0o120000:
                raise ValueError(f'signing archive contains a symlink: {item.filename}')
            target = directory / name
            if item.is_dir():
                target.mkdir(parents=True, exist_ok=True)
            else:
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_bytes(source.read(item))
                target.chmod(mode & 0o777 or 0o644)


def sign_archive(archive, platform):
    with tempfile.TemporaryDirectory(prefix='ownframe-sign-') as temp:
        stage = Path(temp)
        unpack(archive, stage)
        if platform == 'macos':
            require('APPLE_DEVELOPER_ID', 'NOTARY_KEY_ID', 'NOTARY_ISSUER_ID', 'NOTARY_KEY')
            apps = sorted(stage.glob('*.app'))
            if not apps:
                raise ValueError(f'no app in {archive.name}')
            for app in apps:
                run('codesign', '--deep', '--force', '--options', 'runtime', '--sign', os.environ['APPLE_DEVELOPER_ID'], str(app))
                run('codesign', '--verify', '--deep', '--strict', str(app))
            submission = stage / 'submission.zip'
            pack(stage, submission, [app.name for app in apps], 0)
            run('xcrun', 'notarytool', 'submit', str(submission), '--key', os.environ['NOTARY_KEY'], '--key-id', os.environ['NOTARY_KEY_ID'], '--issuer', os.environ['NOTARY_ISSUER_ID'], '--wait')
            submission.unlink()
            for app in apps:
                run('xcrun', 'stapler', 'staple', str(app))
                run('xcrun', 'stapler', 'validate', str(app))
        else:
            require('AZURE_TENANT_ID', 'AZURE_CLIENT_ID', 'AZURE_CLIENT_SECRET', 'AZURE_VAULT_URL', 'AZURE_CERT_NAME')
            exes = sorted(stage.glob('*.exe'))
            if not exes:
                raise ValueError(f'no executable in {archive.name}')
            for exe in exes:
                run('AzureSignTool', 'sign', '-kvu', os.environ['AZURE_VAULT_URL'], '-kvi', os.environ['AZURE_CLIENT_ID'], '-kvt', os.environ['AZURE_TENANT_ID'], '-kvs', os.environ['AZURE_CLIENT_SECRET'], '-kvc', os.environ['AZURE_CERT_NAME'], '-tr', 'http://timestamp.digicert.com', '-td', 'sha256', str(exe))
        replacement = stage / 'signed.zip'
        names = [entry.name for entry in sorted(stage.iterdir())]
        pack(stage, replacement, names, int(os.environ.get('SOURCE_DATE_EPOCH', '0')))
        pending = archive.with_suffix('.signed-tmp')
        try:
            shutil.copyfile(replacement, pending)
            pending.replace(archive)
        finally:
            pending.unlink(missing_ok=True)


def sign(directory, selected):
    candidates = []
    for platform in ('macos', 'windows'):
        if selected in ('all', platform):
            candidates.extend((archive, platform) for archive in sorted(directory.glob(f'*-{platform}-*.zip')))
    if not candidates:
        raise ValueError('sign-release: no signable archives for selected platform')
    # Clear stale authentication before modifying archive bytes. A failed run
    # must not leave a checksum signature that appears to cover signed output.
    for suffix in ('.sig', '.pem'):
        (directory / ('SHA256SUMS' + suffix)).unlink(missing_ok=True)
    try:
        for archive, platform in candidates:
            sign_archive(archive, platform)
    finally:
        checksums(directory)
    return len(candidates)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--directory', type=Path, required=True)
    parser.add_argument('--platform', choices=('all', 'macos', 'windows'), default='all')
    args = parser.parse_args()
    print(f'sign-release: {sign(args.directory, args.platform)} archives signed; checksums refreshed')


if __name__ == '__main__':
    try:
        main()
    except (ValueError, OSError) as error:
        sys.exit(str(error))
