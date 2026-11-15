import React, { useEffect, useRef, useState } from 'react';
import ExampleCode from './ExampleCode';
import { builderFiles, helloBuilderFiles, helloFiles, starterFiles } from './starter';
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

function exampleFiles(pattern, starter) {
  if (pattern === 'builder' && starter === 'hello') return helloBuilderFiles;
  if (starter === 'hello') return helloFiles;
  if (starter === 'split') return starterFiles;
  return builderFiles;
}

function exampleCommand(pattern, starter) {
  if (pattern === 'builder' && starter === 'hello') return 'go run ./examples/frontend/hello-builder';
  return 'go run .';
}

const installCommand = 'go get github.com/chinmay-sawant/ownframe@v0.0.3';

const values = [
  {
    title: 'No embedded browser',
    text: 'An in-process renderer lays out HTML and CSS and paints it directly. No WebView, no Chromium.',
    icon: <path d="M4 5h16v14H4zM4 9h16" />,
  },
  {
    title: 'Go handlers, HTML templates',
    text: 'Write the page as a template or build the tree in Go. Handlers update state, the page repaints.',
    icon: <path d="M8 8l-4 4 4 4M16 8l4 4-4 4M13.5 6l-3 12" />,
  },
  {
    title: 'One codebase, desktop and Android',
    text: 'Run in a native window on Linux, macOS, and Windows, and bind the same page to an Android view.',
    icon: <path d="M3 6h12v9H3zM8 19h2M17 8h4v11h-4z" />,
  },
];

function Picker({ label, value, onChange, options }) {
  return <div className="h-picker" role="group" aria-label={label}>
    {options.map(([id, text, hint]) => <button key={id} type="button" aria-pressed={value === id}
      disabled={Boolean(hint)} title={hint} onClick={() => onChange(id)}>{text}</button>)}
  </div>;
}

export default function Home() {
  const [copyStatus, setCopyStatus] = useState('Copy');
  const [starter, setStarter] = useState('hello');
  const [builderPattern, setBuilderPattern] = useState(false);
  // The builder pattern exists only for Hello World.
  const pattern = starter === 'hello' && builderPattern ? 'builder' : 'template';
  const copyTimer = useRef(null);
  const demoUrl = `${import.meta.env.BASE_URL}demo/index.html`;

  useEffect(() => () => clearTimeout(copyTimer.current), []);

  async function copyCommand() {
    clearTimeout(copyTimer.current);
    try {
      await navigator.clipboard.writeText(installCommand);
      setCopyStatus('Copied');
    } catch {
      setCopyStatus('Select to copy');
    }
    copyTimer.current = setTimeout(() => setCopyStatus('Copy'), 2500);
  }

  const installBox = (label) => <div className="h-install">
    <code>{installCommand}</code>
    <button type="button" onClick={copyCommand} aria-label={label}>{copyStatus}</button>
    <span className="visually-hidden" role="status">{copyStatus === 'Copied' ? 'Install command copied.' : copyStatus === 'Select to copy' ? 'Clipboard unavailable. Select the command to copy it.' : ''}</span>
  </div>;

  return <div className="home">
    <section className="h-hero container" aria-labelledby="hero-title">
      <div className="h-hero-copy">
        <p className="eyebrow">Go logic · HTML and CSS interfaces</p>
        <h1 id="hero-title">Native apps in Go, styled with HTML and CSS.</h1>
        <p className="lede">Desktop and Android interfaces with Go handlers and an in-process renderer. No embedded browser.</p>
        <div className="h-actions">
          <a href="#start" className="btn btn-primary">Get started</a>
          <a href="#live-demo" className="btn">Try the demo</a>
        </div>
        {installBox('Copy install command')}
        <p className="h-platforms">Linux · macOS · Windows · Android · MIT</p>
      </div>
      <img className="h-mascot" src={gopher} width="300" height="300" alt="The blue ownframe mascot waving hello" />
    </section>

    <section id="live-demo" className="h-demo container-wide" aria-labelledby="live-title">
      <h2 id="live-title" className="visually-hidden">Live demo and source</h2>
      <div className="h-demo-grid">
        <div className="h-source">
          <div className="h-controls">
            <Picker label="Choose a pattern" value={pattern} onChange={(id) => setBuilderPattern(id === 'builder')}
              options={[['template', 'HTML template'], ['builder', 'Builder', starter === 'hello' ? '' : 'The builder pattern is available for Hello World only.']]} />
            <Picker label="Choose an example" value={starter} onChange={setStarter} options={[['hello', 'Hello World'], ['split', 'Split with friends'], ['builder', 'Builder basics']]} />
          </div>
          <ExampleCode key={`${starter}-${pattern}`} files={exampleFiles(pattern, starter)} initialFile={0} path="" />
        </div>
        <div className="h-preview">
          <div className="h-preview-bar">
            <span>Launch board</span>
            <span className="h-faint">Go · WebAssembly</span>
          </div>
          <iframe className="h-frame" src={demoUrl} loading="lazy"
            title="Interactive Launch board: complete tasks, filter the list, and switch themes"
            width="560" height="640" />
          <div className="h-preview-foot">
            <span>Click a task. Try the filters and theme switch.</span>
            <a href={demoUrl}>Open in its own page <span aria-hidden="true">↗</span></a>
          </div>
        </div>
      </div>
      <p className="h-note h-center">This preview loads about 53 MB of WebAssembly. Desktop and Android apps run in native windows.</p>
    </section>

    <section className="h-values container" aria-label="Why ownframe">
      <ul>
        {values.map(v => <li key={v.title}>
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">{v.icon}</svg>
          <h3>{v.title}</h3>
          <p>{v.text}</p>
        </li>)}
      </ul>
    </section>

    <section id="demos" className="section container-wide" aria-labelledby="demos-title">
      <div className="h-head">
        <p className="eyebrow">Examples</p>
        <h2 id="demos-title">See what you can build</h2>
        <p className="lede">Desktop tools, phone apps, and small experiments. Every example includes its Go source.</p>
      </div>
      <div className="h-grid">
        {showcases.map(demo => <article className={`h-card${demo.phone ? ' h-card-phone' : ''}`} key={demo.title}>
          <div className="h-card-media">
            <img src={demo.image} alt={demo.alt} width={demo.phone ? 540 : 1500} height={demo.phone ? 1095 : 900} loading="lazy" decoding="async" />
          </div>
          <div className="h-card-body">
            <span className="pill">{demo.tag}</span>
            <h3>{demo.title}</h3>
            <p>{demo.description}</p>
            <div className="h-card-links">
              <a href={demo.href}>{demo.link}</a>
              {demo.liveHref && <a href={demo.liveHref}>{demo.liveLink}</a>}
            </div>
          </div>
        </article>)}
      </div>
      <p className="h-note"><a href="#docs/showcase/code-view">Browse all examples in documentation</a>. Product-inspired showcases are unofficial technical demonstrations.</p>
    </section>

    <section id="start" className="section h-start container" aria-labelledby="start-title">
      <div className="h-head">
        <p className="eyebrow">Get started</p>
        <h2 id="start-title">Your first native window</h2>
        <p className="lede">{pattern === 'template'
          ? 'Three short files and one command. The viewer above shows each one.'
          : 'The page is a Go tree. ui.Mount renders it, the host holds the state, and View.Update rebuilds it after a click.'}</p>
      </div>
      <ol className="h-steps">
        <li>
          <h3>Install ownframe</h3>
          <p>Run this in your Go module.</p>
          {installBox('Copy starter install command')}
        </li>
        <li>
          <h3>{pattern === 'template' ? 'Keep the three files together' : 'Build the tree in Go'}</h3>
          <p>{pattern === 'template'
            ? <>Save <code>main.go</code>, <code>template.html</code>, and <code>style.css</code> in one directory. Go embeds the HTML and CSS into your app.</>
            : <><code>main.go</code> is the whole page, with no template file.</>}</p>
          <a href="#live-demo" className="h-link">Open the source viewer</a>
        </li>
        <li>
          <h3>Run it</h3>
          <p>{starter === 'hello' ? 'Click the button to see a Go handler update the page.' : starter === 'split' ? 'Choose friends, enter a bill, and calculate each share.' : 'Flip the toggle and switch to see host-owned state re-render the card.'}</p>
          <div className="h-install"><code>{exampleCommand(pattern, starter)}</code></div>
        </li>
      </ol>
    </section>

    <section id="mobile" className="h-mobile container" aria-labelledby="mobile-title">
      <div>
        <p className="eyebrow">Android</p>
        <h2 id="mobile-title">Take the same interface to Android</h2>
      </div>
      <div>
        <p className="lede">BindMobile connects your page to Ebitengine&apos;s Android view. Reuse your HTML, CSS, and Go handlers across desktop and mobile, with touch scrolling, camera and photo picker integration, and soft keyboard insets.</p>
        <a href="#docs/showcase/telegram" className="h-link">Build the Android demo <span aria-hidden="true">→</span></a>
      </div>
    </section>

    <section className="h-cta container" aria-labelledby="cta-title">
      <h2 id="cta-title">Build your first window today.</h2>
      <div className="h-actions h-actions-center">
        <a href="#start" className="btn btn-primary">Get started</a>
        <a href={repository} className="btn">View on GitHub</a>
      </div>
    </section>
  </div>;
}
