# OwnFrame website

The homepage introduces the native Go framework and embeds the playable public WebAssembly demo beside its source. A separate link opens the demo on its own. It then shows desktop and Android examples, explains the engine, and links historical measurements to their methodology. The Gopher mascot and yellow GitHub star remain part of the design.

## Run locally

Use Node.js 22.13 or newer.

```sh
cd /home/chinmay/ChinmayPersonalProjects/ownframe/frontend
npm ci
npm run dev
```

Open the `/ownframe/` address Vite prints. The site supports dark and light themes and remembers your choice. All pages use the same page width. On narrow screens, the header has a collapsible menu and the demo preview appears before expandable source code. The header stays visible while you scroll.

## Edit the site

- `src/Home.jsx` contains the homepage copy, code sample, and demo cards.
- `src/starter.js` loads Hello World from `examples/frontend/hello/` and Split with friends from `examples/frontend/budget/`. Its Live demo tab embeds the WASM app, followed by `main.go`, `template.html`, and `style.css` source tabs.
- `src/Documentation.jsx` renders the example and comparison pages and handles documentation navigation.
- `src/showcase.js` defines the website's Example apps content and media dimensions.
- `src/content.js` groups the example and comparison Markdown documents and resolves their links and images.
- `src/main.jsx` contains the shared header, footer, mobile menu, and route selection.
- `src/Team.jsx` introduces the human maintainer and lists AI tools separately.
- `public/demo/` holds the standalone Launch board WASM demo and its dark and light preview images. The homepage embeds this page at the full width of its demo pane and keeps a separate standalone link.
- `public/demos/hello/` and `public/demos/expense-tracker/` hold the starter WASM demos. `build:demo` rebuilds them from the same files displayed in the tabs.
- `src/style.css` styles the homepage and documentation for desktop and mobile.

The docs section shows About, Examples, and Comparisons. The About page is sourced from `../documentation/why-ownframe.md`. Guides use one page scroll. Scrollbar rails are hidden, while wheel, touch, and keyboard scrolling remain available.

Example and comparison pages come from `../documentation/compare-*.md` and `../examples/readme.md`. The website's Example apps page uses `src/showcase.js`; `../showcase.md` remains a separate repository document and does not set the website media sizes. Relative guide links stay on the site; source links open GitHub. Existing `#docs/<slug>/<heading>` links for examples and comparisons still work.

Keep the engine roles clear in the copy. Blinkless parses HTML, applies CSS, and calculates layout. OwnFrame renders its display list and connects the page to Ebitengine's window and input loop. This is an app UI engine with a supported HTML and CSS subset. It does not run website JavaScript.

## Check and build

```sh
npm test
npm run lint
npm run build
npm run check:site
```

`npm test` checks comparison display and the GitHub star cache. `npm run lint` runs ESLint 10 over the source with the flat config in `eslint.config.js`. `check:site` opens the production build in headless Chrome and checks desktop and mobile layouts, both themes, documentation navigation, and image loading. It requires `google-chrome` on PATH, or `CHROME_PATH` pointing to a Chromium executable. Screenshots go to the repository's ignored `temp/frontend-check/` folder. The check also verifies the mobile menu, expandable source, screenshot framing, links to the standalone demo, and mouse interaction in the embedded demo at desktop and phone widths.

After changing the example sources, rebuild the WASM demos. If the Launch board or Spotify preview changed, refresh its captured images too:

```sh
npm run build:demo
npm run capture:demo
npm run build
npm run check:site
```

`capture:demo` opens the public demo in Chrome, exercises theme switching, task completion, and filtering, and saves both preview images under `public/demo/`. It also captures a still Spotify gallery image under `public/demos/`, so the homepage gallery does not autoplay. It uses the same Chrome prerequisites as `check:site`. Review the captured images before publishing.

The repository's `make lint` runs this same `npm run lint` after the Go linters, so one command covers the Go code and the website. There is no separate make target for the frontend.

`eslint-plugin-react` still declares a peer range that stops at ESLint 9, and its version detection calls an API ESLint 10 removed. The `overrides` entry in `package.json` resolves the peer range to the installed ESLint, and `settings.react.version` in the config skips that detection. Two rules are off on purpose: `react/prop-types`, because the site has no PropTypes layer, and `react-hooks/set-state-in-effect`, because the site does not run the React Compiler and two effects reset state on purpose.

Reverify after any `npm install` or ESLint bump: `npm ls eslint` must resolve to one ESLint 10, `npm run lint` must pass, and no major version may move. The override stays invisible in `package-lock.json` (lockfileVersion 3 records no top-level overrides key), so re-check `package.json` itself. Remove the workaround only when `eslint-plugin-react` declares ESLint 10 support.

`build:demo` rebuilds Launch board, Hello World, Split with friends, Workspace Inspector, Telegram, and Spotify Player. Spotify includes bundled audio and exceeds GitHub's 100 MiB file limit as a raw WASM file. Its build strips debug symbols and writes `public/demos/spotify-player/ownframe.wasm.gz`; the browser loader unpacks that asset with `DecompressionStream` when the server does not already send `Content-Encoding: gzip`. Browsers decode that HTTP encoding automatically, so the loader uses the decoded response directly in that case. The raw Spotify WASM remains ignored. Browsers without gzip stream decompression get an explanatory message. Telegram and Workspace Inspector WASM files have explicit Git ignore exceptions so their public and generated copies ship.

The build writes to the root `docs/` folder, including the public demo assets. It preserves `docs/go.mod` so the generated site stays outside the library module archive. `npm run preview` serves the generated site.

## Publish to GitHub Pages

Build before publishing and include both `frontend/` changes and generated `docs/` files. Pages serves `docs/`, not the React source. Configure Pages to deploy from the publishing branch's `/docs` folder.

The default asset base is `/ownframe/`. For a domain hosted at its root, use `npm run build -- --base=/`.

Repository links default to `chinmay-sawant/ownframe`. `VITE_GITHUB_REPOSITORY` overrides that path. The optional GitHub star count uses the stable repository ID, caches successful responses for six hours, and backs off after failures. It retains the previous count when GitHub is unavailable.

Run the examples locally from the repository root with `go run ./examples/frontend`, `go run ./examples/frontend/hello`, or `go run ./examples/frontend/budget`. Each directory contains the Go source and its HTML and CSS.
