import { execFileSync } from 'node:child_process';
import { chmodSync, copyFileSync, existsSync, mkdirSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { join } from 'node:path';

const root = fileURLToPath(new URL('../../', import.meta.url));
const output = join(root, 'frontend/public/demo');
mkdirSync(output, { recursive: true });
execFileSync('go', ['build', '-p', '1', '-o', join(output, 'demo.wasm'), './examples/frontend'], {
  cwd: root,
  env: { ...process.env, GOOS: 'js', GOARCH: 'wasm' },
  stdio: 'inherit',
});
const goroot = execFileSync('go', ['env', 'GOROOT'], { cwd: root, encoding: 'utf8' }).trim();
const runtime = ['lib/wasm/wasm_exec.js', 'misc/wasm/wasm_exec.js']
  .map((path) => join(goroot, path)).find(existsSync);
if (!runtime) throw new Error('Could not locate wasm_exec.js in GOROOT');
const destination = join(output, 'wasm_exec.js');
if (existsSync(destination)) chmodSync(destination, 0o644);
copyFileSync(runtime, destination);
chmodSync(destination, 0o644);
