import React from 'react';

const base = import.meta.env.BASE_URL;

const team = [
  {
    title: 'Muse',
    role: 'AI development tool',
    description: 'Used during AI-assisted development. The maintainer reviews the resulting changes.',
    image: `${base}team/news.webp`,
    alt: 'Muse tool artwork.',
  },
  {
    title: 'Codex',
    role: 'AI coding tool',
    description: 'Used for implementation, code exploration, and verification tasks.',
    image: `${base}team/codex.webp`,
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
  },
  {
    title: 'Claude',
    role: 'Not used',
    description: 'No Claude models were used during development. The repo tag refers to the Space Bunny commit, not Claude AI.',
    image: `${base}team/claude.png`,
    alt: 'Meme artwork stating Claude is not used here.',
  },
];

export default function Team() {
  return <div className="home">
    <section id="team" className="team-section" aria-labelledby="team-title">
      <p className="eyebrow">People & tools</p>
      <h1 id="team-title">Built by a developer, with AI tools.</h1>
      <p>Ownframe is Chinmay Sawant&apos;s open-source experiment in building native interfaces with Go, HTML, and CSS.</p>
      <article className="maintainer-card">
        <img className="maintainer-avatar" src={`${base}team/chinmay-x.jpg`}
          alt="Chinmay Sawant's X profile photo" width="72" height="72" />
        <div><h2>Chinmay Sawant</h2><p>Creator and maintainer</p>
          <div className="maintainer-links">
            <a href="https://x.com/chinmay_sawant_">@chinmay_sawant_ on X <span aria-hidden="true">↗</span></a>
            <a href="https://github.com/chinmay-sawant">GitHub <span aria-hidden="true">↗</span></a>
          </div></div>
      </article>
      <h2 className="tools-title">Tools in the workflow</h2>
      <p>AI-assisted development is part of the experiment. These are tools, with generated code reviewed and tested before use.</p>
      <div className="team-grid">{team.map((member) => <article
        className="team-card" key={member.title}>
        <div className={`team-media${member.light ? ' team-media-light' : ''}`}>
          <img className="team-image" src={member.image} alt={member.alt} width="480" height="270" loading="lazy" decoding="async" />
        </div>
        <div className="team-card-content">
          <div className="team-card-header">
            <h3>{member.title}</h3>
            <span className="team-role">{member.role}</span>
          </div>
          <p>{member.description}</p>
        </div>
      </article>)}</div>
    </section>
  </div>;
}
