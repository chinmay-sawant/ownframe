// Compare the native sticky fixture with Chrome's viewport geometry and hits.
// No npm dependencies; Node's WebSocket speaks the Chrome DevTools protocol.
import assert from 'node:assert/strict';
import { spawn, execFileSync } from 'node:child_process';
import { mkdir, mkdtemp, readFile, writeFile, rm } from 'node:fs/promises';
import { createHash } from 'node:crypto';
import { resolve } from 'node:path';
import { tmpdir } from 'node:os';
import { setTimeout as pause } from 'node:timers/promises';

const root = resolve(import.meta.dirname, '..');
const output = resolve(process.argv[2] || `${root}/temp/v004-performance/browser-reference`);
await mkdir(output, { recursive: true });
const profile = await mkdtemp(`${tmpdir()}/ownframe-sticky-chrome-`);
const fixturePath = `${root}/internal/window/testdata/sticky.html`;
const fixture = await readFile(fixturePath, 'utf8');
const chrome = spawn(process.env.CHROME_PATH || 'google-chrome', [
  '--headless', '--no-sandbox', '--disable-dev-shm-usage', '--hide-scrollbars',
  '--no-first-run', '--no-default-browser-check', '--remote-debugging-port=0',
  `--user-data-dir=${profile}`, 'about:blank',
], { stdio: 'ignore' });
let socket;
try {
  let port;
  for (let attempt = 0; attempt < 100; attempt++) {
    try { port = (await readFile(`${profile}/DevToolsActivePort`, 'utf8')).split('\n')[0]; break; }
    catch { await pause(100); }
  }
  assert.ok(port, 'Chrome did not publish its debugging port');
  const targets = await (await fetch(`http://127.0.0.1:${port}/json`)).json();
  socket = new WebSocket(targets.find(target => target.type === 'page').webSocketDebuggerUrl);
  await new Promise((resolve, reject) => {
    socket.addEventListener('open', resolve, { once: true });
    socket.addEventListener('error', reject, { once: true });
  });
  let sequence = 0;
  const pending = new Map();
  socket.addEventListener('message', ({ data }) => {
    const reply = JSON.parse(data);
    const call = pending.get(reply.id);
    if (!call) return;
    pending.delete(reply.id);
    clearTimeout(call.timer);
    if (reply.error) call.reject(new Error(JSON.stringify(reply.error)));
    else call.resolve(reply.result);
  });
  function send(method, params = {}) {
    return new Promise((resolve, reject) => {
      const id = ++sequence;
      const timer = setTimeout(() => { pending.delete(id); reject(new Error(`CDP timeout: ${method}`)); }, 10000);
      pending.set(id, { resolve, reject, timer });
      socket.send(JSON.stringify({ id, method, params }));
    });
  }
  async function evaluate(expression) {
    const reply = await send('Runtime.evaluate', { expression, returnByValue: true, awaitPromise: true });
    assert.ok(!reply.exceptionDetails, JSON.stringify(reply.exceptionDetails));
    return reply.result.value;
  }
  await send('Emulation.setDeviceMetricsOverride', { width: 200, height: 100, deviceScaleFactor: 1, mobile: false });
  await send('Page.navigate', { url: `data:text/html,${encodeURIComponent(fixture.replace('{{.}}', 'red'))}` });
  for (let attempt = 0; attempt < 100; attempt++) {
    if (await evaluate('!!document.getElementById("head")')) break;
    await pause(50);
  }
  const results = [];
  for (const [step, scroll, expectedY] of [[0,0,80],[1,100,0],[2,200,0],[3,360,-20],[4,400,-60],[5,200,0],[6,200,0],[9,0,80]]) {
    if (step === 6) await evaluate('document.getElementById("head").style.background="yellow"');
    await evaluate(`scrollTo(0,${scroll});new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve)))`);
    const actual = await evaluate(`({scrollY, width:innerWidth,height:innerHeight,y:document.getElementById('head').getBoundingClientRect().y,headHit:document.elementFromPoint(10,10)?.id,overlapHit:document.elementFromPoint(175,10)?.id})`);
    assert.equal(actual.y, expectedY, `sticky Y at scroll ${scroll}`);
    assert.equal(actual.scrollY, scroll);
    assert.equal(actual.width, 200);
    assert.equal(actual.height, 100);
    if (step === 2) { assert.equal(actual.headHit, 'head'); assert.equal(actual.overlapHit, 'cover'); }
    const shot = await send('Page.captureScreenshot', { format: 'png', captureBeyondViewport: false });
    await writeFile(`${output}/chrome-sticky-${step}.png`, Buffer.from(shot.data, 'base64'));
    results.push({ step, scroll, expectedY, actual });
  }
  // Independent height-query boundary check, including reverse resize.
  await evaluate(`document.body.innerHTML='<style>#media{height:30px}@media(min-height:500px){#media{height:90px}}@media(max-height:400px){#media{height:60px}}</style><div id="media"></div>'`);
  const heights = [];
  for (const [height, expected] of [[499,30],[500,90],[501,90],[499,30],[400,60],[399,60],[401,30]]) {
    await send('Emulation.setDeviceMetricsOverride', { width: 800, height, deviceScaleFactor: 1, mobile: false });
    const actual = await evaluate('document.getElementById("media").getBoundingClientRect().height');
    assert.equal(actual, expected);
    heights.push({ viewportHeight: height, expected, actual });
  }
  const report = { status: 'passed', sourceCommit: execFileSync('git', ['rev-parse','HEAD'], { cwd: root, encoding: 'utf8' }).trim(),
    fixture: 'internal/window/testdata/sticky.html', fixtureSHA256: createHash('sha256').update(fixture).digest('hex'),
    browser: await send('Browser.getVersion'), viewport: [200,100], sticky: results, heightQueries: heights,
    scope: 'Chrome reference for unzoomed native steps; native zoom has its own framebuffer and hit assertions' };
  await writeFile(`${output}/reference.json`, `${JSON.stringify(report, null, 2)}\n`);
  console.log(`PASS Chrome sticky geometry, overlap hit targets, return, and height boundaries: ${output}`);
} finally {
  socket?.close();
  chrome.kill('SIGTERM');
  await new Promise(resolve => { if (chrome.exitCode !== null) resolve(); else chrome.once('exit', resolve); });
  await rm(profile, { recursive: true, force: true });
}
