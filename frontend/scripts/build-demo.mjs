import { execFileSync } from 'node:child_process';
import { chmodSync, copyFileSync, existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { join } from 'node:path';
import { gzipSync } from 'node:zlib';

const root = fileURLToPath(new URL('../../', import.meta.url));
const goroot = execFileSync('go', ['env', 'GOROOT'], { cwd: root, encoding: 'utf8' }).trim();
const runtime = ['lib/wasm/wasm_exec.js', 'misc/wasm/wasm_exec.js']
  .map((path) => join(goroot, path)).find(existsSync);
if (!runtime) throw new Error('Could not locate wasm_exec.js in GOROOT');
for (const [example, folder, binary] of [
  ['frontend', 'demo', 'demo.wasm'],
  ['frontend/workspace', 'demos/workspace-inspector', 'demo.wasm'],
  ['telegram', 'demos/telegram', 'ownframe.wasm'],
  ['spotify-player', 'demos/spotify-player', 'ownframe.wasm'],
]) {
  const output = join(root, 'frontend/public', folder);
  mkdirSync(output, { recursive: true });
  execFileSync('go', ['build', '-p', '1',
    ...(example === 'spotify-player' ? ['-ldflags=-s -w'] : []),
    '-o', join(output, binary), `./examples/${example}`], {
    cwd: root,
    env: { ...process.env, GOOS: 'js', GOARCH: 'wasm' },
    stdio: 'inherit',
  });
  if (example === 'spotify-player') {
    writeFileSync(join(output, `${binary}.gz`), gzipSync(readFileSync(join(output, binary)), { level: 9 }));
  }
  const destination = join(output, 'wasm_exec.js');
  if (existsSync(destination)) chmodSync(destination, 0o644);
  copyFileSync(runtime, destination);
  chmodSync(destination, 0o644);
}
