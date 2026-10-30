import React, { useEffect, useRef, useState } from 'react';
import ExampleCode from './ExampleCode';
import { starterFiles } from './starter';
import { repository } from './project';
import gopher from '../../assets/brand/gopher.webp';
import telegramPreview from '../../assets/screenshots/android/telegram/telegram-thread.webp';
import wisprPreview from '../../assets/screenshots/demos/wispr-flow.webp';

const showcases = [
  {
    title: 'Code Review',
    tag: 'Desktop Tooling',
    description: 'Browse a workspace, inspect commit history, and compare word-level diffs in a native Go window.',
    image: `${import.meta.env.BASE_URL}demos/code-view.png`,
    alt: 'Code Review application showing git commit graph, file tree, and syntax-highlighted editor.',
    href: `${repository}/tree/master/examples/code-view`,
    link: 'View the Go example',
  },
  {
    title: 'Telegram on Android',
    tag: 'Mobile Android',
    description: 'A phone-first chat demo with a pinned composer, Android keyboard support, and light and dark themes.',
    image: telegramPreview,
    alt: 'Telegram demo conversation running on an Android Pixel device.',
    phone: true,
    href: '#docs/showcase/telegram',
    link: 'Build the Android demo',
    liveHref: `${import.meta.env.BASE_URL}demos/telegram/index.html`,
    liveLink: 'Try the live demo (51 MB)',
  },
  {
    title: 'Spotify Player',
    tag: 'Media Application',
    description: 'Play bundled audio, browse your library, and watch CSS animations respond to playback.',
    image: `${import.meta.env.BASE_URL}demos/spotify-still.png`,
    alt: 'Spotify Player desktop demo with a sidebar and music library.',
    href: `${repository}/tree/master/examples/spotify-player`,
    link: 'View the Go example',
    liveHref: `${import.meta.env.BASE_URL}demos/spotify-player/index.html`,
    liveLink: 'Try the live demo (includes bundled audio)',
  },
  {
    title: 'Wispr Flow Dashboard',
    tag: 'Productivity Shell',
    description: 'A desktop productivity dashboard with activity charts, a streak calendar, and 13 connected screens.',
    image: wisprPreview,
    alt: 'Wispr Flow dashboard showing speech typing statistics and activity calendar.',
    href: '#docs/showcase/wispr-flow',
    link: 'Explore showcase',
    liveHref: `${import.meta.env.BASE_URL}demos/wispr-flow/index.html`,
    liveLink: 'Try the live demo (52 MB)',
  },

];

export default function Home() {
  const [copyStatus, setCopyStatus] = useState('Copy');
  const [sourceOpen, setSourceOpen] = useState(() => !window.matchMedia('(max-width: 1024px)').matches);
  const copyTimer = useRef(null);
  const demoUrl = `${import.meta.env.BASE_URL}demo/index.html`;

  useEffect(() => () => clearTimeout(copyTimer.current), []);

  async function copyCommand() {
    clearTimeout(copyTimer.current);
    try {
      await navigator.clipboard.writeText('go get github.com/chinmay-sawant/ownframe@v0.0.2');
      setCopyStatus('Copied');
    } catch {
      setCopyStatus('Select to copy');
    }
    copyTimer.current = setTimeout(() => setCopyStatus('Copy'), 2500);
  }

  return <div className="home">
    <section className="hero" aria-labelledby="hero-title">
      <div className="hero-copy">
        <p className="eyebrow">Go logic. HTML and CSS interfaces.</p>
        <h1 id="hero-title">Build native Go apps<br className="hero-break" /> with HTML and CSS.</h1>
        <p>Create desktop and Android interfaces with Go handlers and an in-process renderer. No embedded browser.</p>
        <div className="hero-actions links">
          <a href="#start" className="button-primary">Get started <span aria-hidden="true">→</span></a>
          <a href="#live-demo" className="button-secondary">Try the demo <span aria-hidden="true">↓</span></a>
        </div>
        <div className="hero-install">
          <code>go get github.com/chinmay-sawant/ownframe@v0.0.2</code>
          <button type="button" onClick={copyCommand} aria-label="Copy install command">{copyStatus}</button>
          <span className="visually-hidden" role="status">{copyStatus === 'Copied' ? 'Install command copied.' : copyStatus === 'Select to copy' ? 'Clipboard unavailable. Select the command to copy it.' : ''}</span>
        </div>
        <p className="hero-platforms">Open source · Linux, macOS, Windows, and Android</p>
      </div>
      <img className="hero-mascot" src={gopher} width="224" height="224"
        alt="The blue ownframe mascot waving hello" />

      <div id="live-demo" className="hero-preview-card">
        <details className="source-disclosure" open={sourceOpen}
          onToggle={(event) => setSourceOpen(event.currentTarget.open)}>
          <summary>Explore the Go, HTML, and CSS source</summary>
          <ExampleCode />
        </details>
        <div className="preview-pane preview-window-pane">
          <div className="preview-header preview-header-app">
            <span className="app-window-title">Launch board</span>
            <span className="preview-label">Interactive Go · WebAssembly demo</span>
          </div>
          <iframe className="launch-demo" src={demoUrl}
            title="Interactive Launch board: complete tasks, filter the list, and switch themes"
            width="560" height="640" />
          <div className="preview-caption">
            <span>Click a task. Try the filters and theme switch.</span>
            <a href={demoUrl}>Open demo in its own page <span aria-hidden="true">↗</span></a>
          </div>
        </div>
      </div>
      <p className="preview-note">This live preview loads about 53 MB of WebAssembly. Desktop and Android apps run in native windows.</p>
    </section>


    <section id="demos" className="demo-section" aria-labelledby="demos-title">
      <h2 id="demos-title">See what you can build</h2>
      <p>Explore desktop tools, phone apps, and small experiments. Every example includes its Go source.</p>
      <div className="demo-grid">
        {showcases.map((demo) => <article
          className={`demo-card${demo.phone ? ' demo-card-phone' : ''}${demo.title === 'Code Review' ? ' demo-card-featured' : ''}`} key={demo.title}>
          <div className={`demo-media${demo.phone ? ' demo-media-phone' : ''}`}>
            <img className="demo-image" src={demo.image} alt={demo.alt} width={demo.phone ? 540 : 1500} height={demo.phone ? 1095 : 900} loading="lazy" decoding="async" />
          </div>
          <div className="demo-card-content">
            <div className="card-top-row">
              <span className="card-category-tag">{demo.tag}</span>
            </div>
            <h3>{demo.title}</h3>
            <p>{demo.description}</p>
            <a href={demo.href}>{demo.link}</a>
            {demo.liveHref && <>
              {' '}
              <a href={demo.liveHref}>{demo.liveLink}</a>
            </>}
          </div>
        </article>)}
      </div>
      <p className="section-note"><a href="#docs/showcase/code-view">Browse all examples in documentation</a>. Product-inspired showcases are unofficial technical demonstrations.</p>
    </section>

    {/* Architecture Section */}
    <section id="engine" className="engine-section" aria-labelledby="engine-title">
      <h2 id="engine-title">Your interface, inside your Go process</h2>
      <p>Write the interface in HTML and CSS. Keep state and event handlers in Go. The layout engine and renderer run in the same application process.</p>
      <div className="engine-flow" aria-label="How an ownframe screen reaches the window">
        <div className="engine-step">
          <div className="step-num">01</div>
          <h3>Describe your screen</h3>
          <p>Use Go HTML templates and the supported CSS subset. Pass your application state to the template with SetData.</p>
        </div>
        <div className="engine-step">
          <div className="step-num">02</div>
          <h3>Lay it out in Go</h3>
          <p>Blinkless parses the template, applies CSS, and calculates layout. It returns a display list for ownframe to render.</p>
        </div>
        <div className="engine-step">
          <div className="step-num">03</div>
          <h3>Open a native window</h3>
          <p>Ownframe connects rendering, input, and Go handlers through Ebitengine. Dirty regions and retained layers reduce repeated painting.</p>
        </div>
      </div>
      <p className="section-note">Ownframe supports a subset of HTML and CSS. Website JavaScript does not run inside your app.</p>
    </section>

    <section id="compare" className="comparison-section" aria-labelledby="compare-title">
      <p className="eyebrow">Measured, with context</p>
      <h2 id="compare-title">Small updates should stay small.</h2>
      <p>Retained operations let an animation update its paint without rebuilding the whole document. These historical measurements cover specific fixtures, not complete GPU presentation or every application.</p>
      <div className="metrics-grid">
        <div className="metric-card">
          <div className="metric-number">176 ns</div>
          <h3 className="metric-title">Animated card update</h3>
          <p className="metric-detail">The recorded animated-card frame operation, after replacing a full redraw.</p>
        </div>
        <div className="metric-card">
          <div className="metric-number">0 allocs</div>
          <h3 className="metric-title">For that card update</h3>
          <p className="metric-detail">Heap allocations per operation in the same animated-card measurement.</p>
        </div>
        <div className="metric-card">
          <div className="metric-number">143 µs</div>
          <h3 className="metric-title">Small-fixture warm redraw</h3>
          <p className="metric-detail">The recorded RedrawWarm stage benchmark from the v0.0.2 baseline.</p>
        </div>
      </div>
      <p className="section-note">Historical v0.0.2 results on the author&apos;s machine. Workload, commit, environment, and limits are recorded in the <a href={`${repository}/blob/master/documentation/performance-baseline.md`}>performance baseline</a>. See the <a href={`${repository}/blob/master/documentation/performance.md`}>measurement procedure</a> to reproduce the checks.</p>
      <div className="comparison-scroll" role="region" aria-label="Runtime architecture comparison" tabIndex="0">
        <table className="comparison-table">
          <caption className="visually-hidden">Runtime architecture of ownframe and browser-based approaches</caption>
          <thead><tr><th scope="col">Inside the app</th><th scope="col">Electron</th><th scope="col">Wails v2 / Tauri</th><th scope="col">ownframe</th></tr></thead>
          <tbody>
            <tr><th scope="row">Layout</th><td>Bundled Chromium</td><td>System WebView</td><td><strong>Blinkless, in Go</strong></td></tr>
            <tr><th scope="row">UI event handlers</th><td>JavaScript</td><td>JavaScript and a native bridge</td><td><strong>Go functions</strong></td></tr>
            <tr><th scope="row">Web compatibility</th><td>Chromium web platform</td><td>Host WebView web platform</td><td><strong>Supported HTML and CSS subset</strong></td></tr>
          </tbody>
        </table>
      </div>
      <p><a href="#docs/compare-electron">Read the Electron comparison</a> or explore <a href="#docs/why-ownframe">why we built ownframe</a>.</p>
    </section>

    <section id="start" className="start-section" aria-labelledby="start-title">
      <div className="start-intro">
        <div className="start-copy">
          <p className="eyebrow">Your first native window</p>
          <h2 id="start-title">Start with a page<br /> and a Go handler.</h2>
          <p>Try the running example. Then explore the Go, HTML, and CSS that make it work.</p>
          <div className="start-file-labels" aria-label="Three source files">
            <span>Go logic</span><span>HTML template</span><span>CSS styles</span>
          </div>
        </div>
        <ol className="start-steps">
          <li>
            <div className="start-step-content">
              <h3>Install ownframe</h3>
              <p>Run this in your Go module.</p>
              <div className="start-command">
                <code>go get github.com/chinmay-sawant/ownframe@v0.0.2</code>
                <button type="button" onClick={copyCommand} aria-label="Copy starter install command">{copyStatus}</button>
              </div>
            </div>
          </li>
          <li>
            <div className="start-step-content">
              <h3>Keep the three files together</h3>
              <p>Save <code>main.go</code>, <code>template.html</code>, and <code>style.css</code> in one directory. Go embeds the HTML and CSS into your app.</p>
            </div>
          </li>
        </ol>
      </div>
      <div className="core-example">
        <div className="starter-preview-heading">
          <span>Workspace Inspector</span><span>Interactive example</span>
        </div>
        <ExampleCode files={starterFiles} initialFile={0} path="" />
      </div>
      <div className="start-run">
        <span>Open your native window</span><code>go run .</code>
        <p>Click Rescan Workspace to see the Go handler update the page.</p>
      </div>
    </section>

    {/* Mobile Section - className mobile-section */}
    <section id="mobile" className="mobile-section" aria-labelledby="mobile-title">
      <h2 id="mobile-title">Take the same interface to Android</h2>
      <p>BindMobile connects your page to Ebitengine&apos;s Android view. Reuse your HTML, CSS, and Go handlers across desktop and mobile without WebView dependencies.</p>
      <p>Phone apps support touch momentum scrolling, camera and photo picker integration, and soft keyboard input with dynamic IME insets.</p>
      <div className="mobile-guides links">
        <a href="#docs/showcase/telegram">Build the Android demo</a>
        <a href="#docs/showcase/telegram">See phone examples</a>
      </div>
    </section>
  </div>;
}
