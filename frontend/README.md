# ownframe website

A React and Vite site with the README demos and the repository's Markdown documentation. It uses plain CSS and native video controls. The site defaults to dark mode; the header toggle switches to light mode and remembers the choice in this browser. The Gopher mascot lives in `assets/brand/gopher.webp`.

## Run locally

Use Node.js 22.12 or newer.

```sh
cd frontend
npm ci
npm run dev
```

Vite prints the local address under `/ownframe/`. `npm test` checks the GitHub cache. `npm run build` writes the static site to the repository's root `docs/` folder; Vite restores `docs/go.mod` after each build so generated files stay outside the library module archive. Each build replaces the generated output in `docs/`. `npm run preview` serves that build.

## GitHub Pages

Run `npm run build` from `frontend/`. The build includes the videos, images, and `.nojekyll` file, and uses `/ownframe/` for asset paths. Keep the generated `docs/` folder in the repository when publishing. In the repository's Settings > Pages, choose **Deploy from a branch**, select the publishing branch, and select **/docs**.

For a custom domain hosted at the root, build with `npm run build -- --base=/`.

## Content

The site imports `../documentation/*.md`, the example index, desktop cat guide, Android guide, and `../showcase.md` at build time. Edit those files to update the website. Relative documentation links stay in the site; source-code and plan links open the repository on GitHub.

The three MP4s in `public/demos/` are the original videos from their linked X posts. The Spotify Player preview is a 960 by 520, 12 fps animated WebP at `public/demos/videos/spotify.webp`; it is silent and loops in the browser. The Wispr Flow, desktop cat, and Teams demos use preview images under `assets/screenshots/demos/`. The Teams clip comes from [this post](https://x.com/chinmay_sawant_/status/2107548496194895993/video/1). MP4 videos load metadata only and do not autoplay.

The Dino Run card uses `assets/screenshots/demos/dino.webp` and links to `examples/dino` on GitHub. From the repository root, run it with `go run ./examples/dino`.

## GitHub stars

The top-right link reads `stargazers_count` from GitHub's public repository API using the repository's stable numeric ID. The count keeps working across the rename. Successful counts are cached in localStorage for six hours. Failed requests back off for at least one hour and retain the last count. The loader honors longer `Retry-After` and rate-limit reset times, limits requests to ten seconds, and shares an in-flight request within the page. Web Locks coordinate tabs where supported. If browser storage is blocked, the page still uses its memory cache. Existing theme and count preferences are read from the previous storage keys during migration.

This cache is per browser, not shared across visitors. There is no polling or client-side token. A shared cache at the hosting layer can be added when deployment is chosen.

API reference: [GitHub repository endpoint](https://docs.github.com/en/rest/repos/repos#get-a-repository). Build setup: [Vite documentation](https://vite.dev/guide/).

During the rename, GitHub links can point to the current repository with `VITE_GITHUB_REPOSITORY=chinmay-sawant/go-gpui npm run dev`. Production builds default to `chinmay-sawant/ownframe`. This override changes repository links; the star endpoint continues to use the same numeric ID.
