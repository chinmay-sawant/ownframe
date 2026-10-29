import React, { useRef, useState } from 'react';
import highlight from 'highlight.js/lib/core';
import go from 'highlight.js/lib/languages/go';
import xml from 'highlight.js/lib/languages/xml';
import css from 'highlight.js/lib/languages/css';
import mainSource from '../../examples/frontend/main.go?raw';
import appSource from '../../examples/frontend/app.go?raw';
import tasksSource from '../../examples/frontend/tasks.go?raw';
import templateSource from '../../examples/frontend/template.html?raw';
import styleSource from '../../examples/frontend/style.css?raw';

highlight.registerLanguage('go', go);
highlight.registerLanguage('xml', xml);
highlight.registerLanguage('css', css);

const files = [
  { name: 'main.go', language: 'go', source: mainSource },
  { name: 'app.go', language: 'go', source: appSource },
  { name: 'tasks.go', language: 'go', source: tasksSource },
  { name: 'template.html', language: 'xml', source: templateSource },
  { name: 'style.css', language: 'css', source: styleSource },
].map(file => ({ ...file, html: highlight.highlight(file.source, { language: file.language }).value }));

export default function ExampleCode() {
  const [selected, setSelected] = useState(0);
  const tabs = useRef([]);

  function select(index) {
    setSelected(index);
    tabs.current[index].focus();
  }

  function onKeyDown(event, index) {
    let next;
    switch (event.key) {
      case 'ArrowRight': next = (index + 1) % files.length; break;
      case 'ArrowLeft': next = (index + files.length - 1) % files.length; break;
      case 'Home': next = 0; break;
      case 'End': next = files.length - 1; break;
      default: return;
    }
    event.preventDefault();
    select(next);
    tabs.current[next].scrollIntoView({ block: 'nearest', inline: 'nearest' });
  }

  return <div className="preview-pane preview-code-pane">
    <div className="example-file-tabs" role="tablist" aria-label="Example source files">
      {files.map((file, index) => <button
        key={file.name}
        type="button"
        role="tab"
        id={`example-tab-${index}`}
        aria-selected={selected === index}
        aria-controls={`example-panel-${index}`}
        tabIndex={selected === index ? 0 : -1}
        ref={element => { tabs.current[index] = element; }}
        onClick={() => select(index)}
        onKeyDown={event => onKeyDown(event, index)}
      >
        <span className={`example-file-icon ${file.language}`} aria-hidden="true">{file.language === 'go' ? 'Go' : file.language === 'css' ? '#' : '<>'}</span>
        {file.name}
      </button>)}
    </div>
    {files.map((file, index) => <div
      key={file.name}
      id={`example-panel-${index}`}
      role="tabpanel"
      aria-labelledby={`example-tab-${index}`}
      hidden={selected !== index}
      tabIndex={0}
      className="preview-code-body"
    >
      <pre><code className={`hljs language-${file.language}`}
        dangerouslySetInnerHTML={{ __html: file.html }} /></pre>
    </div>)}
    <div className="example-file-path">examples/frontend/{files[selected].name}</div>
  </div>;
}
