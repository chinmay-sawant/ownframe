import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { displayTextForDocument } from './compareDisplay.js';

const root = fileURLToPath(new URL('../../', import.meta.url));

test('comparison display hides the repeated OwnFrame section but keeps the source intact', () => {
  const source = readFileSync(`${root}documentation/compare-fyne.md`, 'utf8');
  assert.ok(source.includes('### OwnFrame'));
  const shown = displayTextForDocument('compare-fyne', source);
  assert.ok(!shown.includes('### OwnFrame'));
  assert.ok(!shown.includes('Where the code lives'));
  assert.ok(shown.includes('widget.NewLabel("Hello world")'));
});

test('wails display keeps the closing paragraph after the table', () => {
  const source = readFileSync(`${root}documentation/compare-wails.md`, 'utf8');
  const shown = displayTextForDocument('compare-wails', source);
  assert.ok(!shown.includes('### OwnFrame'));
  assert.ok(shown.includes('Wails shows its interface through the system web view.'));
});

test('other documents pass through unchanged', () => {
  const source = readFileSync(`${root}documentation/compare-rust-gpui.md`, 'utf8');
  assert.equal(displayTextForDocument('compare-rust-gpui', source), source);
});
