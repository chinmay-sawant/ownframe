import React, { Suspense, lazy, useEffect, useRef, useState } from 'react';
import { createRoot } from 'react-dom/client';
import Home from './Home';
import Team from './Team';
import Architecture from './Architecture';
import Benchmarks from './Benchmarks';
import ThemeToggle from './ThemeToggle';
import { loadStars } from './stars';
import { repository } from './project';
import './styles/base.css';
import './styles/code.css';
import './styles/home.css';
import './styles/pages.css';
import './styles/docs.css';

const Documentation = lazy(() => import('./Documentation'));

function App() {
  const [hash, setHash] = useState(location.hash);
  const [stars, setStars] = useState(null);
  const [menuOpen, setMenuOpen] = useState(false);
  const menuButton = useRef(null);
  const isDocs = hash.startsWith('#docs');
  const isTeam = hash === '#team';
  const isArchitecture = hash === '#engine';
  const isBenchmarks = hash === '#compare';

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
    if (isTeam || isArchitecture || isBenchmarks || !hash || hash === '#home') window.scrollTo(0, 0);
    else document.getElementById(hash.slice(1))?.scrollIntoView();
  }, [hash, isDocs, isTeam, isArchitecture, isBenchmarks]);

  return <div className={isDocs ? 'docs-shell' : undefined}>
    <a className="skip" href="#main" onClick={(event) => {
      event.preventDefault();
      document.getElementById('main').focus();
    }}>Skip to content</a>
    <header className="site-header" onKeyDown={(event) => {
      if (event.key === 'Escape' && menuOpen) {
        setMenuOpen(false);
        menuButton.current.focus();
      }
    }}>
      <div className="site-header-inner">
        <a className="brand" href="#home">
          <svg className="brand-glyph" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <rect x="3" y="3" width="18" height="18" rx="3" />
            <path d="M3 9h18" />
            <path d="M9 21V9" />
          </svg>
          <span className="brand-name">ownframe</span>
        </a>
        <nav className="site-nav" id="site-navigation" aria-label="Main navigation" data-open={menuOpen}
          onClick={(event) => { if (event.target.closest('a')) setMenuOpen(false); }}>
          <a href="#demos">Examples</a>
          <a href="#engine" aria-current={isArchitecture ? 'page' : undefined}>How it works</a>
          <a href="#compare" aria-current={isBenchmarks ? 'page' : undefined}>Benchmarks</a>
          <a href="#docs/why-ownframe" aria-current={isDocs ? 'page' : undefined}>Docs</a>
          <a href="#team" aria-current={isTeam ? 'page' : undefined}>About</a>
        </nav>
        <div className="site-actions">
          <ThemeToggle />
          <a className="github" href={repository} title="View ownframe on GitHub">
            <svg className="github-star" viewBox="0 0 16 16" width="14" height="14" fill="currentColor" aria-hidden="true">
              <path d="M8 .25a.75.75 0 0 1 .673.418l1.882 3.815 4.21.612a.75.75 0 0 1 .416 1.279l-3.046 2.97.719 4.192a.75.75 0 0 1-1.088.791L8 12.347l-3.766 1.98a.75.75 0 0 1-1.088-.79l.72-4.194L.818 6.374a.75.75 0 0 1 .416-1.28l4.21-.611L7.327.668A.75.75 0 0 1 8 .25Z" />
            </svg>
            <span className="github-text">GitHub <span className="github-count" aria-live="polite">{stars === null ? '' : stars.toLocaleString()}</span></span>
          </a>
        </div>
        <button className="site-menu-toggle" ref={menuButton} type="button"
          aria-expanded={menuOpen} aria-controls="site-navigation"
          onClick={() => setMenuOpen(!menuOpen)}>{menuOpen ? 'Close' : 'Menu'}</button>
      </div>
    </header>
    <main id="main" className={isDocs ? 'docs-main' : undefined} tabIndex="-1">
      {isDocs ? <Suspense fallback={<p role="status">Loading documentation...</p>}><Documentation hash={hash} /></Suspense> : isTeam ? <Team /> : isArchitecture ? <Architecture /> : isBenchmarks ? <Benchmarks /> : <Home />}
    </main>
    <footer className="site-footer">
      <div className="site-footer-inner">
        <div className="footer-blurb">
          <a className="brand" href="#home"><span className="brand-name">ownframe</span></a>
          <p>Native Go desktop and Android apps styled with HTML and CSS.</p>
        </div>
        <div className="footer-col">
          <h2>Product</h2>
          <a href="#demos">Examples</a>
          <a href="#engine">How it works</a>
          <a href="#compare">Benchmarks</a>
          <a href="#start">Get started</a>
        </div>
        <div className="footer-col">
          <h2>Resources</h2>
          <a href="#docs/why-ownframe">Documentation</a>
          <a href="#docs/showcase/wispr-flow">Example apps</a>
          <a href="#docs/compare-electron">Comparisons</a>
        </div>
        <div className="footer-col">
          <h2>Project</h2>
          <a href={repository}>GitHub</a>
          <a href="#team">About</a>
        </div>
      </div>
      <div className="footer-bottom">
        <span>MIT licensed. Pure Go, open source.</span>
        <span>Linux · macOS · Windows · Android</span>
      </div>
    </footer>
  </div>;
}

createRoot(document.getElementById('root')).render(<React.StrictMode><App /></React.StrictMode>);
