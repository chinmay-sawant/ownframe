#!/usr/bin/env python3
"""Verify packaged archives and checksums without extracting executables."""
import argparse
import hashlib
from pathlib import Path, PurePosixPath
import re
import sys
import tarfile
import zipfile


def safe_name(name):
    parts = PurePosixPath(name).parts
    if not parts or name.startswith('/') or '\\' in name or '..' in parts or ':' in name:
        raise ValueError(f'unsafe archive path: {name}')


def verify(directory, required_platforms=()):
    checksums = directory / 'SHA256SUMS'
    rows = checksums.read_text().splitlines()
    seen = set()
    if not rows:
        raise ValueError('empty checksum file')
    for row in rows:
        match = re.fullmatch(r'([a-f0-9]{64})  ([^/\\]+)', row)
        if not match:
            raise ValueError(f'malformed checksum row: {row}')
        digest, name = match.groups()
        safe_name(name)
        if name in seen:
            raise ValueError(f'duplicate checksum: {name}')
        seen.add(name)
        path = directory / name
        if path.is_symlink() or hashlib.sha256(path.read_bytes()).hexdigest() != digest:
            raise ValueError(f'checksum mismatch: {name}')
        if name.endswith('.zip'):
            with zipfile.ZipFile(path) as archive:
                entries = archive.infolist()
                if len({entry.filename for entry in entries}) != len(entries):
                    raise ValueError(f'duplicate ZIP entries: {name}')
                for entry in entries:
                    safe_name(entry.filename)
                    if (entry.external_attr >> 16) & 0o170000 == 0o120000:
                        raise ValueError(f'ZIP symlink: {entry.filename}')
                bad = archive.testzip()
                if bad:
                    raise ValueError(f'corrupt ZIP member: {bad}')
                files = {entry.filename for entry in entries if not entry.is_dir()}
                if name.endswith('-wasm.zip'):
                    if files != {'ownframe.wasm', 'wasm_exec.js', 'index.html'}:
                        raise ValueError(f'incomplete wasm archive: {files}')
                    if not archive.read('ownframe.wasm').startswith(b'\x00asm'):
                        raise ValueError('wasm archive has no wasm executable')
                elif not any(file.endswith('.exe') or '.app/Contents/MacOS/' in file for file in files):
                    raise ValueError(f'no desktop executable: {name}')
        elif name.endswith('.tar.gz'):
            with tarfile.open(path, 'r:gz') as archive:
                entries = archive.getmembers()
                files = [entry for entry in entries if entry.isfile()]
                for entry in entries:
                    safe_name(entry.name)
                    if not entry.isfile() and not entry.isdir():
                        raise ValueError(f'unsupported tar member: {entry.name}')
                names = {entry.name for entry in files}
                if len(names) != len(files) or not any(n.endswith('.desktop') for n in names) or not any(n.endswith('/README.md') for n in names):
                    raise ValueError(f'incomplete Linux archive: {name}')
                if not any(entry.mode & 0o111 and entry.size > 0 for entry in files):
                    raise ValueError(f'no executable in Linux archive: {name}')
        else:
            raise ValueError(f'unsupported archive: {name}')
    actual = {path.name for path in directory.iterdir() if path.name.endswith(('.zip', '.tar.gz'))}
    if actual != seen:
        raise ValueError(f'archives and checksums disagree: {actual ^ seen}')
    if not any(name.endswith('-wasm.zip') for name in seen) or not any(not name.endswith('-wasm.zip') for name in seen):
        raise ValueError('bundle needs a desktop archive and a wasm preview')
    for platform in required_platforms:
        if not any(f'-{platform}-' in name for name in seen):
            raise ValueError(f'missing required platform archive: {platform}')
    return len(seen)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('directory', type=Path)
    parser.add_argument('--require-platform', action='append', choices=('linux', 'windows', 'macos'), default=[])
    args = parser.parse_args()
    print(f'release-check: {verify(args.directory, args.require_platform)} archives verified; no release published')


if __name__ == '__main__':
    try:
        main()
    except (ValueError, OSError, tarfile.TarError, zipfile.BadZipFile) as error:
        sys.exit(str(error))
