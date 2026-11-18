import React from 'react';

const steps = [
  { n: '01', title: 'Describe', body: 'Use Go HTML templates and the supported CSS subset. Pass your application state to the template with SetData.' },
  { n: '02', title: 'Lay out', body: 'Blinkless parses the template, applies CSS, and calculates layout. It returns a display list for ownframe to render.' },
  { n: '03', title: 'Window', body: 'Ownframe connects rendering, input, and Go handlers through Ebitengine. Dirty regions and retained layers reduce repeated painting.' },
];

const stages = ['Go state', 'Template', 'Blinkless layout', 'Display list', 'Ebitengine window'];

export default function Architecture() {
  return <div className="p-page">
    <section id="engine" className="container p-head" aria-labelledby="engine-title">
      <p className="eyebrow">How it works</p>
      <h1 id="engine-title" className="p-title">Your interface, inside your Go process</h1>
      <p className="lede">Write the interface in HTML and CSS. Keep state and event handlers in Go. The layout engine and renderer run in the same application process.</p>
    </section>
    <section className="container p-block" aria-label="How an ownframe screen reaches the window">
      <ol className="p-stages" aria-label="Pipeline stages">
        {stages.map((s, i) => <li key={s} className="p-stage">
          <span className="p-stage-label">{s}</span>
          {i < stages.length - 1 && <svg className="p-arrow" width="28" height="12" viewBox="0 0 28 12" aria-hidden="true" focusable="false"><path d="M0 6h26M21 1l5 5-5 5" fill="none" stroke="currentColor" strokeWidth="1.25" strokeLinecap="round" strokeLinejoin="round" /></svg>}
        </li>)}
      </ol>
      <ol className="p-steps">
        {steps.map((s) => <li key={s.n} className="p-step">
          <span className="p-step-num">{s.n}</span>
          <h2 className="p-step-title">{s.title}</h2>
          <p>{s.body}</p>
        </li>)}
      </ol>
      <p className="p-note">Ownframe supports a subset of HTML and CSS. Website JavaScript does not run inside your app.</p>
    </section>
  </div>;
}
