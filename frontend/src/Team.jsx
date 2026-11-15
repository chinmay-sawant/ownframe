import React from 'react';

const base = import.meta.env.BASE_URL;

const team = [
  {
    title: 'Muse Spark 1.3',
    role: 'via OpenCode',
    description: 'Used during development. The maintainer reviews the resulting changes.',
    image: `${base}team/news.webp`,
    alt: 'Muse tool artwork.',
    light: true,
  },
  {
    title: 'Codex',
    role: 'AI coding tool',
    description: 'Used for implementation, code exploration, and verification tasks.',
    image: `${base}team/codex-white.png`,
    alt: 'Codex tool artwork.',
    light: true,
  },
  {
    title: 'DeepSeek',
    role: 'AI development tool',
    description: 'Part of the AI-assisted development workflow.',
    image: `${base}team/deepsea.webp`,
    alt: 'DeepSeek tool artwork.',
  },
  {
    title: 'OpenCode',
    role: 'Coding agent interface',
    description: 'A tool for working with coding agents during development.',
    image: `${base}team/opencode.png`,
    alt: 'OpenCode tool artwork.',
    dark: true,
  },
  {
    title: 'Claude',
    role: 'AI coding tool',
    description: 'Used for implementation and the website redesign. The maintainer reviews the resulting changes.',
  },
];

export default function Team() {
  return <div className="p-page">
    <section id="team" className="container p-head" aria-labelledby="team-title">
      <p className="eyebrow">People &amp; tools</p>
      <h1 id="team-title" className="p-title">Built by a solo developer, with AI tools.</h1>
      <p className="lede">Ownframe is Chinmay Sawant&apos;s open-source experiment in building native interfaces with Go, HTML, and CSS.</p>
    </section>
    <section className="container p-block">
      <article className="card p-maint">
        <img className="p-avatar" src={`${base}team/chinmay-x.jpg`}
          alt="Chinmay Sawant's X profile photo" width="72" height="72" />
        <div><h2 className="p-maint-name">Chinmay Sawant</h2><p className="p-muted">Creator and maintainer</p>
          <div className="p-maint-links">
            <a href="https://x.com/chinmay_sawant_">@chinmay_sawant_ on X <span aria-hidden="true">↗</span></a>
            <a href="https://github.com/chinmay-sawant">GitHub <span aria-hidden="true">↗</span></a>
          </div></div>
      </article>
      <h2 className="p-sub">Tools in the workflow</h2>
      <p className="p-muted p-sub-lede">AI-assisted development is part of the experiment. These are tools, with generated code reviewed and tested before use.</p>
      <div className="p-tools">{team.map((member) => <article className="p-tool" key={member.title}>
        <div className={`p-tool-media${member.light ? ' is-light' : ''}${member.dark ? ' is-dark' : ''}`}>
          {member.image && <img src={member.image} alt={member.alt} width="480" height="270" decoding="async" />}
        </div>
        <div className="p-tool-head"><h3>{member.title}</h3><span className="p-tool-role">{member.role}</span></div>
        <p className="p-muted">{member.description}</p>
      </article>)}</div>
    </section>
  </div>;
}
