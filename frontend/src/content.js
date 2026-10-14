import { defaultUrlTransform } from 'react-markdown';
import { repository } from './project';

const sources = import.meta.glob('../../documentation/*.md', {
  query: '?raw', import: 'default', eager: true,
});
const extras = import.meta.glob([
  '../../examples/readme.md', '../../examples/desktop-cat/readme.md',
  '../../examples/telegram/android/README.md', '../../showcase.md',
], { query: '?raw', import: 'default', eager: true });

export const documents = Object.entries({ ...sources, ...extras }).map(([path, text]) => ({
  path: path.replace('../../', ''),
  slug: path.startsWith('../../documentation/')
    ? path.split('/').pop().replace('.md', '').toLowerCase()
    : path.replace('../../', '').replace(/\.md$/, '').replaceAll('/', '-').toLowerCase(),
  title: text.match(/^# (.+)/m)?.[1] ?? path,
  text,
})).sort((a, b) => {
  if (a.slug === 'readme') return -1;
  if (b.slug === 'readme') return 1;
  return a.title.localeCompare(b.title);
});

const navigation = [
  { title: 'Start here', slugs: ['readme'], standalone: true },
  {
    title: 'Build a screen',
    slugs: ['window', 'screen', 'forms', 'binding', 'theming', 'interaction', 'editing', 'pointer', 'keys', 'scrolling', 'navigation'],
  },
  {
    title: 'Run your app',
    slugs: ['platforms', 'web', 'android-app-creation', 'android-build', 'android-runtime', 'packaging', 'hot-reload'],
  },
  {
    title: 'Reference',
    slugs: ['features', 'accessibility', 'accessibility-native', 'clipboard', 'drag-drop', 'crash', 'fetch', 'ipc', 'printing', 'emoji', 'desktop-shell-integrations'],
  },
  {
    title: 'Advanced topics',
    slugs: ['frames', 'repaint', 'devtools', 'performance', 'performance-baseline', 'runtime-behavior', 'styleinvalid'],
  },
  {
    title: 'Examples and comparisons',
    slugs: ['showcase', 'examples-readme', 'examples-desktop-cat-readme', 'examples-telegram-android-readme', 'features-examples', 'compare-electron', 'compare-go-app-frameworks', 'compare-rust-gpui', 'compare-syntax', 'migrate-experimental', 'rebrand', 'tricks-tips'],
  },
];

export const documentGroups = navigation.map(({ title, slugs, standalone }) => ({
  title,
  standalone,
  documents: slugs.map((slug) => documents.find((entry) => entry.slug === slug)).filter(Boolean),
})).filter((group) => group.documents.length);

const groupedSlugs = new Set(documentGroups.flatMap((group) => group.documents.map((entry) => entry.slug)));
const otherDocuments = documents.filter((entry) => !groupedSlugs.has(entry.slug));
if (otherDocuments.length) documentGroups.push({ title: 'More topics', documents: otherDocuments });

const images = import.meta.glob([
  '../../assets/**/*.webp', '../../assets/**/*.jpg', '../../assets/**/*.png',
], { import: 'default', eager: true });
export const preview = images['../../assets/screenshots/demos/ownframe-preview.webp'];
export const catPreview = images['../../assets/screenshots/demos/desktop-cat.webp'];
const sourceBase = `${repository}/blob/master/`;
const sourcePath = new URL(sourceBase).pathname;

export function contentUrl(href, document) {
  const safe = defaultUrlTransform(href);
  if (!safe) return '';
  if (/^(https?:|mailto:)/.test(safe) || safe.startsWith('//')) return safe;
  const url = new URL(safe, sourceBase + document.path);
  const path = decodeURIComponent(url.pathname.replace(sourcePath, ''));
  const target = documents.find((entry) => entry.path === path);
  if (target) return `#docs/${target.slug}${url.hash ? '/' + url.hash.slice(1) : ''}`;
  if (path.startsWith('frontend/public/')) {
    return `${import.meta.env.BASE_URL}${path.slice('frontend/public/'.length)}`;
  }
  const image = images[`../../${path}`];
  return image || url.href;
}
