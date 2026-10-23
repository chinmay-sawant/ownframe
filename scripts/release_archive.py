"""Write release archives with fixed ordering, timestamps, owners, and modes."""
import argparse
import gzip
import hashlib
import os
from pathlib import Path
import stat
import tarfile
import time
import zipfile


def pack(root, output, names, epoch):
    root, output = Path(root), Path(output)
    paths = set()
    for name in names:
        path = root / name
        if path.is_symlink() or not path.exists() or not path.resolve().is_relative_to(root.resolve()):
            raise ValueError(f'unsafe or missing release entry: {name}')
        paths.add(path)
        if path.is_dir():
            paths.update(path.rglob('*'))
    ordered = sorted(paths, key=lambda path: path.relative_to(root).as_posix())
    for path in ordered:
        if path.is_symlink() or not (path.is_file() or path.is_dir()):
            raise ValueError(f'unsupported release entry: {path}')
    if output.name.endswith('.tar.gz'):
        with output.open('wb') as raw, gzip.GzipFile(filename='', mode='wb', fileobj=raw, mtime=epoch) as compressed:
            with tarfile.open(fileobj=compressed, mode='w', format=tarfile.USTAR_FORMAT) as archive:
                for path in ordered:
                    info = archive.gettarinfo(str(path), path.relative_to(root).as_posix())
                    info.uid = info.gid = 0
                    info.uname = info.gname = ''
                    info.mtime = epoch
                    info.mode = 0o755 if path.is_dir() or path.stat().st_mode & 0o111 else 0o644
                    if path.is_file():
                        with path.open('rb') as stream:
                            archive.addfile(info, stream)
                    else:
                        archive.addfile(info)
    elif output.suffix == '.zip':
        stamp = time.gmtime(max(epoch, 315532800))[:6]
        with zipfile.ZipFile(output, 'w', compression=zipfile.ZIP_DEFLATED, compresslevel=9) as archive:
            for path in ordered:
                directory = path.is_dir()
                name = path.relative_to(root).as_posix() + ('/' if directory else '')
                info = zipfile.ZipInfo(name, stamp)
                info.create_system = 3
                mode = 0o755 if directory or path.stat().st_mode & 0o111 else 0o644
                info.external_attr = ((stat.S_IFDIR if directory else stat.S_IFREG) | mode) << 16
                info.compress_type = zipfile.ZIP_DEFLATED
                archive.writestr(info, b'' if directory else path.read_bytes())
    else:
        raise ValueError(f'unsupported release archive: {output}')


def checksums(directory):
    directory = Path(directory)
    archives = sorted([*directory.glob('*.zip'), *directory.glob('*.tar.gz')])
    if not archives:
        raise ValueError('no release archives')
    (directory / 'SHA256SUMS').write_text(''.join(
        f'{hashlib.sha256(path.read_bytes()).hexdigest()}  {path.name}\n' for path in archives))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest='command', required=True)
    create = sub.add_parser('pack')
    create.add_argument('root', type=Path)
    create.add_argument('output', type=Path)
    create.add_argument('names', nargs='+')
    sums = sub.add_parser('checksums')
    sums.add_argument('directory', type=Path)
    args = parser.parse_args()
    if args.command == 'pack':
        pack(args.root, args.output, args.names, int(os.environ.get('SOURCE_DATE_EPOCH', '0')))
    else:
        checksums(args.directory)


if __name__ == '__main__':
    main()
