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
// Pin the exact toolchain from go.work so every machine produces
// identical demo binaries: CI installs this version via setup-go while
// its GOTOOLCHAIN=local would otherwise block the upgrade, and a local
// auto toolchain could drift newer and poison the checked-in output.
let toolchain;
try {
  const match = readFileSync(join(root, 'go.work'), 'utf8').match(/^go (\d+\.\d+(?:\.\d+)?)/m);
  if (match) toolchain = `go${match[1]}`;
} catch {
  // Keep the default toolchain when go.work is unreadable.
}
for (const [example, folder, binary] of [
  ['frontend', 'demo', 'demo.wasm'],
  ['frontend/workspace', 'demos/workspace-inspector', 'demo.wasm'],
  ['frontend/hello', 'demos/hello', 'demo.wasm'],
  ['frontend/budget', 'demos/expense-tracker', 'demo.wasm'],
  ['telegram', 'demos/telegram', 'ownframe.wasm'],
  ['spotify-player', 'demos/spotify-player', 'ownframe.wasm'],
]) {
  const output = join(root, 'frontend/public', folder);
  mkdirSync(output, { recursive: true });
  execFileSync('go', ['build', '-p', '1', '-trimpath', '-buildvcs=false',
    ...(example === 'spotify-player' ? ['-ldflags=-s -w'] : []),
    '-o', join(output, binary), `./examples/${example}`], {
    cwd: root,
    env: { ...process.env, GOOS: 'js', GOARCH: 'wasm', ...(toolchain ? { GOTOOLCHAIN: toolchain } : {}) },
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
