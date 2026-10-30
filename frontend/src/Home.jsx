import React, { useEffect, useRef, useState } from 'react';
import ExampleCode from './ExampleCode';
import { helloFiles, starterFiles } from './starter';
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
  const [starter, setStarter] = useState('hello');
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
        <h1 id="hero-title">Build native apps in Go<br className="hero-break" /> with HTML and CSS.</h1>
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
      <div className={`core-example ${starter === 'hello' ? 'hello-example' : 'split-example'}`}>
        <div className="starter-picker" aria-label="Choose an example">
          <button type="button" aria-pressed={starter === 'hello'} onClick={() => setStarter('hello')}>Hello World</button>
          <button type="button" aria-pressed={starter === 'split'} onClick={() => setStarter('split')}>Split with friends</button>
        </div>
        <ExampleCode key={starter} files={starter === 'hello' ? helloFiles : starterFiles} initialFile={0} path="" />
      </div>
      <div className="start-run">
        <span>Open your native window</span><code>go run .</code>
        <p>{starter === 'hello' ? 'Click the button to see a Go handler update the page.' : 'Choose friends, enter a bill, and calculate each share.'}</p>
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
