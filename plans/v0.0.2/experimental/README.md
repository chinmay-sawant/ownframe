# Chrome-like Behavior Track (Experimental Pro)

Goal: add a retained/reactive layer above `Page` so ownframe changes
subtrees instead of re-painting the whole picture, while still using Go
templates, HTML/CSS layout, and canvas rendering.

Status convention: `- [ ]` open, `- [x]` done.

## Sections

1. [`01-reactive-state.md`](./01-reactive-state.md) - state and reactivity
2. [`02-dom-element-model.md`](./02-dom-element-model.md) - retained elements
3. [`03-css-layout-invalidation.md`](./03-css-layout-invalidation.md) - scoped style/layout
4. [`04-patching-display-management.md`](./04-patching-display-management.md) - operation patching
5. [`05-browser-runtime-behaviors.md`](./05-browser-runtime-behaviors.md) - runtime behaviors

## Constraints carried from the current design

- Keep templates and Go as the app-language surface
- Do not add a JavaScript engine
- Preserve page-level invalidation compatibility for old templates
- Keep the current direct and integration test entry points passing