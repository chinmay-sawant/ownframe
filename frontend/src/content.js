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
  const image = images[`../../${path}`];
  return image || url.href;
}
