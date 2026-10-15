# Rename and migration

ownframe was formerly known as go-gpui. The product and root Go package are
named `ownframe`. Documentation and examples use the full name.

## Go imports

The root module is `github.com/chinmay-sawant/ownframe`. Import it directly:

```go
import "github.com/chinmay-sawant/ownframe"
```

Use `ownframe.New`, `ownframe.Config`, and `ownframe.Run` in place of the
previous `gpui` selectors.
Internal packages and example imports use the new module prefix too. The
examples module is `github.com/chinmay-sawant/ownframe/examples`.

This changes the public module and package names. Existing applications must
update their imports and selectors when adopting the renamed release. Go
imports under the two module paths are separate package identities; keep one
version of the library in an application.

The committed `go.work` joins both local modules. `examples/go.mod` replaces
`github.com/chinmay-sawant/ownframe` with the parent checkout until a release
tag exists. Local `make test`, `make build`, and example commands work before
the GitHub rename. Existing release tags retain their original
module declarations; a repository redirect does not rewrite those files.
External installs need a new release that declares the new module path.

## Templates and settings

Library-managed HTML attributes now use `data-ownframe-*`, including
`data-ownframe-field`, `data-ownframe-focus`, `data-ownframe-placeholder`, and
`data-ownframe-selection`.
Update custom CSS selectors that target the previous `data-gpui-*` attributes.
The included templates and styles have been updated together.

Debug settings use `OWNFRAME_FILEPICK_DEBUG` and
`OWNFRAME_BROWSER_PORT`.
The previous `GPUI_*` spellings for those settings remain accepted. The GPU
parity test accepts `OWNFRAME_REPLAY_GPU_TEST=1` and the previous spelling. New crash
reports and music downloads use the `ownframe` storage directories; previous
directories are left intact.

The website keeps the existing Gopher mascot, demos, themes, documentation,
and star emoji. It reads saved theme and star preferences from the previous
storage keys. The star API uses repository ID `1400382069`, which identifies
the same repository before and after its rename.

## GitHub rename and release

The source changes are prepared locally. The remote repository name,
description, topics, release tags, and Git remote are unchanged.

1. Rename `chinmay-sawant/go-gpui` to `chinmay-sawant/ownframe` in GitHub
   repository settings when ready to publish the source migration.
2. Keep the existing Go GUI, platform, and rendering topics. A description
   that fits the project is: "Native apps with HTML, CSS, and Go. No browser
   runtime."
3. Update local remotes after the rename:

   ```sh
   git remote set-url origin https://github.com/chinmay-sawant/ownframe.git
   ```

4. Publish a new root release containing the renamed module declaration.
   Update `examples/go.mod` to require that version, then remove the parent
   `replace` after the version is available. Keep earlier tags
   intact and publish the examples tag for the corresponding release.
5. Build the website with `cd frontend && npm run build`. GitHub Pages serves
   the generated root `docs/` directory, with assets under `/ownframe/`.
   Check the new Pages address after publication.

GitHub redirects repository traffic and Git operations after a rename, but
project-site URLs are excluded. See [GitHub's repository rename
guide](https://docs.github.com/en/repositories/creating-and-managing-repositories/renaming-a-repository).
