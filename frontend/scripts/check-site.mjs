import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { mkdir, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

const frontend = fileURLToPath(new URL('..', import.meta.url));
const output = fileURLToPath(new URL('../../temp/frontend-check/', import.meta.url));
await mkdir(output, { recursive: true });
const processes = [];
let socket;
const pause = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const errors = [];

function start(command, args) {
  const process = spawn(command, args, { cwd: frontend, stdio: 'ignore' });
  process.on('error', (error) => errors.push(error.message));
  processes.push(process);
}

async function ready(url) {
  for (let attempt = 0; attempt < 100; attempt++) {
    if (errors.length) throw new Error(errors.join('\n'));
    try {
      const response = await fetch(url);
      if (response.ok) return response;
    } catch {}
    await pause(100);
  }
  throw new Error(`Could not connect to ${url}`);
}

try {
  start(process.execPath, ['node_modules/vite/bin/vite.js', 'preview', '--host', '127.0.0.1', '--port', '4175', '--strictPort']);
  await ready('http://127.0.0.1:4175/ownframe/');
  start(process.env.CHROME_PATH || 'google-chrome', ['--headless', '--no-sandbox', '--disable-dev-shm-usage',
    '--remote-debugging-port=9232', `--user-data-dir=${output}chrome`, 'about:blank']);
  const targets = await (await ready('http://127.0.0.1:9232/json')).json();
  socket = new WebSocket(targets.find((target) => target.type === 'page').webSocketDebuggerUrl);
  await new Promise((resolve) => socket.addEventListener('open', resolve, { once: true }));
  let nextId = 0;
  const pending = new Map();
  socket.addEventListener('message', ({ data }) => {
    const message = JSON.parse(data);
    if (message.method === 'Runtime.exceptionThrown') errors.push(message.params.exceptionDetails.text);
    if (message.method === 'Network.responseReceived' && message.params.response.status >= 400
      && message.params.response.url.startsWith('http://127.0.0.1:4175/')) errors.push(message.params.response.url);
    const request = pending.get(message.id);
    if (!request) return;
    pending.delete(message.id);
    if (message.error) request.reject(new Error(JSON.stringify(message.error)));
    else request.resolve(message.result);
  });
  function send(method, params = {}) {
    return new Promise((resolve, reject) => {
      const id = ++nextId;
      pending.set(id, { resolve, reject });
      socket.send(JSON.stringify({ id, method, params }));
    });
  }
  async function evaluate(expression) {
    const result = await send('Runtime.evaluate', { expression, returnByValue: true, awaitPromise: true });
    assert.ok(!result.exceptionDetails, JSON.stringify(result.exceptionDetails));
    return result.result.value;
  }
  async function navigate(hash) {
    await send('Page.navigate', { url: `http://127.0.0.1:4175/ownframe/${hash}` });
    for (let attempt = 0; attempt < 50; attempt++) {
      await pause(100);
      if (await evaluate('!!document.querySelector("main h1")')) return;
    }
    throw new Error(`Page did not render ${hash}`);
  }
  async function screenshot(name, full = false) {
    const clip = full ? (await send('Page.getLayoutMetrics')).cssContentSize : undefined;
    const shot = await send('Page.captureScreenshot', { format: 'png', captureBeyondViewport: full,
      ...(clip ? { clip: { ...clip, scale: 1 } } : {}) });
    await writeFile(`${output}${name}.png`, Buffer.from(shot.data, 'base64'));
  }
  await send('Runtime.enable');
  await send('Network.enable');
  const metrics = [];
  for (const [width, height] of [[320, 740], [390, 844], [768, 1024], [844, 390], [1440, 1000]]) {
    await send('Emulation.setDeviceMetricsOverride', { width, height, deviceScaleFactor: 1, mobile: width < 1000 });
    for (const theme of ['dark', 'light']) {
      for (const hash of ['#home', '#docs/readme', '#docs/binding', '#docs/platforms', '#docs/css-motion', '#docs/showcase/telegram', '#docs/compare-rust-gpui', '#docs/examples-readme', '#docs/compare-electron', '#docs/compare-go-app-frameworks', '#docs/compare-syntax']) {
        await navigate(hash);
        if (hash === '#docs/css-motion') {
          assert.equal(await evaluate(`Array.from(document.querySelectorAll('.docs-nav a')).find(link => link.getAttribute('href') === '#docs/css-motion')?.textContent`), 'Animations and transitions');
          assert.equal(await evaluate('document.querySelector("main h1").textContent'), 'CSS animations and transitions');
        }
        if (await evaluate('document.documentElement.dataset.theme') !== theme) {
          await evaluate('document.querySelector(".theme-toggle").click()');
        }
        const layout = await evaluate(`({width: innerWidth, pageWidth: document.documentElement.scrollWidth,
          mainWidth: document.querySelector('main').getBoundingClientRect().width,
          headerWidth: document.querySelector('header').getBoundingClientRect().width,
          scale: Number(getComputedStyle(document.documentElement).getPropertyValue('--site-scale')),
          theme: document.documentElement.dataset.theme, heading: document.querySelector('main h1').textContent})`);
        assert.ok(layout.pageWidth <= width, `Horizontal overflow: ${JSON.stringify({ width, hash, ...layout })}`);
        assert.equal(layout.theme, theme);
        assert.ok(Math.abs(layout.mainWidth - Math.min(width, 1440 * layout.scale)) < 2, 'Page content width differs between routes');
        assert.ok(Math.abs(layout.headerWidth - layout.mainWidth) < 2, 'Header width differs from page content');
        metrics.push({ width, height, hash, ...layout });
        if (hash.startsWith('#docs')) {
          assert.ok(await evaluate(`Array.from(document.querySelectorAll('.markdown-table-scroll')).every(wrapper => {
            const table = wrapper.querySelector('table');
            return Math.abs(wrapper.getBoundingClientRect().width - wrapper.closest('article').getBoundingClientRect().width) < 2
              && table.getBoundingClientRect().width >= wrapper.getBoundingClientRect().width - 2;
          })`), 'A documentation table does not fill the available article width');
        }
        if (hash === '#docs/examples-readme' && [390, 1440].includes(width)) {
          await evaluate('document.querySelector("#feature-examples").scrollIntoView()');
          assert.ok(await evaluate('Math.abs(document.querySelector("header").getBoundingClientRect().top) < 1'), 'Header disappears above the example table');
          await screenshot(`example-tables-${width}-${theme}`);
        }
        if (hash === '#home' && [390, 1440].includes(width)) {
          await evaluate(`Promise.all(Array.from(document.querySelectorAll('main img')).map(image => {
            image.loading = 'eager';
            return image.decode();
          }))`);
          await screenshot(`home-${width}-${theme}`);
          if (width === 1440) await screenshot(`home-full-${theme}`, true);
        }
        if (hash === '#home') {
          assert.ok(await evaluate(`['start-section', 'comparison-section', 'mobile-section'].every(className => {
            const section = document.querySelector('.' + className);
            return Array.from(section.querySelectorAll(':scope > p')).every(paragraph =>
              Math.abs(paragraph.getBoundingClientRect().width - section.getBoundingClientRect().width) < 2);
          })`), 'Setup, comparison, or mobile paragraphs do not use the full content width');
        }
        if (hash === '#docs/compare-rust-gpui') {
          assert.ok(await evaluate('!!document.querySelector(".language-rust .hljs-string")'), 'Rust highlighting missing');
          assert.ok(await evaluate('!!document.querySelector(".language-go .hljs-string")'), 'Guide Go highlighting missing');
        }
        if (hash.startsWith('#docs/compare-')) {
          assert.ok(await evaluate(`Array.from(document.querySelectorAll('.docs-content pre code')).every(code =>
            code.textContent.includes('Hello world'))`), 'A comparison code example shows different content');
        }
        if (hash === '#docs/readme' && [390, 1440].includes(width)) await screenshot(`docs-${width}-${theme}`);
      }
    }
  }
  await send('Emulation.setDeviceMetricsOverride', { width: 390, height: 844, deviceScaleFactor: 1, mobile: true });
  await navigate('#docs/readme');
  await evaluate('document.querySelector(".docs-menu-toggle").click()');
  assert.equal(await evaluate('document.querySelector(".docs-menu-toggle").getAttribute("aria-expanded")'), 'true');
  await evaluate(`(() => {
    let parent = document.querySelector('.docs-nav a[href="#docs/binding"]').parentElement;
    while (parent) {
      if (parent.tagName === 'DETAILS') parent.open = true;
      parent = parent.parentElement;
    }
  })()`);
  await evaluate(`document.querySelector('.docs-nav a[href="#docs/binding"]').click()`);
  await pause(200);
  assert.equal(await evaluate('location.hash'), '#docs/binding');
  assert.equal(await evaluate('document.querySelector(".docs-menu-toggle").getAttribute("aria-expanded")'), 'false');
  await evaluate('location.hash = "#docs/binding/example"');
  await pause(200);
  assert.ok(await evaluate('Math.abs(document.querySelector("#example").getBoundingClientRect().top - document.querySelector("header").getBoundingClientRect().height - 16 * Number(getComputedStyle(document.documentElement).getPropertyValue("--site-scale"))) < 4'), 'Mobile anchor did not clear the sticky header');
  await screenshot('mobile-docs');
  await navigate('#docs/showcase/pet-care');
  assert.equal(await evaluate('document.querySelector(".docs-content h2").textContent'), 'Pet Care');
  await navigate('#docs/unknown-document');
  assert.equal(await evaluate('document.querySelector("main h1").textContent'), 'Document not found');
  await navigate('#home');
  await evaluate(`Promise.all(Array.from(document.querySelectorAll('main img')).map(image => {
    image.loading = 'eager';
    return image.decode();
  }))`);
  assert.ok(await evaluate('Array.from(document.querySelectorAll("main img")).every(image => image.naturalWidth > 0)'), 'Broken homepage image');
  assert.equal(await evaluate('getComputedStyle(document.querySelector(".github")).backgroundColor'), 'rgba(0, 0, 0, 0)', 'GitHub background is still visible');
  assert.equal(await evaluate('getComputedStyle(document.querySelector(".github-count")).borderLeftWidth'), '0px', 'GitHub count divider is still visible');
  assert.equal(await evaluate('document.querySelector(".home > section:nth-child(2)").id'), 'demos');
  assert.ok(await evaluate('document.querySelector(".hero-mascot").naturalWidth > 0'), 'Gopher missing');
  assert.ok(await evaluate('!!document.querySelector(".core-example .language-go .hljs-keyword")'), 'Go highlighting missing');
  assert.equal(await evaluate('getComputedStyle(document.documentElement).scrollbarWidth'), 'none');
  await evaluate('window.scrollTo(0, 0)');
  await send('Input.dispatchMouseEvent', { type: 'mouseWheel', x: 190, y: 600, deltaY: 400, deltaX: 0 });
  await pause(200);
  assert.ok(await evaluate('scrollY > 0'), 'Page cannot scroll with hidden scrollbar');
  assert.ok(await evaluate('Math.abs(document.querySelector("header").getBoundingClientRect().top) < 1'), 'Header does not stay visible while scrolling');
  await evaluate(`document.querySelector('header a[href="#start"]').click()`);
  await pause(200);
  assert.ok(await evaluate('Math.abs(document.querySelector("#start").getBoundingClientRect().top - document.querySelector("header").getBoundingClientRect().height - 24 * Number(getComputedStyle(document.documentElement).getPropertyValue("--site-scale"))) < 4'), 'Get started link did not clear the sticky header');
  await evaluate(`document.querySelector('header a[href="#docs/readme"]').click()`);
  await pause(200);
  assert.equal(await evaluate('location.hash'), '#docs/readme');
  assert.equal(await evaluate('getComputedStyle(document.querySelector(".docs-content")).overflowY'), 'visible');
  await evaluate('document.querySelector(".skip").click()');
  assert.equal(await evaluate('location.hash'), '#docs/readme');
  assert.equal(await evaluate('document.activeElement.id'), 'main');
  await evaluate('document.querySelector(".docs-menu-toggle").click()');
  await evaluate('document.querySelector(".docs-menu-toggle").focus()');
  await send('Input.dispatchKeyEvent', { type: 'keyDown', key: 'Escape', code: 'Escape', windowsVirtualKeyCode: 27 });
  assert.equal(await evaluate('document.querySelector(".docs-menu-toggle").getAttribute("aria-expanded")'), 'false');
  assert.deepEqual(errors, [], 'Browser errors');
  await writeFile(`${output}metrics.json`, JSON.stringify(metrics, null, 2));
  console.log(`PASS: ${metrics.length} layout checks, themes, mobile menu, anchors, showcase, keyboard focus, and images.`);
} finally {
  socket?.close();
  for (const process of processes) process.kill();
}
