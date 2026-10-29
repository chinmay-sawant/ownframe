import React from 'react';

const base = import.meta.env.BASE_URL;

const team = [
  {
    title: 'Muse',
    description: 'Muse.',
    image: `${base}team/news.webp`,
    alt: 'Muse logo provided by the team.',
  },
  {
    title: 'Codex',
    description: 'Codex.',
    image: `${base}team/codex.webp`,
    alt: 'Codex logo provided by the team.',
    light: true,
  },
  {
    title: 'DeepSeek',
    description: 'DeepSeek.',
    image: `${base}team/deepsea.webp`,
    alt: 'DeepSeek image provided by the team.',
  },
  {
    title: 'OpenCode',
    description: 'OpenCode.',
    image: `${base}team/opencode.png`,
    alt: 'OpenCode wordmark provided by the team.',
  },
  {
    title: 'Claude',
    description: 'Claude.',
    image: `${base}team/claude.png`,
    alt: 'Claude image provided by the team.',
  },
];

export default function Team() {
  return <div className="home">
    <section id="team" className="team-section" aria-labelledby="team-title">
      <h1 id="team-title">This is our team</h1>
      <p>This is our team.</p>
      <div className="team-grid">{team.map((member) => <article
        className="team-card" key={member.title}>
        <div className={`team-media${member.light ? ' team-media-light' : ''}`}>
          <img className="team-image" src={member.image} alt={member.alt} loading="lazy" decoding="async" />
        </div>
        <div className="team-card-content">
          <h3>{member.title}</h3>
          <p>{member.description}</p>
        </div>
      </article>)}</div>
    </section>
  </div>;
}
