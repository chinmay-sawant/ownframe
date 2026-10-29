import React, { useState } from 'react';
import ExampleCode from './ExampleCode';
import highlight from 'highlight.js/lib/core';
import go from 'highlight.js/lib/languages/go';
import { repository } from './project';
import gopher from '../../assets/brand/gopher.webp';
import teamsPreview from '../../assets/screenshots/demos/teams.webp';
import telegramPreview from '../../assets/screenshots/android/telegram/telegram-thread.webp';
import wisprPreview from '../../assets/screenshots/demos/wispr-flow.webp';
import desktopCatPreview from '../../assets/screenshots/demos/desktop-cat.webp';

highlight.registerLanguage('go', go);

function GoCode({ source }) {
  return <pre><code className="hljs language-go"
    dangerouslySetInnerHTML={{ __html: highlight.highlight(source, { language: 'go' }).value }} /></pre>;
}

export const quickStart = `package main

import (
    "context"
    "log"

    "github.com/chinmay-sawant/ownframe"
)

type Counter struct {
    Count int
}

func main() {
    page, err := ownframe.New(ownframe.Config{
        Title:  "Counter",
        Width:  480,
        Height: 320,
        HTML: ` + "`" + `
            <div style="padding: 24px; font-family: sans-serif;">
                <h1>Count: {{.Count}}</h1>
                <button data-action="inc">+1</button>
            </div>` + "`" + `,
    })
    if err != nil {
        log.Fatal(err)
    }

    state := Counter{Count: 0}
    page.SetData(&state)

    page.Handle(ownframe.Handlers{
        Click: func(ctx context.Context, box ownframe.Box) error {
            if box.Action == "inc" {
                state.Count++
                page.SetData(&state)
            }
            return nil
        },
    })

    if err := ownframe.Run(context.Background(), page); err != nil {
        log.Fatal(err)
    }
}`;

const coreExample = `page, err := ownframe.New(ownframe.Config{
    Title:  "Workspace Inspector",
    Width:  640,
    Height: 480,
    HTML: ` + "`" + `<div style="padding: 24px;">
            <h1>{{.Workspace}}</h1>
            <p>Status: {{.Status}} ({{.Files}} files)</p>
            <button data-action="rescan">Rescan Workspace</button>
          </div>` + "`" + `,
})
if err != nil {
    log.Fatal(err)
}
view := struct {
    Workspace string
    Files     int
    Status    string
}{Workspace: "main", Files: 42, Status: "Ready"}
page.SetData(&view)
page.Handle(ownframe.Handlers{
    Click: func(ctx context.Context, box ownframe.Box) error {
        if box.Action == "rescan" {
            view.Status = "Updated"
            page.SetData(&view)
        }
        return nil
    },
})
if err := ownframe.Run(context.Background(), page); err != nil {
    log.Fatal(err)
}`;

const showcases = [
  {
    title: 'Code Review',
    tag: 'Desktop Tooling',
    description: 'A native dark multi-pane IDE with syntax highlighting in pure Go, visual git commit graph, and word-level diff pairing with 26 ns idle ticks.',
    image: `${import.meta.env.BASE_URL}demos/code-view.png`,
    alt: 'Code Review application showing git commit graph, file tree, and syntax-highlighted editor.',
    href: `${repository}/tree/master/examples/code-view`,
    link: 'View the Go example',
  },
  {
    title: 'Spotify Player',
    tag: 'Media Application',
    description: 'Multi-screen desktop music player with offline CC0 audio playback, 60 FPS CSS keyframe equalizer, and sub-millisecond retained display list updates.',
    image: `${import.meta.env.BASE_URL}demos/videos/spotify.webp`,
    alt: 'Spotify Player demo switching between desktop music screens.',
    href: `${repository}/tree/master/examples/spotify-player`,
    link: 'View the Go example',
  },
  {
    title: 'Telegram on Android',
    tag: 'Mobile Android',
    description: 'Native mobile app handling dynamic soft keyboard insets, pinned composer, touch momentum, and dark theme in a 29 MB standalone APK.',
    image: telegramPreview,
    alt: 'Telegram demo conversation running on an Android Pixel device.',
    phone: true,
    href: '#docs/showcase/telegram',
    link: 'Build the Android demo',
  },
  {
    title: 'Wispr Flow Dashboard',
    tag: 'Productivity Shell',
    description: '13-package productivity app shell featuring typing speed radial gauges, heat-calendar streak grid, and responsive desktop layout.',
    image: wisprPreview,
    alt: 'Wispr Flow dashboard showing speech typing statistics and activity calendar.',
    href: '#docs/showcase/wispr-flow',
    link: 'Explore showcase',
    liveHref: `${import.meta.env.BASE_URL}demos/wispr-flow/index.html`,
    liveLink: 'Try the live demo (52 MB)',
  },
  {
    title: 'Desktop Cat Overlay',
    tag: 'System Companion',
    description: 'Transparent click-through desktop companion that passes mouse events through clear pixels, receiving notifications over a local Unix domain socket.',
    image: desktopCatPreview,
    alt: 'Desktop cat overlay floating with transparent speech bubble.',
    href: '#docs/showcase/desktop-overlays',
    link: 'Explore showcase',
  },
  {
    title: 'Teams Collaboration',
    tag: 'Enterprise Chat',
    description: 'Multi-screen collaboration app with channel sidebar, thread conversations, and responsive message composer written in HTML, CSS, and Go.',
    image: teamsPreview,
    alt: 'Teams collaboration demo with channel sidebar and message composer.',
    href: `${repository}/tree/master/examples/teams`,
    link: 'View the Go example',
  },
];

export default function Home() {
  const [copied, setCopied] = useState(false);

  const copyCommand = () => {
    navigator.clipboard?.writeText('go get github.com/chinmay-sawant/ownframe@v0.0.2');
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return <div className="home">
    {/* Hero Section */}
    <section className="hero" aria-labelledby="hero-title">
      <div className="hero-copy">
        <h1 id="hero-title">HTML and CSS interfaces. Compiled Go runtime. Zero browser engine.</h1>
        <p>ownframe renders declarative HTML templates and CSS styling directly into native windows through Ebitengine and Blinkless, an in-process layout engine written in pure Go. No Chromium, no WebKit, no WebView, and no JavaScript runtime.</p>
        <div className="hero-actions links">
          <a href="#start" className="button-primary">Quickstart Guide</a>
          <a href="#engine" className="button-secondary">Architecture</a>
          <button type="button" className="command-chip" onClick={copyCommand} title="Copy install command">
            <span className="command-prompt">$</span>
            <code>go get github.com/chinmay-sawant/ownframe@v0.0.2</code>
            <span className="copy-label">{copied ? 'Copied' : 'Copy'}</span>
          </button>
        </div>
      </div>
      <img className="hero-mascot" src={gopher} width="150" height="150"
        alt="The blue ownframe mascot waving hello" />

      {/* Interactive Hero Preview */}
      <div className="hero-preview-card" aria-label="Interactive live preview of ownframe application">
        <ExampleCode />

        <div className="preview-pane preview-window-pane">
          <div className="preview-header preview-header-app">
            <span className="app-window-title">Launch board</span>
            <div className="devtools-hud" aria-label="Demo runtime">
              <span className="hud-metric">Go · WebAssembly preview</span>
            </div>
          </div>
          <div className="preview-wasm-container">
            <iframe
              title="Interactive launch board rendered by ownframe"
              src={`${import.meta.env.BASE_URL}demo/index.html`}
              className="hero-wasm-iframe"
              loading="eager"
            />
          </div>
        </div>
      </div>
    </section>

    {/* Showcase Section - Child 2 (Required by check-site.mjs: child 2 id must be "demos") */}
    <section id="demos" className="demo-section" aria-labelledby="demos-title">
      <h2 id="demos-title">Real applications built on ownframe</h2>
      <p>Every example below runs directly from the repository as a compiled Go program in native desktop windows or Android activities.</p>
      <div className="demo-grid">
        {showcases.map((demo) => <article
          className={`demo-card${demo.phone ? ' demo-card-phone' : ''}`} key={demo.title}>
          <div className={`demo-media${demo.phone ? ' demo-media-phone' : ''}`}>
            <img className="demo-image" src={demo.image} alt={demo.alt} loading="lazy" decoding="async" />
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
      <p><a href="#docs/showcase/code-view">Browse all examples in documentation</a>. Product-inspired showcases are unofficial technical demonstrations.</p>
    </section>

    {/* Architecture Section */}
    <section id="engine" className="engine-section" aria-labelledby="engine-title">
      <h2 id="engine-title">A single Go binary from template to GPU canvas</h2>
      <p>Unlike hybrid desktop frameworks that orchestrate multiple IPC processes, ownframe executes entirely within your Go application process.</p>
      <div className="engine-flow" aria-label="How an ownframe screen reaches the window">
        <div className="engine-step">
          <div className="step-num">01</div>
          <h3>Declarative Markup & State</h3>
          <p>Describe screens with standard HTML templates and CSS styling. Bind Go structs directly with SetData. Zero IPC serialization and zero bridge overhead.</p>
        </div>
        <div className="engine-step">
          <div className="step-num">02</div>
          <h3>In-Process Blinkless Engine</h3>
          <p>Pure Go tokenizer, cascade resolver, and layout solver built from scratch. Emits vector display lists without external C dependencies or browser runtimes.</p>
        </div>
        <div className="engine-step">
          <div className="step-num">03</div>
          <h3>Ebitengine GPU Presentation</h3>
          <p>Hardware accelerated rendering via modern graphics APIs. 256 px tile caching, partial repaints, and retained GPU layers ensure 60 FPS motion with zero heap allocations.</p>
        </div>
      </div>
      <p>No embedded browser, no WebKit, and no JavaScript runtime.</p>
    </section>

    {/* Benchmarks Section - className comparison-section */}
    <section id="compare" className="comparison-section" aria-labelledby="compare-title">
      <h2 id="compare-title">Engine performance measured in microseconds</h2>
      <p>Measurements taken from production release builds running on Linux and macOS desktop targets. Real runtime pipeline latency with zero marketing inflation.</p>
      
      {/* Metric Cards */}
      <div className="metrics-grid">
        <div className="metric-card">
          <div className="metric-number">176 ns</div>
          <div className="metric-title">Compositor Latency</div>
          <p className="metric-detail">Retained GPU layer transforms and opacity tracks bypass document relayout entirely with 0 heap allocs.</p>
        </div>
        <div className="metric-card">
          <div className="metric-number">15 MB</div>
          <div className="metric-title">Idle Go Heap</div>
          <p className="metric-detail">Lightweight single binary memory footprint compared to 150 MB to 500 MB for Chromium runtimes.</p>
        </div>
        <div className="metric-card">
          <div className="metric-number">163 µs</div>
          <div className="metric-title">Warm Redraw Time</div>
          <p className="metric-detail">Template execution, cascade resolution, layout computation, and vector display list emission.</p>
        </div>
        <div className="metric-card">
          <div className="metric-number">0.0%</div>
          <div className="metric-title">Idle CPU Utilization</div>
          <p className="metric-detail">Clean desktop windows park in native OS event wait, waking instantly on user input or frame deadlines.</p>
        </div>
      </div>

      <div className="comparison-scroll">
        <table className="comparison-table">
          <caption className="visually-hidden">Runtime architecture comparison between ownframe, Electron, and Wails</caption>
          <thead>
            <tr>
              <th scope="col">Dimension</th>
              <th scope="col">Electron (Chromium)</th>
              <th scope="col">Wails v2 / Tauri</th>
              <th scope="col">ownframe 1.0</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <th scope="row">OS Processes</th>
              <td>4 to 8 processes (Main, Renderer, GPU)</td>
              <td>1 host process + OS WebView helpers</td>
              <td>1 single compiled Go binary</td>
            </tr>
            <tr>
              <th scope="row">Layout Engine</th>
              <td>Chromium Blink (C++)</td>
              <td>System WebKit / WebView2</td>
              <td>Blinkless (pure Go, in-process)</td>
            </tr>
            <tr>
              <th scope="row">Scripting Engine</th>
              <td>V8 / Node.js JavaScript engines</td>
              <td>JavaScript engine in WebView</td>
              <td>Zero JavaScript; pure Go handlers</td>
            </tr>
            <tr>
              <th scope="row">IPC Overhead</th>
              <td>JSON serialization over IPC pipes</td>
              <td>WebKit JS bridge serialization</td>
              <td>Zero IPC; direct Go memory struct mutation</td>
            </tr>
            <tr>
              <th scope="row">Idle Memory</th>
              <td>90 MB to 220 MB RSS</td>
              <td>40 MB to 95 MB RSS</td>
              <td>15 MB to 17 MB Heap (~28 MB RSS)</td>
            </tr>
            <tr>
              <th scope="row">Cold Redraw</th>
              <td>12 ms to 45 ms</td>
              <td>8 ms to 24 ms</td>
              <td>0.16 ms to 0.63 ms (163 µs warm)</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p><a href="#docs/compare-electron">Read the Electron comparison</a> or explore <a href="#docs/why-ownframe">why we built ownframe</a>.</p>
    </section>

    {/* Quickstart Section - className start-section */}
    <section id="start" className="start-section" aria-labelledby="start-title">
      <h2 id="start-title">Simple, explicit Go primitives</h2>
      <p>Install ownframe in your Go module:</p>
      <pre className="setup-command"><code>go get github.com/chinmay-sawant/ownframe@v0.0.2</code></pre>
      <p>Create a page and open its window. Standard Go structs back the template:</p>
      <div className="core-example">
        <GoCode source={coreExample} />
      </div>
      <details className="full-program">
        <summary>Full runnable program</summary>
        <p>Save this as <code>main.go</code>, then run <code>go run .</code>:</p>
        <div className="code-example"><GoCode source={quickStart} /></div>
      </details>
      <p>Add CSS directly in your template or load external stylesheet files. Connect interactive controls with data-action attributes and Go handlers.</p>
    </section>

    {/* Mobile Section - className mobile-section */}
    <section id="mobile" className="mobile-section" aria-labelledby="mobile-title">
      <h2 id="mobile-title">Bring your screen to mobile</h2>
      <p>BindMobile connects your page to Ebitengine&apos;s Android view. Reuse your HTML, CSS, and Go handlers across desktop and mobile without WebView dependencies.</p>
      <p>Phone apps support touch momentum scrolling, camera and photo picker integration, and soft keyboard input with dynamic IME insets.</p>
      <div className="mobile-guides links">
        <a href="#docs/showcase/telegram">Build the Android demo</a>
        <a href="#docs/showcase/telegram">See phone examples</a>
      </div>
    </section>
  </div>;
}
