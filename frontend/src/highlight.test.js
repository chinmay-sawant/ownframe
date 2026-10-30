import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { highlightSource } from './highlight.js';

const root = fileURLToPath(new URL('../../', import.meta.url));

function readSample(path) {
  return readFileSync(`${root}examples/frontend/${path}`, 'utf8');
}

test('go sample highlights keywords, strings, and types', () => {
  const html = highlightSource(readSample('hello/main.go'), 'go');
  assert.match(html, /hljs-keyword/);
  assert.match(html, /hljs-string/);
  assert.match(html, /hljs-type/);
  assert.match(html, /hljs-comment/);
});

test('css sample highlights properties, selectors, and values', () => {
  const html = highlightSource(readSample('hello/style.css'), 'css');
  assert.match(html, /hljs-attribute/);
  assert.match(html, /hljs-selector-class/);
  assert.match(html, /hljs-number/);
});

test('html sample keeps go templates intact instead of splitting attributes', () => {
  const html = highlightSource(readSample('template.html'), 'xml');
  assert.match(html, /hljs-template-tag/);
  assert.doesNotMatch(html, /hljs-attr">all</);
  assert.ok(html.includes('{{if eq .Filter &quot;all&quot;}}') || html.includes('{{if eq .Filter'));
});
