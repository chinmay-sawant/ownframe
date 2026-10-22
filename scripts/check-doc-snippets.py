#!/usr/bin/env python3
"""Compile every registered Go fence; additions and exclusions require review."""
import json
from pathlib import Path
import re
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parent.parent


def snippets():
    found = {}
    methods = set()
    testing_methods = set()
    for path in ROOT.glob('*.go'):
        if not path.name.endswith('_test.go'):
            methods.update(re.findall(r'func \(p \*Page\) (\w+)\(', path.read_text()))
            testing_methods.update(re.findall(r'func \(t Testing\) (\w+)\(', path.read_text()))
    for path in [ROOT / 'README.md', *sorted((ROOT / 'documentation').glob('*.md'))]:
        text = path.read_text()
        unknown = set(re.findall(r'\bPage\.([A-Z]\w*)', text)) - methods
        unknown_testing = set(re.findall(r'\bPage\.Testing\(\)\.([A-Z]\w*)', text)) - testing_methods
        if unknown or unknown_testing:
            raise ValueError(f'{path}: unknown Page methods {sorted(unknown)}; unknown Testing methods {sorted(unknown_testing)}')
        for index, match in enumerate(re.finditer(r'^```go[^\n]*\n(.*?)^```', path.read_text(), re.M | re.S), 1):
            key = f'{path.relative_to(ROOT)}#{index}'
            found[key] = (match.group(1), path, path.read_text()[:match.start(1)].count('\n') + 1)
    path = ROOT / 'frontend/src/Home.jsx'
    for name in ['quickStart', 'coreExample']:
        match = re.search(r'(?:export )?const ' + name + r' = `(.*?)`;', path.read_text(), re.S)
        if match is None:
            raise ValueError(f'{path}: missing {name}')
        found[f'frontend/src/Home.jsx#{name}'] = (match.group(1), path, path.read_text()[:match.start(1)].count('\n') + 1)
    return found


def main():
    found = snippets()
    manifest = json.loads((ROOT / 'documentation/snippets.json').read_text())
    if found.keys() != manifest.keys():
        raise ValueError(f'unregistered snippets: {sorted(found.keys() - manifest.keys())}; stale entries: {sorted(manifest.keys() - found.keys())}')
    checked = skipped = 0
    with tempfile.TemporaryDirectory(prefix='.doc-snippets-', dir=ROOT) as scratch:
        packages = []
        for index, (key, (body, path, line)) in enumerate(found.items()):
            spec = manifest[key]
            if spec['mode'] == 'illustration':
                if 'ownframe.' in body:
                    raise ValueError(f'{key}: ownframe API snippets must compile')
                if not spec.get('reason'):
                    raise ValueError(f'{key}: excluded illustration needs a reason')
                print(f'illustration: {key}: {spec["reason"]}', flush=True)
                skipped += 1
                continue
            imports = spec.get('imports', [])
            if any('/internal/' in imp for imp in imports) and spec.get('audience') != 'internal':
                raise ValueError(f'{key}: public snippet imports internal implementation')
            header = 'package main\n' if spec['mode'] == 'file' and body.startswith('package main') else 'package snippet\n'
            if body.startswith('package '):
                source = body
            else:
                source = header
                if imports:
                    source += 'import (\n' + '\n'.join(json.dumps(imp) for imp in imports) + '\n)\n'
                source += spec.get('before', '') + '\n'
                if spec['mode'] in ('body', 'expression'):
                    source += f'func snippet({spec.get("params", "")}) error {{\n'
                source += f'//line {path}:{line}\n'
                if spec['mode'] == 'expression':
                    source += '_ = '
                source += body + '\n'
                source += spec.get('after', '') + '\n'
                if spec['mode'] in ('body', 'expression'):
                    source += '}\n'
            package = Path(scratch) / f'snippet{index:03d}'
            package.mkdir()
            (package / 'snippet.go').write_text(source)
            packages.append('./' + str(package.relative_to(ROOT)))
            checked += 1
        subprocess.run(['go', 'vet', '-p', '1', *packages], cwd=ROOT, check=True)
    print(f'docs-check: {checked} snippets compiled; {skipped} external illustrations listed explicitly')


if __name__ == '__main__':
    try:
        main()
    except subprocess.CalledProcessError as error:
        sys.exit(error.returncode)
    except ValueError as error:
        sys.exit(str(error))
