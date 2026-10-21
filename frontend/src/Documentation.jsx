import React, { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import rehypeSlug from 'rehype-slug';
import rehypeHighlight from 'rehype-highlight';
import { documents, documentGroups, contentUrl } from './content';
import { docsOverview } from './docs-overview';
import { showcaseIntro, showcaseSections } from './showcase';

// useReducedMotion reports the reduced motion preference and follows it when
// the reader changes it while the page is open.
function useReducedMotion() {
  const [reduced, setReduced] = useState(() => Boolean(window.matchMedia?.('(prefers-reduced-motion: reduce)')?.matches));
  useEffect(() => {
    const query = window.matchMedia?.('(prefers-reduced-motion: reduce)');
    if (!query) return undefined;
    const onChange = (event) => setReduced(event.matches);
    query.addEventListener('change', onChange);
    return () => query.removeEventListener('change', onChange);
  }, []);
  return reduced;
}

// ShowcasePreview shows a preview clip that plays on its own. The clip gets a
// pause control, it stays still for a reader who asked for reduced motion,
// and it stays in the accessibility tree because it carries the preview.
function ShowcasePreview({ media, doc, onOpen }) {
  const videoRef = useRef(null);
  const reduced = useReducedMotion();
  const [playing, setPlaying] = useState(!reduced);

  useEffect(() => {
    if (reduced) setPlaying(false);
  }, [reduced]);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;
    if (!playing) {
      video.pause();
      return;
    }
    video.play().catch(() => setPlaying(false));
  }, [playing]);

  return <>
    <button type="button" className="gallery-open" aria-label={`Open media: ${media.alt}`}
      aria-haspopup="dialog" onClick={onOpen}>
      <video ref={videoRef} src={`${import.meta.env.BASE_URL}${media.video}`}
        poster={contentUrl(media.src, doc)} autoPlay={!reduced} muted loop playsInline
        preload={reduced ? 'none' : 'auto'} aria-label={media.alt}
        width={media.width} height={media.height}
        style={{ width: media.nativeDisplay
          ? `min(100%, calc(${media.width}px / var(--site-scale)))`
          : `min(100%, ${media.width}px)` }} />
    </button>
    <button type="button" className="showcase-media-control" onClick={() => setPlaying(!playing)}>
      {playing ? 'Pause preview' : 'Play preview'}
    </button>
  </>;
}

export default function Documentation({ hash }) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [lightbox, setLightbox] = useState(null);
  const contentPane = useRef(null);
  const menuButton = useRef(null);
  const [, slug = 'readme', anchor] = hash.split('/');
  const doc = documents.find((entry) => entry.slug === slug);
  const isShowcase = doc?.slug === 'showcase';
  const activeShowcase = showcaseSections.some((entry) => entry.id === anchor) ? anchor : showcaseSections[0].id;
  const showcase = isShowcase ? showcaseSections.find((entry) => entry.id === activeShowcase) : null;
  const text = slug === 'readme' && !anchor ? docsOverview : doc?.text;
  const mediaResolution = (src, width, height) => setLightbox((box) => {
    if (!width || !height || !box || box.items[box.index]?.src !== src) return box;
    const items = box.items.slice();
    items[box.index] = { ...items[box.index], width, height, resolution: `${width} × ${height}` };
    return { ...box, items };
  });

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
  }, [hash, anchor, isShowcase]);

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

  const openLightbox = (index) => {
    const items = showcase.media.map((media) => {
      const poster = contentUrl(media.src, doc);
      const src = media.video ? `${import.meta.env.BASE_URL}${media.video}` : poster;
      return {
        ...media,
        src,
        poster,
        type: media.video ? 'video' : 'image',
        width: media.video ? undefined : media.width,
        height: media.video ? undefined : media.height,
        resolution: media.video ? '' : `${media.width} × ${media.height}`,
      };
    });
    if (items.length) setLightbox({ items, index });
  };

  const components = {
    table: ({ children, className }) => <div className={`markdown-table-scroll${className ? ` ${className}-scroll` : ''}`}>
      <table className={className}>{children}</table>
    </div>,
    img: ({ src, alt = '', title, width, height }) => {
      const w = Number(width);
      const h = Number(height);
      const fixed = Number.isFinite(w) && w > 0 && Number.isFinite(h) && h > 0;
      return <img src={src} alt={alt} title={title}
        width={fixed ? w : undefined} height={fixed ? h : undefined}
        style={fixed ? { maxWidth: `min(${w}px, 100%)`, maxHeight: `${h}px`, margin: '0 auto' } : undefined}
        loading="lazy" decoding="async" />;
    },
  };

  const markdownProps = {
    remarkPlugins: [remarkGfm],
    rehypePlugins: [rehypeRaw, rehypeSlug, rehypeHighlight],
    urlTransform: (url) => contentUrl(url, doc),
    components,
  };
  const lightboxItem = lightbox?.items[lightbox.index];

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
                <a href={entry.slug === 'showcase' ? `#docs/showcase/${showcaseSections[0].id}` : `#docs/${entry.slug}`}
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
          aria-current={activeShowcase === section.id ? 'true' : undefined}>
          {section.title}
        </a>)}
      </nav>}
      {doc ? isShowcase ? <>
        <h1>{doc.title}</h1>
        <p>{showcaseIntro}</p>
        <h2>{showcase.title}</h2>
        <Markdown {...markdownProps}>{showcase.before}</Markdown>
        <div className={`showcase-media-grid showcase-media-grid--${showcase.layout}`}>
          {showcase.media.map((media, index) => <figure key={media.src}
            className={`showcase-media${media.wide ? ' showcase-media--wide' : ''}`}>
            {media.autoplayPreview ? <ShowcasePreview media={media} doc={doc}
              onOpen={() => openLightbox(index)} /> : <button type="button" className="gallery-open"
              aria-label={`Open media: ${media.alt}`} aria-haspopup="dialog" onClick={() => openLightbox(index)}>
              <img src={contentUrl(media.src, doc)} alt={media.alt} width={media.width} height={media.height}
                loading="lazy" decoding="async" style={{ width: media.nativeDisplay
                  ? `min(100%, calc(${media.width}px / var(--site-scale)))`
                  : `min(100%, ${media.width}px)` }} />
            </button>}
            <figcaption>{media.caption}</figcaption>
          </figure>)}
        </div>
        <Markdown {...markdownProps}>{showcase.after}</Markdown>
      </> : <Markdown {...markdownProps}>{text}</Markdown>
        : <><h1>Document not found</h1><p><a href="#docs/readme">Open the documentation overview</a></p></>}
    </article>
    {lightboxItem && createPortal(<div className="lightbox-backdrop" onClick={() => setLightbox(null)}>
      <div className="lightbox-card" role="dialog" aria-modal="true"
        aria-label={`Expanded media: ${lightboxItem.alt}`}>
        {lightboxItem.type === 'video'
          ? <video key={lightboxItem.src} src={lightboxItem.src}
            poster={lightboxItem.poster} autoPlay controls muted loop playsInline preload="auto"
            width={lightboxItem.width} height={lightboxItem.height}
            aria-label={lightboxItem.alt}
            onLoadedMetadata={(event) => mediaResolution(lightboxItem.src,
              event.currentTarget.videoWidth, event.currentTarget.videoHeight)}
            onClick={(event) => event.stopPropagation()} />
          : <img key={lightboxItem.src}
            src={lightboxItem.src} alt={lightboxItem.alt}
            width={lightboxItem.width} height={lightboxItem.height}
            onClick={(event) => event.stopPropagation()} />}
        <div className="lightbox-bar" onClick={(event) => event.stopPropagation()}>
          {lightbox.items.length > 1 && <button type="button" className="lightbox-prev" aria-label="Show previous media"
            onClick={() => setLightbox((box) => box && {
              ...box, index: (box.index - 1 + box.items.length) % box.items.length,
            })}>‹</button>}
          <p className="lightbox-caption">{lightboxItem.alt}
            {lightbox.items.length > 1 && ` · ${lightbox.index + 1} / ${lightbox.items.length}`}</p>
          {lightboxItem.resolution && <span className="lightbox-resolution">{lightboxItem.resolution}</span>}
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
