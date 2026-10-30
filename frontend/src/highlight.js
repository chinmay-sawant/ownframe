import highlight from 'highlight.js/lib/core';
import go from 'highlight.js/lib/languages/go';
import xml from 'highlight.js/lib/languages/xml';
import css from 'highlight.js/lib/languages/css';

highlight.registerLanguage('go', go);
highlight.registerLanguage('xml', xml);
highlight.registerLanguage('css', css);

const TEMPLATE_PATTERN = /\{\{[\s\S]*?\}\}/g;
const PLACEHOLDER_PREFIX = '__OWNFRAME_TMPL_';

function escapeHtml(value) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;');
}

// Go template actions break XML attribute highlighting because a quoted
// string inside {{...}} ends the attribute early. Protect each action with a
// placeholder before highlighting, then restore it as its own span.
function highlightXmlWithTemplates(source) {
  const templates = [];
  const protectedSource = source.replace(TEMPLATE_PATTERN, (match) => {
    const slot = `${PLACEHOLDER_PREFIX}${templates.length}__`;
    templates.push(match);
    return slot;
  });
  const html = highlight.highlight(protectedSource, { language: 'xml' }).value;
  if (templates.length === 0) return html;
  return templates.reduce(
    (out, template, index) =>
      out.replaceAll(
        `${PLACEHOLDER_PREFIX}${index}__`,
        `<span class="hljs-template-tag">${escapeHtml(template)}</span>`,
      ),
    html,
  );
}

export function highlightSource(source, language) {
  if (language === 'xml' || language === 'html') return highlightXmlWithTemplates(source);
  return highlight.highlight(source, { language }).value;
}
