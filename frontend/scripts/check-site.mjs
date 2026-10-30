import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { mkdir, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

const frontend = fileURLToPath(new URL('..', import.meta.url));
const output = fileURLToPath(new URL('../../temp/frontend-check/', import.meta.url));
await mkdir(output, { recursive: true });
const captureDemo = process.argv.includes('--capture-demo');
const processes = [];
const shutdown = { stopping: false };
let socket;
const pause = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const errors = [];

function start(command, args, capture = false) {
  const child = spawn(command, args, { cwd: frontend, stdio: capture ? ['ignore', 'pipe', 'pipe'] : 'ignore' });
  const entry = { process: child, output: '' };
  if (capture) {
    const keep = (chunk) => {
      entry.output = (entry.output + chunk.toString()).slice(-4000);
    };
    child.stdout.on('data', keep);
    child.stderr.on('data', keep);
  }
  child.on('error', (error) => errors.push(`${command}: ${error.message}`));
  child.on('exit', (code, signal) => {
    if (shutdown.stopping || code === 0) return;
    if (code === null) errors.push(`${command} stopped by signal ${signal}${entry.output ? `: ${entry.output}` : ''}`);
    else errors.push(`${command} exited with code ${code}${entry.output ? `: ${entry.output}` : ''}`);
  });
  processes.push(child);
  return entry;
}

async function ready(url, attempts = 100) {
  for (let attempt = 0; attempt < attempts; attempt++) {
    if (errors.length) throw new Error(errors.join('\n'));
    try {
      const response = await fetch(url);
      if (response.ok) return response;
    } catch {
      // The server is not up yet, so the loop keeps trying.
    }
    await pause(100);
  }
  throw new Error(`Could not connect to ${url}`);
}

try {
  start(process.execPath, ['node_modules/vite/bin/vite.js', ...(captureDemo ? [] : ['preview']), '--host', '127.0.0.1', '--port', '4175', '--strictPort'], true);
  await ready('http://127.0.0.1:4175/ownframe/');
  const chrome = start(process.env.CHROME_PATH || 'google-chrome', ['--headless', '--no-sandbox', '--disable-dev-shm-usage',
    '--no-first-run', '--no-default-browser-check', '--enable-unsafe-swiftshader',
    '--remote-debugging-port=9232', `--user-data-dir=${output}chrome`, 'about:blank'], true);
  let targets;
  try {
    targets = await (await ready('http://127.0.0.1:9232/json', 300)).json();
  } catch (error) {
    throw new Error(`Chrome did not open remote debugging: ${error.message}${chrome.output ? `\nchrome output: ${chrome.output}` : ''}`, { cause: error });
  }
  socket = new WebSocket(targets.find((target) => target.type === 'page').webSocketDebuggerUrl);
  await new Promise((resolve) => socket.addEventListener('open', resolve, { once: true }));
  let nextId = 0;
  const pending = new Map();
  socket.addEventListener('message', ({ data }) => {
    const message = JSON.parse(data);
    if (message.method === 'Runtime.exceptionThrown') errors.push(message.params.exceptionDetails.text);
    if (message.method === 'Runtime.consoleAPICalled' && message.params.type === 'error') {
      errors.push(message.params.args.map(argument => argument.description || argument.value || '').join(' '));
    }
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
  if (captureDemo) {
    await send('Emulation.setDeviceMetricsOverride', { width: 560, height: 640, deviceScaleFactor: 1, mobile: false });
    await send('Page.navigate', { url: 'http://127.0.0.1:4175/ownframe/demo/index.html' });
    let loaded = false;
    for (let attempt = 0; attempt < 300; attempt++) {
      await pause(100);
      loaded = await evaluate('!!document.querySelector("canvas") && document.querySelector("#loader")?.classList.contains("loaded")');
      if (loaded) break;
    }
    assert.ok(loaded, 'The standalone WASM demo did not load');
    await pause(1500);
    async function click(x, y) {
      await send('Input.dispatchMouseEvent', { type: 'mouseMoved', x, y });
      await send('Input.dispatchMouseEvent', { type: 'mousePressed', x, y, button: 'left', clickCount: 1 });
      await send('Input.dispatchMouseEvent', { type: 'mouseReleased', x, y, button: 'left', clickCount: 1 });
      await send('Input.dispatchMouseEvent', { type: 'mouseMoved', x: 550, y: 550 });
      await pause(400);
    }
    const dark = (await send('Page.captureScreenshot', { format: 'png' })).data;
    await click(480, 50);
    await click(280, 110);
    const light = (await send('Page.captureScreenshot', { format: 'png' })).data;
    assert.notEqual(light, dark, 'The WASM theme control did not change the rendered app');
    await click(480, 50);
    await click(280, 110);
    await click(240, 417);
    const completed = (await send('Page.captureScreenshot', { format: 'png' })).data;
    assert.notEqual(completed, dark, 'Completing a task did not change the rendered app');
    await screenshot('demo-task-completed');
    await click(215, 305);
    const filtered = (await send('Page.captureScreenshot', { format: 'png' })).data;
    assert.notEqual(filtered, completed, 'Filtering tasks did not change the rendered app');
    await screenshot('demo-filtered');
    const demoOutput = fileURLToPath(new URL('../public/demo/', import.meta.url));
    await writeFile(`${demoOutput}preview-dark.png`, Buffer.from(dark, 'base64'));
    await writeFile(`${demoOutput}preview-light.png`, Buffer.from(light, 'base64'));
    await send('Emulation.setDeviceMetricsOverride', { width: 780, height: 520, deviceScaleFactor: 1, mobile: false });
    await evaluate(`(async () => {
      const image = new Image(780, 520);
      image.src = '/ownframe/demos/videos/spotify.webp';
      image.style.cssText = 'position:fixed;inset:0;width:780px;height:520px;z-index:100';
      await image.decode();
      document.body.append(image);
    })()`);
    await pause(100);
    const poster = await send('Page.captureScreenshot', { format: 'png' });
    await writeFile(fileURLToPath(new URL('../public/demos/spotify-still.png', import.meta.url)), Buffer.from(poster.data, 'base64'));
    assert.deepEqual(errors, [], 'Browser errors in the WASM demo');
    console.log('PASS: standalone WASM loading, theme, completion, and filter rendering. Saved public demo previews.');
  } else {
  for (const demo of ['telegram', 'spotify-player']) {
    await send('Emulation.setDeviceMetricsOverride', { width: 1440, height: 1000, deviceScaleFactor: 1, mobile: false });
    await send('Page.navigate', { url: `http://127.0.0.1:4175/ownframe/demos/${demo}/index.html` });
    let loaded = false;
    for (let attempt = 0; attempt < 600; attempt++) {
      await pause(100);
      loaded = await evaluate('!!document.querySelector("canvas") && !document.querySelector("[role=status]")');
      if (loaded) break;
      if (await evaluate('document.querySelector("[role=status]")?.textContent.includes("could not load")')) break;
    }
    if (!loaded) await screenshot(`linked-demo-${demo}-failure`);
    assert.ok(loaded, `${demo} WASM did not start: ${errors.join('; ')}`);
    await pause(1000);
    await screenshot(`linked-demo-${demo}`);
  }
  if (process.argv.includes('--linked-demos-only')) {
    assert.deepEqual(errors, [], 'Linked demo browser errors');
    console.log('PASS: Telegram and Spotify Player WASM demos loaded.');
  } else {
  const metrics = [];
  const viewports = process.argv.includes('--interactions-only') ? [] : [[320, 740], [390, 844], [768, 1024], [844, 390], [1440, 1000]];
  for (const [width, height] of viewports) {
    await send('Emulation.setDeviceMetricsOverride', { width, height, deviceScaleFactor: 1, mobile: width < 1000 });
    for (const theme of ['dark', 'light']) {
      for (const hash of ['#home', '#engine', '#compare', '#docs/showcase/wispr-flow', '#docs/showcase/telegram', '#docs/showcase/pet-care', '#docs/examples-readme', '#docs/compare-rust-gpui', '#docs/compare-electron', '#docs/compare-wails', '#docs/compare-fyne', '#docs/compare-gio', '#docs/compare-webview', '#docs/compare-ebitengine']) {
        await navigate(hash);
        if (hash === '#docs/showcase/wispr-flow') {
          assert.equal(await evaluate(`Array.from(document.querySelectorAll('.docs-nav a')).find(link => link.getAttribute('href') === '#docs/showcase/wispr-flow')?.textContent`), 'Example apps');
          assert.equal(await evaluate('document.querySelector("main h1").textContent'), 'Example apps');
        }
        if (await evaluate('document.documentElement.dataset.theme') !== theme) {
          await evaluate('document.querySelector(".theme-toggle").click()');
          await pause(180);
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
        if (hash === '#engine' || hash === '#compare') {
          assert.equal(await evaluate('document.querySelectorAll("main section").length'), 1, `${hash} displays another section`);
          assert.ok(await evaluate(`!!document.querySelector('${hash}')`), `Missing ${hash} page`);
          assert.equal(await evaluate(`document.querySelector('header a[href="${hash}"]').getAttribute('aria-current')`), 'page');
        }
        if (hash === '#home' && [390, 1440].includes(width)) {
          let loaded = false;
          for (let attempt = 0; attempt < 300; attempt++) {
            await pause(100);
            loaded = await evaluate(`(() => {
              const doc = document.querySelector('.launch-demo').contentDocument;
              return !!doc?.querySelector('canvas') && doc.querySelector('#loader')?.classList.contains('loaded');
            })()`);
            if (loaded) break;
          }
          assert.ok(loaded, `Embedded WASM did not load at ${width}px`);
          await pause(500);
          assert.ok(await evaluate(`(() => {
            const frame = document.querySelector('.launch-demo');
            return Math.abs(frame.getBoundingClientRect().width - frame.parentElement.getBoundingClientRect().width) < 2;
          })()`), 'Playable demo does not fill its pane');
          await evaluate(`Promise.all(Array.from(document.querySelectorAll('main img')).map(image => {
            image.loading = 'eager';
            return image.decode();
          }))`);
          await screenshot(`home-${width}-${theme}`);
          if (width === 1440) await screenshot(`home-full-${theme}`, true);
        }
        if (hash === '#home') {
          assert.equal(await evaluate('!!document.querySelector("#engine, #compare")'), false, 'Home includes a separate page');
          assert.ok(await evaluate('!!document.querySelector("#demos") && !!document.querySelector("#start")'), 'Home content is missing');
          assert.deepEqual(await evaluate(`Array.from(document.querySelectorAll('.hello-example [role="tab"]')).map(tab => tab.textContent.trim().replace(/^(Go|<>|#|▶)\\s*/, ''))`), ['Live demo', 'main.go', 'template.html', 'style.css']);
          for (const index of [1, 2, 3, 0]) {
            await evaluate(`document.querySelectorAll('.hello-example [role="tab"]')[${index}].click()`);
            assert.equal(await evaluate(`document.querySelector('.hello-example [role="tab"][aria-selected="true"]').getAttribute('aria-controls')`), await evaluate(`document.querySelector('.hello-example [role="tabpanel"]:not([hidden])').id`));
          }
          assert.equal(await evaluate(`new Set(Array.from(document.querySelectorAll('[id]')).map(element => element.id)).size === document.querySelectorAll('[id]').length`), true, 'Duplicate IDs between code examples');
          assert.equal(await evaluate('document.querySelectorAll("iframe.launch-demo").length'), 1, 'Homepage is missing the playable WASM demo');
          assert.ok(await evaluate(`Boolean(document.querySelector('.preview-caption a[href$="demo/index.html"]'))`), 'Missing public WASM link');
          assert.equal(await evaluate('getComputedStyle(document.querySelector(".demo-card-phone .demo-image")).objectFit'), 'contain');
          const navigation = await evaluate(`({
            bottom: document.querySelector('header nav').getBoundingClientRect().bottom,
            headerBottom: document.querySelector('header').getBoundingClientRect().bottom,
            visible: getComputedStyle(document.querySelector('header nav')).display !== 'none'
          })`);
          assert.ok(!navigation.visible || navigation.bottom <= navigation.headerBottom + 1, 'Navigation escapes its header');
          if (width <= 1024) {
            assert.ok(await evaluate('document.querySelector(".preview-window-pane").getBoundingClientRect().top < document.querySelector(".source-disclosure").getBoundingClientRect().top'), 'Mobile source appears before the demo');
            assert.equal(await evaluate('document.querySelector(".source-disclosure").open'), false, 'Mobile source should start collapsed');
          }
        }
        if (hash === '#docs/compare-rust-gpui') {
          assert.ok(await evaluate('!!document.querySelector(".language-rust .hljs-string")'), 'Rust highlighting missing');
          assert.ok(await evaluate('!!document.querySelector(".language-go .hljs-string")'), 'Guide Go highlighting missing');
        }
        if (hash.startsWith('#docs/compare-')) {
          assert.ok(await evaluate(`Array.from(document.querySelectorAll('.docs-content pre code')).every(code =>
            code.textContent.includes('Hello world'))`), 'A comparison code example shows different content');
        }
        if (hash === '#docs/showcase/wispr-flow' && [390, 1440].includes(width)) await screenshot(`docs-${width}-${theme}`);
      }
    }
  }
  await send('Emulation.setDeviceMetricsOverride', { width: 390, height: 844, deviceScaleFactor: 1, mobile: true });
  for (const width of [1440, 390]) {
    await send('Emulation.setDeviceMetricsOverride', { width, height: 900, deviceScaleFactor: 1, mobile: width < 1000 });
    await navigate('#home');
    await send('Page.reload');
    for (let attempt = 0; attempt < 300; attempt++) {
      await pause(100);
      if (await evaluate(`document.querySelector('.launch-demo')?.contentDocument?.querySelector('#loader')?.classList.contains('loaded')`)) break;
    }
    await evaluate('document.querySelector("#live-demo").scrollIntoView()');
    await pause(500);
    const frame = await evaluate(`document.querySelector('.launch-demo').getBoundingClientRect().toJSON()`);
    const before = (await send('Page.captureScreenshot', { format: 'png' })).data;
    async function clickDemo(x, y) {
      const point = { x: frame.x + x, y: frame.y + y };
      await send('Input.dispatchMouseEvent', { type: 'mousePressed', ...point, button: 'left', clickCount: 1 });
      await send('Input.dispatchMouseEvent', { type: 'mouseReleased', ...point, button: 'left', clickCount: 1 });
      await pause(300);
    }
    await clickDemo(160, 417);
    const completed = (await send('Page.captureScreenshot', { format: 'png' })).data;
    assert.notEqual(completed, before, 'Embedded task click did not change the rendered page');
    await screenshot(`home-demo-completed-${width}`);
    await clickDemo(width === 390 ? 192 : 215, 305);
    const filtered = (await send('Page.captureScreenshot', { format: 'png' })).data;
    assert.notEqual(filtered, completed, 'Embedded filter click did not change the rendered page');
    await screenshot(`home-demo-filtered-${width}`);
    for (const selector of ['.hello-example', '.split-example']) {
      await evaluate(`document.querySelectorAll('.starter-picker button')[${selector === '.hello-example' ? 0 : 1}].click()`);
      await evaluate(`document.querySelector('${selector}').scrollIntoView()`);
      let loaded = false;
      for (let attempt = 0; attempt < 300; attempt++) {
        await pause(100);
        loaded = await evaluate(`(() => {
          const doc = document.querySelector('${selector} iframe')?.contentDocument;
          return !!doc?.querySelector('canvas') && doc.querySelector('#loader')?.classList.contains('loaded');
        })()`);
        if (loaded) break;
      }
      assert.ok(loaded, `${selector} WASM did not load`);
      await screenshot(`${selector.slice(1)}-${width}`);
    }
    await evaluate(`document.querySelectorAll('.starter-picker button')[0].click()`);
    for (const index of [1, 2, 3]) {
      await evaluate(`document.querySelectorAll('.hello-example [role=tab]')[${index}].click()`);
      await screenshot(`starter-source-${index}-${width}`);
    }

  }
  await navigate('#home');
  await send('Page.reload');
  for (let attempt = 0; attempt < 50; attempt++) {
    await pause(100);
    if (await evaluate('!!document.querySelector(".hero-install button")')) break;
  }
  await evaluate(`navigator.clipboard.writeText = async (text) => { window.copiedCommand = text; };
    document.querySelector('.hero-install button').click()`);
  await pause(50);
  assert.equal(await evaluate('window.copiedCommand'), 'go get github.com/chinmay-sawant/ownframe@v0.0.2');
  assert.equal(await evaluate('document.querySelector(".hero-install button").textContent'), 'Copied');
  await evaluate(`navigator.clipboard.writeText = async () => { throw new Error('Clipboard unavailable'); };
    document.querySelector('.hero-install button').click()`);
  await pause(50);
  assert.equal(await evaluate('document.querySelector(".hero-install button").textContent'), 'Select to copy');
  await evaluate('document.querySelector(".site-menu-toggle").click()');
  assert.equal(await evaluate('getComputedStyle(document.querySelector("header nav")).display'), 'grid');
  assert.ok(await evaluate('document.querySelector("header nav").getBoundingClientRect().bottom <= document.querySelector("header").getBoundingClientRect().bottom'));
  await evaluate('document.querySelector(".site-menu-toggle").focus()');
  await send('Input.dispatchKeyEvent', { type: 'keyDown', key: 'Escape', code: 'Escape', windowsVirtualKeyCode: 27 });
  assert.equal(await evaluate('document.querySelector(".site-menu-toggle").getAttribute("aria-expanded")'), 'false');
  assert.equal(await evaluate('document.activeElement.className'), 'site-menu-toggle');
  await evaluate('document.querySelector(".source-disclosure summary").click()');
  await pause(50);
  assert.equal(await evaluate('document.querySelector(".source-disclosure").open'), true);
  await evaluate('document.querySelector(".example-file-tabs [role=tab]").focus()');
  await send('Input.dispatchKeyEvent', { type: 'keyDown', key: 'ArrowRight', code: 'ArrowRight', windowsVirtualKeyCode: 39 });
  assert.equal(await evaluate('document.activeElement.getAttribute("aria-selected")'), 'true');
  await navigate('#docs/showcase/wispr-flow');
  await evaluate('document.querySelector(".docs-menu-toggle").click()');
  assert.equal(await evaluate('document.querySelector(".docs-menu-toggle").getAttribute("aria-expanded")'), 'true');
  await evaluate(`(() => {
    let parent = document.querySelector('.docs-nav a[href="#docs/examples-readme"]').parentElement;
    while (parent) {
      if (parent.tagName === 'DETAILS') parent.open = true;
      parent = parent.parentElement;
    }
  })()`);
  await evaluate(`document.querySelector('.docs-nav a[href="#docs/examples-readme"]').click()`);
  await pause(200);
  assert.equal(await evaluate('location.hash'), '#docs/examples-readme');
  assert.equal(await evaluate('document.querySelector(".docs-menu-toggle").getAttribute("aria-expanded")'), 'false');
  await evaluate('location.hash = "#docs/examples-readme/feature-examples"');
  await pause(200);
  assert.ok(await evaluate('Math.abs(document.querySelector("#feature-examples").getBoundingClientRect().top - document.querySelector("header").getBoundingClientRect().height - 16 * Number(getComputedStyle(document.documentElement).getPropertyValue("--site-scale"))) < 4'), 'Mobile anchor did not clear the sticky header');
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
  await navigate('#team');
  assert.equal(await evaluate('document.querySelector("#team-title").textContent'), 'Built by a developer, with AI tools.');
  assert.ok(await evaluate('document.querySelector(".maintainer-card").textContent.includes("Chinmay Sawant")'));
  assert.equal(await evaluate('document.querySelectorAll(".team-card").length'), 5);
  await evaluate(`Promise.all(Array.from(document.querySelectorAll('.team-card img')).map(image => {
    image.loading = 'eager';
    return image.decode();
  }))`);
  assert.ok(await evaluate('Array.from(document.querySelectorAll(".team-card img")).every(image => image.naturalWidth > 0)'), 'Broken team image');
  await evaluate('document.querySelector(".maintainer-avatar").decode()');
  assert.equal(await evaluate('document.querySelector(".maintainer-links a").href'), 'https://x.com/chinmay_sawant_');
  for (const width of [390, 1440]) {
    await send('Emulation.setDeviceMetricsOverride', { width, height: 1000, deviceScaleFactor: 1, mobile: width < 1000 });
    await screenshot(`team-x-${width}`);
    assert.ok(await evaluate('document.documentElement.scrollWidth <= innerWidth'), 'Maintainer card overflows');
  }
  await send('Emulation.setDeviceMetricsOverride', { width: 390, height: 844, deviceScaleFactor: 1, mobile: true });
  assert.equal(await evaluate('document.querySelectorAll(".home > section").length'), 1, 'Team page shows more than the team');
  await navigate('#home');
  assert.ok(await evaluate('document.querySelector(".hero-mascot").naturalWidth > 0'), 'Gopher missing');
  assert.ok(await evaluate('!!document.querySelector(".core-example .language-go .hljs-keyword")'), 'Go highlighting missing');
  assert.equal(await evaluate('getComputedStyle(document.documentElement).scrollbarWidth'), 'none');
  await evaluate('window.scrollTo(0, 0)');
  const wheelPoint = await evaluate(`(() => {
    const header = document.querySelector('header').getBoundingClientRect();
    return { x: Math.round(header.x + header.width / 2), y: Math.round(header.y + header.height / 2) };
  })()`);
  assert.ok(await evaluate(`(() => { const el = document.elementFromPoint(${wheelPoint.x}, ${wheelPoint.y}); return !!el && !el.closest('iframe'); })()`), 'Wheel test point landed inside an iframe');
  await send('Input.dispatchMouseEvent', { type: 'mouseWheel', x: wheelPoint.x, y: wheelPoint.y, deltaY: 400, deltaX: 0 });
  let scrolled = false;
  for (let attempt = 0; attempt < 20 && !scrolled; attempt++) {
    await pause(100);
    scrolled = await evaluate('scrollY > 0');
  }
  assert.ok(scrolled, 'Page cannot scroll with hidden scrollbar');
  assert.ok(await evaluate('Math.abs(document.querySelector("header").getBoundingClientRect().top) < 1'), 'Header does not stay visible while scrolling');
  await evaluate(`document.querySelector('.site-menu-toggle').click(); document.querySelector('header a[href="#start"]').click()`);
  await pause(200);
  assert.ok(await evaluate('Math.abs(document.querySelector("#start").getBoundingClientRect().top - document.querySelector("header").getBoundingClientRect().height - 24 * Number(getComputedStyle(document.documentElement).getPropertyValue("--site-scale"))) < 4'), 'Get started link did not clear the sticky header');
  await evaluate(`document.querySelector('header a[href="#docs/showcase/wispr-flow"]').click()`);
  await pause(200);
  assert.equal(await evaluate('location.hash'), '#docs/showcase/wispr-flow');
  assert.equal(await evaluate('getComputedStyle(document.querySelector(".docs-content")).overflowY'), 'visible');
  await evaluate('document.querySelector(".skip").click()');
  assert.equal(await evaluate('location.hash'), '#docs/showcase/wispr-flow');
  assert.equal(await evaluate('document.activeElement.id'), 'main');
  await evaluate('document.querySelector(".docs-menu-toggle").click()');
  await evaluate('document.querySelector(".docs-menu-toggle").focus()');
  await send('Input.dispatchKeyEvent', { type: 'keyDown', key: 'Escape', code: 'Escape', windowsVirtualKeyCode: 27 });
  assert.equal(await evaluate('document.querySelector(".docs-menu-toggle").getAttribute("aria-expanded")'), 'false');
  assert.deepEqual(errors, [], 'Browser errors');
  await writeFile(`${output}metrics.json`, JSON.stringify(metrics, null, 2));
  console.log(`PASS: ${metrics.length} layout checks, themes, mobile menu, anchors, showcase, keyboard focus, and images.`);
  }
  }
} finally {
  socket?.close();
  shutdown.stopping = true;
  for (const process of processes) process.kill();
}
