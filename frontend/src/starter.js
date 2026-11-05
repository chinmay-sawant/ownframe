import helloGo from '../../examples/frontend/hello/main.go?raw';
import helloHTML from '../../examples/frontend/hello/template.html?raw';
import helloCSS from '../../examples/frontend/hello/style.css?raw';
import budgetGo from '../../examples/frontend/budget/main.go?raw';
import budgetHTML from '../../examples/frontend/budget/template.html?raw';
import budgetCSS from '../../examples/frontend/budget/style.css?raw';
import builderGo from '../../examples/frontend/builder/main.go?raw';
import builderHTML from '../../examples/frontend/builder/template.html?raw';
import builderCSS from '../../examples/frontend/builder/style.css?raw';
import helloBuilderGo from '../../examples/frontend/hello-builder/main.go?raw';
import builderMountGo from '../../examples/frontend/builder-mount/main.go?raw';
import { highlightSource } from './highlight';

function files(demo, main, template, style) {
  return [
    { name: 'Live demo', url: `${import.meta.env.BASE_URL}demos/${demo}/index.html` },
    ...[
      { name: 'main.go', language: 'go', source: main },
      { name: 'template.html', language: 'xml', source: template },
      { name: 'style.css', language: 'css', source: style },
    ].map(file => ({ ...file, html: highlightSource(file.source, file.language) })),
  ];
}

function goFile(source) {
  const file = { name: 'main.go', language: 'go', source };
  return [{ ...file, html: highlightSource(file.source, file.language) }];
}

export const starterFiles = files('expense-tracker', budgetGo, budgetHTML, budgetCSS);

export const helloFiles = files('hello', helloGo, helloHTML, helloCSS);

export const builderFiles = files('builder', builderGo, builderHTML, builderCSS);

export const helloBuilderFiles = goFile(helloBuilderGo);

export const builderMountFiles = goFile(builderMountGo);
