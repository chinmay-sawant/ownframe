import React, { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
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
  { id: 'eyes', title: 'Eyes' },
  { id: 'depth-alert', title: 'Depth Alert' },
];

const showcaseVideoPaths = {
  'Animated preview of the Teams chat and collaboration app': 'teams.mp4',
  'Desktop cat overlay preview': 'desktop-cat.mp4',
  'Wispr Flow example preview': 'preview.mp4',
};

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
  const [lightbox, setLightbox] = useState(null);
  const contentPane = useRef(null);
  const menuButton = useRef(null);
  const [, slug = 'readme', anchor] = hash.split('/');
  const doc = documents.find((entry) => entry.slug === slug);
  const isShowcase = doc?.slug === 'showcase';
  const activeShowcase = showcaseSections.some((entry) => entry.id === anchor) ? anchor : 'telegram';
  const text = isShowcase ? selectShowcase(doc.text, activeShowcase)
    : slug === 'readme' && !anchor ? docsOverview : doc?.text;

  useEffect(() => {
    setMenuOpen(false);
    setLightbox(null);
    if (isShowcase) {
      window.scrollTo(0, 0);
      return;
    }
    const pane = contentPane.current;
    const target = anchor && pane.querySelector(`#${CSS.escape(anchor)}`);
    if (target) target.scrollIntoView();
    else window.scrollTo(0, 0);
  }, [hash, anchor]);

  useEffect(() => {
    if (!lightbox) return;
    const step = (direction) => setLightbox((box) => box && {
      ...box, index: (box.index + direction + box.items.length) % box.items.length,
    });
    const onKey = (event) => {
      if (event.key === 'Escape') setLightbox(null);
      if (event.key === 'ArrowRight') step(1);
      if (event.key === 'ArrowLeft') step(-1);
    };
    window.addEventListener('keydown', onKey);
    const previous = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      window.removeEventListener('keydown', onKey);
      document.body.style.overflow = previous;
    };
  }, [lightbox]);

  const openLightbox = (event) => {
    const showcase = event.currentTarget.closest('.showcase-page');
    const buttons = showcase ? Array.from(showcase.querySelectorAll('button.gallery-open')) : [];
    const items = buttons.map((button) => {
      const image = button.querySelector('img');
      const poster = image?.currentSrc || image?.src;
      const video = button.dataset.videoSrc;
      return { src: video || poster, poster, type: video ? 'video' : 'image', alt: image?.alt || '' };
    });
    const index = Math.max(0, buttons.indexOf(event.currentTarget));
    if (items.length) setLightbox({ items, index });
  };

  const components = {
    table: ({ children, className }) => <div className={`markdown-table-scroll${className ? ` ${className}-scroll` : ''}`}>
      <table className={className}>{children}</table>
    </div>,
    img: ({ src, alt = '', title, width, height, 'data-video': video }) => {
      const w = Number(width);
      const h = Number(height);
      const fixed = Number.isFinite(w) && w > 0 && Number.isFinite(h) && h > 0;
      const videoSrc = video ? contentUrl(video, doc) : showcaseVideoPaths[alt]
        ? `${import.meta.env.BASE_URL}demos/${showcaseVideoPaths[alt]}` : undefined;
      const image = <img src={src} alt={alt} title={title}
        className={src?.includes('/spotify.webp') ? 'spotify-window-preview' : undefined}
        width={fixed ? w : undefined} height={fixed ? h : undefined}
        style={fixed ? { maxWidth: `min(${w}px, 100%)`, maxHeight: `${h}px`, margin: '0 auto' } : undefined}
        loading="lazy" decoding="async" />;
      if (!isShowcase) return image;
      return <button type="button" className="gallery-open" aria-label={`Open media: ${alt}`}
        data-video-src={videoSrc}
        onClick={openLightbox}>{image}</button>;
    },
    ...(isShowcase ? {
      ul: ({ children }) => {
        const extra = activeShowcase === 'pet-care' || activeShowcase === 'eyes' || activeShowcase === 'depth-alert'
          ? ` ${activeShowcase}-gallery` : '';
        return <ul className={`showcase-gallery${extra}`}>{children}</ul>;
      },
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
      {doc ? <Markdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeRaw, rehypeSlug, rehypeHighlight]}
        urlTransform={(url) => contentUrl(url, doc)} components={components}>{text}</Markdown>
        : <><h1>Document not found</h1><p><a href="#docs/readme">Open the documentation overview</a></p></>}
    </article>
    {lightbox && createPortal(<div className="lightbox-backdrop" onClick={() => setLightbox(null)}>
      <div className="lightbox-card" role="dialog" aria-modal="true"
        aria-label={`Expanded media: ${lightbox.items[lightbox.index].alt}`}>
        {lightbox.items[lightbox.index].type === 'video'
          ? <video key={lightbox.items[lightbox.index].src} src={lightbox.items[lightbox.index].src}
            poster={lightbox.items[lightbox.index].poster} autoPlay controls muted loop playsInline preload="auto"
            aria-label={lightbox.items[lightbox.index].alt}
            onClick={(event) => event.stopPropagation()} />
          : <img key={lightbox.items[lightbox.index].src}
            src={lightbox.items[lightbox.index].src} alt={lightbox.items[lightbox.index].alt}
            onClick={(event) => event.stopPropagation()} />}
        <div className="lightbox-bar" onClick={(event) => event.stopPropagation()}>
          {lightbox.items.length > 1 && <button type="button" className="lightbox-prev" aria-label="Show previous media"
            onClick={() => setLightbox((box) => box && {
              ...box, index: (box.index - 1 + box.items.length) % box.items.length,
            })}>‹</button>}
          <p className="lightbox-caption">{lightbox.items[lightbox.index].alt}
            {lightbox.items.length > 1 && ` · ${lightbox.index + 1} / ${lightbox.items.length}`}</p>
          {lightbox.items.length > 1 && <button type="button" className="lightbox-next" aria-label="Show next media"
            onClick={() => setLightbox((box) => box && {
              ...box, index: (box.index + 1) % box.items.length,
            })}>›</button>}
          <button type="button" className="lightbox-close" aria-label="Close expanded media"
            autoFocus onClick={() => setLightbox(null)}>×</button>
        </div>
      </div>
    </div>, document.body)}
  </div>;
}
