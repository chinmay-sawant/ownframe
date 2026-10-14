## Summary

The homepage made it difficult to see what OwnFrame builds and who owns its rendering engine. Put desktop and Android examples first, explain the Blinkless engine built from scratch in Go, and shorten setup and comparison guides. Keep the Gopher mascot, mobile documentation, and all reference guides accessible.

## Motivation / context

- Clarify Blinkless layout, OwnFrame page and input handling, and Ebitengine windows and drawing.
- Make the website easier to read on desktop and phones, with consistent widths and a smaller default scale.

## Changes

### Homepage and layout

- Split the homepage and documentation into separate components; load documentation on demand.
- Move examples ahead of setup, retain Telegram on Android, and show a short highlighted Go example with the complete program in a disclosure.
- Use a common page width and 80% default scale, a sticky header with measured anchor offsets, and touch targets that retain their size after scaling.
- Use natural page scrolling and full-width tables with horizontal scrolling on narrow screens.
- Keep the yellow GitHub star while removing its badge background and count divider; add a favicon.

### Documentation and comparisons

- Add a short overview and six primary guide links, with separate Examples and Comparisons menus and a collapsed All guides menu.
- Preserve reference routes and the app showcase; move its app selection into the article.
- Compare the same plain Hello world UI across Rust GPUI, Electron, and Go libraries. Label snippets as UI excerpts and link complete setup separately.
- Explain our own Go HTML and CSS view without promising browser DOM or JavaScript support.

### Build and verification

- Declare the existing highlight.js 11.11.2 version directly for Go syntax highlighting.
- Add a headless Chrome check using the browser debugging protocol, with screenshots in ignored temp output.
- Rebuild the committed docs output served by GitHub Pages and preserve its Go module boundary.

## Impact

| Area | Impact |
| --- | --- |
| Performance | Documentation JavaScript loads when a guide is opened. |
| Memory | No Go runtime changes. |
| Behavior / correctness | Examples lead the homepage; headers, tables, themes, and mobile guide navigation remain usable at 80% scale. |
| API / CLI | No library API changes. The frontend adds npm run check:site. |
| Dependencies | Direct frontend declaration of the already-used highlight.js version; no Go module changes. |
| Binary size / build time | Regenerates static website assets; no Go binary changes. |

## Breaking changes / migration

| Item | Migration |
| --- | --- |
| None | Existing documentation routes remain available. |

## Test plan

- [x] Frontend unit tests, production build, and browser interaction checks.
- [x] Root and examples Go tests.
- [x] Go compile and vet checks.
- [x] Authored source whitespace checks; generated output validated by the production browser check.
- [x] Unrelated local example files excluded from staging.

### Commands

```sh
cd frontend
npm test
npm run build
npm run check:site
cd ..
make test TEST_P=4
make build BUILD_P=4
git diff master...HEAD --check -- frontend documentation plans/PR
```

## Screenshots / sample output

```text
PASS: 100 layout checks, themes, mobile menu, anchors, showcase, keyboard focus, and images.
```

The browser check covers five viewport sizes in both themes, including example tables and all comparison guides. It also checks sticky-header scrolling, guide anchors, menu selection, showcase selection, highlighted code, and image loading. Screenshots can be regenerated with npm run check:site.

## Related issues

None. This implements the requested website simplification.

## PR metadata checklist

- Self-assign the opening author with --assignee @me.
- Apply enhancement and documentation labels.
- Commit this PR record and generate its diff statistics from master.

## Follow-ups

None required for this change.

## Reviewer checklist

- [x] Homepage retains desktop examples and the original Telegram Android demo.
- [x] No unrelated example work staged.
- [x] Public library APIs remain unchanged.
- [x] Generated GitHub Pages assets included.

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.css` | 3 | 140 | 75 |
| `.html` | 2 | 6 | 4 |
| `.js` | 5 | 2741 | 2939 |
| `.json` | 2 | 4 | 1 |
| `.jsx` | 3 | 335 | 208 |
| `.md` | 6 | 276 | 420 |
| `.mjs` | 1 | 199 | 0 |
| `.svg` | 2 | 8 | 0 |
| **Total** | **24** | **3709** | **3647** |
