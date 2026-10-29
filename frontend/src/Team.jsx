import React from 'react';

const base = import.meta.env.BASE_URL;

const team = [
  {
    title: 'Muse',
    role: 'System Architecture',
    description: 'Pipeline design, layout abstractions, and host runtime coordination.',
    image: `${base}team/news.webp`,
    alt: 'Muse architecture partner.',
  },
  {
    title: 'Codex',
    role: 'Core Go Engine',
    description: 'Blinkless parser integration, template execution, and Go API interfaces.',
    image: `${base}team/codex.webp`,
    alt: 'Codex implementation partner.',
    light: true,
  },
  {
    title: 'DeepSeek',
    role: 'Performance & Benchmarks',
    description: 'Tile cache optimization, partial repaint algorithms, and GPU layer compositing.',
    image: `${base}team/deepsea.webp`,
    alt: 'DeepSeek performance partner.',
  },
  {
    title: 'OpenCode',
    role: 'Developer Tooling',
    description: 'Syntax highlighting engine, DevTools HUD, and code viewer showcase.',
    image: `${base}team/opencode.png`,
    alt: 'OpenCode tooling partner.',
  },
  {
    title: 'Claude',
    role: 'Verification & Quality',
    description: 'Cross-platform test suites, memory auditing, and API contract verification.',
    image: `${base}team/claude.png`,
    alt: 'Claude quality assurance partner.',
  },
];

export default function Team() {
  return <div className="home">
    <section id="team" className="team-section" aria-labelledby="team-title">
      <h1 id="team-title">This is our team</h1>
      <p>The specialized intelligence and engineering systems driving ownframe development.</p>
      <div className="team-grid">{team.map((member) => <article
        className="team-card" key={member.title}>
        <div className={`team-media${member.light ? ' team-media-light' : ''}`}>
          <img className="team-image" src={member.image} alt={member.alt} loading="lazy" decoding="async" />
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
