import highlight from 'highlight.js/lib/core';
import go from 'highlight.js/lib/languages/go';
import xml from 'highlight.js/lib/languages/xml';
import css from 'highlight.js/lib/languages/css';
import helloGo from '../../examples/frontend/hello/main.go?raw';
import helloHTML from '../../examples/frontend/hello/template.html?raw';
import helloCSS from '../../examples/frontend/hello/style.css?raw';
import budgetGo from '../../examples/frontend/budget/main.go?raw';
import budgetHTML from '../../examples/frontend/budget/template.html?raw';
import budgetCSS from '../../examples/frontend/budget/style.css?raw';

highlight.registerLanguage('go', go);
highlight.registerLanguage('xml', xml);
highlight.registerLanguage('css', css);

function files(demo, main, template, style) {
  return [
    { name: 'Live demo', url: `${import.meta.env.BASE_URL}demos/${demo}/index.html` },
    ...[
      { name: 'main.go', language: 'go', source: main },
      { name: 'template.html', language: 'xml', source: template },
      { name: 'style.css', language: 'css', source: style },
    ].map(file => ({ ...file, html: highlight.highlight(file.source, { language: file.language }).value })),
  ];
}

export const starterFiles = files('expense-tracker', budgetGo, budgetHTML, budgetCSS);

export const helloFiles = files('hello', helloGo, helloHTML, helloCSS);
