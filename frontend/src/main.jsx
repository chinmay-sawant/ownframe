import React, { useEffect, useRef, useState } from 'react';
import { createRoot } from 'react-dom/client';
import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeSlug from 'rehype-slug';
import rehypeHighlight from 'rehype-highlight';
import { documents, contentUrl, catPreview } from './content';
import { loadStars } from './stars';
import { repository } from './project';
import './style.css';
import ThemeToggle from './ThemeToggle';
import DocPagination from './DocPagination';
import gopher from '../../assets/brand/gopher.webp';
import wisprFlowPreview from '../../assets/screenshots/demos/wispr-flow.webp';
import teamsPreview from '../../assets/screenshots/demos/teams.webp';
import dinoPreview from '../../assets/screenshots/demos/dino.webp';

const demos = [
  {
    title: 'Spotify Player',
    description: 'A Go desktop player rendered by ownframe from HTML and CSS. Ebitengine supplies its window and audio. It uses no WebView; phone examples use the Ebitengine Mobile binding provided by ownframe.',
    animatedWebp: `${import.meta.env.BASE_URL}demos/videos/spotify.webp`,
    alt: 'Animated recording of the Spotify Player screens, including Home, Browse, Profile, and Library.',
    example: `${repository}/tree/master/examples/spotify-player`,
  },
  {
    title: 'Wispr Flow', poster: wisprFlowPreview,
    description: 'HTML and CSS screens rendered by a layout engine written in ownframe.',
    video: `${import.meta.env.BASE_URL}demos/preview.mp4`,
    post: 'https://x.com/chinmay_sawant_/status/2106788230154871126',
  },
  {
    title: 'Desktop cat overlay', poster: catPreview,
    description: 'A transparent, click-through cat that reports what is happening inside opencode.',
    video: `${import.meta.env.BASE_URL}demos/desktop-cat.mp4`,
    post: 'https://x.com/chinmay_sawant_/status/2106829998409789929',
  },
  {
    title: 'Teams demo', poster: teamsPreview,
    description: 'A native chat and collaboration app built with Go, HTML, and CSS using ownframe.',
    video: `${import.meta.env.BASE_URL}demos/teams.mp4`,
    post: 'https://x.com/chinmay_sawant_/status/2107548496194895993/video/1',
  },
  {
    title: 'Dino Run', poster: dinoPreview,
    description: 'An endless runner built with HTML, CSS, and Go using ownframe.',
    example: `${repository}/tree/master/examples/dino`,
  },
];
const showcaseSections = [
  { id: 'telegram', title: 'Telegram' },
  { id: 'teams', title: 'Teams' },
  { id: 'spotify-player', title: 'Spotify Player' },
  { id: 'pet-care', title: 'Pet Care' },
  { id: 'desktop-overlays', title: 'Desktop overlays' },
  { id: 'wispr-flow', title: 'Wispr Flow' },
  { id: 'dino-run', title: 'Dino Run' },
];
const quickStart = `\`\`\`go
package main

import (
\t"context"
\t"log"

\t"github.com/chinmay-sawant/ownframe"
)

func main() {
\tpage, err := ownframe.New(ownframe.Config{
\t\tTitle:  "Hello",
\t\tHTML:   "<h1>{{.Title}}</h1>",
\t\tWidth:  480,
\t\tHeight: 640,
\t})
\tif err != nil {
\t\tlog.Fatal(err)
\t}
\tpage.SetData(struct{ Title string }{"Hello"})
\tif err := ownframe.Run(context.Background(), page); err != nil {
\t\tlog.Fatal(err)
\t}
}
\`\`\``;

function App() {
  const [hash, setHash] = useState(location.hash);
  const [stars, setStars] = useState(null);
  const contentPane = useRef(null);
  useEffect(() => {
    let mounted = true;
    loadStars().then((value) => { if (mounted) setStars(value); });
    const update = () => setHash(location.hash);
    window.addEventListener('hashchange', update);
    return () => { mounted = false; window.removeEventListener('hashchange', update); };
  }, []);
  const [, slug = 'readme', anchor] = hash.split('/');
  const isDocs = hash.startsWith('#docs');
  const document = documents.find((entry) => entry.slug === slug);
  const isShowcase = document?.path === 'showcase.md';
  const activeShowcase = showcaseSections.some((entry) => entry.id === anchor) ? anchor : 'telegram';
  const markdownComponents = isShowcase ? {
    ul: ({ children }) => <ul className={`showcase-gallery${activeShowcase === 'pet-care' ? ' pet-care-gallery' : ''}`}>{children}</ul>,
    img: ({ src, alt = '', title }) => <img src={src} alt={alt} title={title}
      className={src?.includes('/spotify.webp') ? 'spotify-window-preview' : undefined}
      loading="lazy" decoding="async" />,
  } : undefined;
  const showcaseText = isShowcase
    ? selectShowcase(document.text, activeShowcase)
    : document?.text;
  useEffect(() => {
    if (isDocs) {
      const pane = contentPane.current;
      if (pane) {
        const target = anchor && pane.querySelector(`#${CSS.escape(anchor)}`);
        if (target) pane.scrollTop += target.getBoundingClientRect().top - pane.getBoundingClientRect().top;
        else pane.scrollTop = 0;
      }
    } else if (hash === '' || hash === '#home') window.scrollTo(0, 0);
    else documentElement(hash.slice(1))?.scrollIntoView();
  }, [hash, isDocs, anchor]);

  return <div className={isDocs ? 'docs-shell' : undefined}>
    <a className="skip" href="#main">Skip to content</a>
    <header>
      <a className="brand" href="#home">ownframe</a>
      <nav aria-label="Main navigation">
        <a href="#demos">Demos</a>
        <a href="#docs/readme" aria-current={isDocs ? 'page' : undefined}>Documentation</a>
      </nav>
      <ThemeToggle />
      <a className="github" href={repository} title="View ownframe on GitHub">
        <span className="github-star" aria-hidden="true">⭐</span>
        GitHub <span aria-live="polite">{stars === null ? '' : stars.toLocaleString() + ' stars'}</span>
      </a>
    </header>
    <main id="main" className={isDocs ? 'docs-main' : undefined} tabIndex="-1">
      {isDocs ? <div className="docs-layout">
        <aside>
          <h2>Documentation</h2>
          <nav className="docs-nav" aria-label="Documentation topics">{documents.map((entry) =>
            <div key={entry.slug} className="docs-nav-entry">
              <a href={entry.slug === 'showcase' ? '#docs/showcase/telegram' : `#docs/${entry.slug}`}
                aria-current={slug === entry.slug && entry.slug !== 'showcase' ? 'page' : undefined}
                className={slug === 'showcase' && entry.slug === 'showcase' ? 'docs-nav-parent active' : 'docs-nav-parent'}>
                {entry.title}
              </a>
              {entry.slug === 'showcase' && <div className="showcase-subnav" role="group" aria-label="Showcase sections">
                {showcaseSections.map((section) => <a key={section.id}
                  href={`#docs/showcase/${section.id}`}
                  aria-current={slug === 'showcase' && activeShowcase === section.id ? 'page' : undefined}>
                  {section.title}
                </a>)}
              </div>}
            </div>
          )}</nav>
        </aside>
        <article ref={contentPane} className={`markdown docs-content${isShowcase ? ' showcase-page' : ''}`}>
          {document ? <><Markdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeSlug, rehypeHighlight]}
            urlTransform={(url) => contentUrl(url, document)} components={markdownComponents}>{showcaseText}</Markdown>
            <DocPagination documents={documents} slug={slug} /></>
            : <><h1>Document not found</h1><a href="#docs/readme">Open the documentation index</a></>}
        </article>
      </div> : <>
        <section className="intro">
          <img className="mascot" src={gopher} width="180" height="180" alt="A cheerful blue Gopher waving hello" />
          <h1>HTML screens. Go logic.</h1>
          <p>ownframe was formerly known as go-gpui.</p>
          <p>Write your screen in HTML and CSS, handle its data and actions in Go, and run it on desktop, in the browser through WebAssembly, or on a phone.</p>
          <p>ownframe uses the Blinkless HTML and CSS engine for layout and Ebitengine for its native window. It ships no WebView, Chromium, WebKit, or JavaScript runtime.</p>
          <div className="links"><a href="#start">Get started</a><a href="#docs/readme">Read the documentation</a></div>
        </section>
        <section id="demos">
          <h2>Demos</h2>
          <div className="demos">{demos.map((demo) => <article key={demo.title}>
            <h3>{demo.title}</h3><p>{demo.description}</p>
            <div className="demo-media">
              {demo.video ? <video controls playsInline preload="metadata" poster={demo.poster} aria-label={demo.title}>
                <source src={demo.video} type="video/mp4" />
                Your browser does not support this video. Use the original post below.
              </video> : demo.animatedWebp ? <img className="demo-image" src={demo.animatedWebp}
                width="780" height="520" alt={demo.alt} loading="lazy" decoding="async" />
                : <img className="demo-image" src={demo.poster} loading="lazy" alt="Dino Run in a native desktop window" />}
            </div>
            <a href={demo.post || demo.example}>{demo.post ? 'Watch the original video on X' : `View ${demo.title} on GitHub`}</a>
          </article>)}</div>
        </section>
        <section id="start">
          <h2>Get started</h2>
          <p>Install the module with <code>go get github.com/chinmay-sawant/ownframe</code>, then create a Go file:</p>
          <Markdown rehypePlugins={[rehypeHighlight]}>{quickStart}</Markdown>
          <p className="get-started-note">The login demo accepts <code>secret</code> for both fields. <a href="#docs/examples-readme">Browse the examples</a> or read about <a href="#docs/window">opening a window</a>.</p>
        </section>
        <section>
          <h2>Documentation</h2>
          <p>The guides cover <a href="#docs/binding">data binding</a>, <a href="#docs/theming">theming</a>, <a href="#docs/platforms">platforms</a>, and <a href="#docs/packaging">packaging</a>.</p>
          <p className="wide-note">There is no live DOM or JavaScript engine, and the library does not support video, canvas, or WebGL. See the <a href="#docs/features">feature list and limits</a>.</p>
        </section>
      </>}
    </main>
    <footer><a href={repository}>Source on GitHub</a><a href="#docs/readme">Documentation</a></footer>
  </div>;
}

function documentElement(id) { return window.document.getElementById(id); }

function selectShowcase(markdown, id) {
  const [intro, ...sections] = markdown.split(/(?=^## )/m);
  const selected = sections.find((section) => {
    const title = section.match(/^## (.+)$/m)?.[1];
    return title?.toLowerCase().replace(/\s+/g, '-') === id;
  });
  return [intro, selected].filter(Boolean).join('\n');
}

createRoot(document.getElementById('root')).render(<React.StrictMode><App /></React.StrictMode>);
