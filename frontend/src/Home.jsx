import React from 'react';
import highlight from 'highlight.js/lib/core';
import go from 'highlight.js/lib/languages/go';
import { repository } from './project';
import gopher from '../../assets/brand/gopher.webp';
import teamsPreview from '../../assets/screenshots/demos/teams.webp';
import telegramPreview from '../../assets/screenshots/android/telegram/telegram-thread.webp';
import eyesPreview from '../../assets/screenshots/android/eyes/eyes-idle.webp';
import depthAlertPreview from '../../assets/screenshots/android/depth-alert/depth-alert-pixel7.webp';

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

func main() {
    page, err := ownframe.New(ownframe.Config{
        Title: "Hello",
        HTML:  "<div>Hello world</div>",
        Width: 480, Height: 320,
    })
    if err != nil {
        log.Fatal(err)
    }
    if err := ownframe.Run(context.Background(), page); err != nil {
        log.Fatal(err)
    }
}`;

const coreExample = `page, err := ownframe.New(ownframe.Config{
    HTML: "<div>Hello world</div>", Width: 480, Height: 320,
})
if err != nil {
    log.Fatal(err)
}
if err := ownframe.Run(context.Background(), page); err != nil {
    log.Fatal(err)
}`;

const demos = [
  {
    title: 'Spotify player',
    description: 'A desktop music player with several screens and bundled audio.',
    image: `${import.meta.env.BASE_URL}demos/videos/spotify.webp`,
    alt: 'Spotify player demo switching between its desktop screens.',
    href: `${repository}/tree/master/examples/spotify-player`,
    link: 'View the Go example',
  },
  {
    title: 'Teams',
    description: 'Chat and collaboration screens written in HTML, CSS, and Go.',
    image: teamsPreview,
    alt: 'Teams demo with a sidebar, conversations, and a message composer.',
    href: `${repository}/tree/master/examples/teams`,
    link: 'View the Go example',
  },
  {
    title: 'Telegram on Android',
    description: 'A phone app with conversations, themes, and soft keyboard input.',
    image: telegramPreview,
    alt: 'Telegram demo conversation running on Android.',
    phone: true,
    href: '#docs/examples-telegram-android-readme',
    link: 'Build the Android demo',
  },
  {
    title: 'Depth Alert on Android',
    description: 'An experimental ARCore depth and haptics prototype. It reports unknown when the camera cannot confirm its view; it is not a mobility aid.',
    image: depthAlertPreview,
    alt: 'Depth Alert on a Pixel 7 showing its live camera preview, status, and depth summary.',
    phone: true,
    href: '#docs/examples-depth-alert-readme',
    link: 'Read about the prototype',
  },
  {
    title: 'Eyes on Android',
    description: 'Animated robot eyes with 21 expressions, touch controls, and motion response.',
    image: eyesPreview,
    alt: 'Robot eyes open on a black Android screen.',
    href: `${repository}/raw/refs/heads/master/examples/eyes/android/artifacts/eyes-arm64-release.apk`,
    link: 'Download arm64 APK',
  },
];

export default function Home() {
  return <div className="home">
    <section className="hero" aria-labelledby="hero-title">
      <div className="hero-copy">
        <p className="eyebrow">Go apps with an HTML and CSS interface</p>
        <h1 id="hero-title">HTML screens. Go logic.</h1>
        <p>Write your interface in HTML and CSS. Handle data and actions in Go.
          Build desktop and mobile apps with Blinkless, the engine we built from scratch.</p>
        <div className="hero-actions links">
          <a href="#start">Get started</a>
          <a href="#demos">Explore examples</a>
        </div>
      </div>
      <img className="hero-mascot" src={gopher} width="180" height="180"
        alt="The blue OwnFrame Gopher waving hello" />
    </section>

    <section id="demos" className="demo-section" aria-labelledby="demos-title">
      <h2 id="demos-title">Examples</h2>
      <p>Desktop screens and a phone app, built with the same engine.</p>
      <div className="demo-grid">{demos.map((demo) => <article
        className={`demo-card${demo.phone ? ' demo-card-phone' : ''}`} key={demo.title}>
        <div className={`demo-media${demo.phone ? ' demo-media-phone' : ''}`}>
          <img className="demo-image" src={demo.image} alt={demo.alt} loading="lazy" decoding="async" />
        </div>
        <div className="demo-card-content">
          <h3>{demo.title}</h3>
          <p>{demo.description}</p>
          <a href={demo.href}>{demo.link}</a>
        </div>
      </article>)}</div>
      <p><a href="#docs/showcase/telegram">Browse all examples</a>. Product-inspired demos are unofficial.</p>
    </section>

    <section id="engine" className="engine-section" aria-labelledby="engine-title">
      <h2 id="engine-title">One engine across your screens</h2>
      <p><a href="https://github.com/chinmay-sawant/blinkless">Blinkless</a> is our HTML and CSS engine,
        built from scratch in Go. Every OwnFrame screen uses it for layout. OwnFrame connects
        your Go code to the screen, and Ebitengine supplies the window and drawing.</p>
      <div className="engine-flow" aria-label="How an OwnFrame screen reaches the window">
        <div className="engine-step">
          <h3>Your HTML and CSS</h3>
          <p>Describe the screen in markup and styles.</p>
        </div>
        <div className="engine-step">
          <h3>Blinkless</h3>
          <p>Parse HTML, apply CSS, and calculate layout.</p>
        </div>
        <div className="engine-step">
          <h3>OwnFrame + Ebitengine</h3>
          <p>Draw the result and connect input to Go handlers.</p>
        </div>
      </div>
      <p>No embedded browser or JavaScript runtime.
        {' '}<a href="#docs/features">See supported features and limits</a>.</p>
    </section>

    <section id="start" className="start-section" aria-labelledby="start-title">
      <h2 id="start-title">Start with one HTML screen</h2>
      <p>Install OwnFrame in a Go module:</p>
      <pre className="setup-command"><code>go get github.com/chinmay-sawant/ownframe</code></pre>
      <p>Create a page, then open its window. These are the core calls inside <code>main</code>:</p>
      <div className="core-example">
        <GoCode source={coreExample} />
      </div>
      <details className="full-program">
        <summary>Full runnable program</summary>
        <p>Save this as <code>main.go</code>, then run <code>go run .</code>.</p>
        <div className="code-example"><GoCode source={quickStart} /></div>
      </details>
      <p>Add CSS to your HTML or load a template file. Connect changing data with
        {' '}<a href="#docs/binding"><code>SetData</code> and Go handlers</a>.
        The <a href="#docs/window">window guide</a> covers setup and sizing.</p>
    </section>

    <section id="compare" className="comparison-section" aria-labelledby="compare-title">
      <h2 id="compare-title">Compare with Rust GPUI</h2>
      <p>GPUI builds the interface in Rust code. OwnFrame uses HTML and CSS for the screen,
        with Go for application logic.</p>
      <div className="comparison-scroll">
        <table className="comparison-table">
          <caption className="visually-hidden">How you write an interface in OwnFrame and Rust GPUI</caption>
          <thead><tr><th scope="col">You write</th><th scope="col">OwnFrame</th><th scope="col">Rust GPUI</th></tr></thead>
          <tbody>
            <tr><th scope="row">The screen</th><td>HTML templates</td><td>Rust elements and builder calls</td></tr>
            <tr><th scope="row">The styles</th><td>CSS</td><td>Rust methods such as <code>.flex()</code></td></tr>
            <tr><th scope="row">The app logic</th><td>Go data and handlers</td><td>Rust state and event handlers</td></tr>
            <tr><th scope="row">The renderer</th><td>Our own Go web view, with HTML/CSS layout powered by Blinkless</td><td>GPUI's renderer for Rust-defined interfaces</td></tr>
          </tbody>
        </table>
      </div>
      <p>Blinkless lays out the supported HTML and CSS. OwnFrame handles input and Go updates;
        Ebitengine supplies the window and drawing.</p>
      <p><a href="#docs/compare-rust-gpui">Read the comparison</a> or see the
        {' '}<a href="https://www.gpui.rs/">official GPUI example</a>.</p>
    </section>

    <section id="mobile" className="mobile-section" aria-labelledby="mobile-title">
      <h2 id="mobile-title">Bring your screen to mobile</h2>
      <p><code>BindMobile</code> connects your page to Ebitengine's Android and iOS view.
        Reuse your HTML, CSS, and Go handlers. Android has a complete Telegram demo project;
        iOS builds use the mobile binding and require Xcode.</p>
      <p>Phone apps support touch scrolling and soft keyboard input. The clipboard uses an
        in-memory copy, and native screen-reader integration is still unavailable.
        Check the platform limits before shipping.</p>
      <div className="mobile-guides links">
        <a href="#docs/android-app-creation">Create an Android app</a>
        <a href="#docs/platforms/phone">Android and iOS setup</a>
        <a href="#docs/android-runtime">Android runtime limits</a>
      </div>
    </section>
  </div>;
}
