import { defaultUrlTransform } from 'react-markdown';
import { repository } from './project';
import { displayTextForDocument } from './compareDisplay';

const sources = import.meta.glob('../../documentation/*.md', {
  query: '?raw', import: 'default', eager: true,
});
const extras = import.meta.glob([
  '../../examples/readme.md', '../../examples/desktop-cat/readme.md',
  '../../examples/telegram/android/README.md',
], { query: '?raw', import: 'default', eager: true });

const labels = {
  showcase: 'Example apps',
  'why-ownframe': 'Why ownframe exists',
  'examples-readme': 'Feature example',
  'compare-rust-gpui': 'Rust GPUI',
  'compare-electron': 'Electron',
  'compare-wails': 'Wails',
  'compare-fyne': 'Fyne',
  'compare-gio': 'Gio',
  'compare-webview': 'webview_go',
  'compare-ebitengine': 'Ebitengine',
};

const keptSlugs = new Set([
  'showcase', 'examples-readme', 'why-ownframe',
  'compare-rust-gpui', 'compare-electron',
  'compare-wails', 'compare-fyne', 'compare-gio', 'compare-webview',
  'compare-ebitengine',
]);

const markdownDocuments = Object.entries({ ...sources, ...extras }).map(([path, text]) => {
  const slug = path.startsWith('../../documentation/')
    ? path.split('/').pop().replace('.md', '').toLowerCase()
    : path.replace('../../', '').replace(/\.md$/, '').replaceAll('/', '-').toLowerCase();
  return {
    path: path.replace('../../', ''),
    slug,
    title: text.match(/^# (.+)/m)?.[1] ?? path,
    text: displayTextForDocument(slug, text),
  };
}).map((entry) => ({ ...entry, title: labels[entry.slug] ?? entry.title }));

export const documents = [...markdownDocuments.filter((entry) => keptSlugs.has(entry.slug)), {
  path: 'showcase.md',
  slug: 'showcase',
  title: labels.showcase,
  text: '',
}].sort((a, b) => a.title.localeCompare(b.title));

const navigation = [
  {
    title: 'About',
    slugs: ['why-ownframe'],
  },
  {
    title: 'Examples',
    slugs: ['showcase', 'examples-readme'],
  },
  {
    title: 'Comparisons',
    slugs: ['compare-rust-gpui', 'compare-electron', 'compare-wails', 'compare-fyne', 'compare-gio', 'compare-webview', 'compare-ebitengine'],
  },
];

export const documentGroups = navigation.map(({ title, slugs }) => ({
  title,
  documents: slugs.map((slug) => documents.find((entry) => entry.slug === slug)).filter(Boolean),
})).filter((group) => group.documents.length);

const images = import.meta.glob([
  '../../assets/screenshots/**/*.webp', '../../assets/screenshots/**/*.jpg', '../../assets/screenshots/**/*.png',
], { import: 'default', eager: true });
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
