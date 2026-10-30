import React from 'react';
import { repository } from './project';

export default function Benchmarks() {
  return <div className="home detail-page">
    <section id="compare" className="comparison-section" aria-labelledby="compare-title">
      <p className="eyebrow">Measured, with context</p>
      <h1 id="compare-title">Small updates should stay small.</h1>
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

  </div>;
}
