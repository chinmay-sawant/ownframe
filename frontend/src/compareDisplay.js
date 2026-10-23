// Presentation-only overrides for the comparison docs.
//
// The markdown sources under documentation/ stay complete: they keep the
// OwnFrame snippet and the "Where the code lives" table so docs-check and
// the Go snippet manifest keep compiling. The website hides that repeated
// section for a small set of comparisons and shows only the third-party
// example. This file owns that display rule, so future edits happen here
// instead of in the markdown.

const STRIPPED_COMPARE_SLUGS = new Set([
  'compare-ebitengine',
  'compare-fyne',
  'compare-gio',
  'compare-wails',
  'compare-webview',
]);

// Remove the "### OwnFrame" section and the "## Where the code lives"
// table, keeping anything that follows the table (for example the Wails
// closing paragraph). Returns the input unchanged for other documents.
export function displayTextForDocument(slug, text) {
  if (!STRIPPED_COMPARE_SLUGS.has(slug)) return text;
  const sectionStart = text.indexOf('\n### OwnFrame');
  if (sectionStart === -1) return text;
  const tableHeading = '\n## Where the code lives';
  const tableStart = text.indexOf(tableHeading, sectionStart);
  if (tableStart === -1) return text.slice(0, sectionStart).trimEnd() + '\n';
  const afterHeading = tableStart + tableHeading.length;
  const lines = text.slice(afterHeading).split('\n');
  let index = 0;
  while (index < lines.length && lines[index].trim() === '') index += 1;
  while (index < lines.length && lines[index].trimStart().startsWith('|')) index += 1;
  const rest = lines.slice(index).join('\n').replace(/^\n+/, '');
  const head = text.slice(0, sectionStart).trimEnd();
  return rest ? `${head}\n\n${rest}` : `${head}\n`;
}
