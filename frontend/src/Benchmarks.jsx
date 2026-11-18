import React from 'react';
import { repository } from './project';

const metrics = [
  { n: '176 ns', t: 'Animated card update', d: 'The recorded animated-card frame operation, after replacing a full redraw.' },
  { n: '0 allocs', t: 'For that card update', d: 'Heap allocations per operation in the same animated-card measurement.' },
  { n: '143 µs', t: 'Small-fixture warm redraw', d: 'The recorded RedrawWarm stage benchmark from the v0.0.2 baseline.' },
];

export default function Benchmarks() {
  return <div className="p-page">
    <section id="compare" className="container p-head" aria-labelledby="compare-title">
      <p className="eyebrow">Measured, with context</p>
      <h1 id="compare-title" className="p-title">Small updates should stay small.</h1>
      <p className="lede">Retained operations let an animation update its paint without rebuilding the whole document. These historical measurements cover specific fixtures, not complete GPU presentation or every application.</p>
    </section>
    <section className="container p-block">
      <ul className="p-metrics">
        {metrics.map((m) => <li key={m.n} className="p-metric">
          <span className="p-metric-num">{m.n}</span>
          <h2 className="p-metric-title">{m.t}</h2>
          <p>{m.d}</p>
        </li>)}
      </ul>
      <p className="p-note">Historical v0.0.2 results on the author&apos;s machine. Workload, commit, environment, and limits are recorded in the <a href={`${repository}/blob/master/documentation/performance-baseline.md`}>performance baseline</a>. See the <a href={`${repository}/blob/master/documentation/performance.md`}>measurement procedure</a> to reproduce the checks.</p>
      <div className="p-table-wrap" role="region" aria-label="Runtime architecture comparison" tabIndex="0">
        <table className="p-table">
          <caption className="visually-hidden">Runtime architecture of ownframe and browser-based approaches</caption>
          <thead><tr><th scope="col">Inside the app</th><th scope="col">Electron</th><th scope="col">Wails v2 / Tauri</th><th scope="col" className="p-col-hl">ownframe</th></tr></thead>
          <tbody>
            <tr><th scope="row">Layout</th><td>Bundled Chromium</td><td>System WebView</td><td className="p-col-hl"><strong>Blinkless, in Go</strong></td></tr>
            <tr><th scope="row">UI event handlers</th><td>JavaScript</td><td>JavaScript and a native bridge</td><td className="p-col-hl"><strong>Go functions</strong></td></tr>
            <tr><th scope="row">Web compatibility</th><td>Chromium web platform</td><td>Host WebView web platform</td><td className="p-col-hl"><strong>Supported HTML and CSS subset</strong></td></tr>
          </tbody>
        </table>
      </div>
      <p className="p-links"><a href="#docs/compare-electron">Read the Electron comparison</a> <span aria-hidden="true">·</span> <a href="#docs/why-ownframe">Why we built ownframe</a></p>
    </section>
  </div>;
}
