import { execFileSync } from 'node:child_process';
import { chmodSync, copyFileSync, existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { fileURLToPath } from 'node:url';
import { join } from 'node:path';
import { gzipSync } from 'node:zlib';

const root = fileURLToPath(new URL('../../', import.meta.url));
const version = readFileSync(join(root, 'go.work'), 'utf8').match(/^go (\S+)/m)?.[1];
if (!version) throw new Error('Could not locate the Go version in go.work');
const env = { ...process.env, GOOS: 'js', GOARCH: 'wasm', GOTOOLCHAIN: `go${version}` };
const goroot = execFileSync('go', ['env', 'GOROOT'], { cwd: root, env, encoding: 'utf8' }).trim();
const runtime = ['lib/wasm/wasm_exec.js', 'misc/wasm/wasm_exec.js']
  .map((path) => join(goroot, path)).find(existsSync);
if (!runtime) throw new Error('Could not locate wasm_exec.js in GOROOT');
for (const [example, folder, binary] of [
  ['frontend', 'demo', 'demo.wasm'],
  ['frontend/workspace', 'demos/workspace-inspector', 'demo.wasm'],
  ['frontend/hello', 'demos/hello', 'demo.wasm'],
  ['frontend/budget', 'demos/expense-tracker', 'demo.wasm'],
  ['frontend/builder', 'demos/builder', 'demo.wasm'],
  ['builder/ui-builder', 'demos/ui-builder', 'demo.wasm'],
  ['telegram', 'demos/telegram', 'ownframe.wasm'],
  ['spotify-player', 'demos/spotify-player', 'ownframe.wasm'],
  ['wispr-flow-dashboard', 'demos/wispr-flow', 'ownframe.wasm'],
]) {
  const output = join(root, 'frontend/public', folder);
  mkdirSync(output, { recursive: true });
  const compressed = example === 'spotify-player';
  const scratch = compressed ? mkdtempSync(join(tmpdir(), 'ownframe-wasm-')) : null;
  const built = join(scratch ?? output, binary);
  try {
    execFileSync('go', ['build', '-p', '1', '-trimpath', '-buildvcs=false',
      ...(compressed ? ['-ldflags=-s -w'] : []),
      '-o', built, `./examples/${example}`], { cwd: root, env, stdio: 'inherit' });
    if (compressed) {
      writeFileSync(join(output, `${binary}.gz`), gzipSync(readFileSync(built), { level: 9 }));
      // Remove the raw copy left by earlier versions of this builder.
      rmSync(join(output, binary), { force: true });
    }
  } finally {
    if (scratch) rmSync(scratch, { recursive: true, force: true });
  }
  const destination = join(output, 'wasm_exec.js');
  if (existsSync(destination)) chmodSync(destination, 0o644);
  copyFileSync(runtime, destination);
  chmodSync(destination, 0o644);
}
