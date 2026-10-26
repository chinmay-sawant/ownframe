# Secrets

ownframe never touches your credentials: templates, themes, and page
state are trusted author content, but anything baked into a binary or
the wasm bundle is public. This page collects the rules for app secrets
such as API keys, pairing keys, passwords, and tokens.

## Never bake secrets into artifacts

`-ldflags -X`, `//go:embed`, and the wasm bundle are trivially
extractable with `strings`. The wasm build is fully public by design:
it ships to every browser that opens the page.

Rules:

- Keep API keys server side. The app calls your backend; the backend
  calls the third party. No key ships in the desktop binary, the APK,
  or `ownframe.wasm`.
- Treat the TV pairing key (`examples/lg-remote` store file, mode
  `0600`) and demo logins (`secret`/`secret` in `examples/login`) as
  local-only development values, never as production auth.
- Run the release scan before tagging:
  `sh scripts/scan-secrets.sh`. It greps tracked source for
  high-confidence token patterns and runs `strings` over `dist/`
  artifacts. Install `gitleaks` for the full rule set; the shared
  config is `.gitleaks.toml`, with demo credentials allowlisted by
  path, never by blanket exclude.

## OS keychain for app secrets

Apps that need a persistent token (teams/sqlite, download manager)
should store it in the OS keychain, not in a plaintext file:

- macOS Keychain, Windows Credential Manager, Linux Secret Service
  through one API: `github.com/zalando/go-keyring`.
- Fall back to a `0600` file under the app config dir only when no
  keyring is available, and say so in the UI.

Sketch (apps add the dependency themselves; the library ships none):

```go
import "github.com/zalando/go-keyring"

const service = "com.example.myapp"

func saveToken(account, token string) error {
    return keyring.Set(service, account, token)
}

func loadToken(account string) (string, error) {
    return keyring.Get(service, account)
}
```

## Redacting secret types

`ownframe.Secret` (the public alias of `internal/secret.Secret`) is a string that renders as `[REDACTED]`
through `String`, `GoString`, `Format`, and `slog.LogValuer`. Use it
for every in-memory password, token, and pairing key:

- `examples/login` keeps the undo-history password as a `Secret`.
- `examples/lg-remote` keeps the paired TV key as a `Secret` in
  memory; only the `0600` store file on disk holds the raw value.

Call `Reveal()` at the single point of use (filling a field, sending
a pairing message, authorizing a request). Never log, `%v`-print, or
template-render the raw value.
