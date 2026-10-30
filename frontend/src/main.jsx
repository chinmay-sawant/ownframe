import React, { Suspense, lazy, useEffect, useLayoutEffect, useRef, useState } from 'react';
import { createRoot } from 'react-dom/client';
import Home from './Home';
import Team from './Team';
import ThemeToggle from './ThemeToggle';
import { loadStars } from './stars';
import { repository } from './project';
import './style.css';

const Documentation = lazy(() => import('./Documentation'));

function App() {
  const [hash, setHash] = useState(location.hash);
  const [stars, setStars] = useState(null);
  const [menuOpen, setMenuOpen] = useState(false);
  const menuButton = useRef(null);
  const isDocs = hash.startsWith('#docs');
  const isTeam = hash === '#team';

  useLayoutEffect(() => {
    const header = document.querySelector('header');
    if (!header) return;
    const measure = () => document.documentElement.style.setProperty('--header-height', `${header.offsetHeight}px`);
    measure();
    const observer = new ResizeObserver(measure);
    observer.observe(header);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    let mounted = true;
    loadStars().then((value) => { if (mounted) setStars(value); });
    const update = () => {
      setHash(location.hash);
      setMenuOpen(false);
    };
    window.addEventListener('hashchange', update);
    return () => {
      mounted = false;
      window.removeEventListener('hashchange', update);
    };
  }, []);

  useEffect(() => {
    if (isDocs) return;
    if (isTeam || !hash || hash === '#home') window.scrollTo(0, 0);
    else document.getElementById(hash.slice(1))?.scrollIntoView();
  }, [hash, isDocs, isTeam]);

  return <div className={isDocs ? 'docs-shell' : undefined}>
    <a className="skip" href="#main" onClick={(event) => {
      event.preventDefault();
      document.getElementById('main').focus();
    }}>Skip to content</a>
    <header onKeyDown={(event) => {
      if (event.key === 'Escape' && menuOpen) {
        setMenuOpen(false);
        menuButton.current.focus();
      }
    }}>
      <a className="brand" href="#home">
        <svg className="brand-glyph" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
          <rect x="3" y="3" width="18" height="18" rx="2" />
          <path d="M3 9h18" />
          <path d="M9 21V9" />
        </svg>
        <span className="brand-name">ownframe</span>
      </a>
      <button className="site-menu-toggle" ref={menuButton} type="button"
        aria-expanded={menuOpen} aria-controls="site-navigation"
        onClick={() => setMenuOpen(!menuOpen)}>{menuOpen ? 'Close' : 'Menu'}</button>
      <nav id="site-navigation" aria-label="Main navigation" data-open={menuOpen}
        onClick={(event) => { if (event.target.closest('a')) setMenuOpen(false); }}>
        <a href="#docs/showcase/wispr-flow" aria-current={hash.startsWith('#docs/showcase') ? 'page' : undefined}>Example apps</a>
        <a href="#engine">Architecture</a>
        <a href="#compare">Benchmarks</a>
        <a href="#docs/why-ownframe" aria-current={hash.startsWith('#docs/why-ownframe') ? 'page' : undefined}>Documentation</a>
        <a href="#team" aria-current={isTeam ? 'page' : undefined}>People & tools</a>
        <a href="#start">Get started</a>
      </nav>
      <ThemeToggle />
      <a className="github" href={repository} title="View ownframe on GitHub">
        <svg className="github-star" viewBox="0 0 16 16" width="16" height="16" fill="currentColor" aria-hidden="true">
          <path d="M8 .25a.75.75 0 0 1 .673.418l1.882 3.815 4.21.612a.75.75 0 0 1 .416 1.279l-3.046 2.97.719 4.192a.75.75 0 0 1-1.088.791L8 12.347l-3.766 1.98a.75.75 0 0 1-1.088-.79l.72-4.194L.818 6.374a.75.75 0 0 1 .416-1.28l4.21-.611L7.327.668A.75.75 0 0 1 8 .25Z" />
        </svg>
        <span className="github-text">GitHub <span className="github-count" aria-live="polite">{stars === null ? '' : stars.toLocaleString() + ' stars'}</span></span>
      </a>
    </header>
    <main id="main" className={isDocs ? 'docs-main' : undefined} tabIndex="-1">
      {isDocs ? <Suspense fallback={<p role="status">Loading documentation...</p>}><Documentation hash={hash} /></Suspense> : isTeam ? <Team /> : <Home />}
    </main>
    <footer>
      <div className="footer-content">
        <div className="footer-brand-col">
          <div className="footer-logo">
            <span className="brand-name">ownframe</span>
          </div>
          <p className="footer-tagline">Native Go desktop and Android apps styled with HTML and CSS.</p>
        </div>
        <div className="footer-nav">
          <a href={repository}>Source on GitHub</a>
          <a href="#demos">Showcase</a>
          <a href="#engine">Architecture</a>
          <a href="#compare">Benchmarks</a>
          <a href="#docs/why-ownframe">Documentation</a>
          <a href="#team">People & tools</a>
        </div>
      </div>
      <div className="footer-bottom">
        <span>MIT Licensed pure Go open-source project.</span>
        <span>Linux · macOS · Windows · Android</span>
      </div>
    </footer>
  </div>;
}

createRoot(document.getElementById('root')).render(<React.StrictMode><App /></React.StrictMode>);
