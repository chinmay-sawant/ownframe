# OwnFrame website

The homepage shows desktop and mobile examples first, then explains the shared engine, a small Go app, and the GPUI comparison. The Gopher mascot and yellow GitHub star remain part of the design.

## Run locally

Use Node.js 22.12 or newer.

```sh
cd /home/chinmay/ChinmayPersonalProjects/ownframe/frontend
npm ci
npm run dev
```

Open the `/ownframe/` address Vite prints. The site supports dark and light themes and remembers your choice. All pages use an 80% default scale and the same page width. The header stays visible while you scroll.

## Edit the site

- `src/Home.jsx` contains the homepage copy, code sample, and demo cards.
- `src/Documentation.jsx` renders the guides and handles documentation navigation.
- `src/showcase.js` defines the website's Example apps content and media dimensions.
- `src/docs-overview.js` contains the short starting guide.
- `src/content.js` groups Markdown documents and resolves their links and images.
- `src/main.jsx` contains the shared header, footer, and route selection.
- `src/style.css` styles the homepage and documentation for desktop and mobile.

The documentation overview links six main guides. Examples and comparisons have separate menus; the remaining reference sits under All guides. Guides use one page scroll. Scrollbar rails are hidden, while wheel, touch, and keyboard scrolling remain available.

Guides come from `../documentation/*.md`, `../examples/readme.md`, the desktop cat guide, and the Android guide. The website's Example apps page uses `src/showcase.js`; `../showcase.md` remains a separate repository document and does not set the website media sizes. Relative guide links stay on the site; source links open GitHub. Existing `#docs/<slug>/<heading>` links still work.

Keep the engine roles clear in the copy. Blinkless parses HTML, applies CSS, and calculates layout. OwnFrame renders its display list and connects the page to Ebitengine's window and input loop. This is an app UI engine with a supported HTML and CSS subset. It does not run website JavaScript.

## Check and build

```sh
npm test
npm run lint
npm run build
npm run check:site
```

`npm test` checks the GitHub star cache. `npm run lint` runs ESLint 10 over the source with the flat config in `eslint.config.js`. `check:site` opens the production build in headless Chrome and checks desktop and mobile layouts, both themes, documentation navigation, and image loading. It requires `google-chrome` on PATH, or `CHROME_PATH` pointing to a Chromium executable. Screenshots go to the repository's ignored `temp/frontend-check/` folder.

The repository's `make lint` runs this same `npm run lint` after the Go linters, so one command covers the Go code and the website. There is no separate make target for the frontend.

`eslint-plugin-react` still declares a peer range that stops at ESLint 9, and its version detection calls an API ESLint 10 removed. The `overrides` entry in `package.json` resolves the peer range to the installed ESLint, and `settings.react.version` in the config skips that detection. Two rules are off on purpose: `react/prop-types`, because the site has no PropTypes layer, and `react-hooks/set-state-in-effect`, because the site does not run the React Compiler and two effects reset state on purpose.

The build writes to the root `docs/` folder, including the public demo assets. It preserves `docs/go.mod` so the generated site stays outside the library module archive. `npm run preview` serves the generated site.

## Publish to GitHub Pages

Build before publishing and include both `frontend/` changes and generated `docs/` files. Pages serves `docs/`, not the React source. Configure Pages to deploy from the publishing branch's `/docs` folder.

The default asset base is `/ownframe/`. For a domain hosted at its root, use `npm run build -- --base=/`.

Repository links default to `chinmay-sawant/ownframe`. `VITE_GITHUB_REPOSITORY` overrides that path. The optional GitHub star count uses the stable repository ID, caches successful responses for six hours, and backs off after failures. It retains the previous count when GitHub is unavailable.
