import React, { useEffect, useRef, useState } from 'react';
import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeSlug from 'rehype-slug';
import rehypeHighlight from 'rehype-highlight';
import { documents, documentGroups, contentUrl } from './content';
import { docsOverview } from './docs-overview';

const showcaseSections = [
  { id: 'telegram', title: 'Telegram' },
  { id: 'teams', title: 'Teams' },
  { id: 'spotify-player', title: 'Spotify Player' },
  { id: 'pet-care', title: 'Pet Care' },
  { id: 'digital-wellbeing', title: 'Digital Wellbeing' },
  { id: 'desktop-overlays', title: 'Desktop overlays' },
  { id: 'wispr-flow', title: 'Wispr Flow' },
  { id: 'dino-run', title: 'Dino Run' },
];

function selectShowcase(markdown, id) {
  const [intro, ...sections] = markdown.split(/(?=^## )/m);
  const selected = sections.find((section) => {
    const title = section.match(/^## (.+)$/m)?.[1];
    return title?.toLowerCase().replace(/\s+/g, '-') === id;
  });
  return [intro, selected].filter(Boolean).join('\n');
}

export default function Documentation({ hash }) {
  const [menuOpen, setMenuOpen] = useState(false);
  const contentPane = useRef(null);
  const menuButton = useRef(null);
  const [, slug = 'readme', anchor] = hash.split('/');
  const document = documents.find((entry) => entry.slug === slug);
  const isShowcase = document?.slug === 'showcase';
  const activeShowcase = showcaseSections.some((entry) => entry.id === anchor) ? anchor : 'telegram';
  const text = isShowcase ? selectShowcase(document.text, activeShowcase)
    : slug === 'readme' && !anchor ? docsOverview : document?.text;

  useEffect(() => {
    setMenuOpen(false);
    const pane = contentPane.current;
    const target = anchor && pane.querySelector(`#${CSS.escape(anchor)}`);
    if (target) target.scrollIntoView();
    else window.scrollTo(0, 0);
  }, [hash, anchor]);

  const components = {
    table: ({ children }) => <div className="markdown-table-scroll"><table>{children}</table></div>,
    img: ({ src, alt = '', title }) => <img src={src} alt={alt} title={title}
      className={src?.includes('/spotify.webp') ? 'spotify-window-preview' : undefined}
      loading="lazy" decoding="async" />,
    ...(isShowcase ? {
      ul: ({ children }) => <ul className={`showcase-gallery${activeShowcase === 'pet-care' ? ' pet-care-gallery' : ''}`}>{children}</ul>,
    } : {}),
  };

  return <div className="docs-layout">
    <aside data-open={menuOpen} onKeyDown={(event) => {
      if (event.key === 'Escape' && menuOpen) {
        setMenuOpen(false);
        menuButton.current.focus();
      }
    }}>
      <button ref={menuButton} className="docs-menu-toggle" type="button" aria-expanded={menuOpen}
        aria-controls="documentation-topics" onClick={() => setMenuOpen(!menuOpen)}>
        Browse documentation
      </button>
      <nav id="documentation-topics" className="docs-nav" aria-label="Documentation topics"
        onClick={(event) => { if (event.target.closest('a')) setMenuOpen(false); }}>
        {documentGroups.map((group) => group.standalone
          ? <a key={group.title} href={`#docs/${group.documents[0].slug}`}
            aria-current={slug === group.documents[0].slug ? 'page' : undefined}
            className="docs-nav-parent">{group.documents[0].title}</a>
          : <details key={`${group.title}/${slug}`} className="docs-nav-group"
            open={!group.keepClosed && group.documents.some((entry) => entry.slug === slug)}>
            <summary>{group.title}</summary>
            <div className="docs-nav-links">
              {group.documents.map((entry) => <div key={entry.slug} className="docs-nav-entry">
                <a href={entry.slug === 'showcase' ? '#docs/showcase/telegram' : `#docs/${entry.slug}`}
                  aria-current={slug === entry.slug && entry.slug !== 'showcase' ? 'page' : undefined}
                  className={slug === 'showcase' && entry.slug === 'showcase' ? 'docs-nav-parent active' : 'docs-nav-parent'}>
                  {entry.title}
                </a>
              </div>)}
            </div>
          </details>)}
      </nav>
    </aside>
    <article ref={contentPane} className={`markdown docs-content${isShowcase ? ' showcase-page' : ''}`}>
      {isShowcase && <nav className="showcase-picker" aria-label="Example apps">
        {showcaseSections.map((section) => <a key={section.id}
          href={`#docs/showcase/${section.id}`}
          aria-current={activeShowcase === section.id ? 'page' : undefined}>
          {section.title}
        </a>)}
      </nav>}
      {document ? <Markdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeSlug, rehypeHighlight]}
        urlTransform={(url) => contentUrl(url, document)} components={components}>{text}</Markdown>
        : <><h1>Document not found</h1><p><a href="#docs/readme">Open the documentation overview</a></p></>}
    </article>
  </div>;
}
