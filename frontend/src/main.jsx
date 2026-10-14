import React, { Suspense, lazy, useEffect, useLayoutEffect, useState } from 'react';
import { createRoot } from 'react-dom/client';
import Home from './Home';
import ThemeToggle from './ThemeToggle';
import { loadStars } from './stars';
import { repository } from './project';
import './style.css';

const Documentation = lazy(() => import('./Documentation'));

function App() {
  const [hash, setHash] = useState(location.hash);
  const [stars, setStars] = useState(null);
  const isDocs = hash.startsWith('#docs');

  useLayoutEffect(() => {
    const header = document.querySelector('header');
    const measure = () => document.documentElement.style.setProperty('--header-height', `${header.offsetHeight}px`);
    measure();
    const observer = new ResizeObserver(measure);
    observer.observe(header);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    let mounted = true;
    loadStars().then((value) => { if (mounted) setStars(value); });
    const update = () => setHash(location.hash);
    window.addEventListener('hashchange', update);
    return () => {
      mounted = false;
      window.removeEventListener('hashchange', update);
    };
  }, []);

  useEffect(() => {
    if (isDocs) return;
    if (!hash || hash === '#home') window.scrollTo(0, 0);
    else document.getElementById(hash.slice(1))?.scrollIntoView();
  }, [hash, isDocs]);

  return <div className={isDocs ? 'docs-shell' : undefined}>
    <a className="skip" href="#main" onClick={(event) => {
      event.preventDefault();
      document.getElementById('main').focus();
    }}>Skip to content</a>
    <header>
      <a className="brand" href="#home">ownframe</a>
      <nav aria-label="Main navigation">
        <a href="#demos">Examples</a>
        <a href="#compare">Comparison</a>
        <a href="#start">Get started</a>
        <a href="#docs/readme" aria-current={isDocs ? 'page' : undefined}>Docs</a>
      </nav>
      <ThemeToggle />
      <a className="github" href={repository} title="View ownframe on GitHub">
        <svg className="github-star" viewBox="0 0 24 24" width="18" height="18" fill="currentColor" aria-hidden="true">
          <path d="m12 3 2.8 5.7 6.2.9-4.5 4.4 1.1 6.2-5.6-3-5.6 3 1.1-6.2L3 9.6l6.2-.9Z" />
        </svg>
        GitHub <span className="github-count" aria-live="polite">{stars === null ? '' : stars.toLocaleString() + ' stars'}</span>
      </a>
    </header>
    <main id="main" className={isDocs ? 'docs-main' : undefined} tabIndex="-1">
      {isDocs ? <Suspense fallback={<p role="status">Loading documentation...</p>}><Documentation hash={hash} /></Suspense> : <Home />}
    </main>
    <footer>
      <a href={repository}>Source on GitHub</a>
      <a href="#docs/readme">Documentation</a>
      <a href="#docs/platforms">Desktop and mobile</a>
    </footer>
  </div>;
}

createRoot(document.getElementById('root')).render(<React.StrictMode><App /></React.StrictMode>);
