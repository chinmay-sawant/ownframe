import { defaultUrlTransform } from 'react-markdown';
import { repository } from './project';

const sources = import.meta.glob('../../documentation/*.md', {
  query: '?raw', import: 'default', eager: true,
});
const extras = import.meta.glob([
  '../../examples/readme.md', '../../examples/desktop-cat/readme.md',
  '../../examples/telegram/android/README.md', '../../examples/depth-alert/README.md',
  '../../showcase.md',
], { query: '?raw', import: 'default', eager: true });

const labels = {
  readme: 'Overview',
  showcase: 'Example apps',
  window: 'Open a window',
  screen: 'Rendering',
  binding: 'Data and actions',
  theming: 'Styling and themes',
  platforms: 'Desktop and mobile',
  features: 'Features and limits',
  'android-app-creation': 'Create an Android app',
  'android-build': 'Build for Android',
  'android-runtime': 'Android runtime',
  'examples-telegram-android-readme': 'Telegram on Android',
  'examples-depth-alert-readme': 'Depth Alert on Android',
  'examples-readme': 'Run an example',
  'examples-desktop-cat-readme': 'Desktop cat',
  'features-examples': 'Feature examples',
  'compare-rust-gpui': 'Rust GPUI',
  'compare-electron': 'Electron',
  'compare-go-app-frameworks': 'Go frameworks',
  'compare-syntax': 'Hello world syntax',
  'accessibility-native': 'Native accessibility',
  'desktop-shell-integrations': 'Desktop integration',
  'performance-baseline': 'Performance measurements',
  'runtime-behavior': 'Runtime behavior',
  styleinvalid: 'Style invalidation',
  'migrate-experimental': 'Migrate an older app',
  rebrand: 'Project rename',
  'tricks-tips': 'Tips',
};

export const documents = Object.entries({ ...sources, ...extras }).map(([path, text]) => ({
  path: path.replace('../../', ''),
  slug: path.startsWith('../../documentation/')
    ? path.split('/').pop().replace('.md', '').toLowerCase()
    : path.replace('../../', '').replace(/\.md$/, '').replaceAll('/', '-').toLowerCase(),
  title: text.match(/^# (.+)/m)?.[1] ?? path,
  text,
})).map((entry) => ({ ...entry, title: labels[entry.slug] ?? entry.title })).sort((a, b) => {
  if (a.slug === 'readme') return -1;
  if (b.slug === 'readme') return 1;
  return a.title.localeCompare(b.title);
});

const navigation = [
  { title: 'Start here', slugs: ['readme'], standalone: true },
  { title: 'Window', slugs: ['window'], standalone: true },
  { title: 'Data', slugs: ['binding'], standalone: true },
  { title: 'Styling', slugs: ['theming'], standalone: true },
  { title: 'Platforms', slugs: ['platforms'], standalone: true },
  { title: 'Features', slugs: ['features'], standalone: true },
  {
    title: 'Examples',
    slugs: ['showcase', 'examples-readme', 'examples-telegram-android-readme', 'examples-depth-alert-readme', 'examples-desktop-cat-readme', 'features-examples'],
  },
  {
    title: 'Comparisons',
    slugs: ['compare-rust-gpui', 'compare-electron', 'compare-go-app-frameworks', 'compare-syntax'],
  },
];

export const documentGroups = navigation.map(({ title, slugs, standalone }) => ({
  title,
  standalone,
  documents: slugs.map((slug) => documents.find((entry) => entry.slug === slug)).filter(Boolean),
})).filter((group) => group.documents.length);

const groupedSlugs = new Set(documentGroups.flatMap((group) => group.documents.map((entry) => entry.slug)));
const otherDocuments = documents.filter((entry) => !groupedSlugs.has(entry.slug));
if (otherDocuments.length) documentGroups.push({ title: 'All guides', documents: otherDocuments, keepClosed: true });

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
