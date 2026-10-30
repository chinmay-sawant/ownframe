import React from 'react';

export default function Architecture() {
  return <div className="home detail-page">
    <section id="engine" className="engine-section" aria-labelledby="engine-title">
      <h1 id="engine-title">Your interface, inside your Go process</h1>
      <p>Write the interface in HTML and CSS. Keep state and event handlers in Go. The layout engine and renderer run in the same application process.</p>
      <div className="engine-flow" aria-label="How an ownframe screen reaches the window">
        <div className="engine-step">
          <div className="step-num">01</div>
          <h3>Describe your screen</h3>
          <p>Use Go HTML templates and the supported CSS subset. Pass your application state to the template with SetData.</p>
        </div>
        <div className="engine-step">
          <div className="step-num">02</div>
          <h3>Lay it out in Go</h3>
          <p>Blinkless parses the template, applies CSS, and calculates layout. It returns a display list for ownframe to render.</p>
        </div>
        <div className="engine-step">
          <div className="step-num">03</div>
          <h3>Open a native window</h3>
          <p>Ownframe connects rendering, input, and Go handlers through Ebitengine. Dirty regions and retained layers reduce repeated painting.</p>
        </div>
      </div>
      <p className="section-note">Ownframe supports a subset of HTML and CSS. Website JavaScript does not run inside your app.</p>
    </section>

  </div>;
}
