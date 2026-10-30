import highlight from 'highlight.js/lib/core';
import go from 'highlight.js/lib/languages/go';
import xml from 'highlight.js/lib/languages/xml';
import css from 'highlight.js/lib/languages/css';
import mainSource from '../../examples/frontend/workspace/main.go?raw';
import templateSource from '../../examples/frontend/workspace/template.html?raw';
import styleSource from '../../examples/frontend/workspace/style.css?raw';

highlight.registerLanguage('go', go);
highlight.registerLanguage('xml', xml);
highlight.registerLanguage('css', css);

export const starterFiles = [
  { name: 'Live demo', url: `${import.meta.env.BASE_URL}demos/workspace-inspector/index.html` },
  ...[
    { name: 'main.go', language: 'go', source: mainSource },
    { name: 'template.html', language: 'xml', source: templateSource },
    { name: 'style.css', language: 'css', source: styleSource },
  ].map(file => ({ ...file, html: highlight.highlight(file.source, { language: file.language }).value })),
];
